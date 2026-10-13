package core

import (
	"errors"
	"reflect"
	"testing"
)

func TestInventory(t *testing.T) {
	herb40 := ItemKey{ItemID: "herb", Quality: 40}
	herb41 := ItemKey{ItemID: "herb", Quality: 41}
	water40 := ItemKey{ItemID: "water", Quality: 40}
	tests := []struct {
		name    string
		add     []Stack
		remove  Stack
		want    []Stack
		wantErr error
	}{
		{"TC-FR-ITM-03-01 同じ種類と品質をまとめる", []Stack{{herb40, 2}, {herb40, 3}}, Stack{}, []Stack{{herb40, 5}}, nil},
		{"TC-FR-ITM-03-02 品質または種類が違えば分ける", []Stack{{herb40, 1}, {herb41, 2}, {water40, 3}}, Stack{}, []Stack{{herb40, 1}, {herb41, 2}, {water40, 3}}, nil},
		{"TC-FR-ITM-03-03 数が足りない削除は状態を保つ", []Stack{{herb40, 2}}, Stack{herb40, 3}, []Stack{{herb40, 2}}, ErrNotEnoughItems},
		{"TC-FR-ITM-03-04 全数を削除すると在庫単位が消える", []Stack{{herb40, 2}}, Stack{herb40, 2}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inv Inventory
			for _, stack := range tt.add {
				if err := inv.Add(stack.Key, stack.Count); err != nil {
					t.Fatal(err)
				}
			}
			if tt.remove.Count != 0 {
				if err := inv.Remove(tt.remove.Key, tt.remove.Count); !errors.Is(err, tt.wantErr) {
					t.Fatalf("Remove error = %v, want %v", err, tt.wantErr)
				}
			}
			if got := inv.Stacks(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Stacks = %v, want %v", got, tt.want)
			}
			for _, stack := range tt.want {
				if got := inv.Count(stack.Key); got != stack.Count {
					t.Errorf("Count(%v) = %d, want %d", stack.Key, got, stack.Count)
				}
			}
		})
	}
}

func TestInventoryInvalid(t *testing.T) {
	tests := []struct {
		name   string
		key    ItemKey
		count  int
		remove bool
		want   error
	}{
		{"TC-FR-ITM-02-04 品質範囲外は追加できない", ItemKey{"herb", 201}, 1, false, ErrInvalidQuality},
		{"TC-FR-ITM-03-05 非正の追加はできない", ItemKey{"herb", 1}, 0, false, ErrInvalidCount},
		{"TC-FR-ITM-03-06 非正の削除はできない", ItemKey{"herb", 1}, -1, true, ErrInvalidCount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inv Inventory
			var err error
			if tt.remove {
				err = inv.Remove(tt.key, tt.count)
			} else {
				err = inv.Add(tt.key, tt.count)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if len(inv.Stacks()) != 0 {
				t.Fatalf("失敗後に在庫が変化した: %v", inv.Stacks())
			}
		})
	}
}

func TestInventoryOverflow(t *testing.T) {
	t.Run("TC-FR-ITM-03-07 個数の桁あふれは状態を保つ", func(t *testing.T) {
		var inv Inventory
		key := ItemKey{"herb", 1}
		maxInt := int(^uint(0) >> 1)
		if err := inv.Add(key, maxInt); err != nil {
			t.Fatal(err)
		}
		if err := inv.Add(key, 1); !errors.Is(err, ErrInvalidCount) {
			t.Fatalf("Add error = %v", err)
		}
		if got := inv.Count(key); got != maxInt {
			t.Fatalf("Count = %d, want %d", got, maxInt)
		}
	})
}

func TestQuality(t *testing.T) {
	tests := []struct {
		name    string
		stacks  []Stack
		want    int
		wantErr error
	}{
		{"TC-FR-ITM-04-01 個数で重み付けして切り捨てる", []Stack{{ItemKey{"herb", 40}, 2}, {ItemKey{"water", 31}, 1}}, 37, nil},
		{"TC-FR-ITM-04-02 上下限の平均", []Stack{{ItemKey{"low", 1}, 1}, {ItemKey{"high", 200}, 1}}, 100, nil},
		{"TC-FR-ITM-04-03 素材なしは誤り", nil, 0, ErrInvalidCount},
		{"TC-FR-ITM-04-04 非正の個数は誤り", []Stack{{ItemKey{"herb", 40}, 0}}, 0, ErrInvalidCount},
		{"TC-FR-ITM-04-05 合計の桁あふれは誤り", []Stack{{ItemKey{"herb", 200}, int(^uint(0) >> 1)}}, 0, ErrInvalidCount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AverageQuality(tt.stacks)
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Fatalf("AverageQuality = (%d, %v), want (%d, %v)", got, err, tt.want, tt.wantErr)
			}
		})
	}
	for _, tt := range []struct {
		name        string
		input, want int
	}{
		{"TC-FR-ITM-02-01 品質範囲内", 100, 100},
		{"TC-FR-ITM-02-02 下限に収める", -5, 1},
		{"TC-FR-ITM-02-03 上限に収める", 220, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampQuality(tt.input); got != tt.want {
				t.Fatalf("ClampQuality = %d, want %d", got, tt.want)
			}
		})
	}
}
