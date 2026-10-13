package core

import (
	"fmt"
	"sort"
)

// Buy は店の固定品質・固定価格の商品を買う（DES-ECO-02、DES-WLD-01）。
func (w *World) Buy(offerIndex, count int) error {
	if offerIndex < 0 || offerIndex >= len(w.catalog.Game.Shop) {
		return fmt.Errorf("商品番号 %d: %w", offerIndex, ErrUnknownID)
	}
	if count <= 0 {
		return fmt.Errorf("購入数 %d: %w", count, ErrInvalidCount)
	}
	offer := w.catalog.Game.Shop[offerIndex]
	if _, ok := w.catalog.Item(offer.Item); !ok {
		return fmt.Errorf("商品 %s: %w", offer.Item, ErrUnknownID)
	}
	if offer.Price < 0 || offer.Price > int(^uint(0)>>1)/count {
		return ErrInvalidCount
	}
	cost := offer.Price * count
	if cost > w.Money {
		return ErrNotEnoughMoney
	}
	key := ItemKey{offer.Item, offer.Quality}
	if err := w.Inventory.Add(key, count); err != nil {
		return err
	}
	w.Money -= cost
	return nil
}

// SellPrice は品質と施設効果から 1 個の売値を各段階で切り捨てて求める（DES-ECO-01）。
// 未知の種類・範囲外の品質では 0 を返す。
func (w *World) SellPrice(key ItemKey) int {
	item, ok := w.catalog.Item(key.ItemID)
	if !ok || key.Quality < 1 || key.Quality > 200 || item.BasePrice < 0 {
		return 0
	}
	rate := w.effectTotal("sell_rate")
	if rate < -100 {
		return 0
	}
	max := int(^uint(0) >> 1)
	factor := 50 + key.Quality
	if item.BasePrice > max/factor {
		return 0
	}
	price := item.BasePrice * factor / 100
	factor = 100 + rate
	if factor < 0 || (factor > 0 && price > max/factor) {
		return 0
	}
	return price * factor / 100
}

// Sell は在庫の製品を売り、お金を加える（DES-ECO-01、DES-WLD-01）。
func (w *World) Sell(key ItemKey, count int) (int, error) {
	if count <= 0 {
		return 0, ErrInvalidCount
	}
	if _, ok := w.catalog.Item(key.ItemID); !ok {
		return 0, fmt.Errorf("アイテム %s: %w", key.ItemID, ErrUnknownID)
	}
	if key.Quality < 1 || key.Quality > 200 {
		return 0, ErrInvalidQuality
	}
	if w.Inventory.Count(key) < count {
		return 0, ErrNotEnoughItems
	}
	price := w.SellPrice(key)
	max := int(^uint(0) >> 1)
	if price < 0 || (price > 0 && count > max/price) {
		return 0, ErrInvalidCount
	}
	earned := price * count
	if earned > max-w.Money {
		return 0, ErrInvalidCount
	}
	if err := w.Inventory.Remove(key, count); err != nil {
		return 0, err
	}
	w.Money += earned
	return earned, nil
}

// BuildFacility は次段階の費用を払い、施設を強化する（DES-TWN-01、DES-WLD-01）。
func (w *World) BuildFacility(facilityID string) error {
	facility, ok := w.catalog.Facility(facilityID)
	if !ok {
		return fmt.Errorf("施設 %s: %w", facilityID, ErrUnknownID)
	}
	level := w.Facilities[facilityID]
	if level < 0 {
		return ErrInvalidCount
	}
	if level >= len(facility.Levels) {
		return ErrFacilityMaxLevel
	}
	next := facility.Levels[level]
	if next.Money < 0 {
		return ErrInvalidCount
	}
	if w.Money < next.Money {
		return ErrNotEnoughMoney
	}
	// 品質を問わない材料は、種類ごとに低品質から仮の在庫で消費する。
	trial := cloneInventory(w.Inventory)
	for _, material := range next.Materials {
		if material.Count <= 0 {
			return ErrInvalidCount
		}
		remaining := material.Count
		for _, stack := range trial.Stacks() {
			if stack.Key.ItemID != material.Item {
				continue
			}
			n := stack.Count
			if n > remaining {
				n = remaining
			}
			if n > 0 {
				if err := trial.Remove(stack.Key, n); err != nil {
					return err
				}
				remaining -= n
			}
			if remaining == 0 {
				break
			}
		}
		if remaining > 0 {
			return fmt.Errorf("施設材料 %s: %w", material.Item, ErrNotEnoughItems)
		}
	}
	w.Inventory = trial
	w.Money -= next.Money
	if w.Facilities == nil {
		w.Facilities = make(map[string]int)
	}
	w.Facilities[facilityID] = level + 1
	return nil
}

// GolemSlots は基本枠と施設効果を合計する（DES-TWN-02）。
func (w *World) GolemSlots() int { return w.catalog.Game.BaseGolemSlots + w.effectTotal("golem_slots") }

// effectTotal は建設済みの各段階の効果を合計する（DES-TWN-01）。
func (w *World) effectTotal(kind string) int {
	total := 0
	for id, level := range w.Facilities {
		facility, ok := w.catalog.Facility(id)
		if !ok {
			continue
		}
		if level > len(facility.Levels) {
			level = len(facility.Levels)
		}
		for i := 0; i < level; i++ {
			for _, effect := range facility.Levels[i].Effects {
				if effect.Kind == kind {
					total += effect.Value
				}
			}
		}
	}
	return total
}

func cloneInventory(src Inventory) Inventory {
	dst := Inventory{stacks: make(map[ItemKey]int, len(src.stacks))}
	for key, n := range src.stacks {
		dst.stacks[key] = n
	}
	return dst
}

// gatherStacks は抽選結果を在庫順に整える（DES-GAT-01）。
func gatherStacks(results map[ItemKey]int) []Stack {
	stacks := make([]Stack, 0, len(results))
	for key, n := range results {
		stacks = append(stacks, Stack{key, n})
	}
	sort.Slice(stacks, func(i, j int) bool {
		if stacks[i].Key.ItemID != stacks[j].Key.ItemID {
			return stacks[i].Key.ItemID < stacks[j].Key.ItemID
		}
		return stacks[i].Key.Quality < stacks[j].Key.Quality
	})
	return stacks
}
