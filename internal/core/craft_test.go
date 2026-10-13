package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func TestCraftSlotNumber(t *testing.T) {
	herb := craftStack("herb", 40, 2)
	water := craftStack("water", 31, 1)
	for _, tt := range []struct {
		name      string
		selection Selection
		reason    error
		wantSlot  string
	}{
		{"TC-FR-ALC-02-05 一番目の枠は個数不足", craftSelection(nil, []Stack{water}), ErrInvalidCount, "入力枠 1"},
		{"TC-FR-ALC-02-05 二番目の枠は個数不足", craftSelection([]Stack{herb}, nil), ErrInvalidCount, "入力枠 2"},
		{"TC-FR-ALC-02-05 二番目の枠は品質不正", craftSelection([]Stack{herb}, []Stack{craftStack("water", 0, 1)}), ErrInvalidQuality, "入力枠 2"},
		{"TC-FR-ALC-02-05 二番目の枠は種類不適合", craftSelection([]Stack{herb}, []Stack{craftStack("herb", 40, 1)}), ErrInvalidSelection, "入力枠 2"},
		{"TC-FR-ALC-02-05 二番目の枠は個数不正", craftSelection([]Stack{herb}, []Stack{craftStack("water", 31, 0)}), ErrInvalidCount, "入力枠 2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(craftCatalog(), 1)
			w.Inventory.Add(herb.Key, herb.Count)
			w.Inventory.Add(water.Key, water.Count)
			before := w.Inventory.Stacks()
			_, err := w.Craft("potion", tt.selection)
			if !errors.Is(err, tt.reason) || !strings.Contains(err.Error(), tt.wantSlot) {
				t.Fatalf("error = %v", err)
			}
			if !reflect.DeepEqual(before, w.Inventory.Stacks()) {
				t.Fatal("失敗後に在庫が変化")
			}
		})
	}
}

func craftCatalog() defs.Catalog {
	return defs.Catalog{
		Items: map[string]defs.ItemDef{
			"herb":   {ID: "herb", Attributes: []string{"草", "癒し"}},
			"weed":   {ID: "weed", Attributes: []string{"草"}},
			"water":  {ID: "water", Attributes: []string{"水"}},
			"potion": {ID: "potion"}, "strong": {ID: "strong"},
		},
		Recipes: map[string]defs.RecipeDef{
			"potion":      {ID: "potion", Output: "potion", OutputCount: 1, Inputs: []defs.InputSlot{{Attribute: "草", Count: 2}, {Item: "water", Count: 1}}, Bonuses: []defs.Bonus{{Attribute: "癒し", MinCount: 2, Add: 15}}},
			"potion_alt":  {ID: "potion_alt", Output: "potion", OutputCount: 2, Inputs: []defs.InputSlot{{Item: "herb", Count: 1}}},
			"double_herb": {ID: "double_herb", Output: "potion", OutputCount: 1, Inputs: []defs.InputSlot{{Item: "herb", Count: 1}, {Attribute: "草", Count: 1}}},
			"strong":      {ID: "strong", Output: "strong", OutputCount: 1, Inputs: []defs.InputSlot{{Item: "potion", Count: 1}}},
		},
	}
}

func craftStack(id string, q, n int) Stack {
	return Stack{Key: ItemKey{ItemID: id, Quality: q}, Count: n}
}
func craftSelection(slots ...[]Stack) Selection { return Selection{Slots: slots} }

func TestCraftCases(t *testing.T) {
	herb := craftStack("herb", 40, 2)
	water := craftStack("water", 31, 1)
	weed := craftStack("weed", 60, 2)
	standard := craftSelection([]Stack{herb}, []Stack{water})
	tests := []struct {
		name, recipe string
		stock        []Stack
		selection    Selection
		want         Stack
		wantErr      error
		preview      bool
		wantAverage  int
		wantBonuses  int
	}{
		{"TC-FR-ALC-01-01 予告と確定", "potion", []Stack{herb, water}, standard, craftStack("potion", 52, 1), nil, true, 37, 1},
		{"TC-FR-ALC-01-02 必要数ちょうど", "potion", []Stack{herb, water}, standard, craftStack("potion", 52, 1), nil, false, 0, 0},
		{"TC-FR-ALC-01-03 素材不足", "potion", []Stack{craftStack("herb", 40, 1), water}, standard, Stack{}, ErrNotEnoughItems, false, 0, 0},
		{"TC-FR-ALC-01-04 出力先の桁あふれ", "potion", []Stack{herb, water, craftStack("potion", 52, int(^uint(0)>>1))}, standard, Stack{}, ErrInvalidCount, false, 0, 0},
		{"TC-FR-ALC-02-01 枠別の入力", "potion", []Stack{herb, water}, standard, craftStack("potion", 52, 1), nil, false, 0, 0},
		{"TC-FR-ALC-02-02 同じ在庫を二枠で使用", "double_herb", []Stack{herb}, craftSelection([]Stack{craftStack("herb", 40, 1)}, []Stack{craftStack("herb", 40, 1)}), craftStack("potion", 40, 1), nil, false, 0, 0},
		{"TC-FR-ALC-02-03 枠数不一致", "potion", []Stack{herb, water}, craftSelection([]Stack{herb}), Stack{}, ErrInvalidCount, false, 0, 0},
		{"TC-FR-ALC-03-02 レシピ別の出力個数", "potion_alt", []Stack{craftStack("herb", 40, 1)}, craftSelection([]Stack{craftStack("herb", 40, 1)}), craftStack("potion", 40, 2), nil, false, 0, 0},
		{"TC-FR-ALC-03-03 不明なレシピ", "missing", []Stack{herb, water}, standard, Stack{}, ErrUnknownID, false, 0, 0},
		{"TC-FR-ALC-04-01 計算例の薬草", "potion", []Stack{herb, water}, standard, craftStack("potion", 52, 1), nil, true, 37, 1},
		{"TC-FR-ALC-04-02 計算例の雑草", "potion", []Stack{weed, water}, craftSelection([]Stack{weed}, []Stack{water}), craftStack("potion", 50, 1), nil, true, 50, 0},
		{"TC-FR-ALC-04-03 属性が合わない", "potion", []Stack{herb, water}, craftSelection([]Stack{craftStack("water", 31, 2)}, []Stack{water}), Stack{}, ErrInvalidSelection, false, 0, 0},
		{"TC-FR-ALC-04-04 品質の上限", "potion", []Stack{craftStack("herb", 195, 2), craftStack("water", 195, 1)}, craftSelection([]Stack{craftStack("herb", 195, 2)}, []Stack{craftStack("water", 195, 1)}), craftStack("potion", 200, 1), nil, true, 195, 1},
		{"TC-FR-ALC-06-02 製品を全数消費", "strong", []Stack{craftStack("potion", 52, 1)}, craftSelection([]Stack{craftStack("potion", 52, 1)}), craftStack("strong", 52, 1), nil, false, 0, 0},
		{"TC-FR-ALC-06-03 製品が不足", "strong", nil, craftSelection([]Stack{craftStack("potion", 52, 1)}), Stack{}, ErrNotEnoughItems, false, 0, 0},
		{"TC-FR-ITM-05-01 同じ製品と品質差", "potion", []Stack{herb, water}, standard, craftStack("potion", 52, 1), nil, true, 37, 1},
		{"TC-FR-ITM-05-02 ボーナスなしは平均品質", "potion", []Stack{weed, water}, craftSelection([]Stack{weed}, []Stack{water}), craftStack("potion", 50, 1), nil, true, 50, 0},
		{"TC-FR-ITM-05-03 失敗時に出力なし", "potion", []Stack{herb, water}, craftSelection([]Stack{craftStack("herb", 40, 3)}, []Stack{water}), Stack{}, ErrInvalidCount, false, 0, 0},
		{"TC-FR-ALC-02-04 同じ在庫の合計不足", "double_herb", []Stack{craftStack("herb", 40, 1)}, craftSelection([]Stack{craftStack("herb", 40, 1)}, []Stack{craftStack("herb", 40, 1)}), Stack{}, ErrNotEnoughItems, false, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(craftCatalog(), 1)
			for _, s := range tt.stock {
				if err := w.Inventory.Add(s.Key, s.Count); err != nil {
					t.Fatal(err)
				}
			}
			before := w.Inventory.Stacks()
			if tt.preview {
				p, err := w.PreviewCraft(tt.recipe, tt.selection)
				if err != nil {
					t.Fatal(err)
				}
				if p.Quality != tt.want.Key.Quality || p.AverageQuality != tt.wantAverage || len(p.Bonuses) != tt.wantBonuses {
					t.Fatalf("preview = %+v", p)
				}
				if !reflect.DeepEqual(w.Inventory.Stacks(), before) {
					t.Fatal("PreviewCraft が在庫を変更した")
				}
			}
			got, err := w.Craft(tt.recipe, tt.selection)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Craft error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if !reflect.DeepEqual(w.Inventory.Stacks(), before) {
					t.Fatalf("失敗後に在庫が変化: %v", w.Inventory.Stacks())
				}
				return
			}
			if got != tt.want {
				t.Fatalf("Craft = %+v, want %+v", got, tt.want)
			}
			for _, s := range tt.stock {
				used := 0
				for _, slot := range tt.selection.Slots {
					for _, picked := range slot {
						if picked.Key == s.Key {
							used += picked.Count
						}
					}
				}
				if n := w.Inventory.Count(s.Key); n != s.Count-used {
					t.Fatalf("素材 %v の残数 = %d", s.Key, n)
				}
			}
			if n := w.Inventory.Count(got.Key); n != got.Count {
				t.Fatalf("製品数 = %d", n)
			}
		})
	}
}

func TestCraftChain(t *testing.T) {
	t.Run("TC-FR-ALC-06-01 製品を次のレシピへ渡す", func(t *testing.T) {
		w := NewWorld(craftCatalog(), 1)
		for _, s := range []Stack{craftStack("herb", 40, 2), craftStack("water", 31, 1)} {
			if err := w.Inventory.Add(s.Key, s.Count); err != nil {
				t.Fatal(err)
			}
		}
		first, err := w.Craft("potion", craftSelection([]Stack{craftStack("herb", 40, 2)}, []Stack{craftStack("water", 31, 1)}))
		if err != nil {
			t.Fatal(err)
		}
		second, err := w.Craft("strong", craftSelection([]Stack{first}))
		if err != nil {
			t.Fatal(err)
		}
		if second.Key != (ItemKey{ItemID: "strong", Quality: 52}) || w.Inventory.Count(first.Key) != 0 || w.Inventory.Count(second.Key) != 1 {
			t.Fatalf("連鎖した結果: first=%+v second=%+v inventory=%v", first, second, w.Inventory.Stacks())
		}
	})
}

func TestCraftAlternativeRecipes(t *testing.T) {
	t.Run("TC-FR-ALC-03-01 同じ製品の別レシピ", func(t *testing.T) {
		w := NewWorld(craftCatalog(), 1)
		for _, s := range []Stack{craftStack("weed", 60, 2), craftStack("water", 31, 1), craftStack("herb", 50, 1)} {
			if err := w.Inventory.Add(s.Key, s.Count); err != nil {
				t.Fatal(err)
			}
		}
		a, err := w.Craft("potion", craftSelection([]Stack{craftStack("weed", 60, 2)}, []Stack{craftStack("water", 31, 1)}))
		if err != nil {
			t.Fatal(err)
		}
		b, err := w.Craft("potion_alt", craftSelection([]Stack{craftStack("herb", 50, 1)}))
		if err != nil {
			t.Fatal(err)
		}
		if a.Key != b.Key || w.Inventory.Count(a.Key) != 3 {
			t.Fatalf("別レシピの出力が統合されない: %v, %v, %v", a, b, w.Inventory.Stacks())
		}
	})
}

func TestCraftMultipleBonuses(t *testing.T) {
	t.Run("TC-FR-ALC-04-05 ボーナスを合計する", func(t *testing.T) {
		catalog := craftCatalog()
		recipe := catalog.Recipes["potion"]
		recipe.Bonuses = append(recipe.Bonuses, defs.Bonus{Attribute: "水", MinCount: 1, Add: 5})
		catalog.Recipes["potion"] = recipe
		w := NewWorld(catalog, 1)
		herb := craftStack("herb", 40, 2)
		water := craftStack("water", 31, 1)
		for _, s := range []Stack{herb, water} {
			if err := w.Inventory.Add(s.Key, s.Count); err != nil {
				t.Fatal(err)
			}
		}
		got, err := w.PreviewCraft("potion", craftSelection([]Stack{herb}, []Stack{water}))
		if err != nil {
			t.Fatal(err)
		}
		if got.AverageQuality != 37 || got.Quality != 57 || !reflect.DeepEqual(got.Bonuses, recipe.Bonuses) {
			t.Fatalf("PreviewCraft = %+v", got)
		}
	})
}
