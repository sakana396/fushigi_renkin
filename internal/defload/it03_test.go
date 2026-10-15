package defload

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
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

// 読み込み値の取り違えを見分けられるよう、各項目に異なる値を与える（DES-DEF-04〜07）。
var richIT03Data = func() string {
	s := it03Data
	for _, pair := range [][2]string{
		{`range = 0`, `range = 2`},
		{`effects=[{kind="heal", value=1}]`, `effects=[{kind="heal", value=13}]`},
		{`party = [{unit="hero", x=0, y=0}]`, `party = [{unit="hero", x=1, y=2}]`},
		{`id = "hero"
role = "front"
hp = 1
attack = 1
speed = 1
move = 1
range = 1`, `id = "hero"
name = "勇者"
role = "front"
hp = 11
attack = 12
speed = 13
move = 14
range = 15`},
		{`id = "foe"
role = "enemy"
hp = 1
attack = 1
speed = 1
move = 1
range = 1`, `id = "foe"
name = "敵"
role = "enemy"
hp = 21
attack = 22
speed = 23
move = 24
range = 25`},
		{`enemies = [{unit="foe", x=7, y=5}]
rolls = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]`, `enemies = [{unit="foe", x=6, y=4}]
rolls = 3
drops = [{item="herb", weight=4, quality_min=51, quality_max=151}]`},
		{`id = "delivery"
item = "potion"
min_quality = 1
count = 1
reward = 1`, `id = "delivery"
item = "potion"
min_quality = 52
count = 5
reward = 120`},
	} {
		if strings.Count(s, pair[0]) != 1 {
			panic("読み込み検査の入力が変わった: " + pair[0])
		}
		s = strings.Replace(s, pair[0], pair[1], 1)
	}
	return s
}()

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
		{"TC-FR-MOD-02-10 全定義の読み込み", func(string) string { return richIT03Data }, nil},
		{"TC-FR-MOD-04-20 使用の正常", modify(replace(`target = "ally"`, `target = "enemy"`), replace(`range = 0`, `range = 1`)), nil},
		{"TC-FR-MOD-04-21 使用の境界", replace(`range = 0`, `range = 1`), nil},
		{"TC-FR-MOD-04-22 使用の異常", modify(replace(`target = "ally"`, `target = "other"`), replace(`range = 0`, `range = -1`), replace(`radius=0`, `radius=-1`), replace(`kind="heal", value=1`, `kind="other", value=0`)), []string{"item[id=potion].use.target", "item[id=potion].use.range", "item[id=potion].use.stages[0].radius", "item[id=potion].use.stages[0].effects[0].kind", "item[id=potion].use.stages[0].effects[0].value"}},
		{"TC-FR-MOD-04-23 段階の正常", replace(`{min_quality=200, radius=1}`, `{min_quality=50, radius=1}, {min_quality=100}, {min_quality=150}, {min_quality=200}`), nil},
		{"TC-FR-MOD-04-24 段階の境界", replace(`{min_quality=200, radius=1}`, `{min_quality=50, radius=1}`), nil},
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
		{"TC-FR-MOD-04-41 明示された零の戦闘設定", func(string) string { return "[game]\nbattle_width = 0\n" }, []string{"game.battle_width"}},
		{"TC-FR-MOD-04-42 盤面の負座標と高さ境界", modify(replace(`party = [{unit="hero", x=0, y=0}]`, `party = [{unit="hero", x=-1, y=6}]`), replace(`enemies = [{unit="foe", x=7, y=5}]`, `enemies = [{unit="foe", x=-1, y=6}]`)), []string{"game.party[0].x", "game.party[0].y", "gather[id=field].battle.enemies[0].x", "gather[id=field].battle.enemies[0].y"}},
		{"TC-FR-MOD-04-43 四種類の使用効果", replace(`effects=[{kind="heal", value=1}]`, `effects=[{kind="heal", value=1}, {kind="damage", value=2}, {kind="attack_up", value=3}, {kind="speed_up", value=4}]`), nil},
		{"TC-FR-MOD-04-44 仲間の位置重複", replace(`party = [{unit="hero", x=0, y=0}]`, `party = [{unit="hero", x=0, y=0}, {unit="hero", x=0, y=0}]`), []string{"game.party[1]: 初期位置が重複"}},
		{"TC-FR-MOD-04-45 敵の位置重複", replace(`enemies = [{unit="foe", x=7, y=5}]`, `enemies = [{unit="foe", x=7, y=5}, {unit="foe", x=7, y=5}]`), []string{"gather[id=field].battle.enemies[1]: 初期位置が重複"}},
		{"TC-FR-MOD-04-46 戦闘ごとに位置を検査", func(s string) string {
			return s + `
[[gather]]
id = "forest"
rolls = 1
cooldown_seconds = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]
[gather.battle]
enemies = [{unit="foe", x=7, y=5}]
rolls = 1
drops = [{item="herb", weight=1, quality_min=1, quality_max=200}]
`
		}, nil},
		{"TC-FR-MOD-04-47 game なしの戦闘", func(s string) string {
			return strings.Replace(s, "[game]\nbattle_width = 8\nbattle_height = 6\ninjury_seconds = 60\nparty = [{unit=\"hero\", x=0, y=0}]\n", "", 1)
		}, []string{"gather[id=field].battle: 戦闘には game の設定が必要"}},
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
				if strings.HasPrefix(tc.name, "TC-FR-MOD-02-10 ") {
					checkIT03Values(t, catalog)
				}
				if strings.HasPrefix(tc.name, "TC-FR-MOD-04-43 ") {
					want := []defs.UseEffect{{Kind: "heal", Value: 1}, {Kind: "damage", Value: 2}, {Kind: "attack_up", Value: 3}, {Kind: "speed_up", Value: 4}}
					if got := catalog.Items["potion"].Use.Stages[0].Effects; !reflect.DeepEqual(got, want) {
						t.Errorf("効果 = %+v; 欲しい値 = %+v", got, want)
					}
				}
				if strings.HasPrefix(tc.name, "TC-FR-MOD-04-46 ") && !reflect.DeepEqual(catalog.Gathers["forest"].Battle, catalog.Gathers["field"].Battle) {
					t.Errorf("別の採取地の戦闘 = %+v", catalog.Gathers["forest"].Battle)
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

// 品質段階の三つの規則を、それぞれ独立した入力と正確な経路で確かめる（DES-VAL-10）。
func TestIT03StageValidation(t *testing.T) {
	const old = `stages = [{min_quality=1, radius=0, effects=[{kind="heal", value=1}]}, {min_quality=200, radius=1}]`
	cases := []struct {
		name, stages, path, reason string
	}{
		{"TC-FR-MOD-04-25 許されない段階", `stages = [{min_quality=1, effects=[{kind="heal", value=1}]}, {min_quality=49}]`, "item[id=potion].use.stages[1].min_quality", "1・50・100・150・200 のいずれかであること"},
		{"TC-FR-MOD-04-25 許される値の重複", `stages = [{min_quality=1, effects=[{kind="heal", value=1}]}, {min_quality=50}, {min_quality=50}]`, "item[id=potion].use.stages[2].min_quality", "品質段階が重複"},
		{"TC-FR-MOD-04-25 基本段階の欠落", `stages = [{min_quality=50}]`, "item[id=potion].use.stages", "品質 1 の基本の段階が必要"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(files(map[string]string{"it03.toml": strings.Replace(it03Data, old, tc.stages, 1)}))
			if err == nil {
				t.Fatal("誤りがない")
			}
			diagnostics, ok := err.(Errors)
			if !ok {
				t.Fatalf("誤りの型 = %T", err)
			}
			for _, d := range diagnostics {
				if d.File == "it03.toml" && d.Path == tc.path && d.Reason == tc.reason {
					return
				}
			}
			t.Errorf("%s: %s がない: %v", tc.path, tc.reason, err)
		})
	}
}

// 読み込んだ戦闘と依頼の全フィールドを期待値と比較する（DES-DEF-04〜07）。
func checkIT03Values(t *testing.T, c defs.Catalog) {
	t.Helper()
	wantUse := &defs.UseDef{Target: "ally", Range: 2, Stages: []defs.UseStage{
		{MinQuality: 1, Radius: 0, Effects: []defs.UseEffect{{Kind: "heal", Value: 13}}},
		{MinQuality: 200, Radius: 1},
	}}
	if got := c.Items["potion"].Use; !reflect.DeepEqual(got, wantUse) {
		t.Errorf("Use = %+v; 欲しい値 = %+v", got, wantUse)
	}
	wantUnits := map[string]defs.UnitDef{
		"hero": {ID: "hero", Name: "勇者", Role: "front", HP: 11, Attack: 12, Speed: 13, Move: 14, Range: 15},
		"foe":  {ID: "foe", Name: "敵", Role: "enemy", HP: 21, Attack: 22, Speed: 23, Move: 24, Range: 25},
	}
	if !reflect.DeepEqual(c.Units, wantUnits) {
		t.Errorf("Units = %+v; 欲しい値 = %+v", c.Units, wantUnits)
	}
	wantGame := defs.GameDef{BattleWidth: 8, BattleHeight: 6, InjurySeconds: 60, Party: []defs.Placement{{Unit: "hero", X: 1, Y: 2}}}
	if !reflect.DeepEqual(c.Game, wantGame) {
		t.Errorf("Game = %+v; 欲しい値 = %+v", c.Game, wantGame)
	}
	wantBattle := &defs.BattleDef{Enemies: []defs.Placement{{Unit: "foe", X: 6, Y: 4}}, Rolls: 3, Drops: []defs.Drop{{Item: "herb", Weight: 4, QualityMin: 51, QualityMax: 151}}}
	if got := c.Gathers["field"].Battle; !reflect.DeepEqual(got, wantBattle) {
		t.Errorf("Battle = %+v; 欲しい値 = %+v", got, wantBattle)
	}
	wantQuests := map[string]defs.QuestDef{"delivery": {ID: "delivery", Rank: 1, Item: "potion", MinQuality: 52, Count: 5, Reward: 120}}
	if !reflect.DeepEqual(c.Quests, wantQuests) {
		t.Errorf("Quests = %+v; 欲しい値 = %+v", c.Quests, wantQuests)
	}
	if got, ok := c.Unit("hero"); !ok || !reflect.DeepEqual(got, wantUnits["hero"]) {
		t.Errorf("Unit(hero) = %+v, %t", got, ok)
	}
	if got, ok := c.Quest("delivery"); !ok || !reflect.DeepEqual(got, wantQuests["delivery"]) {
		t.Errorf("Quest(delivery) = %+v, %t", got, ok)
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
		unitStart := strings.Index(richIT03Data, "[[unit]]")
		gatherStart := strings.Index(richIT03Data, "[[gather]]")
		questStart := strings.Index(richIT03Data, "[[quest]]")
		c, err := Load(files(map[string]string{
			"game_items.toml": richIT03Data[:unitStart],
			"units.toml":      richIT03Data[unitStart:gatherStart],
			"gathers.toml":    richIT03Data[gatherStart:questStart],
			"quests.toml":     richIT03Data[questStart:],
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(c.Units) != 2 || len(c.Quests) != 1 || c.Gathers["field"].Battle == nil {
			t.Fatalf("定義を読み込めていない: %+v", c)
		}
		checkIT03Values(t, c)
	})
}
