package data_test

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 反復 3 の本体データを設計値と照合する（DES-DAT-01、DES-DEF-05〜07）。
func TestBaseDataIT03(t *testing.T) {
	c, err := defload.Load(data.Base)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("TC-FR-MOD-02-20 消耗品の段階と爆弾レシピ", func(t *testing.T) {
		want := map[string]defs.UseDef{
			"potion": {Target: "ally", Range: 1, Stages: []defs.UseStage{{MinQuality: 1, Effects: []defs.UseEffect{{Kind: "heal", Value: 50}}}, {MinQuality: 50, Radius: 1}}},
			"burger": {Target: "ally", Range: 1, Stages: []defs.UseStage{{MinQuality: 1, Effects: []defs.UseEffect{{Kind: "heal", Value: 30}}}, {MinQuality: 100, Effects: []defs.UseEffect{{Kind: "attack_up", Value: 5}}}, {MinQuality: 150, Effects: []defs.UseEffect{{Kind: "speed_up", Value: 3}}}}},
			"bomb":   {Target: "enemy", Range: 3, Stages: []defs.UseStage{{MinQuality: 1, Effects: []defs.UseEffect{{Kind: "damage", Value: 40}}}, {MinQuality: 50, Radius: 1}, {MinQuality: 100, Effects: []defs.UseEffect{{Kind: "damage", Value: 20}}}}},
		}
		for id, expected := range want {
			got, ok := c.Item(id)
			if !ok || got.Use == nil || !reflect.DeepEqual(*got.Use, expected) {
				t.Errorf("%s の使用定義: got=%+v want=%+v", id, got.Use, expected)
			}
		}
		bomb := c.Items["bomb"]
		if bomb.Name != "爆弾" || bomb.BasePrice != 30 || !reflect.DeepEqual(bomb.Attributes, []string{"鉱石", "発光"}) {
			t.Errorf("爆弾のアイテム: %+v", bomb)
		}
		r := c.Recipes["bomb"]
		if r.Output != "bomb" || !reflect.DeepEqual(r.Inputs, []defs.InputSlot{{Item: "ore", Count: 2}, {Item: "glow_moss", Count: 1}}) || !reflect.DeepEqual(r.Bonuses, []defs.Bonus{{Attribute: "発光", MinCount: 1, Add: 10}}) {
			t.Errorf("爆弾のレシピ: %+v", r)
		}
	})
	t.Run("TC-FR-MOD-02-21 ユニットと仲間の盤面", func(t *testing.T) {
		want := []defs.UnitDef{
			{ID: "swordsman", Name: "剣士", Role: "front", HP: 30, Attack: 6, Speed: 5, Move: 2, Range: 1},
			{ID: "mage", Name: "術士", Role: "back", HP: 18, Attack: 5, Speed: 7, Move: 2, Range: 2},
			{ID: "slime", Name: "スライム", Role: "enemy", HP: 14, Attack: 3, Speed: 4, Move: 1, Range: 1},
			{ID: "wolf", Name: "オオカミ", Role: "enemy", HP: 18, Attack: 5, Speed: 8, Move: 3, Range: 1},
		}
		for _, expected := range want {
			if got := c.Units[expected.ID]; !reflect.DeepEqual(got, expected) {
				t.Errorf("%s: got=%+v want=%+v", expected.ID, got, expected)
			}
		}
		if c.Game.BattleWidth != 8 || c.Game.BattleHeight != 6 || c.Game.InjurySeconds != 60 || !reflect.DeepEqual(c.Game.Party, []defs.Placement{{Unit: "swordsman", X: 1, Y: 2}, {Unit: "mage", X: 0, Y: 3}}) {
			t.Errorf("盤面と仲間: %+v", c.Game)
		}
	})
	t.Run("TC-FR-MOD-02-22 探索と採取の品質", func(t *testing.T) {
		for _, id := range []string{"meadow", "spring_cave"} {
			g := c.Gathers[id]
			if g.Battle == nil || g.Battle.Rolls != 2 || len(g.Battle.Enemies) < 2 || len(g.Battle.Enemies) > 3 || len(g.Battle.Drops) == 0 {
				t.Errorf("%s の戦闘: %+v", id, g.Battle)
				continue
			}
			for _, d := range g.Drops {
				if d.QualityMin != 5 || d.QualityMax != 25 {
					t.Errorf("%s の採取品質: %+v", id, d)
				}
			}
			for _, d := range g.Battle.Drops {
				if d.QualityMin != 40 || d.QualityMax != 90 {
					t.Errorf("%s の探索品質: %+v", id, d)
				}
			}
		}
		if c.Gathers["meadow"].Battle == nil || c.Gathers["spring_cave"].Battle == nil {
			return
		}
		if got := c.Gathers["meadow"].Battle.Enemies; !reflect.DeepEqual(got, []defs.Placement{{Unit: "slime", X: 6, Y: 2}, {Unit: "slime", X: 6, Y: 4}}) {
			t.Errorf("草原の敵: %+v", got)
		}
		if got := c.Gathers["spring_cave"].Battle.Enemies; !reflect.DeepEqual(got, []defs.Placement{{Unit: "slime", X: 6, Y: 2}, {Unit: "wolf", X: 6, Y: 4}}) {
			t.Errorf("洞窟の敵: %+v", got)
		}
	})
	t.Run("TC-FR-MOD-02-23 依頼は製品だけで報酬が売値を超える", func(t *testing.T) {
		want := []defs.QuestDef{
			{ID: "q_burger", Item: "burger", MinQuality: 20, Count: 2, Reward: 120},
			{ID: "q_potion", Item: "potion", MinQuality: 20, Count: 2, Reward: 70},
			{ID: "q_bomb", Item: "bomb", MinQuality: 20, Count: 2, Reward: 80},
			{ID: "q_bread", Item: "bread", MinQuality: 20, Count: 3, Reward: 65},
		}
		products := map[string]bool{}
		for _, r := range c.Recipes {
			products[r.Output] = true
		}
		for _, expected := range want {
			got := c.Quests[expected.ID]
			if !reflect.DeepEqual(got, expected) || !products[got.Item] {
				t.Errorf("%s: got=%+v want=%+v", expected.ID, got, expected)
			}
			item := c.Items[got.Item]
			sell := item.BasePrice * (50 + got.MinQuality) / 100 * got.Count
			if got.Reward <= sell {
				t.Errorf("%s: 報酬 %d <= 売値 %d", got.ID, got.Reward, sell)
			}
		}
	})
	t.Run("TC-FR-MOD-02-24 店は素材だけで品質15", func(t *testing.T) {
		want := []defs.ShopOffer{{Item: "milk", Quality: 15, Price: 16}, {Item: "meat", Quality: 15, Price: 25}, {Item: "water", Quality: 15, Price: 5}}
		if !reflect.DeepEqual(c.Game.Shop, want) {
			t.Errorf("店: got=%+v want=%+v", c.Game.Shop, want)
		}
		products := map[string]bool{}
		for _, r := range c.Recipes {
			products[r.Output] = true
		}
		for _, offer := range c.Game.Shop {
			if products[offer.Item] {
				t.Errorf("店に製品 %s がある", offer.Item)
			}
		}
	})
	t.Run("TC-FR-MOD-02-25 施設費用と序盤の売値", func(t *testing.T) {
		for _, tc := range []struct {
			id    string
			money []int
		}{{"workshop", []int{60, 140}}, {"market", []int{70}}} {
			f := c.Facilities[tc.id]
			if len(f.Levels) != len(tc.money) {
				t.Errorf("%s の段階数: %d", tc.id, len(f.Levels))
				continue
			}
			for i, want := range tc.money {
				if got := f.Levels[i].Money; got != want {
					t.Errorf("%s[%d] の費用: got=%d want=%d", tc.id, i, got, want)
				}
			}
		}
		for _, tc := range []struct {
			item          string
			quality, sell int
		}{{"potion", 20, 19}, {"burger", 20, 38}, {"bomb", 20, 21}} {
			item := c.Items[tc.item]
			if got := item.BasePrice * (50 + tc.quality) / 100; got != tc.sell {
				t.Errorf("%s の品質%dの売値: got=%d want=%d", tc.item, tc.quality, got, tc.sell)
			}
		}
	})
	t.Run("TC-FR-MOD-02-26 序盤の製品品質", func(t *testing.T) {
		// DES-QLT-02 の入力平均は個数で重み付けし、各錬金で端数を切り捨てる。
		for _, tc := range []struct {
			id     string
			inputs []defs.InputSlot
			bonus  defs.Bonus
		}{
			{"potion_herb", []defs.InputSlot{{Attribute: "草", Count: 2}, {Item: "water", Count: 1}}, defs.Bonus{Attribute: "癒し", MinCount: 2, Add: 15}},
			{"bread_whole_wheat", []defs.InputSlot{{Item: "wheat", Count: 3}, {Item: "water", Count: 1}}, defs.Bonus{Attribute: "穀物", MinCount: 3, Add: 8}},
			{"burger_double", []defs.InputSlot{{Item: "bread", Count: 1}, {Item: "meat", Count: 2}}, defs.Bonus{Attribute: "活力", MinCount: 2, Add: 15}},
		} {
			r := c.Recipes[tc.id]
			if !reflect.DeepEqual(r.Inputs, tc.inputs) || !reflect.DeepEqual(r.Bonuses, []defs.Bonus{tc.bonus}) {
				t.Errorf("%s の素材とボーナス: %+v", tc.id, r)
			}
		}
		for _, tc := range []struct {
			name string
			got  int
			want int
		}{
			{"回復薬の下限", (5*2+5)/3 + c.Recipes["potion_herb"].Bonuses[0].Add, 20},
			{"回復薬の上限", (25*2+25)/3 + c.Recipes["potion_herb"].Bonuses[0].Add, 40},
			{"パンの下限", (5*3+5)/4 + c.Recipes["bread_whole_wheat"].Bonuses[0].Add, 13},
			{"パンの上限", (25*3+25)/4 + c.Recipes["bread_whole_wheat"].Bonuses[0].Add, 33},
			{"ハンバーガーの下限", (13+15*2)/3 + c.Recipes["burger_double"].Bonuses[0].Add, 29},
			{"ハンバーガーの上限", (33+15*2)/3 + c.Recipes["burger_double"].Bonuses[0].Add, 36},
			{"爆弾の下限", (5*2+5)/3 + c.Recipes["bomb"].Bonuses[0].Add, 15},
			{"爆弾の上限", (25*2+25)/3 + c.Recipes["bomb"].Bonuses[0].Add, 35},
		} {
			if tc.got != tc.want {
				t.Errorf("%s: got=%d want=%d", tc.name, tc.got, tc.want)
			}
		}
	})
	t.Run("TC-FR-MOD-02-27 設計書の効果計算例", func(t *testing.T) {
		for _, tc := range []struct {
			item    string
			quality int
			radius  int
			effects map[string]int
		}{
			{"bomb", 20, 0, map[string]int{"damage": 8}},
			{"bomb", 60, 1, map[string]int{"damage": 24}},
			{"bomb", 120, 1, map[string]int{"damage": 72}},
			{"burger", 30, 0, map[string]int{"heal": 9}},
			{"burger", 100, 0, map[string]int{"heal": 30, "attack_up": 5}},
			{"burger", 160, 0, map[string]int{"heal": 48, "attack_up": 8, "speed_up": 4}},
			{"potion", 20, 0, map[string]int{"heal": 10}},
		} {
			item := c.Items[tc.item]
			if item.Use == nil {
				t.Errorf("%s に使用定義がない", tc.item)
				continue
			}
			radius := 0
			sums := map[string]int{}
			for _, stage := range item.Use.Stages {
				if stage.MinQuality > tc.quality {
					continue
				}
				if stage.Radius > radius {
					radius = stage.Radius
				}
				for _, effect := range stage.Effects {
					sums[effect.Kind] += effect.Value
				}
			}
			got := map[string]int{}
			for kind, value := range sums {
				got[kind] = value * tc.quality / 100
			}
			if radius != tc.radius || !reflect.DeepEqual(got, tc.effects) {
				t.Errorf("%s Q%d: 範囲=%d 効果=%v、期待は範囲=%d 効果=%v", tc.item, tc.quality, radius, got, tc.radius, tc.effects)
			}
		}
	})
}
