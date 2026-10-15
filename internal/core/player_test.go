package core

import (
	"fmt"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 計算例の地図を作る（DES-FLD-01〜05）。
func playerCatalog() defs.Catalog {
	return defs.Catalog{Maps: map[string]defs.MapDef{
		defs.MapWorkshop: {ID: defs.MapWorkshop, Rows: []string{"########", "#.KS..G#", "#......#", "#......#", "#......#", "####1###"}, StartX: 4, StartY: 3, Exits: []defs.ExitDef{{Mark: '1', To: defs.MapTown, Dir: defs.DirUp, X: 4, Y: 5}}},
		defs.MapTown:     {ID: defs.MapTown, Rows: []string{"TTTTTTT", "TWW1WWT", "TM,,,QT", "TF,,,,T", "TTT2TTT"}, StartX: 3, StartY: 2, Exits: []defs.ExitDef{{Mark: '1', To: defs.MapWorkshop, Dir: defs.DirDown, X: 3, Y: 1}, {Mark: '2', To: defs.ExitGather, Dir: defs.DirUp, X: 3, Y: 4}}},
	}}
}

func TestPlayerExamples(t *testing.T) {
	workshop := defs.MapWorkshop
	town := defs.MapTown
	gather := SceneGather
	down, up, left, right := defs.DirDown, defs.DirUp, defs.DirLeft, defs.DirRight
	type example struct {
		number        int
		id            string
		before        Player
		action        string
		dx, dy, count int
		want          Player
		fixture       string
		found         bool
		wantErr       error
		battle        bool
		noMaps        bool
	}
	cases := []example{
		{1, "TC-FR-FLD-02-01", Player{}, "new", 0, 0, 0, Player{workshop, 432, 336, down}, "", false, nil, false, false},
		{2, "TC-FR-FLD-02-02", Player{workshop, 432, 336, down}, "move", 1, 0, 1, Player{workshop, 439, 336, right}, "", false, nil, false, false},
		{3, "TC-FR-FLD-02-03", Player{workshop, 432, 336, down}, "move", 1, 1, 1, Player{workshop, 437, 341, down}, "", false, nil, false, false},
		{4, "TC-FR-FLD-02-04", Player{workshop, 432, 336, down}, "move", -1, -1, 1, Player{workshop, 427, 331, left}, "", false, nil, false, false},
		{5, "TC-FR-FLD-02-05", Player{workshop, 432, 336, down}, "move", 0, 0, 1, Player{workshop, 432, 336, down}, "", false, nil, false, false},
		{6, "TC-FR-FLD-02-06", Player{workshop, 636, 336, right}, "move", 1, 0, 1, Player{workshop, 640, 336, right}, "", false, nil, false, false},
		{7, "TC-FR-FLD-02-07", Player{workshop, 640, 336, right}, "move", 1, 0, 1, Player{workshop, 640, 336, right}, "", false, nil, false, false},
		{8, "TC-FR-FLD-02-08", Player{workshop, 640, 336, right}, "move", 1, 1, 1, Player{workshop, 640, 341, right}, "", false, nil, false, false},
		{9, "TC-FR-FLD-02-09", Player{workshop, 240, 336, down}, "move", 0, -1, 17, Player{workshop, 240, 224, up}, "", false, nil, false, false},
		{10, "TC-IF-08-01", Player{workshop, 240, 224, up}, "fixture", 0, 0, 0, Player{workshop, 240, 224, up}, defs.FixtureCauldron, true, nil, false, false},
		{11, "TC-IF-08-02", Player{workshop, 288, 224, up}, "fixture", 0, 0, 0, Player{workshop, 288, 224, up}, defs.FixtureCauldron, true, nil, false, false},
		{12, "TC-IF-08-03", Player{workshop, 290, 224, up}, "fixture", 0, 0, 0, Player{workshop, 290, 224, up}, defs.FixtureShelf, true, nil, false, false},
		{13, "TC-IF-08-04", Player{workshop, 240, 224, down}, "fixture", 0, 0, 0, Player{workshop, 240, 224, down}, "", false, nil, false, false},
		{14, "TC-IF-08-05", Player{workshop, 240, 240, up}, "fixture", 0, 0, 0, Player{workshop, 240, 240, up}, "", false, nil, false, false},
		{15, "TC-IF-08-06", Player{workshop, 240, 239, up}, "fixture", 0, 0, 0, Player{workshop, 240, 239, up}, defs.FixtureCauldron, true, nil, false, false},
		{16, "TC-FR-FLD-02-10", Player{workshop, 391, 432, down}, "move", 0, 1, 3, Player{workshop, 391, 448, down}, "", false, nil, false, false},
		{17, "TC-FR-FLD-03-01", Player{workshop, 432, 432, up}, "move", 0, 1, 7, Player{town, 336, 240, down}, "", false, nil, false, false},
		{18, "TC-FR-FLD-03-02", Player{town, 336, 336, up}, "move", 0, 1, 7, Player{gather, 336, 385, down}, "", false, nil, false, false},
		{19, "TC-FR-FLD-03-03", Player{gather, 336, 385, down}, "move", 0, 1, 1, Player{gather, 336, 385, down}, "", false, ErrNotOnMap, false, false},
		{20, "TC-FR-FLD-03-04", Player{gather, 336, 385, down}, "leave", 0, 0, 0, Player{town, 336, 336, up}, "", false, nil, false, false},
		{21, "TC-FR-FLD-03-05", Player{gather, 336, 385, down}, "leave", 0, 0, 0, Player{gather, 336, 385, down}, "", false, ErrInBattle, true, false},
		{22, "TC-FR-FLD-03-06", Player{town, 336, 336, up}, "leave", 0, 0, 0, Player{town, 336, 336, up}, "", false, ErrNotInGather, false, false},
		{23, "TC-IF-08-07", Player{workshop, 432, 336, down}, "move", 2, 0, 1, Player{workshop, 432, 336, down}, "", false, ErrInvalidDirection, false, false},
		{24, "TC-FR-FLD-03-07", Player{}, "move", 1, 0, 1, Player{}, "", false, ErrNotOnMap, false, true},
		{25, "TC-FR-FLD-02-11", Player{workshop, 410, 448, down}, "move", 0, 1, 1, Player{workshop, 416, 448, down}, "", false, nil, false, false},
		{26, "TC-FR-FLD-02-12", Player{workshop, 416, 448, down}, "move", 0, 1, 1, Player{workshop, 416, 455, down}, "", false, nil, false, false},
		{27, "TC-FR-FLD-02-13", Player{workshop, 396, 448, down}, "move", 0, 1, 1, Player{workshop, 403, 448, down}, "", false, nil, false, false},
		{28, "TC-FR-FLD-02-14", Player{workshop, 468, 448, down}, "move", 0, 1, 1, Player{workshop, 461, 448, down}, "", false, nil, false, false},
		{29, "TC-FR-FLD-02-15", Player{workshop, 300, 448, down}, "move", 0, 1, 1, Player{workshop, 300, 448, down}, "", false, nil, false, false},
		{30, "TC-FR-FLD-02-16", Player{workshop, 410, 448, down}, "move", 1, 1, 1, Player{workshop, 415, 448, down}, "", false, nil, false, false},
		{31, "TC-IF-08-08", Player{workshop, 432, 336, down}, "move", 1, 0, 1, Player{workshop, 432, 336, down}, "", false, ErrInBattle, true, false},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s 計算例%d", tc.id, tc.number), func(t *testing.T) {
			c := playerCatalog()
			if tc.noMaps {
				c.Maps = nil
			}
			w := NewWorld(c, 1)
			if tc.action != "new" {
				w.Player = tc.before
			}
			if tc.battle {
				w.Battle = &Battle{}
			}
			var err error
			switch tc.action {
			case "move":
				for i := 0; i < tc.count; i++ {
					err = w.MovePlayer(tc.dx, tc.dy)
					if err != nil {
						break
					}
				}
			case "leave":
				err = w.LeaveGather()
			case "fixture":
				got, ok := w.FacingFixture()
				if got != tc.fixture || ok != tc.found {
					t.Errorf("設備=(%q,%v), 期待=(%q,%v)", got, ok, tc.fixture, tc.found)
				}
			}
			if err != tc.wantErr {
				t.Errorf("誤り=%v, 期待=%v", err, tc.wantErr)
			}
			if w.Player != tc.want {
				t.Errorf("主人公=%+v, 期待=%+v", w.Player, tc.want)
			}
		})
	}
}

func TestPlayerAdditionalRules(t *testing.T) {
	t.Run("TC-IF-08-09 向きの誤りは戦闘と地図より先", func(t *testing.T) {
		w := NewWorld(defs.Catalog{}, 1)
		w.Battle = &Battle{}
		before := w.Player
		if err := w.MovePlayer(2, 0); err != ErrInvalidDirection {
			t.Fatalf("誤り=%v", err)
		}
		if w.Player != before {
			t.Fatalf("失敗で状態が変わった: %+v", w.Player)
		}
	})
	t.Run("TC-IF-08-10 戦闘の誤りは地図より先", func(t *testing.T) {
		w := NewWorld(defs.Catalog{}, 1)
		w.Battle = &Battle{}
		before := w.Player
		if err := w.MovePlayer(0, 0); err != ErrInBattle {
			t.Fatalf("誤り=%v", err)
		}
		if w.Player != before {
			t.Fatalf("失敗で状態が変わった: %+v", w.Player)
		}
	})
	t.Run("TC-FR-FLD-02-17 地図外と未知の文字", func(t *testing.T) {
		c := playerCatalog()
		m := c.Maps[defs.MapWorkshop]
		m.Rows[3] = "#.X....#"
		c.Maps[defs.MapWorkshop] = m
		for _, tc := range []struct {
			p      Player
			dx, dy int
			want   Player
		}{
			{Player{defs.MapWorkshop, 128, 336, defs.DirLeft}, -1, 0, Player{defs.MapWorkshop, 128, 336, defs.DirLeft}},
			{Player{defs.MapWorkshop, 240, 256, defs.DirDown}, 0, 1, Player{defs.MapWorkshop, 240, 256, defs.DirDown}},
		} {
			w := NewWorld(c, 1)
			w.Player = tc.p
			if err := w.MovePlayer(tc.dx, tc.dy); err != nil {
				t.Fatal(err)
			}
			if w.Player != tc.want {
				t.Errorf("主人公=%+v, 期待=%+v", w.Player, tc.want)
			}
		}
	})
	t.Run("TC-FR-FLD-02-18 不明な行き先の出口を通れない", func(t *testing.T) {
		c := playerCatalog()
		m := c.Maps[defs.MapWorkshop]
		m.Exits[0].To = "missing"
		c.Maps[defs.MapWorkshop] = m
		w := NewWorld(c, 1)
		w.Player = Player{defs.MapWorkshop, 432, 448, defs.DirDown}
		if err := w.MovePlayer(0, 1); err != nil {
			t.Fatal(err)
		}
		if w.Player != (Player{defs.MapWorkshop, 432, 448, defs.DirDown}) {
			t.Errorf("主人公=%+v", w.Player)
		}
	})
	t.Run("TC-FR-FLD-02-19 負の座標は地図外", func(t *testing.T) {
		if floorTile(-1) != -1 || floorTile(-96) != -1 || floorTile(-97) != -2 {
			t.Fatal("負の座標の切り捨て")
		}
		w := NewWorld(playerCatalog(), 1)
		m := w.catalog.Maps[defs.MapWorkshop]
		if w.passable(m, -1, 2) {
			t.Fatal("地図外を通れる")
		}
	})
	t.Run("TC-FR-FLD-03-08 街から工房へ戻る", func(t *testing.T) {
		w := NewWorld(playerCatalog(), 1)
		w.Player = Player{defs.MapTown, 336, 193, defs.DirUp}
		if err := w.MovePlayer(0, -1); err != nil {
			t.Fatal(err)
		}
		want := Player{defs.MapWorkshop, 432, 432, defs.DirUp}
		if w.Player != want {
			t.Errorf("主人公=%+v, 期待=%+v", w.Player, want)
		}
	})
	t.Run("TC-FR-FLD-03-09 場面が不明なら移動しない", func(t *testing.T) {
		w := NewWorld(playerCatalog(), 1)
		w.Player.Scene = "missing"
		before := w.Player
		if err := w.MovePlayer(1, 0); err != ErrNotOnMap {
			t.Fatalf("誤り=%v", err)
		}
		if w.Player != before {
			t.Fatalf("失敗で状態が変わった: %+v", w.Player)
		}
	})
	t.Run("TC-IF-08-11 地図外では設備を返さない", func(t *testing.T) {
		w := NewWorld(playerCatalog(), 1)
		w.Player.Scene = SceneGather
		before := w.Player
		got, ok := w.FacingFixture()
		if got != "" || ok || w.Player != before {
			t.Fatalf("設備=(%q,%v), 主人公=%+v", got, ok, w.Player)
		}
	})
	t.Run("TC-IF-08-12 採取地の値は定義と同じ", func(t *testing.T) {
		if SceneGather != defs.ExitGather {
			t.Fatalf("%q != %q", SceneGather, defs.ExitGather)
		}
	})
	t.Run("TC-IF-08-13 Tick と移動は独立", func(t *testing.T) {
		w := NewWorld(playerCatalog(), 1)
		start := w.Player
		for i := 0; i < 10; i++ {
			w.Tick()
		}
		if w.Player != start || w.Now != 10 {
			t.Fatalf("Tick 後: %+v, Now=%d", w.Player, w.Now)
		}
		for i := 0; i < 3; i++ {
			if err := w.MovePlayer(1, 0); err != nil {
				t.Fatal(err)
			}
		}
		if w.Player != (Player{defs.MapWorkshop, 453, 336, defs.DirRight}) || w.Now != 10 {
			t.Fatalf("移動後: %+v, Now=%d", w.Player, w.Now)
		}
	})
	t.Run("TC-IF-08-14 同じ入力なら同じ移動", func(t *testing.T) {
		a, b := NewWorld(playerCatalog(), 7), NewWorld(playerCatalog(), 7)
		for _, d := range [][2]int{{1, 0}, {1, 1}, {0, -1}, {-1, 0}} {
			if err := a.MovePlayer(d[0], d[1]); err != nil {
				t.Fatal(err)
			}
			if err := b.MovePlayer(d[0], d[1]); err != nil {
				t.Fatal(err)
			}
		}
		if a.Player != b.Player {
			t.Fatalf("%+v != %+v", a.Player, b.Player)
		}
	})
}
