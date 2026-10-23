package ui

import (
	"fmt"
	"github.com/sakana396/fushigi_renkin/internal/ui/uimodel"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const tilePixels = uimodel.TilePixels

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
	ox := uimodel.CameraOffset(m.Width()*tilePixels, width, g.world.Player.X/2)
	oy := uimodel.CameraOffset(m.Height()*tilePixels, height-72, g.world.Player.Y/2)
	ground := uimodel.DominantFloor(m)
	x0, x1 := uimodel.VisibleRange(ox, m.Width(), width)
	y0, y1 := uimodel.VisibleRange(oy, m.Height(), height-72)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			px, py := ox+x*tilePixels, oy+y*tilePixels
			baseTown, baseNumber := uimodel.GroundTile(ground)
			g.drawTile(dst, baseTown, baseNumber, px, py)
			ch := m.Cell(x, y)
			if defs.TileInfo(ch).Kind == defs.CellLot {
				continue
			}
			switch ch {
			case 'o':
				_, alpha := uimodel.PickDisplay(g.world.PickRemaining(m.ID, x, y))
				g.drawRock(dst, px, py, alpha)
				continue
			case '~':
				g.drawWater(dst, px, py)
				continue
			case 'E':
				g.drawMapEnemy(dst, m.ID, px, py)
				continue
			}
			town, n := uimodel.MapTile(m, x, y, ch)
			if ch == ground {
				continue
			}
			if ch >= '1' && ch <= '9' {
				for _, exit := range m.Exits {
					if exit.Mark == ch {
						town, n = uimodel.ExitTile(exit.Look)
						break
					}
				}
			}
			alpha := float32(1)
			if defs.TileInfo(ch).Kind == defs.CellPick {
				_, alpha = uimodel.PickDisplay(g.world.PickRemaining(m.ID, x, y))
			}
			g.drawTileAlpha(dst, town, n, px, py, alpha)
			for _, exit := range m.Exits {
				if exit.X == x && exit.Y == y && uimodel.GatherExitMarker(exit.To) {
					arrow := uimodel.ExitArrow(px, py)
					vector.StrokeLine(dst, arrow[0], arrow[1], arrow[2], arrow[3], 3, color.RGBA{255, 235, 93, 255}, false)
					vector.StrokeLine(dst, arrow[4], arrow[5], arrow[2], arrow[3], 3, color.RGBA{255, 235, 93, 255}, false)
					vector.StrokeLine(dst, 2*arrow[0]-arrow[4], arrow[5], arrow[2], arrow[3], 3, color.RGBA{255, 235, 93, 255}, false)
				}
			}
		}
	}
	for _, lot := range m.Lots {
		facility := g.catalog.Facilities[lot.Facility]
		level := g.world.Facilities[lot.Facility]
		look := uimodel.LotLook(facility, level)
		for i, n := range uimodel.LotTiles(look) {
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
			if info.Kind == defs.CellPick {
				name, _ = uimodel.PickDisplay(g.world.PickRemaining(m.ID, x, y))
			}
			if info.Kind == defs.CellEnemy {
				name = uimodel.EnemyName(g.catalog, g.catalog.Gathers[m.ID])
			}
			bright := false
			if info.Kind == defs.CellLot {
				continue
			}
			if info.Kind == defs.CellExit {
				for _, exit := range m.Exits {
					if exit.Mark == ch {
						name = g.exitName(exit.To)
						bright = uimodel.GatherExitMarker(exit.To)
						break
					}
				}
			}
			if name == "" {
				continue
			}
			ink := white
			if bright {
				ink = color.RGBA{255, 235, 93, 255}
			}
			if uimodel.FacingCell(fixture, facing, x, y) && (info.Kind == defs.CellFixture || info.Kind == defs.CellPick || info.Kind == defs.CellEnemy) {
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
		name := uimodel.LotName(g.catalog.Facilities[lot.Facility].Name, g.world.Facilities[lot.Facility])
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
	x, y = uimodel.MapNamePosition(x, y, boxW, boxH, width)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(boxW), float32(boxH), color.RGBA{22, 28, 37, 230}, false)
	g.label(dst, x+3, y+2, name, ink)
}

// drawTile は番号 n の 16×16 px を 3 倍で描く（DES-UI-14、DES-UI-16）。
func (g *Game) drawTile(dst *ebiten.Image, town bool, n, x, y int) {
	g.drawTileAlpha(dst, town, n, x, y, 1)
}

// drawTileAlpha は空の拾う場所だけ同じ絵を半透明にする（DES-UI-26）。
func (g *Game) drawTileAlpha(dst *ebiten.Image, town bool, n, x, y int, alpha float32) {
	sheet := g.dungeonTiles
	if town {
		sheet = g.townTiles
	}
	tile := sheet.SubImage(image.Rect((n%12)*16, (n/12)*16, (n%12+1)*16, (n/12+1)*16)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(3, 3)
	op.ColorScale.ScaleAlpha(alpha)
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(tile, op)
}

// drawWater は通れない水を青いマスと暗い青の縁で描く（DES-UI-26、DES-IT06 7章）。
func (g *Game) drawWater(dst *ebiten.Image, x, y int) {
	vector.DrawFilledRect(dst, float32(x), float32(y), tilePixels, tilePixels, color.RGBA{35, 81, 133, 255}, false)
	vector.DrawFilledRect(dst, float32(x+3), float32(y+3), tilePixels-6, tilePixels-6, color.RGBA{53, 119, 184, 255}, false)
}

// drawMapEnemy は最初の敵の絵、または色の四角と名前の頭文字を描く（DES-UI-26）。
func (g *Game) drawMapEnemy(dst *ebiten.Image, siteID string, x, y int) {
	tile, initial := uimodel.EnemyAppearance(g.catalog, g.catalog.Gathers[siteID])
	if tile >= 0 {
		g.drawTile(dst, false, tile, x, y)
		return
	}
	vector.DrawFilledRect(dst, float32(x+6), float32(y+6), tilePixels-12, tilePixels-12, color.RGBA{163, 70, 75, 255}, false)
	w, h := text.Measure(initial, g.face, 0)
	g.label(dst, x+(tilePixels-int(math.Ceil(w)))/2, y+(tilePixels-int(math.Ceil(h)))/2, initial, white)
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
	points := uimodel.DirectionMarker(x, y, p.Facing)
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

// drawRock は背景のない灰色の岩を床に重ね、空のときは絵全体を薄くする（DES-UI-26、DES-IT06 7章）。
func (g *Game) drawRock(dst *ebiten.Image, x, y int, alpha float32) {
	if g.rockImage == nil {
		g.rockImage = ebiten.NewImage(tilePixels, tilePixels)
		// 大小の岩に暗い縁と左上からの光を付ける。
		for _, rock := range []struct{ x, y, radius float32 }{{15, 32, 10}, {29, 29, 14}, {23, 19, 12}} {
			vector.DrawFilledCircle(g.rockImage, rock.x, rock.y, rock.radius, color.RGBA{66, 69, 75, 255}, false)
			vector.DrawFilledCircle(g.rockImage, rock.x-1, rock.y-1, rock.radius-2, color.RGBA{133, 137, 145, 255}, false)
			vector.DrawFilledCircle(g.rockImage, rock.x-3, rock.y-4, rock.radius/3, color.RGBA{189, 193, 199, 255}, false)
		}
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleAlpha(alpha)
	dst.DrawImage(g.rockImage, op)
}
