package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func golemGatherCatalog() defs.Catalog {
	c := golemCatalog()
	c.Golems["worker"] = defs.GolemDef{ID: "worker", Materials: []defs.ItemAmount{{Item: "potion", Count: 2}}, SpeedPercent: 100}
	c.Gathers = map[string]defs.GatherDef{"meadow": {ID: "meadow", Rolls: 3, CooldownSeconds: 12, Drops: []defs.Drop{
		{Item: "herb", Weight: 3, QualityMin: 30, QualityMax: 50},
		{Item: "water", Weight: 1, QualityMin: 40, QualityMax: 60},
	}}}
	return c
}

func assignMeadow(t *testing.T, w *World, index int) {
	t.Helper()
	if err := w.AssignGolem(index, Assignment{GatherID: "meadow"}); err != nil {
		t.Fatal(err)
	}
}

func runTicks(w *World, count int) {
	for range count {
		w.Tick()
	}
}

func inventoryTotal(w *World) int {
	total := 0
	for _, stack := range w.Inventory.Stacks() {
		total += stack.Count
	}
	return total
}

func TestGolemGatherAssignment(t *testing.T) {
	t.Run("TC-FR-GOL-04-01 採取割り当て", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		g := w.Golems[0]
		if g.State != GolemWorking || g.Progress != 0 || len(g.Holding) != 0 || g.Assignment.GatherID != "meadow" {
			t.Fatalf("割り当て後: %+v", g)
		}
	})
	t.Run("TC-FR-GOL-04-02 検査順と状態不変", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		for _, tc := range []struct {
			index int
			a     Assignment
			want  error
		}{
			{1, Assignment{GatherID: "meadow", RecipeID: "potion"}, ErrUnknownID},
			{0, Assignment{GatherID: "missing", RecipeID: "potion"}, ErrInvalidSelection},
			{0, Assignment{GatherID: "missing", SlotItems: []string{"herb"}}, ErrInvalidSelection},
			{0, Assignment{GatherID: "missing"}, ErrUnknownID},
		} {
			before := w.Golems[0]
			if err := w.AssignGolem(tc.index, tc.a); !errors.Is(err, tc.want) {
				t.Fatalf("%+v: error=%v want=%v", tc.a, err, tc.want)
			}
			if !reflect.DeepEqual(w.Golems[0], before) {
				t.Fatal("失敗後にゴーレムが変化")
			}
		}
	})
	t.Run("TC-FR-GOL-04-03 レシピ作業から割り当て直す", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("herb", 40, 2), craftStack("water", 40, 1))
		assignPotion(t, w, 0)
		w.Tick()
		w.Tick()
		before := w.Inventory.Stacks()
		assignMeadow(t, w, 0)
		g := w.Golems[0]
		if g.State != GolemWorking || g.Progress != 0 || len(g.Holding) != 0 || !reflect.DeepEqual(before, w.Inventory.Stacks()) {
			t.Fatalf("割り当て直し: %+v, %v", g, w.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-GOL-04-04 採取途中で解除", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		runTicks(w, 719)
		if err := w.UnassignGolem(0); err != nil {
			t.Fatal(err)
		}
		w.Tick()
		if w.Golems[0].State != GolemUnassigned || w.Golems[0].Assignment != nil || inventoryTotal(w) != 0 {
			t.Fatalf("解除後: %+v %v", w.Golems[0], w.Inventory.Stacks())
		}
	})
}

func TestGolemGatherRules(t *testing.T) {
	t.Run("TC-FR-GOL-04-05 プレイヤーと同じ抽選", func(t *testing.T) {
		c := golemGatherCatalog()
		player, golem := NewWorld(c, 42), NewWorld(c, 42)
		buildWorker(t, golem)
		assignMeadow(t, golem, 0)
		if _, err := player.Gather("meadow"); err != nil {
			t.Fatal(err)
		}
		runTicks(golem, 720)
		if !reflect.DeepEqual(player.Inventory.Stacks(), golem.Inventory.Stacks()) || player.rng.Uint64() != golem.rng.Uint64() {
			t.Fatalf("採取結果が異なる: %v %v", player.Inventory.Stacks(), golem.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-GOL-04-06 プレイヤーの再使用待ちに干渉しない", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		if _, err := w.Gather("meadow"); err != nil {
			t.Fatal(err)
		}
		ready := w.gatherReady["meadow"]
		assignMeadow(t, w, 0)
		runTicks(w, 720)
		if inventoryTotal(w) != 6 || w.gatherReady["meadow"] != ready {
			t.Fatalf("採取または再使用待ち: %v %v", w.Inventory.Stacks(), w.gatherReady)
		}
		if _, err := w.Gather("meadow"); err != nil {
			t.Fatalf("プレイヤーが採取できない: %v", err)
		}
	})
	t.Run("TC-FR-GOL-04-07 在庫あふれで次 tick に再試行 / TC-NFR-06-06 乱数を戻す", func(t *testing.T) {
		c := golemGatherCatalog()
		c.Gathers["meadow"] = defs.GatherDef{ID: "meadow", Rolls: 1, CooldownSeconds: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}
		w, reference := NewWorld(c, 7), NewWorld(c, 7)
		buildWorker(t, w)
		buildWorker(t, reference)
		assignMeadow(t, w, 0)
		assignMeadow(t, reference, 0)
		max := int(^uint(0) >> 1)
		addGolemStock(t, w, Stack{Key: ItemKey{"herb", 40}, Count: max})
		runTicks(w, 60)
		if w.Inventory.Count(ItemKey{"herb", 40}) != max || w.Golems[0].Progress != 60 || w.Golems[0].State != GolemWorking {
			t.Fatalf("失敗後: %+v %v", w.Golems[0], w.Inventory.Stacks())
		}
		if err := w.Inventory.Remove(ItemKey{"herb", 40}, max); err != nil {
			t.Fatal(err)
		}
		runTicks(reference, 60)
		w.Tick()
		if !reflect.DeepEqual(w.Inventory.Stacks(), reference.Inventory.Stacks()) || w.rng.Uint64() != reference.rng.Uint64() || w.Golems[0].Progress != 0 {
			t.Fatalf("再試行後: %+v %v / %v", w.Golems[0], w.Inventory.Stacks(), reference.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-GOL-04-08 同じ採取地に二体", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		assignMeadow(t, w, 1)
		runTicks(w, 720)
		if inventoryTotal(w) != 6 || len(w.gatherReady) != 0 {
			t.Fatalf("二体の採取: %v %v", w.Inventory.Stacks(), w.gatherReady)
		}
	})
}

func TestGolemGatherTiming(t *testing.T) {
	t.Run("TC-FR-GOL-06-04 719・720・1440 tick", func(t *testing.T) {
		w := NewWorld(golemGatherCatalog(), 1)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		runTicks(w, 719)
		if inventoryTotal(w) != 0 || w.Golems[0].Progress != 719 {
			t.Fatal("719 tick で採取した")
		}
		w.Tick()
		if inventoryTotal(w) != 3 || w.Golems[0].Progress != 0 {
			t.Fatal("720 tick で採取しない")
		}
		runTicks(w, 720)
		if inventoryTotal(w) != 6 || w.Golems[0].Progress != 0 {
			t.Fatal("1440 tick で二回目の採取をしない")
		}
	})
	t.Run("TC-FR-GOL-06-05 速度 150", func(t *testing.T) {
		c := golemGatherCatalog()
		worker := c.Golems["worker"]
		worker.SpeedPercent = 150
		c.Golems["worker"] = worker
		w := NewWorld(c, 1)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		runTicks(w, 479)
		if inventoryTotal(w) != 0 {
			t.Fatal("479 tick で採取した")
		}
		w.Tick()
		if inventoryTotal(w) != 3 {
			t.Fatal("480 tick で採取しない")
		}
	})
	t.Run("TC-FR-GOL-06-06 最小一 tick", func(t *testing.T) {
		c := golemGatherCatalog()
		site := c.Gathers["meadow"]
		site.CooldownSeconds = 0
		c.Gathers["meadow"] = site
		w := NewWorld(c, 1)
		buildWorker(t, w)
		assignMeadow(t, w, 0)
		w.Tick()
		if inventoryTotal(w) != 3 {
			t.Fatal("一 tick で採取しない")
		}
	})
	t.Run("TC-FR-GOL-06-07 後ろのレシピが同 tick に受け取る", func(t *testing.T) {
		c := golemGatherCatalog()
		c.Gathers["meadow"] = defs.GatherDef{ID: "meadow", Rolls: 2, CooldownSeconds: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}
		w := NewWorld(c, 1)
		buildWorker(t, w)
		buildWorker(t, w)
		addGolemStock(t, w, craftStack("water", 31, 1))
		assignMeadow(t, w, 0)
		assignPotion(t, w, 1)
		runTicks(w, 59)
		if w.Golems[1].State != GolemWaiting {
			t.Fatal("採取前に作業開始した")
		}
		w.Tick()
		if w.Golems[1].State != GolemWorking || w.Inventory.Count(ItemKey{"herb", 40}) != 0 {
			t.Fatalf("連鎖しない: %+v %v", w.Golems[1], w.Inventory.Stacks())
		}
	})
}

func TestGolemGatherDeterminism(t *testing.T) {
	t.Run("TC-NFR-06-05 同じシードと入力", func(t *testing.T) {
		c := golemGatherCatalog()
		a, b := NewWorld(c, 99), NewWorld(c, 99)
		for _, w := range []*World{a, b} {
			buildWorker(t, w)
			assignMeadow(t, w, 0)
			runTicks(w, 1440)
		}
		if !reflect.DeepEqual(a.Inventory.Stacks(), b.Inventory.Stacks()) || !reflect.DeepEqual(a.Golems, b.Golems) || a.rng.Uint64() != b.rng.Uint64() {
			t.Fatal("結果が一致しない")
		}
	})
}
