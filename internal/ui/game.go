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
	world         *core.World
	catalog       defs.Catalog
	face          *text.GoXFace
	buttons       []button
	message       string
	recipeID      string
	selection     core.Selection
	craftSlot     int
	selectedStack core.ItemKey
	hasStack      bool
	inventoryPage int
	recipePage    int
	selectedGolem int
	slotItems     []string
	quantity      int
	timeScale     int
}

// NewGame は本体の目録と World を画面に結び付ける（DES-PKG-05）。
func NewGame(world *core.World, catalog defs.Catalog) *Game {
	return &Game{world: world, catalog: catalog, face: text.NewGoXFace(bitmapfont.Face), selectedGolem: -1, quantity: 1, timeScale: 1, message: "採取地をクリックして素材を集めてください"}
}

// Run は画面を開き、ゲームを起動する（DES-PKG-05）。
func Run(world *core.World, catalog defs.Catalog) error {
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("ふしぎ錬金 - プロトタイプ")
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
				b.action()
				break
			}
		}
	}
	for i := 0; i < g.timeScale; i++ {
		g.world.Tick()
	}
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

// Draw は在庫・採取地・店・レシピ・ゴーレム・施設と操作結果を描く（DES-PKG-05、DES-IT01 4 章）。
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bg)
	g.buttons = g.buttons[:0]
	g.label(screen, 14, 10, fmt.Sprintf("ふしぎ錬金   お金: %d   経過: %d 秒   ゴーレム: %d/%d   時間 ×%d", g.world.Money, g.world.Now/60, len(g.world.Golems), g.world.GolemSlots(), g.timeScale), white)
	g.label(screen, 1100, 10, "F12: チート（お金と素材）  F11: 時間 ×10", muted)
	g.label(screen, 800, 10, fmt.Sprintf("売買数: %d", g.quantity), white)
	g.button(screen, 920, 5, 35, 24, "－", false, func() {
		if g.quantity > 1 {
			g.quantity--
		}
	})
	g.button(screen, 960, 5, 35, 24, "＋", false, func() {
		if g.quantity < 99 {
			g.quantity++
		}
	})
	g.label(screen, 14, 35, g.message, white)
	g.drawInventory(screen)
	g.drawShop(screen)
	g.drawGather(screen)
	g.drawFacilities(screen)
	g.drawRecipes(screen)
	g.drawCraft(screen)
	g.drawGolems(screen)
	g.drawAssignment(screen)
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
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), ink, false)
	g.label(dst, x+5, y+5, title, white)
	g.buttons = append(g.buttons, button{x, y, w, h, action})
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

// keys は目録の ID を一定の順で一覧化する（DES-PKG-05）。
func keys[T any](values map[string]T) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// drawInventory は品質別の在庫、素材選択、売値と売却操作を表示する（DES-PKG-05、UC-03、UC-04）。
func (g *Game) drawInventory(dst *ebiten.Image) {
	const x, y, w = 10, 70, 385
	g.box(dst, x, y, w, 855, "在庫（クリックで錬金枠へ 1 個追加）")
	stacks := g.world.Inventory.Stacks()
	const rows = 27
	maxPage := 0
	if len(stacks) > 0 {
		maxPage = (len(stacks) - 1) / rows
	}
	if g.inventoryPage > maxPage {
		g.inventoryPage = maxPage
	}
	for i := g.inventoryPage * rows; i < len(stacks) && i < (g.inventoryPage+1)*rows; i++ {
		stack := stacks[i]
		yy := y + 33 + (i%rows)*26
		title := fmt.Sprintf("%s Q%d ×%d", g.itemName(stack.Key.ItemID), stack.Key.Quality, stack.Count)
		g.button(dst, x+8, yy, w-16, 24, title, g.hasStack && stack.Key == g.selectedStack, func() {
			g.selectedStack = stack.Key
			g.hasStack = true
			g.addCraftItem(stack.Key)
		})
	}
	g.button(dst, x+8, y+747, 45, 25, "前", false, func() {
		if g.inventoryPage > 0 {
			g.inventoryPage--
		}
	})
	g.label(dst, x+62, y+752, fmt.Sprintf("%d / %d", g.inventoryPage+1, maxPage+1), muted)
	g.button(dst, x+139, y+747, 45, 25, "次", false, func() {
		if g.inventoryPage < maxPage {
			g.inventoryPage++
		}
	})
	if g.hasStack {
		g.label(dst, x+8, y+788, fmt.Sprintf("選択: %s Q%d / 売値 %d×%d", g.itemName(g.selectedStack.ItemID), g.selectedStack.Quality, g.world.SellPrice(g.selectedStack), g.quantity), white)
		g.button(dst, x+8, y+814, 140, 25, "指定数を売る", false, func() {
			money, err := g.world.Sell(g.selectedStack, g.quantity)
			g.result(fmt.Sprintf("%d 円で売却", money), err)
		})
	}
}

// addCraftItem は選択中の入力枠へ品質別の在庫を 1 個追加する（DES-ALC-01、DES-PKG-05）。
func (g *Game) addCraftItem(key core.ItemKey) {
	recipe, ok := g.catalog.Recipe(g.recipeID)
	if !ok || g.craftSlot < 0 || g.craftSlot >= len(recipe.Inputs) {
		return
	}
	slot := recipe.Inputs[g.craftSlot]
	item, _ := g.catalog.Item(key.ItemID)
	if slot.Item != "" && slot.Item != key.ItemID || slot.Attribute != "" && !contains(item.Attributes, slot.Attribute) {
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
	const x, y, w = 405, 70, 385
	g.box(dst, x, y, w, 175, "店（クリックで指定数購入）")
	for i, offer := range g.catalog.Game.Shop {
		index := i
		title := fmt.Sprintf("%s Q%d   %d 円×%d = %d 円", g.itemName(offer.Item), offer.Quality, offer.Price, g.quantity, offer.Price*g.quantity)
		g.button(dst, x+8, y+33+i*30, w-16, 27, title, false, func() { g.result("購入しました", g.world.Buy(index, g.quantity)) })
	}
}

// drawGather は採取地と再使用までの時間を表示する（DES-PKG-05、UC-01）。
func (g *Game) drawGather(dst *ebiten.Image) {
	const x, y, w = 405, 255, 385
	g.box(dst, x, y, w, 155, "採取地")
	for i, id := range keys(g.catalog.Gathers) {
		site := g.catalog.Gathers[id]
		gatherID := id
		g.button(dst, x+8, y+33+i*34, w-16, 30, fmt.Sprintf("%s （%d 回抽選）", site.Name, site.Rolls), false, func() {
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
	const x, y, w = 405, 420, 385
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

// selectRecipe は錬金と割り当てに使うレシピを選ぶ（DES-PKG-05、UC-03、UC-06）。
func (g *Game) selectRecipe(id string) {
	recipe := g.catalog.Recipes[id]
	g.recipeID = id
	g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
	g.slotItems = make([]string, len(recipe.Inputs))
	for i, slot := range recipe.Inputs {
		if slot.Item != "" {
			g.slotItems[i] = slot.Item
		} else {
			choices := g.choices(slot)
			if len(choices) > 0 {
				g.slotItems[i] = choices[0]
			}
		}
	}
	g.craftSlot = 0
	g.message = "レシピを選択しました。入力枠を選び、在庫をクリックしてください"
}

// drawRecipes はすべてのレシピへページ送りでアクセスできる一覧を描く（DES-PKG-05、UC-03）。
func (g *Game) drawRecipes(dst *ebiten.Image) {
	const x, y, w = 800, 70, 385
	g.box(dst, x, y, w, 385, "レシピ（クリックで選択）")
	ids := keys(g.catalog.Recipes)
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
		r := g.catalog.Recipes[id]
		g.button(dst, x+8, y+32+(i%rows)*25, w-16, 23, fmt.Sprintf("%s → %s ×%d", id, g.itemName(r.Output), r.OutputCount), id == g.recipeID, func() { g.selectRecipe(id) })
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
		g.button(dst, x+8, yy, w-55, 24, fmt.Sprintf("枠%d: %s ×%d", i+1, condition, slot.Count), i == g.craftSlot, func() { g.craftSlot = index })
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
	g.button(dst, x+8, y+407, 160, 32, "この素材で錬金", false, func() {
		output, err := g.world.Craft(g.recipeID, g.selection)
		if err != nil {
			g.result("", err)
			return
		}
		g.message = fmt.Sprintf("完成: %s Q%d ×%d", g.itemName(output.Key.ItemID), output.Key.Quality, output.Count)
		g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
		g.craftSlot = 0
	})
	g.button(dst, x+179, y+407, 120, 32, "選択を取消", false, func() {
		g.selection = core.Selection{Slots: make([][]core.Stack, len(recipe.Inputs))}
		g.craftSlot = 0
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

// drawGolems は所有数、材料、状態、進捗を表示する（DES-PKG-05、UC-05、UC-08）。
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
	for i, golem := range g.world.Golems {
		if i >= 5 {
			break
		}
		index := i
		yy := y + 100 + i*34
		state := "未割り当て"
		if golem.State == core.GolemWaiting {
			state = "素材待ち"
		} else if golem.State == core.GolemWorking {
			state = fmt.Sprintf("作業中 %d tick", golem.Progress)
		}
		recipe := ""
		if golem.Assignment != nil {
			recipe = " / " + golem.Assignment.RecipeID
		}
		g.button(dst, x+8, yy, w-16, 30, fmt.Sprintf("%d: %s%s", i+1, state, recipe), i == g.selectedGolem, func() { g.selectedGolem = index })
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

// drawAssignment はレシピと枠ごとの種類をゴーレムに割り当てる（DES-GOL-03、DES-PKG-05、UC-06）。
func (g *Game) drawAssignment(dst *ebiten.Image) {
	const x, y, w = 1195, 375, 395
	g.box(dst, x, y, w, 550, "ゴーレムの割り当て")
	if g.selectedGolem < 0 || g.selectedGolem >= len(g.world.Golems) {
		g.label(dst, x+8, y+36, "上のゴーレムを選んでください", muted)
		return
	}
	g.button(dst, x+188, y+425, 170, 32, "割り当てを外す", false, func() { g.result("割り当てを外しました", g.world.UnassignGolem(g.selectedGolem)) })
	if g.recipeID == "" {
		g.label(dst, x+8, y+36, "レシピを選んでください", muted)
		return
	}
	recipe := g.catalog.Recipes[g.recipeID]
	g.label(dst, x+8, y+36, fmt.Sprintf("%d 体目 → %s", g.selectedGolem+1, g.recipeID), white)
	g.label(dst, x+8, y+59, "属性枠をクリックして種類を切り替え", muted)
	for i, slot := range recipe.Inputs {
		index := i
		yy := y + 92 + i*53
		g.label(dst, x+8, yy, fmt.Sprintf("枠%d %s ×%d", i+1, g.slotName(slot), slot.Count), muted)
		g.button(dst, x+8, yy+21, w-16, 26, "使用: "+g.itemName(g.slotItems[i]), false, func() {
			options := g.choices(slot)
			for j, id := range options {
				if id == g.slotItems[index] {
					g.slotItems[index] = options[(j+1)%len(options)]
					break
				}
			}
		})
	}
	g.button(dst, x+8, y+425, 170, 32, "仕事を割り当てる", false, func() {
		assignment := core.Assignment{RecipeID: g.recipeID, SlotItems: append([]string(nil), g.slotItems...)}
		g.result("仕事を割り当てました", g.world.AssignGolem(g.selectedGolem, assignment))
	})
	g.label(dst, x+8, y+477, "時間は操作しなくても進みます", muted)
}
