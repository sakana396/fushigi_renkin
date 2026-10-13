package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSlotCandidates(t *testing.T) {
	stock := []Stack{
		craftStack("weed", 60, 2), craftStack("herb", 80, 2),
		craftStack("water", 30, 1), craftStack("herb", 40, 2),
		craftStack("water", 10, 1), craftStack("potion", 50, 1),
	}
	tests := []struct {
		name   string
		recipe string
		index  int
		stock  []Stack
		want   []Stack
		err    error
		slot   string
	}{
		{"TC-FR-ALC-08-01 属性枠と候補順", "potion", 0, stock, []Stack{craftStack("herb", 40, 2), craftStack("herb", 80, 2), craftStack("weed", 60, 2)}, nil, ""},
		{"TC-FR-ALC-08-02 アイテム枠と品質順", "potion", 1, stock, []Stack{craftStack("water", 10, 1), craftStack("water", 30, 1)}, nil, ""},
		{"TC-FR-ALC-08-03 候補なし", "potion", 0, []Stack{craftStack("water", 30, 1)}, nil, nil, ""},
		{"TC-FR-ALC-08-04 負の枠番号", "potion", -1, stock, nil, ErrUnknownID, "入力枠"},
		{"TC-FR-ALC-08-04 上限の枠番号", "potion", 2, stock, nil, ErrUnknownID, "入力枠 3"},
		{"TC-FR-ALC-08-05 未知のレシピ", "missing", 0, stock, nil, ErrUnknownID, ""},
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
			got, err := w.SlotCandidates(tt.recipe, tt.index)
			if !errors.Is(err, tt.err) || (tt.slot != "" && !strings.Contains(err.Error(), tt.slot)) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("candidates = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(w.Inventory.Stacks(), before) {
				t.Fatal("候補照会で在庫が変化")
			}
		})
	}
}

func TestSlotCandidatesCraftCompatibility(t *testing.T) {
	t.Run("TC-FR-ALC-08-06 候補は必ず錬金できる", func(t *testing.T) {
		for _, candidate := range []Stack{craftStack("herb", 40, 2), craftStack("weed", 60, 2)} {
			w := NewWorld(craftCatalog(), 1)
			for _, s := range []Stack{candidate, craftStack("water", 30, 1)} {
				if err := w.Inventory.Add(s.Key, s.Count); err != nil {
					t.Fatal(err)
				}
			}
			for i, selected := range []Stack{candidate, craftStack("water", 30, 1)} {
				got, err := w.SlotCandidates("potion", i)
				if err != nil || !reflect.DeepEqual(got, []Stack{selected}) {
					t.Fatalf("枠 %d の候補 = %v, err = %v", i, got, err)
				}
			}
			if _, err := w.Craft("potion", craftSelection([]Stack{candidate}, []Stack{craftStack("water", 30, 1)})); err != nil {
				t.Fatalf("候補 %v が Craft に通らない: %v", candidate, err)
			}
		}
	})
}

func TestSlotCandidatesDoNotReserve(t *testing.T) {
	t.Run("TC-FR-ALC-08-07 他の枠の選択数を差し引かない", func(t *testing.T) {
		w := NewWorld(craftCatalog(), 1)
		stack := craftStack("herb", 40, 1)
		if err := w.Inventory.Add(stack.Key, stack.Count); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			got, err := w.SlotCandidates("double_herb", i)
			if err != nil || !reflect.DeepEqual(got, []Stack{stack}) {
				t.Fatalf("枠 %d の候補 = %v, err = %v", i, got, err)
			}
		}
	})
}
