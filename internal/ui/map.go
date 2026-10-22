package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const tilePixels = 48

// drawMap は見えるマスだけをカメラの範囲に描く（DES-UI-15、DES-UI-16）。
func (g *Game) drawMap(screen *ebiten.Image) {
	m, ok := g.catalog.Map(g.world.Player.Scene)
	if !ok || g.townTiles == nil || g.dungeonTiles == nil {
		return
	}
	if g.mapImage == nil {
		g.mapImage = ebiten.NewImage(width, height-72)
	}
	dst := g.mapImage
	dst.Clear()
	ox := cameraOffset(m.Width()*tilePixels, width, g.world.Player.X/2)
	oy := cameraOffset(m.Height()*tilePixels, height-72, g.world.Player.Y/2)
	ground := dominantFloor(m)
	x0, x1 := visibleRange(ox, m.Width(), width)
	y0, y1 := visibleRange(oy, m.Height(), height-72)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			px, py := ox+x*tilePixels, oy+y*tilePixels
			baseTown, baseNumber := groundTile(ground)
			g.drawTile(dst, baseTown, baseNumber, px, py)
			ch := m.Cell(x, y)
			if defs.TileInfo(ch).Kind == defs.CellLot {
				continue
			}
			town, n := mapTile(m, x, y, ch)
			if ch == ground {
				continue
			}
			if ch >= '1' && ch <= '9' {
				for _, exit := range m.Exits {
					if exit.Mark == ch {
						town, n = exitTile(exit.Look)
						break
					}
				}
			}
			g.drawTile(dst, town, n, px, py)
		}
	}
	for _, lot := range m.Lots {
		facility := g.catalog.Facilities[lot.Facility]
		level := g.world.Facilities[lot.Facility]
		look := lotLook(facility, level)
		for i, n := range lotTiles(look) {
			g.drawTile(dst, true, n, ox+(lot.X+i%3)*tilePixels, oy+(lot.Y+i/3)*tilePixels)
		}
	}
	g.drawPlayer(dst, ox, oy)
	g.drawMapNames(dst, m, ox, oy, x0, x1, y0, y1)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 72)
	screen.DrawImage(dst, op)
}

// drawMapNames は全タイルと主人公の上に設備・出口の名を描く（DES-UI-15）。
func (g *Game) drawMapNames(dst *ebiten.Image, m defs.MapDef, ox, oy, x0, x1, y0, y1 int) {
	fixture, facing := g.world.FacingTarget()
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			ch := m.Cell(x, y)
			info := defs.TileInfo(ch)
			name := info.Name
			if info.Kind == defs.CellLot {
				continue
			}
			if info.Kind == defs.CellExit {
				for _, exit := range m.Exits {
					if exit.Mark == ch {
						name = g.exitName(exit.To)
						break
					}
				}
			}
			if name == "" {
				continue
			}
			ink := white
			if facing && info.Fixture != "" && info.Fixture == fixture.ID {
				name = "E: " + name
				ink = color.RGBA{255, 235, 93, 255}
			}
			g.drawMapName(dst, ox+x*tilePixels, oy+y*tilePixels, name, ink)
		}
	}
	for _, lot := range m.Lots {
		if lot.X+3 <= x0 || lot.X >= x1 || lot.Y+3 <= y0 || lot.Y >= y1 {
			continue
		}
		name := lotName(g.catalog.Facilities[lot.Facility].Name, g.world.Facilities[lot.Facility])
		ink := white
		if facing && fixture.Kind == core.TargetLot && fixture.ID == lot.Facility {
			name = "E: " + name
			ink = color.RGBA{255, 235, 93, 255}
		}
		g.drawMapName(dst, ox+lot.X*tilePixels, oy+lot.Y*tilePixels, name, ink)
	}
}

// drawMapName は名前をマスの上端より上に置き、表示枠内に収める（DES-UI-15〜16）。
func (g *Game) drawMapName(dst *ebiten.Image, x, y int, name string, ink color.RGBA) {
	w, h := text.Measure(name, g.face, 0)
	boxW, boxH := int(math.Ceil(w))+6, int(math.Ceil(h))+4
	if x+boxW > width {
		x = width - boxW
	}
	if x < 0 {
		x = 0
	}
	y -= boxH
	if y < 0 {
		y = 0
	}
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(boxW), float32(boxH), color.RGBA{22, 28, 37, 230}, false)
	g.label(dst, x+3, y+2, name, ink)
}

// cameraOffset は小さい地図を中央に置き、大きい地図の端で止める（DES-UI-16）。
func cameraOffset(size, viewport, player int) int {
	if size <= viewport {
		return (viewport - size) / 2
	}
	offset := viewport/2 - player
	if offset > 0 {
		return 0
	}
	if offset < viewport-size {
		return viewport - size
	}
	return offset
}

// visibleRange は表示枠に交わるマスの範囲を求める（DES-UI-16）。
func visibleRange(offset, count, viewport int) (int, int) {
	first := 0
	if offset < 0 {
		first = (-offset) / tilePixels
	}
	last := (viewport - offset + tilePixels - 1) / tilePixels
	if last > count {
		last = count
	}
	return first, last
}

// dominantFloor は下地の床を個数と優先順で選ぶ（DES-UI-15）。
func dominantFloor(m defs.MapDef) byte {
	count := map[byte]int{'.': 0, ',': 0, ':': 0}
	for _, row := range m.Rows {
		for i := 0; i < len(row); i++ {
			if _, ok := count[row[i]]; ok {
				count[row[i]]++
			}
		}
	}
	best := byte('.')
	for _, ch := range []byte{',', ':'} {
		if count[ch] > count[best] {
			best = ch
		}
	}
	return best
}

func groundTile(ch byte) (bool, int) {
	switch ch {
	case ',':
		return true, 0
	case ':':
		return true, 25
	default:
		return false, 48
	}
}

// mapTile は文字と隣の文字を絵番号に対応付ける（DES-UI-15、4.3 節）。
func mapTile(m defs.MapDef, x, y int, ch byte) (bool, int) {
	switch ch {
	case '.':
		return false, 48
	case ',':
		return true, 0
	case ':':
		return true, 25
	case '#':
		return false, 40
	case 'W':
		left, right := m.Cell(x-1, y), m.Cell(x+1, y)
		if left != 'W' && !(left >= '1' && left <= '9') && (right == 'W' || right >= '1' && right <= '9') {
			return true, 72
		}
		if right != 'W' && !(right >= '1' && right <= '9') && (left == 'W' || left >= '1' && left <= '9') {
			return true, 75
		}
		return true, 73
	case '^':
		n := 53
		if m.Cell(x-1, y) != '^' && m.Cell(x+1, y) == '^' {
			n = 52
		}
		if m.Cell(x+1, y) != '^' && m.Cell(x-1, y) == '^' {
			n = 54
		}
		if m.Cell(x, y-1) == '^' {
			n += 12
		}
		return true, n
	case 'T':
		return true, 4
	case 'D':
		return true, 86
	case 'M':
		return true, 57
	case 'Q':
		return true, 83
	}
	return true, 25
}

func exitTile(look string) (bool, int) {
	switch look {
	case "stone_door":
		return false, 45
	case "wood_door":
		return true, 86
	default:
		return true, 25
	}
}

// drawTile は番号 n の 16×16 px を 3 倍で描く（DES-UI-14、DES-UI-16）。
func (g *Game) drawTile(dst *ebiten.Image, town bool, n, x, y int) {
	sheet := g.dungeonTiles
	if town {
		sheet = g.townTiles
	}
	tile := sheet.SubImage(image.Rect((n%12)*16, (n/12)*16, (n%12+1)*16, (n/12+1)*16)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(3, 3)
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(tile, op)
}

// drawPlayer は足元と当たり判定を合わせ、向きの印を描く（DES-UI-17）。
func (g *Game) drawPlayer(dst *ebiten.Image, ox, oy int) {
	p := g.world.Player
	if p.Facing == defs.DirLeft || p.Facing == defs.DirRight {
		g.lastHorizontal = p.Facing
	}
	x, y := ox+p.X/2, oy+p.Y/2
	tile := g.dungeonTiles.SubImage(image.Rect(0, 7*16, 16, 8*16)).(*ebiten.Image) // TD 84
	op := &ebiten.DrawImageOptions{}
	if g.lastHorizontal == defs.DirLeft {
		op.GeoM.Scale(-3, 3)
		op.GeoM.Translate(float64(x+24), float64(y-32))
	} else {
		op.GeoM.Scale(3, 3)
		op.GeoM.Translate(float64(x-24), float64(y-32))
	}
	dst.DrawImage(tile, op)
	ink := color.RGBA{255, 231, 75, 255}
	points := directionMarker(x, y, p.Facing)
	ax, ay, bx, by, cx, cy := points[0], points[1], points[2], points[3], points[4], points[5]
	var path vector.Path
	path.MoveTo(ax, ay)
	path.LineTo(bx, by)
	path.LineTo(cx, cy)
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(ink.R) / 255
		vs[i].ColorG = float32(ink.G) / 255
		vs[i].ColorB = float32(ink.B) / 255
		vs[i].ColorA = 1
	}
	dst.DrawTriangles(vs, is, g.markerImage, &ebiten.DrawTrianglesOptions{})
}

// exitName は出口の行き先を表示する（DES-UI-15）。
func (g *Game) exitName(to string) string {
	if to == defs.ExitGather {
		return "採取地へ"
	}
	if m, ok := g.catalog.Map(to); ok {
		return m.Name + "へ"
	}
	return fmt.Sprintf("%sへ", to)
}
