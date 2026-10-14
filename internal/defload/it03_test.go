package defload

import (
	"strings"
	"testing"
)

const it03Data = `
[game]
battle_width = 8
battle_height = 6
injury_seconds = 60
party = [{unit="hero", x=0, y=0}]
[[item]]
id = "herb"
[[item]]
id = "potion"
[item.use]
target = "ally"
range = 0
stages = [{min_quality=1, radius=0, effects=[{kind="heal", value=1}]}, {min_quality=200, radius=1}]
[[recipe]]
id = "make_potion"
output = "potion"
output_count = 1
seconds = 1
inputs = [{item="herb", count=1}]
[[unit]]
id = "hero"
role = "front"
hp = 1
attack = 1
speed = 1
move = 1
range = 1
[[unit]]
id = "foe"
role = "enemy"
hp = 1
attack = 1
speed = 1
move = 1
range = 1
[[gather]]
id = "field"
rolls = 1
cooldown_seconds = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]
[gather.battle]
enemies = [{unit="foe", x=7, y=5}]
rolls = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]
[[quest]]
id = "delivery"
item = "potion"
min_quality = 1
count = 1
reward = 1
`

// 反復 3 の読み込みと各検証規則を表駆動で確かめる（DES-DEF-04〜07、DES-VAL-09〜13）。
func TestIT03Definitions(t *testing.T) {
	replace := func(old, new string) func(string) string {
		return func(s string) string { return strings.Replace(s, old, new, 1) }
	}
	modify := func(changes ...func(string) string) func(string) string {
		return func(s string) string {
			for _, change := range changes {
				s = change(s)
			}
			return s
		}
	}
	cases := []struct {
		name   string
		change func(string) string
		wants  []string
	}{
		{"TC-FR-MOD-02-10 全定義の読み込み", func(s string) string { return s }, nil},
		{"TC-FR-MOD-04-20 使用の正常", modify(replace(`target = "ally"`, `target = "enemy"`), replace(`range = 0`, `range = 1`)), nil},
		{"TC-FR-MOD-04-21 使用の境界", replace(`range = 0`, `range = 1`), nil},
		{"TC-FR-MOD-04-22 使用の異常", modify(replace(`target = "ally"`, `target = "other"`), replace(`range = 0`, `range = -1`), replace(`radius=0`, `radius=-1`), replace(`kind="heal", value=1`, `kind="other", value=0`)), []string{"item[id=potion].use.target", "item[id=potion].use.range", "item[id=potion].use.stages[0].radius", "item[id=potion].use.stages[0].effects[0].kind", "item[id=potion].use.stages[0].effects[0].value"}},
		{"TC-FR-MOD-04-23 段階の正常", replace(`{min_quality=200, radius=1}`, `{min_quality=50, radius=1}, {min_quality=100}, {min_quality=150}, {min_quality=200}`), nil},
		{"TC-FR-MOD-04-24 段階の境界", replace(`{min_quality=200, radius=1}`, `{min_quality=50, radius=1}`), nil},
		{"TC-FR-MOD-04-25 段階の異常", replace(`stages = [{min_quality=1, radius=0, effects=[{kind="heal", value=1}]}, {min_quality=200, radius=1}]`, `stages = [{min_quality=49, radius=0}, {min_quality=49, radius=1}]`), []string{"item[id=potion].use.stages[0].min_quality", "item[id=potion].use.stages[1].min_quality", "item[id=potion].use.stages"}},
		{"TC-FR-MOD-04-26 ユニットの正常", replace(`role = "front"`, `role = "back"`), nil},
		{"TC-FR-MOD-04-27 ユニットの境界", func(s string) string { return s }, nil},
		{"TC-FR-MOD-04-28 ユニットの異常", modify(replace(`role = "front"`, `role = "other"`), replace(`hp = 1`, `hp = 0`), replace(`attack = 1`, `attack = 0`), replace(`speed = 1`, `speed = 0`), replace(`move = 1`, `move = 0`), replace(`range = 1`, `range = 0`)), []string{"unit[id=hero].role", "unit[id=hero].hp", "unit[id=hero].attack", "unit[id=hero].speed", "unit[id=hero].move", "unit[id=hero].range", "game.party[0].unit"}},
		{"TC-FR-MOD-04-29 戦闘設定の正常", func(s string) string { return s }, nil},
		{"TC-FR-MOD-04-30 戦闘設定の境界", modify(replace(`battle_width = 8`, `battle_width = 2`), replace(`battle_height = 6`, `battle_height = 1`), replace(`injury_seconds = 60`, `injury_seconds = 1`), replace(`enemies = [{unit="foe", x=7, y=5}]`, `enemies = [{unit="foe", x=1, y=0}]`)), nil},
		{"TC-FR-MOD-04-31 戦闘設定の異常", modify(replace(`battle_width = 8`, `battle_width = 0`), replace(`battle_height = 6`, `battle_height = 0`), replace(`injury_seconds = 60`, `injury_seconds = 0`), replace(`party = [{unit="hero", x=0, y=0}]`, `party = []`), replace(`enemies = [{unit="foe", x=7, y=5}]`, `enemies = []`), replace("[gather.battle]\nenemies = []\nrolls = 1\ndrops = [{item=\"herb\", weight=1, quality_min=1, quality_max=200}]", "[gather.battle]\nenemies = []\nrolls = 0\ndrops = []")), []string{"game.battle_width", "game.battle_height", "game.injury_seconds", "game.party", "gather[id=field].battle.enemies", "gather[id=field].battle.rolls", "gather[id=field].battle.drops"}},
		{"TC-FR-MOD-04-32 依頼の正常", func(s string) string { return s }, nil},
		{"TC-FR-MOD-04-33 依頼の境界", replace(`min_quality = 1`, `min_quality = 200`), nil},
		{"TC-FR-MOD-04-34 依頼の異常", modify(replace("item = \"potion\"\nmin_quality", "item = \"herb\"\nmin_quality"), replace(`min_quality = 1`, `min_quality = 201`), replace("\ncount = 1\n", "\ncount = 0\n"), replace(`reward = 1`, `reward = 0`)), []string{"quest[id=delivery].item", "quest[id=delivery].min_quality", "quest[id=delivery].count", "quest[id=delivery].reward"}},
		{"TC-FR-MOD-04-35 複合エラーの集約", modify(replace(`target = "ally"`, `target = "other"`), replace(`role = "front"`, `role = "other"`), replace(`battle_width = 8`, `battle_width = 0`), replace(`min_quality = 1`, `min_quality = 201`)), []string{"item[id=potion].use.target", "unit[id=hero].role", "game.battle_width", "quest[id=delivery].min_quality"}},
		{"TC-FR-MOD-04-36 配置の重複と参照", modify(replace(`enemies = [{unit="foe", x=7, y=5}]`, `enemies = [{unit="missing", x=0, y=0}, {unit="hero", x=8, y=-1}]`)), []string{"gather[id=field].battle.enemies[0].unit", "gather[id=field].battle.enemies[0]:", "gather[id=field].battle.enemies[1].unit", "gather[id=field].battle.enemies[1].x", "gather[id=field].battle.enemies[1].y"}},
		{"TC-FR-MOD-04-37 戦闘報酬の検証", replace(`enemies = [{unit="foe", x=7, y=5}]
rolls = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]`, `enemies = [{unit="foe", x=7, y=5}]
rolls = 1
drops = [{item="missing", weight=0, quality_min=201, quality_max=0}]`), []string{"gather[id=field].battle.drops[0].item", "gather[id=field].battle.drops[0].weight", "gather[id=field].battle.drops[0].quality_min", "gather[id=field].battle.drops[0].quality_max", "gather[id=field].battle.drops[0]:"}},
		{"TC-FR-MOD-04-38 ユニットと依頼の ID 重複", func(s string) string {
			return s + `
[[unit]]
id = "hero"
role = "front"
hp = 1
attack = 1
speed = 1
move = 1
range = 1
[[quest]]
id = "delivery"
item = "potion"
min_quality = 1
count = 1
reward = 1
`
		}, []string{"unit[id=hero].id", "quest[id=delivery].id"}},
		{"TC-FR-MOD-04-39 基本段階の効果なし", replace(`effects=[{kind="heal", value=1}]`, `effects=[]`), []string{"item[id=potion].use.stages[0].effects"}},
		{"TC-FR-MOD-04-40 敵対象の射程 0", replace(`target = "ally"`, `target = "enemy"`), []string{"item[id=potion].use.range"}},
		{"TC-FR-MOD-04-41 明示された零の戦闘設定", replace(`battle_width = 8`, `battle_width = 0`), []string{"game.battle_width"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := Load(files(map[string]string{"it03.toml": tc.change(it03Data)}))
			if len(tc.wants) == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if len(catalog.Units) != 2 || len(catalog.Quests) != 1 || catalog.Items["potion"].Use == nil || catalog.Gathers["field"].Battle == nil {
					t.Fatalf("定義を読み込めていない: %+v", catalog)
				}
				if _, ok := catalog.Unit("hero"); !ok {
					t.Error("Unit が見つからない")
				}
				if _, ok := catalog.Quest("delivery"); !ok {
					t.Error("Quest が見つからない")
				}
				if _, ok := catalog.Unit("missing"); ok {
					t.Error("存在しない Unit が見つかった")
				}
				if _, ok := catalog.Quest("missing"); ok {
					t.Error("存在しない Quest が見つかった")
				}
				return
			}
			if err == nil {
				t.Fatal("誤りがない")
			}
			for _, want := range tc.wants {
				if !strings.Contains(err.Error(), "it03.toml: "+want) {
					t.Errorf("%s がない: %v", want, err)
				}
			}
		})
	}
}

// 新しい定義ファイルがなくても旧データを読めることを確かめる（DES-PKG-03）。
func TestIT03OptionalFiles(t *testing.T) {
	t.Run("TC-FR-MOD-02-11 新規ファイルなし", func(t *testing.T) {
		c, err := Load(mustSub(t))
		if err != nil {
			t.Fatal(err)
		}
		if len(c.Units) != 0 || len(c.Quests) != 0 {
			t.Fatalf("新しい定義がある: %+v", c)
		}
	})
}

// ファイル間の参照を含む新しいファイル名の読み込みを確かめる（DES-PKG-03、DES-DEF-06、07）。
func TestIT03SeparateFiles(t *testing.T) {
	t.Run("TC-FR-MOD-02-12 新規ファイルの分割読み込み", func(t *testing.T) {
		unitStart := strings.Index(it03Data, "[[unit]]")
		gatherStart := strings.Index(it03Data, "[[gather]]")
		questStart := strings.Index(it03Data, "[[quest]]")
		c, err := Load(files(map[string]string{
			"game_items.toml": it03Data[:unitStart],
			"units.toml":      it03Data[unitStart:gatherStart],
			"gathers.toml":    it03Data[gatherStart:questStart],
			"quests.toml":     it03Data[questStart:],
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(c.Units) != 2 || len(c.Quests) != 1 || c.Gathers["field"].Battle == nil {
			t.Fatalf("定義を読み込めていない: %+v", c)
		}
	})
}
