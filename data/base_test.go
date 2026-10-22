package data_test

import (
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 組み込み MOD の規模とレシピの関係を確認する（DES-PKG-06、DES-DEF-02）。
func TestBaseData(t *testing.T) {
	tests := []struct {
		name  string
		check func(*testing.T, defs.Catalog)
	}{
		{"TC-FR-MOD-02-04 本体データの読み込み", func(t *testing.T, c defs.Catalog) {
			materials := []string{"wheat", "water", "milk", "meat", "herb", "ore", "glow_moss"}
			products := []string{"flour", "bread", "cheese", "burger", "deluxe_burger", "potion", "tonic", "glow_crystal", "lantern", "bomb", "iron_ingot", "herb_rope"}
			if len(c.Items) != len(materials)+len(products) || len(c.Recipes) != 15 || len(c.Facilities) != 4 || len(c.Golems) != 1 || len(c.Gathers) != 4 || len(c.Game.Shop) != 3 || len(c.Units) != 6 || len(c.Quests) != 17 || len(c.Maps) != 5 {
				t.Fatalf("定義の数: items=%d recipes=%d facilities=%d golems=%d gathers=%d shop=%d units=%d quests=%d maps=%d", len(c.Items), len(c.Recipes), len(c.Facilities), len(c.Golems), len(c.Gathers), len(c.Game.Shop), len(c.Units), len(c.Quests), len(c.Maps))
			}
			for _, id := range append(materials, products...) {
				if _, ok := c.Item(id); !ok {
					t.Errorf("アイテム %s がない", id)
				}
			}
		}},
		{"TC-FR-ITM-01-01 属性は最大 3 個", func(t *testing.T, c defs.Catalog) {
			hasThree := false
			for _, item := range c.Items {
				if len(item.Attributes) > 3 {
					t.Errorf("%s の属性が %d 個", item.ID, len(item.Attributes))
				}
				if len(item.Attributes) == 3 {
					hasThree = true
				}
			}
			if !hasThree {
				t.Error("属性が 3 個の境界例がない")
			}
		}},
		{"TC-FR-ALC-03-04 3 種類の別レシピ", func(t *testing.T, c defs.Catalog) {
			counts := map[string]int{}
			for _, recipe := range c.Recipes {
				counts[recipe.Output]++
			}
			for _, id := range []string{"bread", "burger", "potion"} {
				if counts[id] != 2 {
					t.Errorf("%s のレシピ数 = %d", id, counts[id])
				}
			}
		}},
		{"TC-FR-ALC-06-04 3 段の製品連鎖", func(t *testing.T, c defs.Catalog) {
			for _, step := range []struct{ input, output string }{{"flour", "bread"}, {"bread", "burger"}, {"burger", "deluxe_burger"}} {
				found := false
				for _, recipe := range c.Recipes {
					if recipe.Output != step.output {
						continue
					}
					for _, input := range recipe.Inputs {
						if input.Item == step.input {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("%s → %s の連鎖がない", step.input, step.output)
				}
			}
		}},
		{"TC-FR-MOD-02-05 複数の品質ボーナス", func(t *testing.T, c defs.Catalog) {
			count := 0
			for _, recipe := range c.Recipes {
				for _, bonus := range recipe.Bonuses {
					if bonus.MinCount > 0 && bonus.Add > 0 {
						count++
					}
				}
			}
			if count < 3 {
				t.Errorf("正の品質ボーナスは %d 個", count)
			}
			for _, id := range []string{"cheese", "potion_herb"} {
				recipe := c.Recipes[id]
				if len(recipe.Bonuses) == 0 {
					t.Errorf("%s にボーナスがない", id)
					continue
				}
				bonus := recipe.Bonuses[0]
				for _, choice := range []struct {
					item string
					want bool
				}{{"herb", true}, {"glow_moss", false}} {
					amount := 0
					choiceUsed := false
					for _, slot := range recipe.Inputs {
						itemID := slot.Item
						if slot.Attribute != "" {
							itemID = choice.item
							choiceUsed = true
						}
						item := c.Items[itemID]
						if slot.Attribute != "" {
							matches := false
							for _, attr := range item.Attributes {
								matches = matches || attr == slot.Attribute
							}
							if !matches {
								t.Errorf("%s は %s の入力枠に合わない", itemID, slot.Attribute)
							}
						}
						for _, attr := range item.Attributes {
							if attr == bonus.Attribute {
								amount += slot.Count
							}
						}
					}
					if !choiceUsed || (amount >= bonus.MinCount) != choice.want {
						t.Errorf("%s に %s を選んだときのボーナス条件: %d/%d", id, choice.item, amount, bonus.MinCount)
					}
				}
			}
		}},
		{"TC-FR-MAT-02-04 店で買って売るだけでは儲からない", func(t *testing.T, c defs.Catalog) {
			// 売値の式（DES-ECO-01）に、すべての施設の sell_rate を最大段階まで足した値で比べる。
			rate := 0
			for _, f := range c.Facilities {
				for _, level := range f.Levels {
					for _, e := range level.Effects {
						if e.Kind == "sell_rate" {
							rate += e.Value
						}
					}
				}
			}
			for _, offer := range c.Game.Shop {
				item, _ := c.Item(offer.Item)
				sell := item.BasePrice * (50 + offer.Quality) / 100 * (100 + rate) / 100
				if sell >= offer.Price {
					t.Errorf("%s: 買値 %d、最大の売値 %d", offer.Item, offer.Price, sell)
				}
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := defload.Load(data.Base)
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, catalog)
		})
	}
}
