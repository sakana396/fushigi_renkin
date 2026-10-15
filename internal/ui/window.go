package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// drawWindow は地図を暗くし、設備に対応する欄を 42 px 下げて描く（DES-UI-19、4.4 節）。
func (g *Game) drawWindow(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 72, width, height-72, color.RGBA{0, 0, 0, 170}, false)
	if g.windowImage == nil {
		g.windowImage = ebiten.NewImage(width, height)
	}
	panel := g.windowImage
	panel.Clear()
	g.windowDrawing = true
	switch g.window {
	case defs.FixtureCauldron:
		g.drawInventory(panel, false)
		g.drawCandidates(panel)
		g.drawRecipes(panel)
		g.drawCraft(panel)
	case defs.FixtureShelf:
		g.drawInventory(panel, false)
	case defs.FixtureWorkbench:
		g.drawGolems(panel)
		g.drawAssignment(panel)
	case defs.FixtureShop:
		g.label(panel, 1195, 100, fmt.Sprintf("売買数: %d", g.quantity), white)
		g.button(panel, 1295, 95, 35, 24, "－", false, func() {
			if g.quantity > 1 {
				g.quantity--
			}
		})
		g.button(panel, 1335, 95, 35, 24, "＋", false, func() {
			if g.quantity < 99 {
				g.quantity++
			}
		})
		g.drawInventory(panel, true)
		g.drawShop(panel)
	case defs.FixtureQuestBoard:
		g.drawQuests(panel)
	case defs.FixtureFacilityBoard:
		g.drawInventory(panel, false)
		g.drawFacilities(panel)
	}
	g.windowDrawing = false
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 42)
	screen.DrawImage(panel, op)
	vector.DrawFilledRect(screen, 0, 72, width, 34, panelColor, false)
	g.button(screen, 10, 76, 120, 28, "閉じる", false, func() { g.window = "" })
	name := g.window
	for _, ch := range []byte{'K', 'S', 'G', 'M', 'Q', 'F'} {
		info := defs.TileInfo(ch)
		if info.Fixture == g.window {
			name = info.Name
			break
		}
	}
	g.label(screen, 140, 82, name, white)
}
