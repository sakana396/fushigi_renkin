package core

import (
	"fmt"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// Selection は入力枠ごとに使用する在庫と個数を保持する（DES-ALC-01）。
type Selection struct{ Slots [][]Stack }

// CraftPreview は手作りの予想品質と成立したボーナスを表す（DES-ALC-03、DES-QLT-02）。
type CraftPreview struct {
	AverageQuality int
	Quality        int
	Bonuses        []defs.Bonus
}

type craftPlan struct {
	output   Stack
	consumed map[ItemKey]int
	preview  CraftPreview
}

// PreviewCraft は入力を検査して品質を計算し、状態を変更しない（DES-ALC-03）。
func (w *World) PreviewCraft(recipeID string, selection Selection) (CraftPreview, error) {
	plan, err := w.planCraft(recipeID, selection)
	if err != nil {
		return CraftPreview{}, err
	}
	return plan.preview, nil
}

// Craft は入力の検査後に素材を消費し、手作りの製品を追加する（DES-ALC-02、DES-QLT-02）。
func (w *World) Craft(recipeID string, selection Selection) (Stack, error) {
	plan, err := w.planCraft(recipeID, selection)
	if err != nil {
		return Stack{}, err
	}
	for key, count := range plan.consumed {
		// planCraft が全数を確認済みなので、この削除は失敗しない。
		if err := w.Inventory.Remove(key, count); err != nil {
			return Stack{}, err
		}
	}
	// 出力の個数と品質も planCraft で検査済み。
	if err := w.Inventory.Add(plan.output.Key, plan.output.Count); err != nil {
		return Stack{}, err
	}
	return plan.output, nil
}

// planCraft は全入力枠と在庫を検査し、共通の平均品質と手作りボーナスを計算する（DES-ALC-02、DES-QLT-01〜03）。
func (w *World) planCraft(recipeID string, selection Selection) (craftPlan, error) {
	recipe, ok := w.catalog.Recipe(recipeID)
	if !ok {
		return craftPlan{}, fmt.Errorf("レシピ %q: %w", recipeID, ErrUnknownID)
	}
	if w.locked(recipe.Rank) {
		return craftPlan{}, lockedError(w.itemName(recipe.Output), recipe.Rank)
	}
	if len(selection.Slots) != len(recipe.Inputs) {
		return craftPlan{}, fmt.Errorf("入力枠の数: %w", ErrInvalidCount)
	}
	consumed := make(map[ItemKey]int)
	selected := make([]Stack, 0)
	maxInt := int(^uint(0) >> 1)
	for idx, slot := range recipe.Inputs {
		total := 0
		for _, stack := range selection.Slots[idx] {
			if stack.Count <= 0 || stack.Count > maxInt-total {
				return craftPlan{}, fmt.Errorf("入力枠 %d の個数: %w", idx+1, ErrInvalidCount)
			}
			item, found := w.catalog.Item(stack.Key.ItemID)
			if !found {
				return craftPlan{}, fmt.Errorf("アイテム %q: %w", stack.Key.ItemID, ErrUnknownID)
			}
			if stack.Key.Quality < 1 || stack.Key.Quality > 200 {
				return craftPlan{}, fmt.Errorf("入力枠 %d の品質: %w", idx+1, ErrInvalidQuality)
			}
			if !slotAcceptsItem(slot, stack.Key.ItemID, item) {
				return craftPlan{}, fmt.Errorf("入力枠 %d に %q は使えません: %w", idx+1, w.itemName(stack.Key.ItemID), ErrInvalidSelection)
			}
			if stack.Count > maxInt-consumed[stack.Key] {
				return craftPlan{}, fmt.Errorf("入力枠 %d の合計個数: %w", idx+1, ErrInvalidCount)
			}
			consumed[stack.Key] += stack.Count
			total += stack.Count
			selected = append(selected, stack)
		}
		if total != slot.Count {
			return craftPlan{}, fmt.Errorf("入力枠 %d の個数 %d (必要数 %d): %w", idx+1, total, slot.Count, ErrInvalidCount)
		}
	}
	for key, count := range consumed {
		if w.Inventory.Count(key) < count {
			return craftPlan{}, fmt.Errorf("%s 品質 %d を %d 個: %w", w.itemName(key.ItemID), key.Quality, count, ErrNotEnoughItems)
		}
	}
	average, err := AverageQuality(selected)
	if err != nil {
		return craftPlan{}, err
	}
	quality := average
	var bonuses []defs.Bonus
	for _, bonus := range recipe.Bonuses {
		count := 0
		for _, stack := range selected {
			item, _ := w.catalog.Item(stack.Key.ItemID)
			if hasAttribute(item, bonus.Attribute) {
				count += stack.Count
			}
		}
		if count >= bonus.MinCount {
			bonuses = append(bonuses, bonus)
			if bonus.Add > 0 && quality > maxInt-bonus.Add {
				quality = maxInt
			} else {
				quality += bonus.Add
			}
		}
	}
	output := Stack{Key: ItemKey{ItemID: recipe.Output, Quality: ClampQuality(quality)}, Count: recipe.OutputCount}
	if output.Count <= 0 {
		return craftPlan{}, fmt.Errorf("出力個数: %w", ErrInvalidCount)
	}
	// 出力と同じ在庫を入力にも使うレシピを考慮して、消費後の個数で桁あふれを調べる。
	remaining := w.Inventory.Count(output.Key) - consumed[output.Key]
	if output.Count > maxInt-remaining {
		return craftPlan{}, fmt.Errorf("出力後の個数: %w", ErrInvalidCount)
	}
	return craftPlan{output: output, consumed: consumed, preview: CraftPreview{AverageQuality: average, Quality: output.Key.Quality, Bonuses: bonuses}}, nil
}

// SlotCandidates は指定した入力枠に合う在庫を種類・品質順に返す（DES-ALC-04）。
func (w *World) SlotCandidates(recipeID string, slotIndex int) ([]Stack, error) {
	recipe, ok := w.catalog.Recipe(recipeID)
	if !ok {
		return nil, fmt.Errorf("レシピ %q: %w", recipeID, ErrUnknownID)
	}
	if slotIndex < 0 {
		return nil, fmt.Errorf("入力枠の番号: %w", ErrUnknownID)
	}
	if slotIndex >= len(recipe.Inputs) {
		return nil, fmt.Errorf("入力枠 %d: %w", slotIndex+1, ErrUnknownID)
	}
	var candidates []Stack
	for _, stack := range w.Inventory.Stacks() {
		item, ok := w.catalog.Item(stack.Key.ItemID)
		if ok && slotAcceptsItem(recipe.Inputs[slotIndex], stack.Key.ItemID, item) {
			candidates = append(candidates, stack)
		}
	}
	return candidates, nil
}

// slotAcceptsItem は入力枠のアイテムまたは属性の条件を判定する（DES-ALC-04）。
func slotAcceptsItem(slot defs.InputSlot, itemID string, item defs.ItemDef) bool {
	return (slot.Item == "" || itemID == slot.Item) && (slot.Attribute == "" || hasAttribute(item, slot.Attribute))
}

// hasAttribute は定義された属性をアイテムが持つか調べる（DES-DEF-02、DES-QLT-03）。
func hasAttribute(item defs.ItemDef, attribute string) bool {
	for _, a := range item.Attributes {
		if a == attribute {
			return true
		}
	}
	return false
}
