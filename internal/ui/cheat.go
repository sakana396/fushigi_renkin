package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sakana396/fushigi_renkin/internal/core"
)

// テスト用のチートの量（Issue #24）。
const (
	cheatMoney    = 10000
	cheatQuality  = 100
	cheatCount    = 20
	fastTimeScale = 10
)

// handleCheatKeys はテスト用のチートのキーを処理する（Issue #24、DES-PKG-05）。
// F12: お金と全アイテムを加える。F11: 時間の倍率を 1 倍と 10 倍で切り替える。
func (g *Game) handleCheatKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		g.grantCheat()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		if g.timeScale == fastTimeScale {
			g.timeScale = 1
		} else {
			g.timeScale = fastTimeScale
		}
		g.message = fmt.Sprintf("チート: 時間 ×%d", g.timeScale)
	}
}

// grantCheat はお金を増やし、目録のすべてのアイテムを同じ品質で在庫に加える（Issue #24）。
func (g *Game) grantCheat() {
	g.world.Money += cheatMoney
	added := 0
	for _, id := range keys(g.catalog.Items) {
		if err := g.world.Inventory.Add(core.ItemKey{ItemID: id, Quality: cheatQuality}, cheatCount); err != nil {
			g.result("", err)
			return
		}
		added++
	}
	g.message = fmt.Sprintf("チート: お金 +%d、全 %d 種類のアイテムを品質 %d で %d 個ずつ追加", cheatMoney, added, cheatQuality, cheatCount)
}
