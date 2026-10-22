package ui

import (
	"github.com/sakana396/fushigi_renkin/internal/ui/uimodel"
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
	case defs.FixtureWorkshopDoor:
		g.workshopTab = uimodel.WorkshopTab(g.workshopTab)
		switch g.workshopTab {
		case "錬金":
			g.drawInventory(panel, false)
			g.drawCandidates(panel)
			g.drawRecipes(panel)
			g.drawCraft(panel)
		case "ゴーレム":
			g.drawGolems(panel)
			g.drawAssignment(panel)
		case "在庫":
			g.drawInventory(panel, false)
		}
	case defs.FixtureShop:
		g.drawInventory(panel, true)
		g.drawShop(panel)
	case defs.FixtureQuestBoard:
		g.drawQuests(panel)
	case "lot":
		g.drawInventory(panel, false)
		g.drawFacilities(panel)
	}
	g.windowDrawing = false
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 42)
	screen.DrawImage(panel, op)
	vector.DrawFilledRect(screen, 0, 72, width, 34, panelColor, false)
	g.button(screen, 10, 76, 120, 28, "閉じる", false, func() { g.window = "" })
	name := ""
	switch g.window {
	case defs.FixtureWorkshopDoor:
		name = "工房"
	case defs.FixtureShop:
		name = "店"
	case defs.FixtureQuestBoard:
		name = "依頼板"
	case "lot":
		name = g.catalog.Facilities[g.windowFacility].Name
	}
	g.label(screen, 140, 82, name, white)
	if g.window == defs.FixtureWorkshopDoor {
		for i, tab := range []string{"錬金", "ゴーレム", "在庫"} {
			g.button(screen, 240+i*120, 76, 110, 28, tab, g.workshopTab == tab, func() { g.workshopTab = tab })
		}
	}
}
