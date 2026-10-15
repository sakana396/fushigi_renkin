package data_test

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// TestBaseMapsIT05 は本体の地図と歩く距離を確認する（DES-DAT-03、DES-FLD-06、DES-VAL-17〜23）。
func TestBaseMapsIT05(t *testing.T) {
	c, err := defload.Load(data.Base)
	t.Run("TC-FR-FLD-04-40 本体地図の検証", func(t *testing.T) {
		if err != nil {
			t.Fatalf("DES-VAL-17〜23: %v", err)
		}
		if len(c.Maps) != 2 {
			t.Fatalf("地図の件数 = %d、期待 2", len(c.Maps))
		}
	})
	if err != nil {
		return
	}
	t.Run("TC-FR-FLD-01-30 工房の地図", func(t *testing.T) {
		m := c.Maps[defs.MapWorkshop]
		wantRows := []string{
			"####################",
			"#..................#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#....S...K...G.....#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#..................#",
			"#########1##########",
		}
		wantExits := []defs.ExitDef{{Mark: '1', To: defs.MapTown, Dir: defs.DirUp, Look: "stone_door", X: 9, Y: 11}}
		if m.ID != defs.MapWorkshop || m.Name != "工房" || m.Width() != 20 || m.Height() != 12 || m.StartX != 9 || m.StartY != 8 || !reflect.DeepEqual(m.Rows, wantRows) || !reflect.DeepEqual(m.Exits, wantExits) {
			t.Errorf("工房の地図: %+v", m)
		}
	})
	t.Run("TC-FR-FLD-01-31 街の地図", func(t *testing.T) {
		m := c.Maps[defs.MapTown]
		wantRows := []string{
			"TTTTTTTTTTTTTTTTTTTTTTTTTTTTTT",
			"T,,,,,,,,,,,,,,,,,,,,,,,,,,,,T",
			"T,,^^^^^,,,,,,,,,,,,,,^^^^^,,T",
			"T,,^^^^^,,,,,,,,,,,,,,^^^^^,,T",
			"T,,WWWWW,,,,,,,,,,,,,,WWWWW,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,T,,,,,,,,,,::,,,,,,,,,,T,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,^^^^^,::,^^^^^,,,,,,,T",
			"T,,,,,,,^^^^^,::,^^^^^,,,,,,,T",
			"T,,,,,,,WW1WW,::,WWWWW,,,,,,,T",
			"T,,,,,,,,::::::::::M,,,,,,,,,T",
			"T,,,,,,,,::::Q::F::,,,,,,,,,,T",
			"T,,,,,,,,::::::::::,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"TTTTTTTTTTTTTT2TTTTTTTTTTTTTTT",
		}
		wantExits := []defs.ExitDef{
			{Mark: '1', To: defs.MapWorkshop, Dir: defs.DirDown, Look: "wood_door", X: 10, Y: 14},
			{Mark: '2', To: defs.ExitGather, Dir: defs.DirUp, Look: "road", X: 14, Y: 19},
		}
		if m.ID != defs.MapTown || m.Name != "街" || m.Width() != 30 || m.Height() != 20 || m.StartX != 10 || m.StartY != 15 || !reflect.DeepEqual(m.Rows, wantRows) || !reflect.DeepEqual(m.Exits, wantExits) {
			t.Errorf("街の地図: %+v", m)
		}
	})
	t.Run("TC-FR-FLD-04-41 工房の歩く距離", func(t *testing.T) {
		m := c.Maps[defs.MapWorkshop]
		checkMapDistances(t, m, []mapPlace{
			{"錬金釜", 'K'}, {"在庫の棚", 'S'}, {"ゴーレムの作業台", 'G'}, {"扉", '1'},
		}, [][]int{{0, 2, 2, 4}, {2, 0, 8, 8}, {2, 8, 0, 8}, {4, 8, 8, 0}})
	})
	t.Run("TC-FR-FLD-04-42 街の歩く距離", func(t *testing.T) {
		m := c.Maps[defs.MapTown]
		checkMapDistances(t, m, []mapPlace{
			{"工房の扉", '1'}, {"店", 'M'}, {"依頼板", 'Q'}, {"施設の看板", 'F'}, {"採取地への出口", '2'},
		}, [][]int{{0, 8, 3, 6, 7}, {8, 0, 5, 2, 7}, {3, 5, 0, 1, 2}, {6, 2, 1, 0, 3}, {7, 7, 2, 3, 0}})
	})
}

type mapPlace struct {
	name string
	mark byte
}

type mapPoint struct{ x, y int }

// checkMapDistances は前の床マスの組から最短距離を求める（DES-FLD-06）。
func checkMapDistances(t *testing.T, m defs.MapDef, places []mapPlace, want [][]int) {
	t.Helper()
	fronts := make([][]mapPoint, len(places))
	for i, p := range places {
		count := 0
		for y, row := range m.Rows {
			for x := range row {
				if row[x] != p.mark {
					continue
				}
				count++
				for _, d := range []mapPoint{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
					q := mapPoint{x + d.x, y + d.y}
					if defs.TileInfo(m.Cell(q.x, q.y)).Kind == defs.CellFloor {
						fronts[i] = append(fronts[i], q)
					}
				}
			}
		}
		if count != 1 || len(fronts[i]) == 0 {
			t.Fatalf("%s: 文字 %c は %d 個、前の床マスは %d 個", p.name, p.mark, count, len(fronts[i]))
		}
		if p.mark >= '1' && p.mark <= '9' {
			var arrival mapPoint
			for _, e := range m.Exits {
				if e.Mark == p.mark {
					arrival = mapPoint{e.X, e.Y}
					switch e.Dir {
					case defs.DirUp:
						arrival.y--
					case defs.DirDown:
						arrival.y++
					case defs.DirLeft:
						arrival.x--
					case defs.DirRight:
						arrival.x++
					default:
						t.Fatalf("%s: 向き %s", p.name, e.Dir)
					}
				}
			}
			if defs.TileInfo(m.Cell(arrival.x, arrival.y)).Kind != defs.CellFloor {
				t.Fatalf("%s: 到着マス %+v が床ではない", p.name, arrival)
			}
			fronts[i] = []mapPoint{arrival}
		}
	}
	for i := range places {
		for j := i + 1; j < len(places); j++ {
			got := shortestFloorWalk(m, fronts[i], fronts[j])
			if got != want[i][j] || got > 8 {
				t.Errorf("%s → %s: %d マス、期待 %d マス（上限 8）", places[i].name, places[j].name, got, want[i][j])
			}
		}
	}
}

// shortestFloorWalk は床だけを上下左右に塗りつぶす（DES-FLD-06）。
func shortestFloorWalk(m defs.MapDef, from, to []mapPoint) int {
	target := map[mapPoint]bool{}
	for _, p := range to {
		target[p] = true
	}
	seen := map[mapPoint]bool{}
	queue := append([]mapPoint(nil), from...)
	for _, p := range from {
		seen[p] = true
	}
	for steps := 0; len(queue) > 0; steps++ {
		next := []mapPoint{}
		for _, p := range queue {
			if target[p] {
				return steps
			}
			for _, d := range []mapPoint{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				q := mapPoint{p.x + d.x, p.y + d.y}
				if !seen[q] && defs.TileInfo(m.Cell(q.x, q.y)).Kind == defs.CellFloor {
					seen[q] = true
					next = append(next, q)
				}
			}
		}
		queue = next
	}
	return -1
}
