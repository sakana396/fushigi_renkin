package core

import (
	"errors"
	"reflect"

	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func economyCatalog() defs.Catalog {
	return defs.Catalog{
		Game:  defs.GameDef{StartMoney: 100, BaseGolemSlots: 1, Shop: []defs.ShopOffer{{Item: "herb", Quality: 40, Price: 12}}},
		Items: map[string]defs.ItemDef{"herb": {ID: "herb", BasePrice: 10}, "potion": {ID: "potion", BasePrice: 101}},
		Gathers: map[string]defs.GatherDef{
			"field": {ID: "field", Rolls: 3, CooldownSeconds: 1, RespawnSeconds: 1, Drops: []defs.Drop{{Item: "herb", Weight: 3, QualityMin: 1, QualityMax: 2}, {Item: "potion", Weight: 1, QualityMin: 50, QualityMax: 50}}},
		},
		Facilities: map[string]defs.FacilityDef{
			"lab": {ID: "lab", Levels: []defs.FacilityLevel{
				{Money: 20, Materials: []defs.ItemAmount{{Item: "herb", Count: 2}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 2}, {Kind: "sell_rate", Value: 10}}},
				{Money: 30, Materials: []defs.ItemAmount{{Item: "herb", Count: 1}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 3}, {Kind: "sell_rate", Value: 5}}},
			}},
		},
	}
}

func TestGather(t *testing.T) {
	t.Run("TC-FR-MAT-01-01 同じシードで重み付き抽選", func(t *testing.T) {
		a, b := NewWorld(economyCatalog(), 42), NewWorld(economyCatalog(), 42)
		x, e1 := pickAtTestSite(a, "field")
		y, e2 := pickAtTestSite(b, "field")
		if e1 != nil || e2 != nil || !reflect.DeepEqual(x, []Stack{{ItemKey{"herb", 2}, 1}, {ItemKey{"potion", 50}, 2}}) || !reflect.DeepEqual(x, y) || !reflect.DeepEqual(a.Inventory.Stacks(), b.Inventory.Stacks()) {
			t.Fatalf("採取結果: %v %v, %v %v", x, e1, y, e2)
		}
		total := 0
		for _, s := range x {
			total += s.Count
		}
		if total != 3 {
			t.Fatalf("抽選数 = %d", total)
		}
		c := economyCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Rolls: 1000, Drops: c.Gathers["field"].Drops}
		weighted, err := pickAtTestSite(NewWorld(c, 42), "field")
		if err != nil {
			t.Fatal(err)
		}
		want := []Stack{{ItemKey{"herb", 1}, 353}, {ItemKey{"herb", 2}, 397}, {ItemKey{"potion", 50}, 250}}
		if !reflect.DeepEqual(weighted, want) {
			t.Fatalf("重み付き抽選=%v want=%v", weighted, want)
		}
	})
	t.Run("TC-FR-MAT-01-02 品質の両端", func(t *testing.T) {
		c := economyCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Rolls: 100, CooldownSeconds: 0, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 1, QualityMax: 2}}}
		w := NewWorld(c, 1)
		got, err := pickAtTestSite(w, "field")
		if err != nil || !reflect.DeepEqual(got, []Stack{{ItemKey{"herb", 1}, 54}, {ItemKey{"herb", 2}, 46}}) {
			t.Fatalf("結果 = %v, %v", got, err)
		}
	})
	t.Run("TC-FR-MAT-01-03 再使用待ち", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		if _, err := pickAtTestSite(w, "field"); err != nil {
			t.Fatal(err)
		}
		before := w.Inventory.Stacks()
		if _, err := pickAtTestSite(w, "field"); !errors.Is(err, ErrPickNotReady) {
			t.Fatalf("error = %v", err)
		}
		if !reflect.DeepEqual(before, w.Inventory.Stacks()) {
			t.Fatal("在庫が変化")
		}
	})
	t.Run("TC-FR-MAT-01-04 未知の採取地", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		if _, err := pickAtTestSite(w, "unknown"); !errors.Is(err, ErrUnknownID) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("TC-FR-MAT-01-05 待ち終了 tick", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		pickAtTestSite(w, "field")
		for i := 0; i < 59; i++ {
			w.Tick()
		}
		if _, err := pickAtTestSite(w, "field"); !errors.Is(err, ErrPickNotReady) {
			t.Fatalf("error = %v", err)
		}
		w.Tick()
		if _, err := pickAtTestSite(w, "field"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("TC-FR-MAT-01-06 在庫あふれは乱数も保つ", func(t *testing.T) {
		c := economyCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Rolls: 1, CooldownSeconds: 1, RespawnSeconds: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 1, QualityMax: 1}}}
		w, other := NewWorld(c, 7), NewWorld(c, 7)
		max := int(^uint(0) >> 1)
		w.Inventory.Add(ItemKey{"herb", 1}, max)
		if _, err := pickAtTestSite(w, "field"); !errors.Is(err, ErrInvalidCount) {
			t.Fatalf("error = %v", err)
		}
		if w.Inventory.Count(ItemKey{"herb", 1}) != max || w.rng.Uint64() != other.rng.Uint64() {
			t.Fatal("失敗後に状態が変化")
		}
		if _, used := w.pickReady[PickKey{"field", 1, 1}]; used {
			t.Fatal("失敗後に再使用待ちが設定された")
		}
	})
}

func TestGatherCooldownRemaining(t *testing.T) {
	for _, tt := range []struct {
		name    string
		seconds int
		ticks   int
		want    string
	}{
		{"TC-FR-MAT-01-07 開始直後は二秒", 2, 0, "あと 2 秒で拾えます: まだ拾えません"},
		{"TC-FR-MAT-01-07 一 tick 後も切り上げ", 2, 1, "あと 2 秒で拾えます: まだ拾えません"},
		{"TC-FR-MAT-01-07 一秒経過後は一秒", 2, 60, "あと 1 秒で拾えます: まだ拾えません"},
		{"TC-FR-MAT-01-07 終了直前も一秒", 2, 119, "あと 1 秒で拾えます: まだ拾えません"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			catalog := economyCatalog()
			site := catalog.Gathers["field"]
			site.RespawnSeconds = tt.seconds
			catalog.Gathers["field"] = site
			w := NewWorld(catalog, 1)
			if _, err := pickAtTestSite(w, "field"); err != nil {
				t.Fatal(err)
			}
			before := w.Inventory.Stacks()
			for i := 0; i < tt.ticks; i++ {
				w.Tick()
			}
			_, err := pickAtTestSite(w, "field")
			if !errors.Is(err, ErrPickNotReady) || err.Error() != tt.want {
				t.Fatalf("error = %v", err)
			}
			if !reflect.DeepEqual(before, w.Inventory.Stacks()) {
				t.Fatal("再使用待ちで在庫が変化")
			}
		})
	}
}

func TestBuy(t *testing.T) {
	for _, tt := range []struct {
		name                string
		money, index, count int
		want                error
		balance, items      int
	}{
		{"TC-FR-MAT-02-01 複数購入", 100, 0, 2, nil, 76, 2},
		{"TC-FR-MAT-02-02 代金ちょうど", 12, 0, 1, nil, 0, 1},
		{"TC-FR-MAT-02-03 所持金不足", 11, 0, 1, ErrNotEnoughMoney, 11, 0},
		{"TC-FR-MAT-02-03 無効な商品番号", 100, 1, 1, ErrUnknownID, 100, 0},
		{"TC-FR-MAT-02-03 非正の個数", 100, 0, 0, ErrInvalidCount, 100, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(economyCatalog(), 1)
			w.Money = tt.money
			err := w.Buy(tt.index, tt.count)
			if !errors.Is(err, tt.want) || w.Money != tt.balance || w.Inventory.Count(ItemKey{"herb", 40}) != tt.items {
				t.Fatalf("結果 = %v, %d, %v", err, w.Money, w.Inventory.Stacks())
			}
		})
	}
}

func TestSellPrice(t *testing.T) {
	for _, tt := range []struct {
		name       string
		key        ItemKey
		rate, want int
	}{
		{"TC-FR-ECO-02-01 各段階で切り捨て", ItemKey{"potion", 51}, 15, 117},
		{"TC-FR-ECO-02-02 品質 50", ItemKey{"potion", 50}, 0, 101},
		{"TC-FR-ECO-02-03 未知の種類", ItemKey{"none", 50}, 0, 0},
		{"TC-FR-ECO-02-03 範囲外の品質", ItemKey{"potion", 201}, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(economyCatalog(), 1)
			if tt.rate != 0 {
				w.Facilities = map[string]int{"lab": 2}
			}
			if got := w.SellPrice(tt.key); got != tt.want {
				t.Fatalf("価格 = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSell(t *testing.T) {
	for _, tt := range []struct {
		name                string
		key                 ItemKey
		stock, count        int
		want                error
		money, left, earned int
	}{
		{"TC-FR-ECO-01-01 複数売却", ItemKey{"potion", 50}, 2, 2, nil, 302, 0, 202},
		{"TC-FR-ECO-01-02 全数売却", ItemKey{"potion", 50}, 1, 1, nil, 201, 0, 101},
		{"TC-FR-ECO-01-03 在庫不足", ItemKey{"potion", 50}, 1, 2, ErrNotEnoughItems, 100, 1, 0},
		{"TC-FR-ECO-01-03 非正の個数", ItemKey{"potion", 50}, 1, 0, ErrInvalidCount, 100, 1, 0},
		{"TC-FR-ECO-01-03 未知の種類", ItemKey{"none", 50}, 0, 1, ErrUnknownID, 100, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(economyCatalog(), 1)
			if tt.stock > 0 {
				w.Inventory.Add(tt.key, tt.stock)
			}
			earned, err := w.Sell(tt.key, tt.count)
			if !errors.Is(err, tt.want) || w.Money != tt.money || w.Inventory.Count(tt.key) != tt.left || earned != tt.earned {
				t.Fatalf("結果 = %d, %v, %d, %v", earned, err, w.Money, w.Inventory.Stacks())
			}
		})
	}
}

func TestBuildFacility(t *testing.T) {
	t.Run("TC-FR-TWN-02-01 段階と効果の合計", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		w.Inventory.Add(ItemKey{"herb", 1}, 3)
		for i := 0; i < 2; i++ {
			if err := w.BuildFacility("lab"); err != nil {
				t.Fatal(err)
			}
		}
		if w.Facilities["lab"] != 2 || w.Money != 50 || w.GolemSlots() != 6 || w.SellPrice(ItemKey{"potion", 51}) != 117 {
			t.Fatalf("状態 = %+v", w)
		}
	})
	t.Run("TC-FR-TWN-02-02 低品質から消費", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		w.Inventory.Add(ItemKey{"herb", 1}, 1)
		w.Inventory.Add(ItemKey{"herb", 200}, 2)
		if err := w.BuildFacility("lab"); err != nil {
			t.Fatal(err)
		}
		if w.Inventory.Count(ItemKey{"herb", 1}) != 0 || w.Inventory.Count(ItemKey{"herb", 200}) != 1 {
			t.Fatal(w.Inventory.Stacks())
		}
	})
	for _, tt := range []struct {
		name  string
		setup func(*World)
		id    string
		want  error
	}{
		{"TC-FR-TWN-02-03 素材不足", func(w *World) { w.Inventory.Add(ItemKey{"herb", 1}, 1) }, "lab", ErrNotEnoughItems},
		{"TC-FR-TWN-02-03 お金不足", func(w *World) { w.Money = 19; w.Inventory.Add(ItemKey{"herb", 1}, 2) }, "lab", ErrNotEnoughMoney},
		{"TC-FR-TWN-02-03 最大段階", func(w *World) { w.Facilities = map[string]int{"lab": 2} }, "lab", ErrFacilityMaxLevel},
		{"TC-FR-TWN-02-03 未知の施設", func(w *World) {}, "none", ErrUnknownID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(economyCatalog(), 1)
			tt.setup(w)
			beforeMoney := w.Money
			beforeItems := w.Inventory.Stacks()
			beforeLevel := w.Facilities["lab"]
			err := w.BuildFacility(tt.id)
			if !errors.Is(err, tt.want) || w.Money != beforeMoney || !reflect.DeepEqual(w.Inventory.Stacks(), beforeItems) || w.Facilities["lab"] != beforeLevel {
				t.Fatalf("結果 = %v, %+v", err, w)
			}
		})
	}
}

func TestEconomyMoneyUse(t *testing.T) {
	t.Run("TC-FR-ECO-04-01 購入と建設で消費", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		if err := w.Buy(0, 2); err != nil {
			t.Fatal(err)
		}
		if err := w.BuildFacility("lab"); err != nil {
			t.Fatal(err)
		}
		if w.Money != 56 {
			t.Fatalf("残金 = %d", w.Money)
		}
	})
	t.Run("TC-FR-ECO-04-02 購入後の残金がゼロ", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		w.Money = 12
		if err := w.Buy(0, 1); err != nil || w.Money != 0 {
			t.Fatalf("結果 = %v, %d", err, w.Money)
		}
	})
	t.Run("TC-FR-ECO-04-03 建設費不足は状態を保つ", func(t *testing.T) {
		w := NewWorld(economyCatalog(), 1)
		w.Money = 19
		w.Inventory.Add(ItemKey{"herb", 1}, 2)
		if err := w.BuildFacility("lab"); !errors.Is(err, ErrNotEnoughMoney) || w.Money != 19 || w.Inventory.Count(ItemKey{"herb", 1}) != 2 || w.Facilities["lab"] != 0 {
			t.Fatalf("結果 = %v, %+v", err, w)
		}
	})
}
