package data_test

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 本体データの解放、品質の経路、戦闘の前提を確認する（DES-DAT-02、DES-DEF-08、DES-VAL-14〜16）。
func TestBaseDataIT04(t *testing.T) {
	c, err := defload.Load(data.Base)
	if err != nil {
		t.Fatalf("DES-VAL-14〜16 を含む本体データの検証: %v", err)
	}
	t.Run("TC-FR-MOD-02-40 評判の閾値と読み込み検証", func(t *testing.T) {
		if !reflect.DeepEqual(c.Game.RankReputation, []int{3, 8, 15}) {
			t.Errorf("閾値 = %v", c.Game.RankReputation)
		}
	})
	t.Run("TC-FR-MOD-02-41 レシピの解放ランク", func(t *testing.T) {
		want := map[string]int{"flour": 1, "bread_flour": 1, "bread_whole_wheat": 1, "cheese": 1, "burger_cheese": 1, "burger_double": 1, "potion_herb": 2, "potion_moss": 2, "glow_crystal": 2, "bomb": 2, "tonic": 3, "lantern": 3, "deluxe_burger": 4}
		if len(c.Recipes) != len(want) {
			t.Errorf("レシピ件数 = %d", len(c.Recipes))
		}
		for id, rank := range want {
			if c.Recipes[id].Rank != rank {
				t.Errorf("%s: rank = %d、期待 %d", id, c.Recipes[id].Rank, rank)
			}
		}
	})
	t.Run("TC-FR-MOD-02-42 依頼のランクと評判", func(t *testing.T) {
		want := map[string]defs.QuestDef{
			"q_burger":  {ID: "q_burger", Rank: 1, Reputation: 2, Item: "burger", MinQuality: 20, Count: 2, Reward: 120},
			"q_bread":   {ID: "q_bread", Rank: 1, Reputation: 1, Item: "bread", MinQuality: 20, Count: 3, Reward: 65},
			"q_cheese":  {ID: "q_cheese", Rank: 1, Reputation: 1, Item: "cheese", MinQuality: 20, Count: 2, Reward: 45},
			"q_potion":  {ID: "q_potion", Rank: 2, Reputation: 3, Item: "potion", MinQuality: 50, Count: 2, Reward: 70},
			"q_bomb":    {ID: "q_bomb", Rank: 2, Reputation: 3, Item: "bomb", MinQuality: 50, Count: 2, Reward: 80},
			"q_tonic":   {ID: "q_tonic", Rank: 3, Reputation: 4, Item: "tonic", MinQuality: 100, Count: 2, Reward: 160},
			"q_lantern": {ID: "q_lantern", Rank: 3, Reputation: 4, Item: "lantern", MinQuality: 100, Count: 1, Reward: 130},
			"q_deluxe":  {ID: "q_deluxe", Rank: 4, Item: "deluxe_burger", MinQuality: 150, Count: 1, Reward: 250},
		}
		if !reflect.DeepEqual(c.Quests, want) {
			t.Errorf("依頼 = %+v、期待 %+v", c.Quests, want)
		}
	})
	t.Run("TC-FR-MOD-02-43 採取地と敵の解放ランク", func(t *testing.T) {
		for id, rank := range map[string]int{"meadow": 1, "spring_cave": 1, "sunny_hill": 2, "crystal_forest": 3} {
			if c.Gathers[id].Rank != rank {
				t.Errorf("%s: rank = %d、期待 %d", id, c.Gathers[id].Rank, rank)
			}
		}
		for _, id := range []string{"beetle", "stone_golem"} {
			if c.Units[id].Role != "enemy" {
				t.Errorf("%s は敵ではない", id)
			}
		}
	})
	t.Run("TC-FR-MOD-02-45 戦闘の能力と配置", func(t *testing.T) {
		for id, want := range map[string]defs.UnitDef{
			"swordsman":   {ID: "swordsman", Name: "剣士", Role: "front", HP: 30, Attack: 6, Speed: 5, Move: 2, Range: 1},
			"mage":        {ID: "mage", Name: "術士", Role: "back", HP: 18, Attack: 5, Speed: 7, Move: 2, Range: 2},
			"slime":       {ID: "slime", Name: "スライム", Role: "enemy", HP: 14, Attack: 3, Speed: 4, Move: 1, Range: 1},
			"beetle":      {ID: "beetle", Name: "甲殻虫", Role: "enemy", HP: 16, Attack: 4, Speed: 3, Move: 1, Range: 1},
			"stone_golem": {ID: "stone_golem", Name: "石ゴーレム", Role: "enemy", HP: 20, Attack: 5, Speed: 2, Move: 1, Range: 1},
		} {
			if c.Units[id] != want {
				t.Errorf("%s: %+v、期待 %+v", id, c.Units[id], want)
			}
		}
		for _, tc := range []struct {
			site    string
			enemies []defs.Placement
		}{
			{"sunny_hill", []defs.Placement{{Unit: "beetle", X: 6, Y: 2}, {Unit: "slime", X: 6, Y: 4}}},
			{"crystal_forest", []defs.Placement{{Unit: "stone_golem", X: 6, Y: 2}, {Unit: "beetle", X: 6, Y: 4}}},
		} {
			if !reflect.DeepEqual(c.Gathers[tc.site].Battle.Enemies, tc.enemies) {
				t.Errorf("%s の敵: %+v", tc.site, c.Gathers[tc.site].Battle.Enemies)
			}
		}
	})
	t.Run("TC-FR-MOD-02-46 品質計算で使う素材の入手元", func(t *testing.T) {
		for _, tc := range []struct {
			site, item string
			battle     bool
			quality    int
		}{
			{"meadow", "wheat", false, 25}, {"meadow", "herb", false, 25},
			{"sunny_hill", "herb", false, 50}, {"sunny_hill", "water", false, 50}, {"sunny_hill", "ore", false, 50}, {"sunny_hill", "glow_moss", false, 50},
			{"crystal_forest", "wheat", true, 160}, {"crystal_forest", "water", true, 160}, {"crystal_forest", "meat", true, 160}, {"crystal_forest", "herb", true, 160}, {"crystal_forest", "glow_moss", true, 160}, {"crystal_forest", "ore", true, 160},
		} {
			g := c.Gathers[tc.site]
			drops := g.Drops
			if tc.battle {
				drops = g.Battle.Drops
			}
			found := false
			for _, d := range drops {
				if d.Item == tc.item && d.QualityMin <= tc.quality && tc.quality <= d.QualityMax {
					found = true
				}
			}
			if !found {
				t.Errorf("%s の %s は品質 %d で得られない", tc.site, tc.item, tc.quality)
			}
		}
	})
	t.Run("TC-FR-MAT-03-10 / TC-FR-MAT-03-11 ランク別の採取と探索の品質", func(t *testing.T) {
		for _, tc := range []struct {
			id                                         string
			gatherMin, gatherMax, battleMin, battleMax int
		}{
			{"sunny_hill", 20, 50, 60, 120}, {"crystal_forest", 40, 80, 90, 160},
		} {
			g := c.Gathers[tc.id]
			if g.Battle == nil || len(g.Drops) == 0 || len(g.Battle.Drops) == 0 {
				t.Errorf("%s の採取・探索報酬がない", tc.id)
				continue
			}
			for _, d := range g.Drops {
				if d.QualityMin != tc.gatherMin || d.QualityMax != tc.gatherMax {
					t.Errorf("%s 採取 %+v", tc.id, d)
				}
			}
			for _, d := range g.Battle.Drops {
				if d.QualityMin != tc.battleMin || d.QualityMax != tc.battleMax {
					t.Errorf("%s 探索 %+v", tc.id, d)
				}
			}
		}
	})
	t.Run("TC-FR-MOD-02-44 / TC-FR-MAT-03-12 品質条件に届く素材とボーナス", func(t *testing.T) {
		for _, tc := range []struct {
			recipe    string
			items     []string
			qualities []int
			want      int
		}{
			{"cheese", []string{"milk", "milk", "herb"}, []int{15, 15, 25}, 28},
			{"bread_whole_wheat", []string{"wheat", "wheat", "wheat", "water"}, []int{25, 25, 25, 15}, 30},
			{"burger_double", []string{"bread", "meat", "meat"}, []int{30, 15, 15}, 35},
			{"potion_herb", []string{"herb", "herb", "water"}, []int{50, 50, 50}, 65},
			{"bomb", []string{"ore", "ore", "glow_moss"}, []int{50, 50, 50}, 60},
			{"potion_herb", []string{"herb", "herb", "water"}, []int{160, 160, 160}, 175},
			{"tonic", []string{"potion", "glow_moss"}, []int{175, 160}, 177},
			{"glow_crystal", []string{"ore", "ore", "glow_moss"}, []int{160, 160, 160}, 172},
			{"lantern", []string{"glow_crystal", "ore"}, []int{172, 160}, 176},
			{"bread_whole_wheat", []string{"wheat", "wheat", "wheat", "water"}, []int{160, 160, 160, 160}, 168},
			{"burger_double", []string{"bread", "meat", "meat"}, []int{168, 160, 160}, 177},
			{"deluxe_burger", []string{"burger", "tonic"}, []int{177, 177}, 192},
		} {
			r := c.Recipes[tc.recipe]
			if len(tc.items) != len(tc.qualities) {
				t.Fatal("テストデータの件数が違う")
			}
			cursor := 0
			for _, slot := range r.Inputs {
				for n := 0; n < slot.Count; n++ {
					if cursor >= len(tc.items) {
						t.Fatalf("%s の素材数が違う", tc.recipe)
					}
					id := tc.items[cursor]
					if slot.Item != "" && slot.Item != id {
						t.Errorf("%s の素材 %d: %s、期待 %s", tc.recipe, cursor, id, slot.Item)
					}
					if slot.Attribute != "" {
						found := false
						for _, attr := range c.Items[id].Attributes {
							found = found || attr == slot.Attribute
						}
						if !found {
							t.Errorf("%s の素材 %s に属性 %s がない", tc.recipe, id, slot.Attribute)
						}
					}
					cursor++
				}
			}
			if cursor != len(tc.items) {
				t.Errorf("%s の素材数: %d、期待 %d", tc.recipe, cursor, len(tc.items))
			}
			total := 0
			for i, q := range tc.qualities {
				total += q
				_ = c.Items[tc.items[i]]
			}
			got := total / len(tc.qualities)
			for _, b := range r.Bonuses {
				count := 0
				for _, id := range tc.items {
					for _, attr := range c.Items[id].Attributes {
						if attr == b.Attribute {
							count++
						}
					}
				}
				if count >= b.MinCount {
					got += b.Add
				}
			}
			if got != tc.want {
				t.Errorf("%s: 品質 %d、期待 %d", tc.recipe, got, tc.want)
			}
		}
	})
}
