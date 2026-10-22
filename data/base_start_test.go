package data_test

import (
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 起動直後に本体の工房の扉を使えることを確かめる（DES-FLD-05・07、E-110）。
func TestBaseInitialFacingTarget(t *testing.T) {
	t.Run("TC-IF-08-80 本体の開始直後に工房の扉を向く", func(t *testing.T) {
		c, err := defload.Load(data.Base)
		if err != nil {
			t.Fatal(err)
		}
		w := core.NewWorld(c, 42)
		before := w.Player
		got, ok := w.FacingTarget()
		want := core.Target{Kind: core.TargetFixture, ID: defs.FixtureWorkshopDoor, X: 10, Y: 14}
		if !ok || got != want || w.Player != before {
			t.Fatalf("target=%+v,%v want=%+v; player=%+v want=%+v", got, ok, want, w.Player, before)
		}
	})
}
