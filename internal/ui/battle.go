package ui

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sakana396/fushigi_renkin/internal/core"
)

const battleX, battleY, cellSize = 20, 140, 85

// drawBattle は盤面、能力、操作、使用品と記録を表示する（DES-UI-08、DES-UI-09）。
func (g *Game) drawBattle(dst *ebiten.Image) {
	b := g.world.Battle
	if b == nil || len(b.Order) == 0 {
		return
	}
	actor := b.Units[b.Order[b.Turn]]
	def, _ := g.catalog.Unit(actor.DefID)
	site, _ := g.catalog.Gather(b.GatherID)
	g.label(dst, battleX, 105, "戦闘: "+site.Name+" / 現在の番: "+actor.Label, white)
	for y := 0; y < g.catalog.Game.BattleHeight; y++ {
		for x := 0; x < g.catalog.Game.BattleWidth; x++ {
			cx, cy := x, y
			ink := panelColor
			if g.battleReachable(actor, def.Move, def.Range, x, y) {
				ink = activeColor
			}
			px, py := battleX+x*cellSize, battleY+y*cellSize
			vector.DrawFilledRect(dst, float32(px), float32(py), cellSize-3, cellSize-3, ink, false)
			g.buttons = append(g.buttons, button{px, py, cellSize - 3, cellSize - 3, func() { g.battleCell(cx, cy) }})
		}
	}
	for i, u := range b.Units {
		if u.HP <= 0 {
			continue
		}
		px, py := battleX+u.X*cellSize, battleY+u.Y*cellSize
		ink := color.RGBA{71, 110, 152, 255}
		if !u.Ally {
			ink = color.RGBA{145, 78, 73, 255}
		}
		if i == b.Order[b.Turn] {
			ink = color.RGBA{199, 145, 45, 255}
		}
		vector.DrawFilledRect(dst, float32(px+5), float32(py+5), cellSize-13, cellSize-13, ink, false)
		initial, _ := utf8.DecodeRuneInString(u.Label)
		g.label(dst, px+12, py+12, fmt.Sprintf("%c HP%d", initial, u.HP), white)
		g.label(dst, px+8, py+40, u.Label, white)
	}
	g.box(dst, 715, 100, 875, 195, "現在の仲間")
	g.label(dst, 725, 135, fmt.Sprintf("%s  HP %d/%d  攻撃 %d+%d  素早さ %d+%d", actor.Label, actor.HP, def.HP, def.Attack, actor.AttackBonus, def.Speed, actor.SpeedBonus), white)
	g.label(dst, 725, 160, fmt.Sprintf("移動 %d / 攻撃射程 %d / 移動済み %t", def.Move, def.Range, b.Moved), muted)
	g.button(dst, 725, 205, 110, 32, "移動", g.battleAction == moveAction, func() { g.battleAction = moveAction })
	g.button(dst, 845, 205, 110, 32, "攻撃", g.battleAction == attackAction, func() { g.battleAction = attackAction })
	g.button(dst, 965, 205, 110, 32, "アイテム", g.battleAction == itemAction, func() { g.battleAction = itemAction })
	g.button(dst, 1085, 205, 110, 32, "待機", false, func() { g.finishBattleAction("待機しました", g.world.BattleWait()) })
	g.label(dst, 725, 255, "操作を選び、盤面のマスを押してください", muted)
	g.drawBattleItems(dst)
	g.drawBattleLog(dst, b.Log, 715, 645, 875, 275)
}

// battleReachable は選択した操作の届くマスを判定する（DES-UI-08）。
func (g *Game) battleReachable(actor core.BattleUnit, move, attackRange, x, y int) bool {
	d := abs(actor.X-x) + abs(actor.Y-y)
	switch g.battleAction {
	case moveAction:
		if g.world.Battle.Moved || d > move {
			return false
		}
		for _, u := range g.world.Battle.Units {
			if u.HP > 0 && u.X == x && u.Y == y {
				return false
			}
		}
		return true
	case attackAction:
		return d <= attackRange && g.battleTarget(x, y, false)
	case itemAction:
		item, ok := g.catalog.Item(g.battleItem.ItemID)
		return ok && item.Use != nil && g.world.Inventory.Count(g.battleItem) > 0 && d <= item.Use.Range && g.battleTarget(x, y, item.Use.Target == "ally")
	}
	return false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// battleTarget はマスに対象となる生存ユニットがいるか調べる（DES-UI-08）。
func (g *Game) battleTarget(x, y int, ally bool) bool {
	for _, u := range g.world.Battle.Units {
		if u.HP > 0 && u.Ally == ally && u.X == x && u.Y == y {
			return true
		}
	}
	return false
}

// battleCell は盤面のクリックを選択中の core 操作につなぐ（DES-UI-08）。
func (g *Game) battleCell(x, y int) {
	var err error
	success := ""
	switch g.battleAction {
	case moveAction:
		err, success = g.world.BattleMove(x, y), "移動しました"
	case attackAction:
		err, success = g.world.BattleAttack(x, y), "攻撃しました"
	case itemAction:
		if g.battleItem.ItemID == "" {
			g.message = "操作失敗: アイテムを選んでください"
			return
		}
		err, success = g.world.BattleUseItem(g.battleItem, x, y), g.itemName(g.battleItem.ItemID)+"を使いました"
	}
	g.finishBattleAction(success, err)
}

// finishBattleAction は戦闘が終われば結果を保存したまま採取地へ戻す（DES-UI-10）。
func (g *Game) finishBattleAction(success string, err error) {
	g.result(success, err)
	if err == nil {
		if g.world.Battle != nil {
			g.battleLogPage = pageLimit(len(g.world.Battle.Log), 8)
		} else if g.world.LastBattle != nil {
			g.battleLogPage = pageLimit(len(g.world.LastBattle.Log), 8)
		}
	}
	if err == nil && g.world.Battle == nil {
		g.scene = gatherScene
		g.battleItem = core.ItemKey{}
	}
}

// drawBattleItems は在庫にある戦闘用アイテムと品質別効果を並べる（DES-UI-08、DES-UI-09）。
func (g *Game) drawBattleItems(dst *ebiten.Image) {
	const x, y, w, rows = 715, 305, 875, 9
	g.box(dst, x, y, w, 330, "使えるアイテム（行を選択）")
	items := make([]core.Stack, 0)
	for _, stack := range g.world.Inventory.Stacks() {
		item, ok := g.catalog.Item(stack.Key.ItemID)
		if ok && item.Use != nil && stack.Count > 0 {
			items = append(items, stack)
		}
	}
	if g.battleItemPage > pageLimit(len(items), rows) {
		g.battleItemPage = pageLimit(len(items), rows)
	}
	for i := g.battleItemPage * rows; i < len(items) && i < (g.battleItemPage+1)*rows; i++ {
		stack := items[i]
		key := stack.Key
		item, _ := g.catalog.Item(key.ItemID)
		title := fmt.Sprintf("%s Q%d ×%d  射程 %d  %s", item.Name, key.Quality, stack.Count, item.Use.Range, g.itemEffectText(key))
		g.button(dst, x+8, y+33+(i%rows)*29, w-16, 27, title, g.battleAction == itemAction && g.battleItem == key, func() { g.battleItem, g.battleAction = key, itemAction })
	}
	if len(items) == 0 {
		g.label(dst, x+8, y+40, "使えるアイテムはありません", muted)
	}
	g.pageControls(dst, x+8, y+295, len(items), rows, &g.battleItemPage)
}

// drawBattleLog は直近の戦闘記録を表示する（DES-UI-08、DES-UI-10）。
func (g *Game) drawBattleLog(dst *ebiten.Image, log []string, x, y, w, h int) {
	g.box(dst, x, y, w, h, "戦闘の記録")
	const rows = 8
	if g.battleLogPage > pageLimit(len(log), rows) {
		g.battleLogPage = pageLimit(len(log), rows)
	}
	start := g.battleLogPage * rows
	for i, entry := range log[start:] {
		if i >= rows {
			break
		}
		g.label(dst, x+8, y+33+i*23, entry, white)
	}
	g.pageControls(dst, x+8, y+h-30, len(log), rows, &g.battleLogPage)
}

// drawBattleResult は前回の勝敗、報酬または負傷、記録を採取地に表示する（DES-UI-10）。
func (g *Game) drawBattleResult(dst *ebiten.Image) {
	report := g.world.LastBattle
	if report == nil {
		return
	}
	const x, y, w = 800, 95, 790
	g.box(dst, x, y, w, 825, "探索の結果")
	site, _ := g.catalog.Gather(report.GatherID)
	if report.Won {
		g.label(dst, x+8, y+35, site.Name+": 勝利", white)
		parts := make([]string, 0, len(report.Rewards))
		for _, stack := range report.Rewards {
			parts = append(parts, fmt.Sprintf("%s Q%d ×%d", g.itemName(stack.Key.ItemID), stack.Key.Quality, stack.Count))
		}
		if len(parts) == 0 {
			parts = append(parts, "なし")
		}
		g.label(dst, x+8, y+60, "報酬: "+strings.Join(parts, "、"), white)
	} else {
		g.label(dst, x+8, y+35, site.Name+": 敗北", white)
		remaining := uint64(0)
		if g.world.Now < report.InjuredUntil {
			remaining = (report.InjuredUntil - g.world.Now + 59) / 60
		}
		g.label(dst, x+8, y+60, fmt.Sprintf("負傷: 残り %d 秒", remaining), white)
	}
	g.drawBattleLog(dst, report.Log, x+8, y+105, w-16, 285)
}
