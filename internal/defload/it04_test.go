package defload

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const rankData = `[game]
rank_reputation = [3, 8, 15]
shop = [{item="herb", quality=20, price=1}]
[[item]]
id = "herb"
attributes = ["plant"]
[[item]]
id = "ore"
[[item]]
id = "essence"
[[item]]
id = "potion"
[[item]]
id = "elixir"
[[recipe]]
id = "essence"
output = "essence"
output_count = 1
seconds = 1
inputs = [{item="herb", count=1}]
[[recipe]]
id = "potion"
output = "potion"
output_count = 1
seconds = 1
inputs = [{item="essence", count=1}]
[[recipe]]
id = "elixir"
output = "elixir"
output_count = 1
seconds = 1
rank = 3
inputs = [{attribute="plant", count=1}]
[[gather]]
id = "mine"
rank = 3
rolls = 1
cooldown_seconds = 1
drops = [{item="ore", weight=1, quality_min=1, quality_max=1}]
[[quest]]
id = "q1"
item = "potion"
min_quality = 1
count = 1
reward = 1
reputation = 1
[[quest]]
id = "q2"
item = "potion"
min_quality = 1
count = 1
reward = 1
[[quest]]
id = "q3"
item = "potion"
min_quality = 1
count = 1
reward = 1
[[quest]]
id = "q4"
item = "potion"
min_quality = 1
count = 1
reward = 1
rank = 2
reputation = 1
[[quest]]
id = "q5"
item = "elixir"
min_quality = 1
count = 1
reward = 1
rank = 3
reputation = 1
[[quest]]
id = "q7"
item = "potion"
min_quality = 1
count = 1
reward = 1
[[quest]]
id = "q8"
item = "potion"
min_quality = 1
count = 1
reward = 1
[[quest]]
id = "q9"
item = "potion"
min_quality = 1
count = 1
reward = 1

`

func TestIT04RankDefinitions(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(string) string
		want  string
		check func(*testing.T, defs.Catalog)
	}{
		{"TC-FR-MOD-02-30 ランクと評判を読む", func(s string) string { return s }, "", func(t *testing.T, c defs.Catalog) {
			if len(c.Game.RankReputation) != 3 || c.Recipes["essence"].Rank != 1 || c.Gathers["mine"].Rank != 3 || c.Quests["q5"].Rank != 3 || c.Quests["q1"].Reputation != 1 {
				t.Fatalf("rank/reputation = %+v", c)
			}
		}},
		{"TC-FR-MOD-02-31 省略したランクは 1", func(s string) string { return s }, "", func(t *testing.T, c defs.Catalog) {
			if c.Recipes["essence"].Rank != 1 || c.Quests["q1"].Rank != 1 {
				t.Fatalf("recipe rank = %d, quest rank = %d", c.Recipes["essence"].Rank, c.Quests["q1"].Rank)
			}
		}},
		{"TC-FR-MOD-04-50 評判の閾値は正で狭義増加", func(s string) string { return strings.Replace(s, "[3, 8, 15]", "[3, 3, 0]", 1) }, "game.rank_reputation[1]", nil},
		{"TC-FR-MOD-04-64 同じ評判閾値は誤り", func(s string) string { return strings.Replace(s, "[3, 8, 15]", "[3, 3, 15]", 1) }, "game.rank_reputation[1]", nil},
		{"TC-FR-MOD-04-65 評判閾値 0 は誤り", func(s string) string { return strings.Replace(s, "[3, 8, 15]", "[0, 8, 15]", 1) }, "game.rank_reputation[0]", nil},
		{"TC-FR-MOD-04-51 明示したランク 0 は誤り", func(s string) string {
			return strings.Replace(s, "id = \"essence\"\noutput", "id = \"essence\"\nrank = 0\noutput", 1)
		}, "recipe[id=essence].rank", nil},
		{"TC-FR-MOD-04-52 上限を超えるランクは誤り", func(s string) string { return strings.Replace(s, "rank = 3\nrolls", "rank = 5\nrolls", 1) }, "gather[id=mine].rank", nil},
		{"TC-FR-MOD-04-53 評判は負にできない", func(s string) string { return strings.Replace(s, "reputation = 1", "reputation = -1", 1) }, "quest[id=q1].reputation", nil},
		{"TC-FR-MOD-04-54 ランク 1 の依頼が 5 件", func(s string) string {
			return strings.Replace(s, "id = \"q3\"\nitem", "id = \"q3\"\nrank = 2\nitem", 1)
		}, "quest: ランク 1 の依頼が 5 件しかない（6 件以上が必要）", nil},
		{"TC-FR-MOD-04-55 ランク 2 に評判を得られる依頼がない", func(s string) string {
			return strings.Replace(s, "rank = 2\nreputation = 1\n[[quest]]", "rank = 2\nreputation = 0\n[[quest]]", 1)
		}, "quest: ランク 2 に評判を得られる依頼がありません", nil},
		{"TC-FR-MOD-04-66 ランク 1 に評判を得られる依頼がない", func(s string) string {
			return strings.Replace(s, "id = \"q1\"\nitem = \"potion\"\nmin_quality = 1\ncount = 1\nreward = 1\nreputation = 1", "id = \"q1\"\nitem = \"potion\"\nmin_quality = 1\ncount = 1\nreward = 1\nreputation = 0", 1)
		}, "quest: ランク 1 に評判を得られる依頼がありません", nil},
		{"TC-FR-MOD-04-67 ランク 3 に評判を得られる依頼がない", func(s string) string {
			return strings.Replace(s, "rank = 3\nreputation = 1\n", "rank = 3\nreputation = 0\n", 1)
		}, "quest: ランク 3 に評判を得られる依頼がありません", nil},
		{"TC-FR-MOD-04-56 同ランクのレシピ連鎖", func(s string) string { return s }, "", nil},
		{"TC-FR-MOD-04-57 上位ランクの採取品は材料にならない", func(s string) string {
			return strings.Replace(s, "inputs = [{item=\"herb\", count=1}]", "inputs = [{item=\"ore\", count=1}]", 1)
		}, "recipe[id=essence].inputs: ランク 1 で材料が手に入りません", nil},
		{"TC-FR-MOD-04-58 属性入力枠に合う店の品", func(s string) string { return s }, "", nil},
		{"TC-FR-MOD-04-68 同ランクの採取品を材料にできる", func(s string) string {
			return strings.Replace(s, "inputs = [{attribute=\"plant\", count=1}]", "inputs = [{item=\"ore\", count=1}]", 1)
		}, "", nil},
		{"TC-FR-MOD-04-69 戦闘報酬だけの品を材料にできる", func(s string) string {
			s = strings.Replace(s, "shop =", "battle_width = 8\nbattle_height = 6\ninjury_seconds = 60\nparty = [{unit=\"hero\", x=0, y=0}]\nshop =", 1)
			s = strings.Replace(s, "[[gather]]", "[[unit]]\nid = \"hero\"\nrole = \"front\"\nhp = 10\nattack = 1\nspeed = 1\nmove = 1\nrange = 1\n[[unit]]\nid = \"foe\"\nrole = \"enemy\"\nhp = 10\nattack = 1\nspeed = 1\nmove = 1\nrange = 1\n[[gather]]", 1)
			s = strings.Replace(s, "[[item]]\nid = \"ore\"", "[[item]]\nid = \"battle_only\"\n[[item]]\nid = \"ore\"", 1)
			s = strings.Replace(s, "inputs = [{attribute=\"plant\", count=1}]", "inputs = [{item=\"battle_only\", count=1}]", 1)
			return strings.Replace(s, "drops = [{item=\"ore\", weight=1, quality_min=1, quality_max=1}]", "drops = [{item=\"ore\", weight=1, quality_min=1, quality_max=1}]\n[gather.battle]\nenemies = [{unit=\"foe\", x=7, y=5}]\nrolls = 1\ndrops = [{item=\"battle_only\", weight=1, quality_min=1, quality_max=1}]", 1)
		}, "", nil},
		{"TC-FR-MOD-04-70 属性に合う入手品がない", func(s string) string {
			return strings.Replace(s, "inputs = [{attribute=\"plant\", count=1}]", "inputs = [{attribute=\"metal\", count=1}]", 1)
		}, "recipe[id=elixir].inputs: ランク 3 で材料が手に入りません", nil},
		{"TC-FR-MOD-04-71 二つ目の入力枠が満たされない", func(s string) string {
			return strings.Replace(s, "inputs = [{item=\"essence\", count=1}]", "inputs = [{item=\"essence\", count=1}, {item=\"ore\", count=1}]", 1)
		}, "recipe[id=potion].inputs: ランク 1 で材料が手に入りません", nil},
		{"TC-FR-MOD-04-59 ランク 2 の依頼に作れない物", func(s string) string {
			return strings.Replace(s, "id = \"q4\"\nitem = \"potion\"", "id = \"q4\"\nitem = \"elixir\"", 1)
		}, "quest[id=q4].item: ランク 2 で手に入りません", nil},
		{"TC-FR-MOD-04-60 ランク定義が空なら進行検証をしない", func(s string) string {
			s = strings.Replace(s, "rank_reputation = [3, 8, 15]\n", "", 1)
			s = strings.ReplaceAll(s, "rank = 2\n", "")
			s = strings.ReplaceAll(s, "rank = 3\n", "")
			s = strings.Replace(s, "reputation = 1", "reputation = 0", 1)
			s = s[:strings.Index(s, "[[quest]]\nid = \"q3\"")]
			s = strings.Replace(s, "[[item]]\nid = \"ore\"", "[[item]]\nid = \"unobtainable\"\n[[item]]\nid = \"ore\"", 1)
			return strings.Replace(s, "inputs = [{item=\"herb\", count=1}]", "inputs = [{item=\"unobtainable\", count=1}]", 1)
		}, "", nil},
		{"TC-FR-MOD-04-61 依頼のランク上限", func(s string) string {
			return strings.Replace(s, "rank = 3\nreputation = 1", "rank = 5\nreputation = 1", 1)
		}, "quest[id=q5].rank", nil},
		{"TC-FR-MOD-04-62 最高ランクの依頼は評判 0 でもよい", func(s string) string {
			return s + "[[quest]]\nid = \"q6\"\nitem = \"elixir\"\nmin_quality = 1\ncount = 1\nreward = 1\nrank = 4\nreputation = 0\n"
		}, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Load(fstest.MapFS{"definitions.toml": &fstest.MapFile{Data: []byte(tt.edit(rankData))}})
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("want %q, got %v", tt.want, err)
				}
				if strings.HasPrefix(tt.name, "TC-FR-MOD-04-50 ") && !strings.Contains(err.Error(), "game.rank_reputation[2]") {
					t.Fatalf("want game.rank_reputation[2], got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func TestIT04LegacyWithoutRankReputation(t *testing.T) {
	t.Run("TC-FR-MOD-04-63 反復 3 の定義は進行検証を受けない", func(t *testing.T) {
		if _, err := Load(fstest.MapFS{"it03.toml": &fstest.MapFile{Data: []byte(richIT03Data)}}); err != nil {
			t.Fatal(err)
		}
	})
}
