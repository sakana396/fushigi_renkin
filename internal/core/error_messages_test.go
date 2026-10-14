package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func TestGolemDuration(t *testing.T) {
	for _, tt := range []struct {
		name    string
		speed   int
		prepare func(*World) error
		index   int
		want    int
		wantErr error
	}{
		{"TC-FR-GOL-04-20 未割り当て", 100, nil, 0, 0, nil},
		{"TC-FR-GOL-04-21 採取地の所要時間", 100, func(w *World) error { return w.AssignGolem(0, Assignment{GatherID: "meadow"}) }, 0, 720, nil},
		{"TC-FR-GOL-04-22 レシピの所要時間", 100, func(w *World) error {
			return w.AssignGolem(0, Assignment{RecipeID: "potion", SlotItems: []string{"herb", "water"}})
		}, 0, 60, nil},
		{"TC-FR-GOL-04-23 範囲外", 100, nil, 1, 0, ErrUnknownID},
		{"TC-FR-GOL-04-24 速度 150 のレシピ", 150, func(w *World) error {
			return w.AssignGolem(0, Assignment{RecipeID: "potion", SlotItems: []string{"herb", "water"}})
		}, 0, 40, nil},
		{"TC-FR-GOL-04-25 速度 150 の採取地", 150, func(w *World) error {
			return w.AssignGolem(0, Assignment{GatherID: "meadow"})
		}, 0, 480, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := golemGatherCatalog()
			c.Gathers["meadow"] = defs.GatherDef{ID: "meadow", Name: "風の草原", CooldownSeconds: 12}
			worker := c.Golems["worker"]
			worker.SpeedPercent = tt.speed
			c.Golems["worker"] = worker
			w := NewWorld(c, 1)
			buildWorker(t, w)
			if tt.prepare != nil {
				if err := tt.prepare(w); err != nil {
					t.Fatal(err)
				}
			}
			before := w.Golems[0]
			got, err := w.GolemDuration(tt.index)
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Fatalf("所要 tick = %d, 誤り = %v", got, err)
			}
			if !reflect.DeepEqual(w.Golems[0], before) {
				t.Fatal("状態が変化")
			}
		})
	}
}

func TestWorldErrorMessages(t *testing.T) {
	c := golemGatherCatalog()
	c.Items["herb"] = defs.ItemDef{ID: "herb", Name: "薬草", Attributes: []string{"草", "癒し"}}
	c.Items["water"] = defs.ItemDef{ID: "water", Name: "水", Attributes: []string{"水"}}
	c.Items["potion"] = defs.ItemDef{ID: "potion", Name: "薬"}
	c.Gathers["meadow"] = defs.GatherDef{ID: "meadow", Name: "風の草原", Rolls: 1, CooldownSeconds: 12, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}
	c.Game.Shop = []defs.ShopOffer{{Item: "herb", Quality: 40, Price: 1}}
	for _, tt := range []struct {
		name   string
		call   func(*World) error
		want   string
		absent string
		kind   error
	}{
		{"TC-IF-02-10 ゴーレム番号", func(w *World) error { return w.AssignGolem(1, Assignment{}) }, "ゴーレム番号 2", "番号 1:", ErrUnknownID},
		{"TC-IF-02-11 商品番号", func(w *World) error { return w.Buy(1, 1) }, "商品番号 2", "番号 1:", ErrUnknownID},
		{"TC-IF-02-12 錬金素材名", func(w *World) error {
			_, err := w.Craft("potion", craftSelection([]Stack{craftStack("water", 40, 2)}, []Stack{craftStack("water", 40, 1)}))
			return err
		}, "水", "\"water\"", ErrInvalidSelection},
		{"TC-IF-02-13 ゴーレム素材名", func(w *World) error {
			return w.AssignGolem(0, Assignment{RecipeID: "potion", SlotItems: []string{"water", "water"}})
		}, "水", "\"water\"", ErrInvalidSelection},
		{"TC-IF-02-14 採取地名", func(w *World) error {
			_, err := w.Gather("meadow")
			if err != nil {
				return err
			}
			_, err = w.Gather("meadow")
			return err
		}, "風の草原", "meadow", ErrGatherCooldown},
		{"TC-IF-02-15 未知の ID", func(w *World) error { _, err := w.Gather("missing"); return err }, "missing", "", ErrUnknownID},
		{"TC-IF-02-16 在庫不足", func(w *World) error {
			w.Inventory = Inventory{}
			if err := w.Inventory.Add(ItemKey{ItemID: "water", Quality: 40}, 1); err != nil {
				return err
			}
			_, err := w.Craft("potion", craftSelection([]Stack{craftStack("herb", 40, 2)}, []Stack{craftStack("water", 40, 1)}))
			return err
		}, "薬草", "herb", ErrNotEnoughItems},
		{"TC-IF-02-19 割り当て解除のゴーレム番号", func(w *World) error { return w.UnassignGolem(1) }, "ゴーレム番号 2", "番号 1:", ErrUnknownID},
		{"TC-IF-02-20 所要時間のゴーレム番号", func(w *World) error { _, err := w.GolemDuration(1); return err }, "ゴーレム番号 2", "番号 1:", ErrUnknownID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(c, 1)
			buildWorker(t, w)
			err := tt.call(w)
			if !errors.Is(err, tt.kind) || !strings.Contains(err.Error(), tt.want) || (tt.absent != "" && strings.Contains(err.Error(), tt.absent)) {
				t.Fatalf("誤り = %v", err)
			}
		})
	}
}

func TestMaterialErrorMessages(t *testing.T) {
	c := golemCatalog()
	c.Items["potion"] = defs.ItemDef{ID: "potion", Name: "薬"}
	w := NewWorld(c, 1)
	buildWorker(t, w)
	t.Run("TC-IF-02-17 ゴーレム材料名", func(t *testing.T) {
		err := w.BuildGolem("worker")
		if !errors.Is(err, ErrNotEnoughItems) || !strings.Contains(err.Error(), "薬") || strings.Contains(err.Error(), "potion") {
			t.Fatalf("誤り = %v", err)
		}
	})
	t.Run("TC-IF-02-18 目録を持たない在庫の誤り文", func(t *testing.T) {
		var inventory Inventory
		err := inventory.Remove(ItemKey{ItemID: "herb", Quality: 40}, 1)
		if !errors.Is(err, ErrNotEnoughItems) || strings.Contains(err.Error(), "herb") {
			t.Fatalf("誤り = %v", err)
		}
	})
	t.Run("TC-IF-02-21 施設材料名", func(t *testing.T) {
		c := economyCatalog()
		herb := c.Items["herb"]
		herb.Name = "薬草"
		c.Items["herb"] = herb
		w := NewWorld(c, 1)
		err := w.BuildFacility("lab")
		if !errors.Is(err, ErrNotEnoughItems) || !strings.Contains(err.Error(), "薬草") || strings.Contains(err.Error(), "herb") {
			t.Fatalf("誤り = %v", err)
		}
	})
}
