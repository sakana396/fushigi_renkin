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

const battleEventFrames = 30

// startBattleEvents は操作後に増えた出来事の先頭から再生する（DES-UI-12）。
func (g *Game) startBattleEvents(start int) {
	var events []core.BattleEvent
	if g.world.Battle != nil {
		events = g.world.Battle.Events
	} else if g.world.LastBattle != nil {
		events = g.world.LastBattle.Events
	}
	if start >= len(events) {
		if g.world.Battle == nil {
			g.endBattleEvents()
		}
		return
	}
	g.battleEvents = events
	g.battleEventIndex = start
	g.battleEventFrames = 0
	g.battleLogPage = pageLimit(start, 8)
}

// advanceBattleEvents は時間倍率に関係なく 30 更新ごとに次の出来事を見せる（DES-UI-12）。
func (g *Game) advanceBattleEvents() {
	g.battleEventFrames++
	if g.battleEventFrames < battleEventFrames {
		return
	}
	g.battleEventFrames = 0
	g.battleEventIndex++
	if g.battleEventIndex == len(g.battleEvents) {
		g.endBattleEvents()
		return
	}
	g.battleLogPage = pageLimit(g.battleEventIndex, 8)
}

// endBattleEvents は再生または飛ばしを終え、勝敗を採取地に示す（DES-UI-12、DES-UI-10）。
func (g *Game) endBattleEvents() {
	g.battleEvents = nil
	g.battleEventIndex = 0
	g.battleEventFrames = 0
	if g.world.Battle == nil && g.world.LastBattle != nil {
		g.battleLogPage = pageLimit(len(g.world.LastBattle.Log), 8)
		if g.world.LastBattle.Won {
			g.message = "勝利しました"
		} else {
			g.message = "敗北しました"
		}
		g.battleItem = core.ItemKey{}
	} else if g.world.Battle != nil {
		g.battleLogPage = pageLimit(len(g.world.Battle.Log), 8)
	}
}

// drawBattle は盤面、能力、操作、使用品と記録を表示する（DES-UI-08、DES-UI-09）。
func (g *Game) drawBattle(dst *ebiten.Image) {
	if len(g.battleEvents) > 0 {
		g.drawBattleEvent(dst)
		return
	}
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
				// 攻撃・アイテムで狙える相手のいるマスは、移動先と見分けやすい明るい色で塗る（DES-UI-08）。
				if g.battleAction != moveAction && g.battleUnitAt(x, y) {
					ink = targetColor
				}
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
		inset := 5
		if g.battleReachable(actor, def.Move, def.Range, u.X, u.Y) {
			inset = 11
		}
		vector.DrawFilledRect(dst, float32(px+inset), float32(py+inset), float32(cellSize-3-2*inset), float32(cellSize-3-2*inset), ink, false)
		initial, _ := utf8.DecodeRuneInString(u.Label)
		g.label(dst, px+12, py+12, fmt.Sprintf("%c HP%d", initial, u.HP), white)
		g.label(dst, px+8, py+40, u.Label, white)
	}
	g.box(dst, 715, 100, 875, 195, "現在の仲間")
	g.label(dst, 725, 135, fmt.Sprintf("%s  HP %d/%d  攻撃 %d+%d  素早さ %d+%d", actor.Label, actor.HP, def.HP, def.Attack, actor.AttackBonus, def.Speed, actor.SpeedBonus), white)
	moved := "まだ"
	if b.Moved {
		moved = "済み"
	}
	g.label(dst, 725, 160, fmt.Sprintf("移動力 %d / 攻撃射程 %d / 移動: %s", def.Move, def.Range, moved), muted)
	g.button(dst, 725, 205, 110, 32, "移動", g.battleAction == moveAction, func() { g.battleAction = moveAction })
	var attack func()
	if len(g.world.AttackTargets()) > 0 {
		attack = func() { g.battleAction = attackAction }
	}
	g.button(dst, 845, 205, 110, 32, "攻撃", g.battleAction == attackAction, attack)
	g.button(dst, 965, 205, 110, 32, "アイテム", g.battleAction == itemAction, func() { g.battleAction = itemAction })
	g.button(dst, 1085, 205, 110, 32, "待機", false, func() {
		start := len(g.world.Battle.Events)
		g.finishBattleAction("待機しました", g.world.BattleWait(), true, start)
	})
	guide := "移動先のマスを押してください"
	switch g.battleAction {
	case attackAction:
		guide = "攻撃する敵のマスを押してください"
	case itemAction:
		guide = "アイテムを使うマスを押してください"
	}
	g.label(dst, 725, 255, guide, muted)
	if len(g.world.AttackTargets()) == 0 {
		if b.Moved {
			g.label(dst, 725, 274, "届く敵がいません。待機で番を終えます", muted)
		} else {
			g.label(dst, 725, 274, "届く敵がいません。移動するか、待機で番を終えます", muted)
		}
	}
	if g.battleError != "" {
		g.label(dst, 725, 291, "操作失敗: "+g.battleError, battleErrorColor)
	}
	g.drawBattleItems(dst)
	g.drawBattleLog(dst, b.Log, 715, 645, 875, 275)
}

// drawBattleEvent は出来事直後の盤面、行動マスと表示済みの記録を描く（DES-UI-12）。
func (g *Game) drawBattleEvent(dst *ebiten.Image) {
	e := g.battleEvents[g.battleEventIndex]
	var gatherID string
	if g.world.Battle != nil {
		gatherID = g.world.Battle.GatherID
	} else if g.world.LastBattle != nil {
		gatherID = g.world.LastBattle.GatherID
	}
	site, _ := g.catalog.Gather(gatherID)
	g.label(dst, battleX, 105, "戦闘: "+site.Name, white)
	for y := 0; y < g.catalog.Game.BattleHeight; y++ {
		for x := 0; x < g.catalog.Game.BattleWidth; x++ {
			px, py := battleX+x*cellSize, battleY+y*cellSize
			ink := panelColor
			if e.Kind != "start" && e.Kind != "win" && e.Kind != "lose" && e.Kind != "reward_lost" && ((x == e.FromX && y == e.FromY && e.Actor >= 0) || (e.Kind != "fall" && x == e.ToX && y == e.ToY && (e.Target >= 0 || e.Kind == "move" || e.Kind == "use")) || (e.Kind == "fall" && e.Target >= 0 && e.Target < len(e.Units) && x == e.Units[e.Target].X && y == e.Units[e.Target].Y)) {
				ink = eventCellColor
			}
			vector.DrawFilledRect(dst, float32(px), float32(py), cellSize-3, cellSize-3, ink, false)
		}
	}
	for i, u := range e.Units {
		if u.HP <= 0 {
			continue
		}
		px, py := battleX+u.X*cellSize, battleY+u.Y*cellSize
		ink := color.RGBA{71, 110, 152, 255}
		if !u.Ally {
			ink = color.RGBA{145, 78, 73, 255}
		}
		if i == e.Actor || i == e.Target {
			ink = eventUnitColor
		}
		vector.DrawFilledRect(dst, float32(px+8), float32(py+8), cellSize-19, cellSize-19, ink, false)
		initial, _ := utf8.DecodeRuneInString(u.Label)
		g.label(dst, px+12, py+12, fmt.Sprintf("%c HP%d", initial, u.HP), white)
		g.label(dst, px+8, py+40, u.Label, white)
	}
	g.box(dst, 715, 100, 875, 195, "戦闘の出来事")
	line := e.Text
	if e.Kind == "start" {
		line = "戦闘開始：最初の配置"
	}
	g.label(dst, 725, 135, line, white)
	g.label(dst, 725, 165, fmt.Sprintf("出来事 %d/%d　クリックで残りを飛ばす", g.battleEventIndex+1, len(g.battleEvents)), muted)
	log := make([]string, 0, g.battleEventIndex)
	for _, shown := range g.battleEvents[:g.battleEventIndex+1] {
		if shown.Kind != "start" {
			log = append(log, shown.Text)
		}
	}
	g.drawBattleLog(dst, log, 715, 645, 875, 275)
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
	start := len(g.world.Battle.Events)
	var err error
	success := ""
	turnEnded := false
	switch g.battleAction {
	case moveAction:
		err, success = g.world.BattleMove(x, y), "移動しました"
	case attackAction:
		err, success = g.world.BattleAttack(x, y), "攻撃しました"
		turnEnded = true
	case itemAction:
		if g.battleItem.ItemID == "" {
			g.message = "操作失敗: アイテムを選んでください"
			g.battleError = "アイテムを選んでください"
			return
		}
		err, success = g.world.BattleUseItem(g.battleItem, x, y), g.itemName(g.battleItem.ItemID)+"を使いました"
		turnEnded = true
	}
	g.finishBattleAction(success, err, turnEnded, start)
}

// finishBattleAction は成功で増えた出来事を再生し、失敗理由を戦闘画面に示す（DES-UI-12、DES-UI-13）。
func (g *Game) finishBattleAction(success string, err error, turnEnded bool, start int) {
	g.result(success, err)
	if err != nil {
		g.battleError = err.Error()
		return
	}
	g.battleError = ""
	if err == nil {
		if g.world.Battle != nil {
			g.battleLogPage = pageLimit(len(g.world.Battle.Log), 8)
			if turnEnded {
				g.battleAction = moveAction
				g.battleItem = core.ItemKey{}
			} else if g.battleItem.ItemID != "" && g.world.Inventory.Count(g.battleItem) == 0 {
				g.battleItem = core.ItemKey{}
			}
		}
	}
	g.startBattleEvents(start)
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
		var selectItem func()
		if len(g.world.UseTargets(key)) > 0 {
			selectItem = func() { g.battleItem, g.battleAction = key, itemAction }
		} else {
			title += "　狙える相手がいません"
		}
		g.button(dst, x+8, y+33+(i%rows)*29, w-16, 27, title, g.battleAction == itemAction && g.battleItem == key, selectItem)
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
		if remaining == 0 {
			g.label(dst, x+8, y+60, "負傷: 治った", white)
		} else {
			g.label(dst, x+8, y+60, fmt.Sprintf("負傷: 残り %d 秒", remaining), white)
		}
	}
	g.drawBattleLog(dst, report.Log, x+8, y+105, w-16, 285)
}

// targetColor は攻撃・アイテムで狙えるマスの色（DES-UI-08）。
var targetColor = color.RGBA{214, 178, 64, 255}

var (
	battleErrorColor = color.RGBA{255, 110, 105, 255}
	eventCellColor   = color.RGBA{214, 178, 64, 255}
	eventUnitColor   = color.RGBA{235, 181, 53, 255}
)

// battleUnitAt は生きているユニットがそのマスにいるかを返す（DES-UI-08）。
func (g *Game) battleUnitAt(x, y int) bool {
	for _, u := range g.world.Battle.Units {
		if u.HP > 0 && u.X == x && u.Y == y {
			return true
		}
	}
	return false
}
