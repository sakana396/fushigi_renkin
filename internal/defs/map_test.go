package defs

import "testing"

// TestMapDefinition は地図の型と固定の文字表を確かめる（DES-DEF-09）。
func TestMapDefinition(t *testing.T) {
	t.Run("TC-FR-FLD-01-01 地図の寸法と地図外", func(t *testing.T) {
		m := MapDef{Rows: []string{".#", "K1"}}
		if m.Width() != 2 || m.Height() != 2 || m.Cell(0, 1) != 'K' || m.Cell(-1, 0) != 0 || m.Cell(2, 0) != 0 || m.Cell(0, 2) != 0 {
			t.Fatalf("map: %+v", m)
		}
	})
	t.Run("TC-FR-FLD-01-02 空の地図", func(t *testing.T) {
		if (MapDef{}).Width() != 0 || (MapDef{}).Cell(0, 0) != 0 {
			t.Fatal("empty map")
		}
	})
	t.Run("TC-FR-FLD-01-03 文字の種類", func(t *testing.T) {
		for _, tc := range []struct {
			chars string
			kind  CellKind
		}{{".,:", CellFloor}, {"#W^T~", CellWall}, {"DMQ", CellFixture}, {"123456789", CellExit}, {"abcdefghi", CellLot}, {"*o", CellPick}, {"E", CellEnemy}, {"0XKSGF", CellUnknown}} {
			for i := 0; i < len(tc.chars); i++ {
				if got := TileInfo(tc.chars[i]).Kind; got != tc.kind {
					t.Errorf("%c: %v", tc.chars[i], got)
				}
			}
		}
	})
	t.Run("TC-FR-FLD-01-04 設備の種類と表示名", func(t *testing.T) {
		for _, tc := range []struct {
			ch         byte
			kind, name string
		}{{'D', FixtureWorkshopDoor, "工房の扉"}, {'M', FixtureShop, "店"}, {'Q', FixtureQuestBoard, "依頼板"}} {
			got := TileInfo(tc.ch)
			if got.Fixture != tc.kind || got.Name != tc.name {
				t.Errorf("%c: %+v", tc.ch, got)
			}
		}
	})
	t.Run("TC-FR-FLD-01-05 IDで地図を引く", func(t *testing.T) {
		c := Catalog{Maps: map[string]MapDef{MapTown: {ID: MapTown}}}
		if m, ok := c.Map(MapTown); !ok || m.ID != MapTown {
			t.Fatal("missing town")
		}
		if _, ok := c.Map("workshop"); ok {
			t.Fatal("unexpected workshop")
		}
	})
}
