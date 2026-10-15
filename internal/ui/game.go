package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/hajimehoshi/bitmapfont/v3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const width, height = 1600, 940

type button struct {
	x, y, w, h int
	action     func()
}

// Game は一覧、選択中の入力、操作結果を保持する（DES-PKG-05、DES-IT01 4 章）。
type Game struct {
	world              *core.World
	catalog            defs.Catalog
	face               *text.GoXFace
	buttons            []button
	message            string
	recipeID           string
	selection          core.Selection
	craftSlot          int
	selectedStack      core.ItemKey
	hasStack           bool
	inventoryPage      int
	candidatePage      int
	recipePage         int
	gatherPage         int
	golemPage          int
	assignmentPage     int
	assignmentRecipeID string
	assignmentGatherID string
	selectedGolem      int
	slotItems          []string
	quantity           int
	timeScale          int
	scene              scene
	battleAction       battleAction
	battleItem         core.ItemKey
	battleItemPage     int
	battleLogPage      int
}

// scene は表示する場面を表す（DES-UI-06）。
type scene int

const (
	workshopScene scene = iota
	townScene
	gatherScene
)

// battleAction は盤面を押したときの操作を表す（DES-UI-08）。
type battleAction int

const (
	moveAction battleAction = iota
	attackAction
	itemAction
)

// NewGame は本体の目録と World を画面に結び付ける（DES-PKG-05）。
func NewGame(world *core.World, catalog defs.Catalog) *Game {
	return &Game{world: world, catalog: catalog, face: text.NewGoXFace(bitmapfont.Face), selectedGolem: -1, quantity: 1, timeScale: 1, message: "上のタブから場面を選んでください"}
}

// Run は画面を開き、ゲームを起動する（DES-PKG-05）。
func Run(world *core.World, catalog defs.Catalog) error {
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("ふしぎ錬金")
	ebiten.SetTPS(60)
	return ebiten.RunGame(NewGame(world, catalog))
}

// Update はクリックとチートのキーを処理し、1 回の更新につき時間の倍率の分だけ tick を進める（DES-PKG-05、DES-TIME-01）。
func (g *Game) Update() error {
	g.handleCheatKeys()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for i := len(g.buttons) - 1; i >= 0; i-- {
			b := g.buttons[i]
			if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
				if b.action != nil {
					b.action()
				}
				break
			}
		}
	}
	for i := 0; i < g.timeScale; i++ {
		g.world.Tick()
	}
	g.showRankUps()
	return nil
}

// Layout は固定サイズの 1 画面を使う（DES-PKG-05）。
func (g *Game) Layout(_, _ int) (int, int) { return width, height }

var (
	bg          = color.RGBA{22, 28, 37, 255}
	panelColor  = color.RGBA{37, 46, 58, 255}
	buttonColor = color.RGBA{59, 75, 91, 255}
	activeColor = color.RGBA{62, 103, 111, 255}
	white       = color.RGBA{232, 237, 238, 255}
	muted       = color.RGBA{177, 193, 201, 255}
)

// Draw は共通表示と選択した場面を描く（DES-UI-06、DES-UI-07）。
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bg)
	g.buttons = g.buttons[:0]
	g.label(screen, 14, 10, fmt.Sprintf("ふしぎ錬金   お金: %d   経過: %d 秒   ゴーレム: %d/%d   時間 ×%d", g.world.Money, g.world.Now/60, len(g.world.Golems), g.world.GolemSlots(), g.timeScale), white)
	g.label(screen, 1100, 10, "F12: チート（お金と素材）  F11: 時間 ×10", muted)
	for i, line := range strings.Split(g.message, "／") {
		prefix := ""
		if i == 0 {
			prefix = "操作結果: "
		}
		g.label(screen, 14, 35+i*14, prefix+line, white)
	}
	if g.scene == workshopScene && g.world.Battle == nil {
		g.label(screen, 700, 10, g.rankText(), white)
	}
	tabY := 55
	if strings.Contains(g.message, "／") {
		tabY = 65
	}
	for i, title := range []string{"工房", "街", "採取地"} {
		chosen := scene(i)
		var action func()
		if g.world.Battle == nil {
			action = func() { g.scene = chosen }
		}
		g.button(screen, 10+i*130, tabY, 120, 25, title, g.scene == chosen && g.world.Battle == nil, action)
	}
	if g.world.Battle != nil {
		g.drawBattle(screen)
		return
	}
	switch g.scene {
	case workshopScene:
		g.drawInventory(screen, false)
		g.drawCandidates(screen)
		g.drawRecipes(screen)
		g.drawCraft(screen)
		g.drawGolems(screen)
		g.drawAssignment(screen)
	case townScene:
		g.label(screen, 1195, 100, fmt.Sprintf("売買数: %d", g.quantity), white)
		g.button(screen, 1295, 95, 35, 24, "－", false, func() {
			if g.quantity > 1 {
				g.quantity--
			}
		})
		g.button(screen, 1335, 95, 35, 24, "＋", false, func() {
			if g.quantity < 99 {
				g.quantity++
			}
		})
		g.drawInventory(screen, true)
		g.drawShop(screen)
		g.drawFacilities(screen)
		g.drawQuests(screen)
	case gatherScene:
		g.drawGather(screen)
		g.drawBattleResult(screen)
	}
}

// drawQuests は依頼の条件・報酬と納品操作を表示する（DES-UI-07、DES-QST-02、UC-14）。
func (g *Game) drawQuests(dst *ebiten.Image) {
	const x, y, w = 1195, 130, 395
	g.box(dst, x, y, w, 440, "依頼板")
	g.label(dst, x+8, y+28, g.rankText(), white)
	for i, id := range g.world.Quests {
		yy := y + 52 + i*125
		if id == "" {
			g.label(dst, x+8, yy, fmt.Sprintf("依頼 %d: 空き", i+1), muted)
			continue
		}
		quest, ok := g.catalog.Quest(id)
		if !ok {
			g.label(dst, x+8, yy, fmt.Sprintf("依頼 %d: 内容を確認できません", i+1), muted)
			continue
		}
		g.label(dst, x+8, yy, fmt.Sprintf("依頼 %d: %s Q%d 以上 ×%d", i+1, g.itemName(quest.Item), quest.MinQuality, quest.Count), white)
		g.label(dst, x+8, yy+22, fmt.Sprintf("報酬: %d 円", quest.Reward), muted)
		eligible := 0
		for _, stack := range g.world.Inventory.Stacks() {
			if stack.Key.ItemID == quest.Item && stack.Key.Quality >= quest.MinQuality {
				eligible += stack.Count
			}
		}
		g.label(dst, x+8, yy+43, fmt.Sprintf("納品できる在庫: %d/%d", eligible, quest.Count), muted)
		index := i
		g.button(dst, x+8, yy+66, 150, 28, "納品する", false, func() {
			g.result("納品しました", g.world.Deliver(index))
			g.showRankUps()
		})
	}
}

// rankText は現在のランクと次のランクに必要な評判を表示する（DES-UI-11）。
func (g *Game) rankText() string {
	if next, ok := g.world.NextRankReputation(); ok {
		return fmt.Sprintf("ランク %d / 評判 %d / %d", g.world.Rank(), g.world.Reputation, next)
	}
	return fmt.Sprintf("ランク %d（最高）/ 評判 %d", g.world.Rank(), g.world.Reputation)
}

// showRankUps は未読の上昇と解放された定義の表示名を操作結果に並べる（DES-UI-11）。
func (g *Game) showRankUps() {
	var notices []string
	for _, rank := range g.world.TakeRankUps() {
		unlocked := g.world.Unlocked(rank)
		var parts []string
		var names []string
		for _, id := range unlocked.Recipes {
			names = append(names, g.itemName(g.catalog.Recipes[id].Output))
		}
		if len(names) > 0 {
			parts = append(parts, "作れるようになった物: "+strings.Join(names, "、"))
		}
		names = nil
		for _, id := range unlocked.Gathers {
			names = append(names, g.catalog.Gathers[id].Name)
		}
		if len(names) > 0 {
			parts = append(parts, "行けるようになった採取地: "+strings.Join(names, "、"))
		}
		names = nil
		for _, id := range unlocked.Quests {
			quest := g.catalog.Quests[id]
			names = append(names, g.itemName(quest.Item))
		}
		if len(names) > 0 {
			parts = append(parts, "新しい依頼が出るようになった: "+strings.Join(names, "、"))
		}
		notice := fmt.Sprintf("工房ランク %d", rank)
		if len(parts) > 0 {
			notice += ": " + strings.Join(parts, "・")
		}
		notices = append(notices, notice)
	}
	if len(notices) > 0 {
		g.message = strings.Join(notices, "／")
	}
}

// availableRecipes は現在のランクで解放済みのレシピを ID 順に返す（DES-UI-11）。
func (g *Game) availableRecipes() []string {
	var ids []string
	for _, id := range keys(g.catalog.Recipes) {
		if g.catalog.Recipes[id].Rank <= g.world.Rank() {
			ids = append(ids, id)
		}
	}
	return ids
}

// availableGathers は現在のランクで解放済みの採取地を ID 順に返す（DES-UI-11）。
func (g *Game) availableGathers() []string {
	var ids []string
	for _, id := range keys(g.catalog.Gathers) {
		if g.catalog.Gathers[id].Rank <= g.world.Rank() {
			ids = append(ids, id)
		}
	}
	return ids
}

// label は日本語グリフを含む bitmapfont を text/v2 で描く（DES-PKG-05）。
func (g *Game) label(dst *ebiten.Image, x, y int, value string, ink color.RGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(dst, value, g.face, op)
}

// box は見出し付きの四角を描く（DES-PKG-05）。
func (g *Game) box(dst *ebiten.Image, x, y, w, h int, title string) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), panelColor, false)
	g.label(dst, x+8, y+7, title, white)
}

// button は表示した四角とクリック処理を対応付ける（DES-PKG-05）。
func (g *Game) button(dst *ebiten.Image, x, y, w, h int, title string, selected bool, action func()) {
	ink := buttonColor
	if selected {
		ink = activeColor
	}
	if action == nil {
		ink = panelColor
	}
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), ink, false)
	g.label(dst, x+5, y+5, title, white)
	if action != nil {
		g.buttons = append(g.buttons, button{x, y, w, h, action})
	}
}

// result は状態変更の成功または失敗を画面に示す（DES-PKG-05、DES-WLD-01）。
func (g *Game) result(success string, err error) {
	if err != nil {
		g.message = "操作失敗: " + err.Error()
	} else {
		g.message = success
	}
}

// itemName は目録にある種類の表示名を返す（DES-PKG-05）。
func (g *Game) itemName(id string) string {
	if item, ok := g.catalog.Item(id); ok {
		return item.Name
	}
	return id
}

// recipeName は製品名と入力枠で同じ製品の別レシピを区別する（DES-UI-05）。
func (g *Game) recipeName(id string) string {
	recipe, ok := g.catalog.Recipe(id)
	if !ok {
		return id
	}
	parts := make([]string, 0, len(recipe.Inputs))
	for _, slot := range recipe.Inputs {
		parts = append(parts, fmt.Sprintf("%s×%d", g.slotName(slot), slot.Count))
	}
	return fmt.Sprintf("%s（%s）", g.itemName(recipe.Output), strings.Join(parts, "、"))
}

// pageLimit は一覧の最終ページを求める（DES-UI-05）。
func pageLimit(count, rows int) int {
	if count == 0 {
		return 0
	}
	return (count - 1) / rows
}

// pageControls は 1 始まりのページ表示と前後の操作を描く（DES-UI-05）。
func (g *Game) pageControls(dst *ebiten.Image, x, y, count, rows int, page *int) {
	maxPage := pageLimit(count, rows)
	if *page > maxPage {
		*page = maxPage
	}
	g.button(dst, x, y, 45, 25, "前", false, func() {
		if *page > 0 {
			*page--
		}
	})
	g.label(dst, x+54, y+5, fmt.Sprintf("%d/%d", *page+1, maxPage+1), muted)
	g.button(dst, x+108, y, 45, 25, "次", false, func() {
		if *page < maxPage {
			*page++
		}
	})
}

// itemAttributes は定義データの属性を表示し、空なら属性なしとする（DES-ITM-01、DES-UI-02）。
func (g *Game) itemAttributes(id string) string {
	item, ok := g.catalog.Item(id)
	if !ok || len(item.Attributes) == 0 {
		return "属性なし"
	}
	return strings.Join(item.Attributes, "、")
}

// itemEffectText は品質を反映した使用効果を表示用に組み立てる（DES-UI-09、DES-QLT-05）。
func (g *Game) itemEffectText(key core.ItemKey) string {
	item, ok := g.catalog.Item(key.ItemID)
	if !ok {
		return ""
	}
	effect, usable := core.UseEffectOf(item, key.Quality)
	if !usable {
		return ""
	}
	parts := []string{fmt.Sprintf("範囲 %d", effect.Radius)}
	for _, value := range effect.Effects {
		name := map[string]string{"heal": "回復 ", "damage": "ダメージ ", "attack_up": "攻撃 +", "speed_up": "素早さ +"}[value.Kind]
		parts = append(parts, fmt.Sprintf("%s%d", name, value.Value))
	}
	return "使用効果: " + strings.Join(parts, "・")
}

// candidateRemaining は全入力枠の選択済み数を在庫数から差し引く（DES-UI-01）。
func (g *Game) candidateRemaining(stack core.Stack) int {
	remaining := stack.Count
	for _, picked := range g.selection.Slots {
		for _, used := range picked {
			if used.Key == stack.Key {
				remaining -= used.Count
			}
		}
	}
	if remaining < 0 {
		return 0
	}
	return remaining
}

// keys は目録の ID を一定の順で一覧化する（DES-PKG-05）。
func keys[T any](values map[string]T) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// drawInventory は品質別の在庫と使用効果を示し、街では売却も扱う（DES-UI-02、DES-UI-03、DES-UI-09）。
func (g *Game) drawInventory(dst *ebiten.Image, canSell bool) {
	const x, y, w = 10, 95, 775
	title := "在庫（クリックで売る物を選択）"
	if !canSell {
		title = "在庫"
	}
	g.box(dst, x, y, w, 555, title)
	stacks := g.world.Inventory.Stacks()
	const rows = 10
	maxPage := 0
	if len(stacks) > 0 {
		maxPage = (len(stacks) - 1) / rows
	}
	if g.inventoryPage > maxPage {
		g.inventoryPage = maxPage
	}
	for i := g.inventoryPage * rows; i < len(stacks) && i < (g.inventoryPage+1)*rows; i++ {
		stack := stacks[i]
		yy := y + 33 + (i%rows)*43
		title := fmt.Sprintf("%s Q%d ×%d  %s", g.itemName(stack.Key.ItemID), stack.Key.Quality, stack.Count, g.itemAttributes(stack.Key.ItemID))
		var selectForSale func()
		if canSell {
			selectForSale = func() {
				g.selectedStack = stack.Key
				g.hasStack = true
			}
		}
		g.button(dst, x+8, yy, w-16, 24, title, canSell && g.hasStack && stack.Key == g.selectedStack, selectForSale)
		if effect := g.itemEffectText(stack.Key); effect != "" {
			g.label(dst, x+13, yy+25, effect, muted)
		}
	}
	g.button(dst, x+8, y+480, 45, 25, "前", false, func() {
		if g.inventoryPage > 0 {
			g.inventoryPage--
		}
	})
	g.label(dst, x+62, y+485, fmt.Sprintf("%d / %d", g.inventoryPage+1, maxPage+1), muted)
	g.button(dst, x+139, y+480, 45, 25, "次", false, func() {
		if g.inventoryPage < maxPage {
			g.inventoryPage++
		}
	})
	if canSell && g.hasStack {
		g.label(dst, x+8, y+510, fmt.Sprintf("選択: %s Q%d / 売値 %d×%d", g.itemName(g.selectedStack.ItemID), g.selectedStack.Quality, g.world.SellPrice(g.selectedStack), g.quantity), white)
		g.button(dst, x+8, y+529, 140, 22, "指定数を売る", false, func() {
			money, err := g.world.Sell(g.selectedStack, g.quantity)
			g.result(fmt.Sprintf("%d 円で売却", money), err)
		})
	}
}

// drawCandidates は選択した入力枠に合う在庫の残数を表示する（DES-ALC-04、DES-UI-01、DES-UI-02）。
func (g *Game) drawCandidates(dst *ebiten.Image) {
	const x, y, w = 10, 655, 385
	g.box(dst, x, y, w, 270, "錬金の素材候補")
	recipe, ok := g.catalog.Recipe(g.recipeID)
	if !ok || g.craftSlot < 0 || g.craftSlot >= len(recipe.Inputs) {
		g.label(dst, x+8, y+36, "レシピと入力枠を選んでください", muted)
		return
	}
	candidates, err := g.world.SlotCandidates(g.recipeID, g.craftSlot)
	if err != nil {
		g.label(dst, x+8, y+36, err.Error(), muted)
		return
	}
	if len(candidates) == 0 {
		g.label(dst, x+8, y+36, "合う素材がありません", muted)
		return
	}
	available := 0
	for _, stack := range candidates {
		if g.candidateRemaining(stack) > 0 {
			available++
		}
	}
	pickedCount := 0
	for _, stack := range g.selection.Slots[g.craftSlot] {
		pickedCount += stack.Count
	}
	if available == 0 && pickedCount < recipe.Inputs[g.craftSlot].Count {
		g.label(dst, x+205, y+7, "合う素材がありません", muted)
	}
	const rows = 8
	maxPage := (len(candidates) - 1) / rows
	if g.candidatePage > maxPage {
		g.candidatePage = maxPage
	}
	for i := g.candidatePage * rows; i < len(candidates) && i < (g.candidatePage+1)*rows; i++ {
		stack := candidates[i]
		remaining := g.candidateRemaining(stack)
		title := fmt.Sprintf("%s Q%d ×%d  %s", g.itemName(stack.Key.ItemID), stack.Key.Quality, remaining, g.itemAttributes(stack.Key.ItemID))
		var action func()
		if remaining > 0 {
			action = func() { g.addCraftItem(stack.Key) }
		}
		g.button(dst, x+8, y+32+(i%rows)*26, w-16, 24, title, false, action)
	}
	g.button(dst, x+8, y+242, 45, 25, "前", false, func() {
		if g.candidatePage > 0 {
			g.candidatePage--
		}
	})
	g.label(dst, x+62, y+247, fmt.Sprintf("%d / %d", g.candidatePage+1, maxPage+1), muted)
	g.button(dst, x+139, y+242, 45, 25, "次", false, func() {
		if g.candidatePage < maxPage {
			g.candidatePage++
		}
	})
}

// addCraftItem は候補から選択中の入力枠へ在庫を 1 個追加する（DES-ALC-01、DES-UI-01）。
func (g *Game) addCraftItem(key core.ItemKey) {
	recipe, ok := g.catalog.Recipe(g.recipeID)
	if !ok || g.craftSlot < 0 || g.craftSlot >= len(recipe.Inputs) {
		return
	}
	slot := recipe.Inputs[g.craftSlot]
	candidates, err := g.world.SlotCandidates(g.recipeID, g.craftSlot)
	if err != nil {
		g.message = err.Error()
		return
	}
	matched := false
	for _, candidate := range candidates {
		if candidate.Key == key {
			matched = true
			break
		}
	}
	if !matched {
		g.message = "入力枠の条件に合わない素材です"
		return
	}
	total := 0
	used := 0
	for _, s := range g.selection.Slots[g.craftSlot] {
		total += s.Count
	}
	for _, stacks := range g.selection.Slots {
		for _, s := range stacks {
			if s.Key == key {
				used += s.Count
			}
		}
	}
	if total >= slot.Count {
		g.message = "この入力枠は満杯です"
		return
	}
	if used >= g.world.Inventory.Count(key) {
		g.message = "選択できる在庫が足りません"
		return
	}
	g.selection.Slots[g.craftSlot] = append(g.selection.Slots[g.craftSlot], core.Stack{Key: key, Count: 1})
	g.message = fmt.Sprintf("%s Q%d を入力枠 %d に追加", g.itemName(key.ItemID), key.Quality, g.craftSlot+1)
	if total+1 == slot.Count && g.craftSlot+1 < len(recipe.Inputs) {
		g.craftSlot++
		g.candidatePage = 0
	}
}

func contains(values []string, wanted string) bool {
	for _, v := range values {
		if v == wanted {
			return true
		}
	}
	return false
}

// drawShop は固定品質・価格の売り物を表示し、指定数を購入する（DES-PKG-05、UC-02）。
func (g *Game) drawShop(dst *ebiten.Image) {
	const x, y, w = 800, 95, 385
	g.box(dst, x, y, w, 175, "店（クリックで指定数購入）")
	for i, offer := range g.catalog.Game.Shop {
		index := i
		title := fmt.Sprintf("%s Q%d   %d 円×%d = %d 円", g.itemName(offer.Item), offer.Quality, offer.Price, g.quantity, offer.Price*g.quantity)
		g.button(dst, x+8, y+33+i*30, w-16, 27, title, false, func() { g.result("購入しました", g.world.Buy(index, g.quantity)) })
	}
}

// drawGather はプレイヤーの採取地をページ送りで表示する（DES-PKG-05、DES-UI-05、UC-01）。
func (g *Game) drawGather(dst *ebiten.Image) {
	const x, y, w = 10, 95, 775
	g.box(dst, x, y, w, 420, "採取地")
	ids := g.availableGathers()
	const rows = 6
	g.pageControls(dst, x+8, y+382, len(ids), rows, &g.gatherPage)
	for i := g.gatherPage * rows; i < len(ids) && i < (g.gatherPage+1)*rows; i++ {
		id := ids[i]
		site := g.catalog.Gathers[id]
		gatherID := id
		yy := y + 34 + (i%rows)*58
		g.label(dst, x+8, yy, site.Name, white)
		g.button(dst, x+250, yy-4, 170, 28, "採取", false, func() {
			stacks, err := g.world.Gather(gatherID)
			if err != nil {
				g.result("", err)
				return
			}
			names := make([]string, 0, len(stacks))
			for _, s := range stacks {
				names = append(names, fmt.Sprintf("%s Q%d ×%d", g.itemName(s.Key.ItemID), s.Key.Quality, s.Count))
			}
			g.message = "採取: " + strings.Join(names, "、")
		})
		if site.Battle != nil {
			var action func()
			reason := ""
			if g.world.Now < g.world.InjuredUntil {
				reason = fmt.Sprintf("負傷中: 残り %d 秒", (g.world.InjuredUntil-g.world.Now+59)/60)
			} else {
				action = func() {
					err := g.world.StartBattle(gatherID)
					g.result("探索を始めました", err)
					if err == nil {
						g.battleAction = moveAction
						g.battleItemPage = 0
						g.battleLogPage = 0
					}
				}
			}
			g.button(dst, x+430, yy-4, 170, 28, "探索する", false, action)
			if reason != "" {
				g.label(dst, x+605, yy+3, reason, muted)
			}
		}
	}
}

func materials(values []defs.ItemAmount, name func(string) string) string {
	if len(values) == 0 {
		return "なし"
	}
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%s×%d", name(v.Item), v.Count))
	}
	return strings.Join(parts, "、")
}

// drawFacilities は次段階の費用・効果を表示して建設する（DES-PKG-05、UC-07）。
func (g *Game) drawFacilities(dst *ebiten.Image) {
	const x, y, w = 800, 280, 385
	g.box(dst, x, y, w, 505, "施設（クリックで次の段階を建設）")
	for i, id := range keys(g.catalog.Facilities) {
		facility := g.catalog.Facilities[id]
		level := g.world.Facilities[id]
		yy := y + 35 + i*175
		title := fmt.Sprintf("%s  段階 %d/%d", facility.Name, level, len(facility.Levels))
		facilityID := id
		g.button(dst, x+8, yy, w-16, 29, title, false, func() { g.result("施設を建設しました", g.world.BuildFacility(facilityID)) })
		if level < len(facility.Levels) {
			next := facility.Levels[level]
			g.label(dst, x+12, yy+34, fmt.Sprintf("費用: %d 円", next.Money), muted)
			g.label(dst, x+12, yy+55, "材料: "+materials(next.Materials, g.itemName), muted)
			for j, e := range next.Effects {
				effect := e.Kind
				if effect == "golem_slots" {
					effect = "ゴーレム枠"
				} else if effect == "sell_rate" {
					effect = "売値"
				}
				g.label(dst, x+12, yy+76+j*20, fmt.Sprintf("効果: %s +%d", effect, e.Value), muted)
			}
		} else {
			g.label(dst, x+12, yy+37, "最大段階", muted)
		}
	}
}

// selectRecipe は錬金だけに使うレシピを選ぶ（DES-PKG-05、DES-UI-04、UC-03）。
func (g *Game) selectRecipe(id string) {
	recipe := g.catalog.Recipes[id]
	g.recipeID = id
	g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
	g.craftSlot = 0
	g.candidatePage = 0
	g.message = "レシピを選択しました。入力枠を選び、素材候補をクリックしてください"
}

// drawRecipes は解放済みのレシピをページ送りで表示する（DES-PKG-05、DES-UI-11、UC-03）。
func (g *Game) drawRecipes(dst *ebiten.Image) {
	const x, y, w = 800, 70, 385
	g.box(dst, x, y, w, 385, "レシピ（クリックで選択）")
	ids := g.availableRecipes()
	const rows = 12
	maxPage := 0
	if len(ids) > 0 {
		maxPage = (len(ids) - 1) / rows
	}
	if g.recipePage > maxPage {
		g.recipePage = maxPage
	}
	for i := g.recipePage * rows; i < len(ids) && i < (g.recipePage+1)*rows; i++ {
		id := ids[i]
		g.button(dst, x+8, y+32+(i%rows)*25, w-16, 23, g.recipeName(id), id == g.recipeID, func() { g.selectRecipe(id) })
	}
	g.button(dst, x+8, y+347, 45, 27, "前", false, func() {
		if g.recipePage > 0 {
			g.recipePage--
		}
	})
	g.label(dst, x+60, y+354, fmt.Sprintf("%d/%d", g.recipePage+1, maxPage+1), muted)
	g.button(dst, x+110, y+347, 45, 27, "次", false, func() {
		if g.recipePage < maxPage {
			g.recipePage++
		}
	})
}

// drawCraft は入力枠と予想品質・成立ボーナスを表示する（DES-ALC-03、DES-PKG-05、UC-03）。
func (g *Game) drawCraft(dst *ebiten.Image) {
	const x, y, w = 800, 465, 385
	g.box(dst, x, y, w, 460, "手作り錬金")
	recipe, ok := g.catalog.Recipe(g.recipeID)
	if !ok {
		g.label(dst, x+8, y+40, "上のレシピを選んでください", muted)
		return
	}
	g.label(dst, x+8, y+34, fmt.Sprintf("%s ×%d / %d 秒", g.itemName(recipe.Output), recipe.OutputCount, recipe.Seconds), white)
	for i, slot := range recipe.Inputs {
		index := i
		yy := y + 62 + i*77
		condition := g.slotName(slot)
		g.button(dst, x+8, yy, w-55, 24, fmt.Sprintf("枠%d: %s ×%d", i+1, condition, slot.Count), i == g.craftSlot, func() {
			g.craftSlot = index
			g.candidatePage = 0
		})
		g.button(dst, x+w-43, yy, 35, 24, "戻", false, func() {
			stacks := g.selection.Slots[index]
			if len(stacks) > 0 {
				g.selection.Slots[index] = stacks[:len(stacks)-1]
			}
		})
		picked := make([]string, 0, len(g.selection.Slots[i]))
		for _, s := range g.selection.Slots[i] {
			picked = append(picked, fmt.Sprintf("%s Q%d", g.itemName(s.Key.ItemID), s.Key.Quality))
		}
		g.label(dst, x+13, yy+29, strings.Join(picked, "、"), muted)
	}
	preview, err := g.world.PreviewCraft(g.recipeID, g.selection)
	bottom := y + 315
	if err != nil {
		g.label(dst, x+8, bottom, "予想: "+err.Error(), muted)
	} else {
		g.label(dst, x+8, bottom, fmt.Sprintf("予想品質: %d （平均 %d）", preview.Quality, preview.AverageQuality), white)
		bonus := make([]string, 0, len(preview.Bonuses))
		for _, b := range preview.Bonuses {
			bonus = append(bonus, fmt.Sprintf("%s %d個で +%d", b.Attribute, b.MinCount, b.Add))
		}
		if len(bonus) == 0 {
			bonus = append(bonus, "なし")
		}
		g.label(dst, x+8, bottom+23, "成立ボーナス: "+strings.Join(bonus, "、"), white)
	}
	var craftAction func()
	if err == nil {
		craftAction = func() {
			output, err := g.world.Craft(g.recipeID, g.selection)
			if err != nil {
				g.result("", err)
				return
			}
			g.message = fmt.Sprintf("完成: %s Q%d ×%d", g.itemName(output.Key.ItemID), output.Key.Quality, output.Count)
			g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
			g.craftSlot = 0
			g.candidatePage = 0
		}
	}
	g.button(dst, x+8, y+407, 160, 32, "この素材で錬金", false, craftAction)
	g.button(dst, x+179, y+407, 120, 32, "選択を取消", false, func() {
		g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
		g.craftSlot = 0
		g.candidatePage = 0
		g.message = "素材の選択を取り消しました"
	})
}

// slotName は入力枠の条件を日本語で表す（DES-PKG-05）。
func (g *Game) slotName(slot defs.InputSlot) string {
	if slot.Item != "" {
		return g.itemName(slot.Item)
	}
	return "属性「" + slot.Attribute + "」"
}

// drawGolems は所有数、材料、割り当て先と進捗をページ送りで表示する（DES-PKG-05、DES-UI-04、DES-UI-05、UC-05、UC-08）。
func (g *Game) drawGolems(dst *ebiten.Image) {
	const x, y, w = 1195, 70, 395
	g.box(dst, x, y, w, 295, "ゴーレム")
	for i, id := range keys(g.catalog.Golems) {
		def := g.catalog.Golems[id]
		golemID := id
		yy := y + 33 + i*59
		g.label(dst, x+8, yy, "材料: "+materials(def.Materials, g.itemName), muted)
		g.button(dst, x+8, yy+23, w-16, 28, "作成: "+def.Name, false, func() { g.result("ゴーレムを作成しました", g.world.BuildGolem(golemID)) })
	}
	const rows = 4
	g.pageControls(dst, x+8, y+262, len(g.world.Golems), rows, &g.golemPage)
	for i := g.golemPage * rows; i < len(g.world.Golems) && i < (g.golemPage+1)*rows; i++ {
		golem := g.world.Golems[i]
		index := i
		yy := y + 100 + (i%rows)*38
		state := "未割り当て"
		if golem.State == core.GolemWaiting {
			state = "素材待ち"
		} else if golem.State == core.GolemWorking {
			state = "作業中"
		}
		destination := ""
		if golem.Assignment != nil {
			if golem.Assignment.GatherID != "" {
				destination = " / " + g.catalog.Gathers[golem.Assignment.GatherID].Name
			} else {
				destination = " / " + g.recipeName(golem.Assignment.RecipeID)
			}
		}
		duration, err := g.world.GolemDuration(i)
		if err != nil {
			g.message = "操作失敗: " + err.Error()
		}
		g.button(dst, x+8, yy, w-16, 35, fmt.Sprintf("%d: %s%s", i+1, state, destination), i == g.selectedGolem, func() { g.selectedGolem = index })
		g.label(dst, x+14, yy+17, fmt.Sprintf("%d / %d tick", golem.Progress, duration), muted)
	}
}

// choices は属性枠に使える種類を一定順に返す（DES-GOL-03、DES-PKG-05）。
func (g *Game) choices(slot defs.InputSlot) []string {
	if slot.Item != "" {
		return []string{slot.Item}
	}
	var out []string
	for _, id := range keys(g.catalog.Items) {
		if contains(g.catalog.Items[id].Attributes, slot.Attribute) {
			out = append(out, id)
		}
	}
	return out
}

// selectAssignmentRecipe は錬金の選択を変えずに割り当て用のレシピを選ぶ（DES-UI-04）。
func (g *Game) selectAssignmentRecipe(id string) {
	recipe := g.catalog.Recipes[id]
	g.assignmentRecipeID = id
	g.assignmentGatherID = ""
	g.slotItems = make([]string, len(recipe.Inputs))
	for i, slot := range recipe.Inputs {
		choices := g.choices(slot)
		if len(choices) > 0 {
			g.slotItems[i] = choices[0]
		}
	}
}

// drawAssignment は専用の一覧からレシピまたは採取地を割り当てる（DES-GOL-03、DES-UI-04、DES-UI-05、UC-06）。
func (g *Game) drawAssignment(dst *ebiten.Image) {
	const x, y, w = 1195, 375, 395
	g.box(dst, x, y, w, 550, "ゴーレムの割り当て")
	recipes := g.availableRecipes()
	gathers := g.availableGathers()
	const rows = 6
	g.pageControls(dst, x+8, y+214, len(recipes)+len(gathers), rows, &g.assignmentPage)
	for i := g.assignmentPage * rows; i < len(recipes)+len(gathers) && i < (g.assignmentPage+1)*rows; i++ {
		yy := y + 33 + (i%rows)*29
		if i < len(recipes) {
			id := recipes[i]
			g.button(dst, x+8, yy, w-16, 27, "錬金: "+g.recipeName(id), id == g.assignmentRecipeID, func() { g.selectAssignmentRecipe(id) })
		} else {
			id := gathers[i-len(recipes)]
			g.button(dst, x+8, yy, w-16, 27, "採取: "+g.catalog.Gathers[id].Name, id == g.assignmentGatherID, func() {
				g.assignmentGatherID = id
				g.assignmentRecipeID = ""
				g.slotItems = nil
			})
		}
	}
	if g.selectedGolem < 0 || g.selectedGolem >= len(g.world.Golems) {
		g.label(dst, x+8, y+263, "上のゴーレムを選んでください", muted)
		return
	}
	g.button(dst, x+188, y+425, 170, 32, "割り当てを外す", false, func() { g.result("割り当てを外しました", g.world.UnassignGolem(g.selectedGolem)) })
	if g.assignmentRecipeID == "" && g.assignmentGatherID == "" {
		g.label(dst, x+8, y+263, "割り当て先を選んでください", muted)
		return
	}
	if g.assignmentGatherID != "" {
		g.label(dst, x+8, y+263, fmt.Sprintf("%d 体目 → %s", g.selectedGolem+1, g.catalog.Gathers[g.assignmentGatherID].Name), white)
	} else {
		recipe := g.catalog.Recipes[g.assignmentRecipeID]
		g.label(dst, x+8, y+252, fmt.Sprintf("%d 体目 → %s", g.selectedGolem+1, g.recipeName(g.assignmentRecipeID)), white)
		g.label(dst, x+8, y+272, "属性枠をクリックして種類を切り替え", muted)
		for i, slot := range recipe.Inputs {
			index := i
			yy := y + 296 + i*40
			g.label(dst, x+8, yy, fmt.Sprintf("枠%d %s ×%d", i+1, g.slotName(slot), slot.Count), muted)
			g.button(dst, x+8, yy+16, w-16, 22, "使用: "+g.itemName(g.slotItems[i])+"  "+g.itemAttributes(g.slotItems[i]), false, func() {
				options := g.choices(slot)
				for j, id := range options {
					if id == g.slotItems[index] {
						g.slotItems[index] = options[(j+1)%len(options)]
						break
					}
				}
			})
		}
	}
	g.button(dst, x+8, y+425, 170, 32, "仕事を割り当てる", false, func() {
		assignment := core.Assignment{RecipeID: g.assignmentRecipeID, GatherID: g.assignmentGatherID, SlotItems: append([]string(nil), g.slotItems...)}
		g.result("仕事を割り当てました", g.world.AssignGolem(g.selectedGolem, assignment))
	})
	g.label(dst, x+8, y+477, "時間は操作しなくても進みます", muted)
}
