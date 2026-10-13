package core

import (
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func TestWorld(t *testing.T) {
	catalog := defs.Catalog{Game: defs.GameDef{StartMoney: 75}}
	t.Run("TC-NFR-06-01 同じシードは同じ乱数列", func(t *testing.T) {
		a, b := NewWorld(catalog, 123), NewWorld(catalog, 123)
		if a.Money != 75 || a.Now != 0 || len(a.Inventory.Stacks()) != 0 {
			t.Fatalf("初期状態が違う: %+v", a)
		}
		for i := 0; i < 10; i++ {
			if x, y := a.rng.Uint64(), b.rng.Uint64(); x != y {
				t.Fatalf("乱数列が違う: %d, %d", x, y)
			}
		}
	})
	t.Run("TC-NFR-06-02 異なるシードは異なる乱数列", func(t *testing.T) {
		a, b := NewWorld(catalog, 123), NewWorld(catalog, 124)
		if a.rng.Uint64() == b.rng.Uint64() {
			t.Fatal("異なるシードで同じ最初の乱数")
		}
	})
	t.Run("TC-NFR-06-03 Tick は時刻だけ進める", func(t *testing.T) {
		w := NewWorld(catalog, 123)
		w.rng.Uint64()
		w.Tick()
		w.Tick()
		if w.Now != 2 || w.Money != 75 || len(w.Inventory.Stacks()) != 0 {
			t.Fatalf("Tick 後の状態が違う: %+v", w)
		}
		other := NewWorld(catalog, 123)
		other.rng.Uint64()
		if w.rng.Uint64() != other.rng.Uint64() {
			t.Fatal("Tick が乱数列を変えた")
		}
	})
}
