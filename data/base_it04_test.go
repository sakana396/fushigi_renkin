package data_test

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/core"
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
		want := map[string]int{"flour": 1, "bread_flour": 1, "bread_whole_wheat": 1, "cheese": 1, "burger_cheese": 1, "burger_double": 1, "potion_herb": 1, "potion_moss": 2, "glow_crystal": 1, "bomb": 2, "tonic": 3, "lantern": 3, "deluxe_burger": 4}
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
			"q_burger":        {ID: "q_burger", Rank: 1, Reputation: 2, Item: "burger", MinQuality: 20, Count: 2, Reward: 120},
			"q_bread":         {ID: "q_bread", Rank: 1, Reputation: 1, Item: "bread", MinQuality: 20, Count: 3, Reward: 65},
			"q_cheese":        {ID: "q_cheese", Rank: 1, Reputation: 1, Item: "cheese", MinQuality: 20, Count: 2, Reward: 80},
			"q_potion":        {ID: "q_potion", Rank: 2, Reputation: 3, Item: "potion", MinQuality: 50, Count: 2, Reward: 70},
			"q_bomb":          {ID: "q_bomb", Rank: 2, Reputation: 3, Item: "bomb", MinQuality: 50, Count: 2, Reward: 80},
			"q_tonic":         {ID: "q_tonic", Rank: 3, Reputation: 4, Item: "tonic", MinQuality: 100, Count: 2, Reward: 160},
			"q_lantern":       {ID: "q_lantern", Rank: 3, Reputation: 4, Item: "lantern", MinQuality: 100, Count: 1, Reward: 130},
			"q_deluxe":        {ID: "q_deluxe", Rank: 4, Item: "deluxe_burger", MinQuality: 150, Count: 1, Reward: 250},
			"q_potion_master": {ID: "q_potion_master", Rank: 4, Item: "potion", MinQuality: 150, Count: 2, Reward: 180},
		}
		for _, id := range []string{"q_burger", "q_bread", "q_cheese"} {
			q := want[id]
			q.ID = id + "_extra"
			want[q.ID] = q
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
		// ランクの検査だけを外し、実際の錬金規則と在庫の検査を通す。
		unlocked := c
		unlocked.Game.RankReputation = nil
		type ingredient struct {
			item    string
			quality int
			from    string
			count   int
		}
		type route struct {
			name, recipe     string
			slots            [][]ingredient
			average, quality int
		}
		routes := []route{
			{"チーズ", "cheese", [][]ingredient{{{"milk", 15, "", 2}}, {{"herb", 25, "", 1}}}, 18, 28},
			{"パン", "bread_whole_wheat", [][]ingredient{{{"wheat", 25, "", 3}}, {{"water", 15, "", 1}}}, 22, 30},
			{"バーガー", "burger_double", [][]ingredient{{{"bread", 0, "パン", 1}}, {{"meat", 15, "", 2}}}, 20, 35},
			{"回復薬・丘", "potion_herb", [][]ingredient{{{"herb", 50, "", 2}}, {{"water", 50, "", 1}}}, 50, 65},
			{"爆弾", "bomb", [][]ingredient{{{"ore", 50, "", 2}}, {{"glow_moss", 50, "", 1}}}, 50, 60},
			{"回復薬・森", "potion_herb", [][]ingredient{{{"herb", 160, "", 2}}, {{"water", 160, "", 1}}}, 160, 175},
			{"元気の薬", "tonic", [][]ingredient{{{"potion", 0, "回復薬・森", 1}}, {{"glow_moss", 160, "", 1}}}, 167, 177},
			{"ひかり結晶", "glow_crystal", [][]ingredient{{{"ore", 160, "", 2}}, {{"glow_moss", 160, "", 1}}}, 160, 172},
			{"ランタン", "lantern", [][]ingredient{{{"glow_crystal", 0, "ひかり結晶", 1}}, {{"ore", 160, "", 1}}}, 166, 176},
			{"パン・森", "bread_whole_wheat", [][]ingredient{{{"wheat", 160, "", 3}}, {{"water", 160, "", 1}}}, 160, 168},
			{"バーガー・森", "burger_double", [][]ingredient{{{"bread", 0, "パン・森", 1}}, {{"meat", 160, "", 2}}}, 162, 177},
			{"ごちそうバーガー", "deluxe_burger", [][]ingredient{{{"burger", 0, "バーガー・森", 1}}, {{"tonic", 0, "元気の薬", 1}}}, 177, 192},
		}
		products := map[string]int{}
		for _, tc := range routes {
			t.Run(tc.name, func(t *testing.T) {
				w := core.NewWorld(unlocked, 1)
				sel := core.Selection{}
				for _, slot := range tc.slots {
					var stacks []core.Stack
					for _, in := range slot {
						quality := in.quality
						if in.from != "" {
							var ok bool
							quality, ok = products[in.from]
							if !ok {
								t.Fatalf("前段 %s がない", in.from)
							}
						}
						key := core.ItemKey{ItemID: in.item, Quality: quality}
						if err := w.Inventory.Add(key, in.count); err != nil {
							t.Fatal(err)
						}
						stacks = append(stacks, core.Stack{Key: key, Count: in.count})
					}
					sel.Slots = append(sel.Slots, stacks)
				}
				got, err := w.PreviewCraft(tc.recipe, sel)
				if err != nil {
					t.Fatal(err)
				}
				if got.AverageQuality != tc.average || got.Quality != tc.quality {
					t.Errorf("平均 %d・完成 %d、期待 平均 %d・完成 %d", got.AverageQuality, got.Quality, tc.average, tc.quality)
				}
				products[tc.name] = got.Quality
			})
		}
	})
}

// 探索の勝算を、本体データと core の戦闘操作で確かめる（DES-DAT-02、DES-BTL-03）。
func TestBaseBattleRoutesIT04(t *testing.T) {
	catalog, err := defload.Load(data.Base)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Game.RankReputation = nil
	type action struct {
		move             bool
		x, y             int
		attack           bool
		targetX, targetY int
		hp               [4]int
	}
	for _, tc := range []struct {
		id, site string
		steps    []action
	}{
		{"TC-FR-MOD-02-47", "sunny_hill", []action{
			{true, 2, 3, false, 0, 0, [4]int{30, 18, 16, 14}},
			{true, 3, 2, false, 0, 0, [4]int{30, 18, 16, 14}},
			{true, 4, 3, true, 5, 2, [4]int{30, 18, 11, 14}},
			{true, 4, 2, true, 5, 2, [4]int{26, 15, 5, 14}},
			{false, 0, 0, true, 5, 3, [4]int{26, 15, 5, 9}},
			{false, 0, 0, true, 5, 2, [4]int{26, 12, 0, 9}},
			{false, 0, 0, true, 5, 3, [4]int{26, 12, 0, 4}},
			{true, 5, 2, true, 5, 3, [4]int{26, 12, 0, 0}},
		}},
		{"TC-FR-MOD-02-48", "crystal_forest", []action{
			{true, 2, 3, false, 0, 0, [4]int{30, 18, 20, 16}},
			{true, 3, 2, false, 0, 0, [4]int{30, 18, 20, 16}},
			{true, 4, 3, true, 5, 2, [4]int{30, 18, 15, 16}},
			{true, 4, 2, true, 5, 2, [4]int{25, 14, 9, 16}},
			{false, 0, 0, true, 5, 3, [4]int{25, 14, 9, 11}},
			{false, 0, 0, true, 5, 2, [4]int{20, 10, 3, 11}},
			{false, 0, 0, true, 5, 3, [4]int{20, 10, 3, 6}},
			{false, 0, 0, true, 5, 2, [4]int{20, 6, 0, 6}},
			{false, 0, 0, true, 5, 3, [4]int{20, 6, 0, 1}},
			{true, 5, 2, true, 5, 3, [4]int{20, 6, 0, 0}},
		}},
	} {
		t.Run(tc.id, func(t *testing.T) {
			w := core.NewWorld(catalog, 1)
			if err := w.StartBattle(tc.site); err != nil {
				t.Fatal(err)
			}
			for i, step := range tc.steps {
				battle := w.Battle
				if step.move {
					if err := w.BattleMove(step.x, step.y); err != nil {
						t.Fatalf("行動 %d の移動: %v", i+1, err)
					}
				}
				if step.attack {
					err = w.BattleAttack(step.targetX, step.targetY)
				} else {
					err = w.BattleWait()
				}
				if err != nil {
					t.Fatalf("行動 %d: %v", i+1, err)
				}
				if w.Battle == nil && !w.LastBattle.Won {
					t.Fatalf("行動 %d で敗北", i+1)
				}
				for j, hp := range step.hp {
					if battle.Units[j].HP != hp {
						t.Errorf("行動 %d の %s: HP %d、期待 %d", i+1, battle.Units[j].Label, battle.Units[j].HP, hp)
					}
				}
			}
			if w.LastBattle == nil || !w.LastBattle.Won || w.Battle != nil {
				t.Fatalf("探索に勝てない: %+v", w.LastBattle)
			}
		})
	}
}
