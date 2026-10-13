package core

import (
	"fmt"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// GolemState はゴーレムの作業状態を表す（DES-GOL-01）。
type GolemState uint8

const (
	GolemUnassigned GolemState = iota
	GolemWaiting
	GolemWorking
)

// Assignment はレシピと各入力枠で使うアイテムの種類を表す（DES-GOL-03）。
type Assignment struct {
	RecipeID  string
	SlotItems []string
}

// Golem は作業状態と取り出し済みの素材を保持する（DES-GOL-01）。
type Golem struct {
	DefID      string
	State      GolemState
	Assignment *Assignment
	Progress   int
	Holding    []Stack
}

// BuildGolem は品質を問わず材料を低品質から消費してゴーレムを作る（DES-GOL-06、DES-WLD-01）。
func (w *World) BuildGolem(golemID string) error {
	def, ok := w.catalog.Golem(golemID)
	if !ok {
		return fmt.Errorf("ゴーレム %q: %w", golemID, ErrUnknownID)
	}
	if len(w.Golems) >= w.GolemSlots() {
		return ErrNoGolemSlots
	}
	trial := cloneInventory(w.Inventory)
	for _, material := range def.Materials {
		if material.Count <= 0 {
			return ErrInvalidCount
		}
		if _, ok := w.catalog.Item(material.Item); !ok {
			return fmt.Errorf("材料 %q: %w", material.Item, ErrUnknownID)
		}
		if _, err := takeLowest(&trial, material.Item, material.Count); err != nil {
			return err
		}
	}
	w.Inventory = trial
	w.Golems = append(w.Golems, Golem{DefID: golemID, State: GolemUnassigned})
	return nil
}

// AssignGolem は入力枠に指定した種類を検査して割り当てる（DES-GOL-01、DES-GOL-03、DES-WLD-01）。
func (w *World) AssignGolem(index int, assignment Assignment) error {
	if index < 0 || index >= len(w.Golems) {
		return fmt.Errorf("ゴーレム番号 %d: %w", index, ErrUnknownID)
	}
	recipe, ok := w.catalog.Recipe(assignment.RecipeID)
	if !ok {
		return fmt.Errorf("レシピ %q: %w", assignment.RecipeID, ErrUnknownID)
	}
	if len(assignment.SlotItems) != len(recipe.Inputs) {
		return fmt.Errorf("入力枠の数: %w", ErrInvalidCount)
	}
	for i, slot := range recipe.Inputs {
		id := assignment.SlotItems[i]
		item, ok := w.catalog.Item(id)
		if !ok {
			return fmt.Errorf("入力枠 %d の種類 %q: %w", i, id, ErrUnknownID)
		}
		if slot.Count <= 0 {
			return fmt.Errorf("入力枠 %d の個数: %w", i, ErrInvalidCount)
		}
		if slot.Item != "" && id != slot.Item || slot.Attribute != "" && !hasAttribute(item, slot.Attribute) {
			return fmt.Errorf("入力枠 %d の種類 %q: %w", i, id, ErrInvalidSelection)
		}
	}
	selected := Assignment{RecipeID: assignment.RecipeID, SlotItems: append([]string(nil), assignment.SlotItems...)}
	w.Golems[index].Assignment = &selected
	w.Golems[index].State = GolemWaiting
	w.Golems[index].Progress = 0
	w.Golems[index].Holding = nil
	return nil
}

// UnassignGolem は作業を中止し、取り出し済みの素材を戻さない（DES-GOL-01）。
func (w *World) UnassignGolem(index int) error {
	if index < 0 || index >= len(w.Golems) {
		return fmt.Errorf("ゴーレム番号 %d: %w", index, ErrUnknownID)
	}
	w.Golems[index].Assignment = nil
	w.Golems[index].State = GolemUnassigned
	w.Golems[index].Progress = 0
	w.Golems[index].Holding = nil
	return nil
}

// takeLowest は指定種類を品質の低い順に仮の在庫から取り出す（DES-GOL-04、DES-GOL-06）。
func takeLowest(inventory *Inventory, itemID string, count int) ([]Stack, error) {
	if count <= 0 {
		return nil, ErrInvalidCount
	}
	remaining := count
	var taken []Stack
	for _, stack := range inventory.Stacks() {
		if stack.Key.ItemID != itemID {
			continue
		}
		n := stack.Count
		if n > remaining {
			n = remaining
		}
		if err := inventory.Remove(stack.Key, n); err != nil {
			return nil, err
		}
		taken = append(taken, Stack{Key: stack.Key, Count: n})
		remaining -= n
		if remaining == 0 {
			return taken, nil
		}
	}
	return nil, fmt.Errorf("材料 %s: %w", itemID, ErrNotEnoughItems)
}

// golemDuration は秒数と速度から切り上げた所要 tick を求める（DES-GOL-02）。
func golemDuration(seconds, speed int) int {
	if seconds <= 0 || speed <= 0 {
		return 1
	}
	max := int(^uint(0) >> 1)
	if seconds > max/6000 {
		return max
	}
	numerator := seconds * 6000
	return 1 + (numerator-1)/speed
}

// stepGolem は順番に 1 体の状態を 1 tick 分進める（DES-GOL-01〜05、DES-QLT-04）。
func (w *World) stepGolem(index int) {
	g := &w.Golems[index]
	if g.Assignment == nil {
		return
	}
	recipe, ok := w.catalog.Recipe(g.Assignment.RecipeID)
	if !ok {
		return
	}
	switch g.State {
	case GolemWaiting:
		w.startGolem(g, recipe)
	case GolemWorking:
		duration := golemDuration(recipe.Seconds, w.golemSpeed(g.DefID))
		if g.Progress < duration {
			g.Progress++
		}
		if g.Progress < duration {
			return
		}
		quality, err := AverageQuality(g.Holding)
		if err != nil {
			return
		}
		if recipe.OutputCount <= 0 {
			return
		}
		if err := w.Inventory.Add(ItemKey{ItemID: recipe.Output, Quality: quality}, recipe.OutputCount); err != nil {
			return
		}
		g.State = GolemWaiting
		g.Progress = 0
		g.Holding = nil
		// 素材待ちに入った tick のうちに入力がそろっていれば、同じ tick で作業中に移る（DES-GOL-01）。
		w.startGolem(g, recipe)
	}
}

// startGolem は入力がそろっていれば在庫から取り出して作業中にする（DES-GOL-01、DES-GOL-03、DES-GOL-04）。
func (w *World) startGolem(g *Golem, recipe defs.RecipeDef) {
	if len(g.Assignment.SlotItems) != len(recipe.Inputs) {
		return
	}
	trial := cloneInventory(w.Inventory)
	var holding []Stack
	for i, slot := range recipe.Inputs {
		stacks, err := takeLowest(&trial, g.Assignment.SlotItems[i], slot.Count)
		if err != nil {
			return
		}
		holding = append(holding, stacks...)
	}
	w.Inventory = trial
	g.Holding = holding
	g.Progress = 0
	g.State = GolemWorking
}

// golemSpeed は定義された作業速度を返す（DES-GOL-02）。
func (w *World) golemSpeed(id string) int {
	def, ok := w.catalog.Golem(id)
	if !ok {
		return 100
	}
	return def.SpeedPercent
}
