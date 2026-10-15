package core

import "github.com/sakana396/fushigi_renkin/internal/defs"

// SceneGather は採取地の一覧を表す（DES-FLD-01）。
const SceneGather = defs.ExitGather

// Player は主人公の場面、当たり判定の中心、向きを持つ（DES-FLD-01）。
type Player struct {
	Scene  string
	X, Y   int
	Facing defs.Dir
}

const playerTile = 96
const playerHalf = 32

// MovePlayer は向きを更新し、地図の上を 1 フレーム分進める（DES-FLD-02、DES-FLD-03、DES-FLD-05、DES-TIME-02）。
func (w *World) MovePlayer(dx, dy int) error {
	if dx < -1 || dx > 1 || dy < -1 || dy > 1 {
		return ErrInvalidDirection
	}
	if w.Battle != nil {
		return ErrInBattle
	}
	m, ok := w.catalog.Map(w.Player.Scene)
	if !ok || w.Player.Scene == SceneGather {
		return ErrNotOnMap
	}
	if dx == 0 && dy == 0 {
		return nil
	}
	p := w.Player
	if dx != 0 && dy != 0 {
		keepsHorizontal := dx < 0 && p.Facing == defs.DirLeft || dx > 0 && p.Facing == defs.DirRight
		keepsVertical := dy < 0 && p.Facing == defs.DirUp || dy > 0 && p.Facing == defs.DirDown
		if !keepsHorizontal && !keepsVertical {
			if dx < 0 {
				p.Facing = defs.DirLeft
			} else {
				p.Facing = defs.DirRight
			}
		}
	} else if dx != 0 {
		if dx < 0 {
			p.Facing = defs.DirLeft
		} else {
			p.Facing = defs.DirRight
		}
	} else {
		if dy < 0 {
			p.Facing = defs.DirUp
		} else {
			p.Facing = defs.DirDown
		}
	}
	step := 7
	if dx != 0 && dy != 0 {
		step = 5
	}
	if dx != 0 {
		next, blocked, lead := w.moveAxis(m, p.X, p.Y, dx*step, true)
		if blocked && dy == 0 {
			if shifted, ok := w.cornerShift(m, p.X, p.Y, lead, true); ok {
				p.Y = shifted
			} else {
				p.X = next
			}
		} else {
			p.X = next
		}
	}
	if dy != 0 {
		next, blocked, lead := w.moveAxis(m, p.Y, p.X, dy*step, false)
		if blocked && dx == 0 {
			if shifted, ok := w.cornerShift(m, p.Y, p.X, lead, false); ok {
				p.X = shifted
			} else {
				p.Y = next
			}
		} else {
			p.Y = next
		}
	}
	if cell := m.Cell(floorTile(p.X), floorTile(p.Y)); defs.TileInfo(cell).Kind == defs.CellExit {
		for _, exit := range m.Exits {
			if exit.Mark != cell {
				continue
			}
			if exit.To == SceneGather {
				p.Scene = SceneGather
				break
			}
			dest, found := w.catalog.Map(exit.To)
			if !found {
				break
			}
			for _, arrival := range dest.Exits {
				if arrival.To == m.ID {
					p = arrivalPlayer(dest, arrival)
					break
				}
			}
			break
		}
	}
	w.Player = p
	return nil
}

// moveAxis は進行方向の先端が通れないマスに触れる位置を返す（DES-FLD-03）。
func (w *World) moveAxis(m defs.MapDef, primary, other, delta int, xAxis bool) (int, bool, int) {
	if delta == 0 {
		return primary, false, 0
	}
	proposed := primary + delta
	lead := floorTile(proposed - playerHalf)
	if delta > 0 {
		lead = floorTile(proposed + playerHalf - 1)
	}
	first, last := floorTile(other-playerHalf), floorTile(other+playerHalf-1)
	blocked := false
	for cross := first; cross <= last; cross++ {
		x, y := lead, cross
		if !xAxis {
			x, y = cross, lead
		}
		if !w.passable(m, x, y) {
			blocked = true
			break
		}
	}
	if !blocked {
		return proposed, false, lead
	}
	if delta > 0 {
		return lead*playerTile - playerHalf, true, lead
	}
	return (lead+1)*playerTile + playerHalf, true, lead
}

// cornerShift は壁に少しだけ触れた直進を通れる側へずらす（DES-FLD-03）。
func (w *World) cornerShift(m defs.MapDef, primary, other, lead int, xAxis bool) (int, bool) {
	first, last := floorTile(other-playerHalf), floorTile(other+playerHalf-1)
	if last != first+1 {
		return 0, false
	}
	pass := [2]bool{}
	for i, cross := range []int{first, last} {
		x, y := lead, cross
		if !xAxis {
			x, y = cross, lead
		}
		pass[i] = w.passable(m, x, y)
	}
	if pass[0] == pass[1] {
		return 0, false
	}
	var overlap, shift int
	if !pass[0] {
		overlap = (first+1)*playerTile - (other - playerHalf)
		shift = 1
	} else {
		overlap = other + playerHalf - (last * playerTile)
		shift = -1
	}
	if overlap > 24 {
		return 0, false
	}
	if overlap > 7 {
		overlap = 7
	}
	return other + shift*overlap, true
}

// passable は地図外、不明な文字、不正な行き先の出口を通行不可にする（DES-FLD-03）。
func (w *World) passable(m defs.MapDef, x, y int) bool {
	cell := m.Cell(x, y)
	switch defs.TileInfo(cell).Kind {
	case defs.CellFloor:
		return true
	case defs.CellExit:
		for _, e := range m.Exits {
			if e.Mark == cell {
				if e.To == SceneGather {
					return true
				}
				_, ok := w.catalog.Map(e.To)
				return ok
			}
		}
	}
	return false
}

// floorTile は負の位置も小さい側へ切り捨てる（DES-FLD-03）。
func floorTile(v int) int {
	if v >= 0 {
		return v / playerTile
	}
	return (v+1)/playerTile - 1
}

// FacingFixture は前の帯に最も広く重なる設備を返す（DES-FLD-04）。
func (w *World) FacingFixture() (string, bool) {
	m, ok := w.catalog.Map(w.Player.Scene)
	if !ok || w.Player.Scene == SceneGather {
		return "", false
	}
	p := w.Player
	x0, x1, y0, y1 := p.X-playerHalf, p.X+playerHalf, p.Y-playerHalf, p.Y+playerHalf
	switch p.Facing {
	case defs.DirUp:
		y1 = y0
		y0 -= 16
	case defs.DirDown:
		y0 = y1
		y1 += 16
	case defs.DirLeft:
		x1 = x0
		x0 -= 16
	case defs.DirRight:
		x0 = x1
		x1 += 16
	default:
		return "", false
	}
	bestArea, bestX, bestY, bestFixture := 0, 0, 0, ""
	for y := floorTile(y0); y <= floorTile(y1-1); y++ {
		for x := floorTile(x0); x <= floorTile(x1-1); x++ {
			info := defs.TileInfo(m.Cell(x, y))
			if info.Kind != defs.CellFixture {
				continue
			}
			area := overlap(x0, x1, x*playerTile, (x+1)*playerTile) * overlap(y0, y1, y*playerTile, (y+1)*playerTile)
			if area > bestArea || area == bestArea && area > 0 && (y < bestY || y == bestY && x < bestX) {
				bestArea, bestX, bestY, bestFixture = area, x, y, info.Fixture
			}
		}
	}
	return bestFixture, bestArea > 0
}

func overlap(a0, a1, b0, b1 int) int {
	if b0 > a0 {
		a0 = b0
	}
	if b1 < a1 {
		a1 = b1
	}
	if a1 <= a0 {
		return 0
	}
	return a1 - a0
}

// LeaveGather は採取地の一覧から出口のある地図へ戻る（DES-FLD-05、DES-WLD-01）。
func (w *World) LeaveGather() error {
	if w.Player.Scene != SceneGather {
		return ErrNotInGather
	}
	if w.Battle != nil {
		return ErrInBattle
	}
	for _, m := range w.catalog.Maps {
		for _, e := range m.Exits {
			if e.To == SceneGather {
				w.Player = arrivalPlayer(m, e)
				return nil
			}
		}
	}
	// 検証済みの地図には gather 行きの出入り口があるため、ここには到達しない（DES-FLD-05）。
	return ErrNotOnMap
}

// arrivalPlayer は到着のマスの中心と向きを決める（DES-FLD-05）。
func arrivalPlayer(m defs.MapDef, e defs.ExitDef) Player {
	x, y := e.X, e.Y
	switch e.Dir {
	case defs.DirUp:
		y--
	case defs.DirDown:
		y++
	case defs.DirLeft:
		x--
	case defs.DirRight:
		x++
	}
	return Player{Scene: m.ID, X: x*playerTile + playerTile/2, Y: y*playerTile + playerTile/2, Facing: e.Dir}
}
