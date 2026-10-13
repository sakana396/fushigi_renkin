package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func golemCatalog() defs.Catalog {
	c := craftCatalog()
	c.Game.BaseGolemSlots = 2
	c.Golems = map[string]defs.GolemDef{"worker": {ID: "worker", Materials: []defs.ItemAmount{{Item: "potion", Count: 2}}, SpeedPercent: 150}, "fast": {ID: "fast", Materials: []defs.ItemAmount{{Item: "potion", Count: 1}}, SpeedPercent: 7000}}
	r := c.Recipes["potion"]
	r.Seconds = 1
	c.Recipes["potion"] = r
	r = c.Recipes["strong"]
	r.Seconds = 1
	c.Recipes["strong"] = r
	return c
}
func addGolemStock(t *testing.T, w *World, stacks ...Stack) {
	t.Helper()
	for _, s := range stacks {
		if err := w.Inventory.Add(s.Key, s.Count); err != nil {
			t.Fatal(err)
		}
	}
}
func buildWorker(t *testing.T, w *World) {
	t.Helper()
	addGolemStock(t, w, craftStack("potion", 10, 2))
	if err := w.BuildGolem("worker"); err != nil {
		t.Fatal(err)
	}
}
func assignPotion(t *testing.T, w *World, index int) {
	t.Helper()
	if err := w.AssignGolem(index, Assignment{RecipeID: "potion", SlotItems: []string{"herb", "water"}}); err != nil {
		t.Fatal(err)
	}
}

func TestBuildGolem(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		slots                         int
		stock                         []Stack
		id                            string
		wantErr                       error
		wantLow, wantHigh, wantGolems int
	}{
		{"TC-FR-GOL-01-01 低品質から材料消費", 2, []Stack{craftStack("potion", 10, 1), craftStack("potion", 90, 2)}, "worker", nil, 0, 1, 1},
		{"TC-FR-GOL-01-03 素材不足", 2, []Stack{craftStack("potion", 10, 1)}, "worker", ErrNotEnoughItems, 1, 0, 0},
		{"TC-FR-GOL-01-03 未知の定義", 2, nil, "missing", ErrUnknownID, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := golemCatalog()
			c.Game.BaseGolemSlots = tc.slots
			w := NewWorld(c, 1)
			addGolemStock(t, w, tc.stock...)
			before := w.Inventory.Stacks()
			err := w.BuildGolem(tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error=%v want=%v", err, tc.wantErr)
			}
			if tc.wantErr != nil && !reflect.DeepEqual(before, w.Inventory.Stacks()) {
				t.Fatal("失敗で在庫が変化")
			}
			if len(w.Golems) != tc.wantGolems || w.Inventory.Count(ItemKey{"potion", 10}) != tc.wantLow || w.Inventory.Count(ItemKey{"potion", 90}) != tc.wantHigh {
				t.Fatalf("state=%+v stock=%v", w.Golems, w.Inventory.Stacks())
			}
			if tc.wantGolems > 0 && w.Golems[0].State != GolemUnassigned {
				t.Fatal("初期状態")
			}
		})
	}
}
func TestGolemSlotsBoundary(t *testing.T) {
	t.Run("TC-FR-GOL-01-02 所有枠の境界", func(t *testing.T) {
		catalog := golemCatalog()
		catalog.Game.BaseGolemSlots = 1
		w := NewWorld(catalog, 1)
		addGolemStock(t, w, craftStack("potion", 10, 4))
		if err := w.BuildGolem("worker"); err != nil {
			t.Fatal(err)
		}
		before := w.Inventory.Stacks()
		if err := w.BuildGolem("worker"); !errors.Is(err, ErrNoGolemSlots) {
			t.Fatalf("error=%v", err)
		}
		if len(w.Golems) != 1 || !reflect.DeepEqual(before, w.Inventory.Stacks()) {
			t.Fatal("枠不足で状態が変化")
		}
	})
}

func TestGolemAssignment(t *testing.T) {
	t.Run("TC-FR-GOL-02-01 割り当てと連続生産", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 4), craftStack("water", 31, 2))
		assignPotion(t, w, 0)
		if w.Golems[0].State != GolemWaiting {
			t.Fatal("素材待ちではない")
		}
		w.Tick()
		if w.Golems[0].State != GolemWorking || w.Golems[0].Progress != 0 {
			t.Fatal("作業開始 tick で進捗した")
		}
		for i := 0; i < 40; i++ {
			w.Tick()
		}
		// 完了した tick のうちに次の入力を取り出し、作業中に移る（DES-GOL-01）。
		if w.Golems[0].State != GolemWorking || w.Golems[0].Progress != 0 || w.Inventory.Count(ItemKey{"potion", 37}) != 1 {
			t.Fatalf("完了: %+v", w.Golems[0])
		}
		if w.Inventory.Count(ItemKey{"herb", 40}) != 0 || w.Inventory.Count(ItemKey{"water", 31}) != 0 {
			t.Fatal("完了と同じ tick で次の入力を取り出していない")
		}
	})
	t.Run("TC-FR-GOL-02-02 作業中の解除", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 2), craftStack("water", 31, 1))
		assignPotion(t, w, 0)
		w.Tick()
		if err := w.UnassignGolem(0); err != nil {
			t.Fatal(err)
		}
		if w.Golems[0].State != GolemUnassigned || w.Golems[0].Assignment != nil || len(w.Golems[0].Holding) != 0 || w.Inventory.Count(ItemKey{"herb", 40}) != 0 {
			t.Fatalf("解除: %+v", w.Golems[0])
		}
	})
	t.Run("TC-FR-GOL-02-02 素材待ちの解除", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		assignPotion(t, w, 0)
		if err := w.UnassignGolem(0); err != nil {
			t.Fatal(err)
		}
		if w.Golems[0].State != GolemUnassigned || w.Golems[0].Assignment != nil {
			t.Fatalf("解除後: %+v", w.Golems[0])
		}
	})
	t.Run("TC-FR-GOL-02-03 不正な割り当て", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		for _, a := range []Assignment{{RecipeID: "missing", SlotItems: []string{"herb", "water"}}, {RecipeID: "potion", SlotItems: []string{"water", "water"}}, {RecipeID: "potion", SlotItems: []string{"herb"}}} {
			before := w.Golems[0]
			if err := w.AssignGolem(0, a); err == nil {
				t.Fatalf("accepted %+v", a)
			}
			if !reflect.DeepEqual(w.Golems[0], before) {
				t.Fatal("失敗で状態が変化")
			}
		}
		if err := w.AssignGolem(1, Assignment{RecipeID: "potion"}); err == nil {
			t.Fatal("不正な番号")
		}
	})
	t.Run("TC-FR-GOL-03-03 属性不適合", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		if err := w.AssignGolem(0, Assignment{RecipeID: "potion", SlotItems: []string{"potion", "water"}}); !errors.Is(err, ErrInvalidSelection) {
			t.Fatalf("error=%v", err)
		}
	})
}
func TestGolemProduction(t *testing.T) {
	t.Run("TC-FR-GOL-03-01 低品質優先と在庫受渡し / TC-FR-ITM-04-06 自動品質", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 1), craftStack("herb", 80, 2), craftStack("water", 31, 1))
		assignPotion(t, w, 0)
		w.Tick()
		if w.Inventory.Count(ItemKey{"herb", 40}) != 0 || w.Inventory.Count(ItemKey{"herb", 80}) != 1 {
			t.Fatal("低品質優先でない")
		}
		for i := 0; i < 40; i++ {
			w.Tick()
		}
		if w.Inventory.Count(ItemKey{"potion", 50}) != 1 || w.Inventory.Count(ItemKey{"potion", 65}) != 0 {
			t.Fatalf("品質=%v", w.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-GOL-03-02 不足から補充", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 2))
		assignPotion(t, w, 0)
		w.Tick()
		if w.Golems[0].State != GolemWaiting || w.Inventory.Count(ItemKey{"herb", 40}) != 2 {
			t.Fatal("不足時に消費")
		}
		addGolemStock(t, w, craftStack("water", 31, 1))
		w.Tick()
		if w.Golems[0].State != GolemWorking {
			t.Fatal("補充後に開始しない")
		}
	})
	t.Run("TC-FR-GOL-06-01 40 tick で完成", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 2), craftStack("water", 31, 1))
		assignPotion(t, w, 0)
		w.Tick()
		for i := 0; i < 39; i++ {
			w.Tick()
		}
		if w.Inventory.Count(ItemKey{"potion", 37}) != 0 {
			t.Fatal("早すぎる完成")
		}
		w.Tick()
		if w.Inventory.Count(ItemKey{"potion", 37}) != 1 {
			t.Fatal("完成しない")
		}
	})
	t.Run("TC-FR-GOL-06-02 最小 1 tick", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		addGolemStock(t, w, craftStack("potion", 10, 1), craftStack("herb", 40, 2), craftStack("water", 31, 1))
		if err := w.BuildGolem("fast"); err != nil {
			t.Fatal(err)
		}
		assignPotion(t, w, 0)
		w.Tick()
		w.Tick()
		if w.Inventory.Count(ItemKey{"potion", 37}) != 1 {
			t.Fatal("1 tick で完成しない")
		}
	})
	t.Run("TC-FR-GOL-06-03 順序どおり連鎖", func(t *testing.T) {
		w := NewWorld(golemCatalog(), 1)
		addGolemStock(t, w, craftStack("potion", 10, 4), craftStack("herb", 40, 2), craftStack("water", 31, 1))
		for i := 0; i < 2; i++ {
			if err := w.BuildGolem("worker"); err != nil {
				t.Fatal(err)
			}
		}
		assignPotion(t, w, 0)
		if err := w.AssignGolem(1, Assignment{RecipeID: "strong", SlotItems: []string{"potion"}}); err != nil {
			t.Fatal(err)
		}
		w.Tick()
		for i := 0; i < 40; i++ {
			w.Tick()
		}
		if w.Golems[1].State != GolemWorking || w.Inventory.Count(ItemKey{"potion", 37}) != 0 {
			t.Fatalf("連鎖: %+v %v", w.Golems[1], w.Inventory.Stacks())
		}
	})
	t.Run("TC-NFR-06-04 同じ入力の決定性", func(t *testing.T) {
		c := golemCatalog()
		a, b := NewWorld(c, 7), NewWorld(c, 7)
		for _, w := range []*World{a, b} {
			buildWorker(t, w)
			addGolemStock(t, w, craftStack("herb", 40, 4), craftStack("water", 31, 2))
			assignPotion(t, w, 0)
			for i := 0; i < 100; i++ {
				w.Tick()
			}
		}
		if a.Now != b.Now || !reflect.DeepEqual(a.Inventory.Stacks(), b.Inventory.Stacks()) || !reflect.DeepEqual(a.Golems, b.Golems) || a.rng.Uint64() != b.rng.Uint64() {
			t.Fatal("結果が一致しない")
		}
	})
}
