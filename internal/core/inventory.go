package core

import (
	"fmt"
	"sort"
)

// ItemKey はアイテムの種類と品質を表す（DES-INV-01）。
type ItemKey struct {
	ItemID  string
	Quality int
}

// Stack は同じ種類・品質のアイテムと個数を表す（DES-INV-01）。
type Stack struct {
	Key   ItemKey
	Count int
}

// Inventory は種類と品質の組ごとに個数を保持する（DES-INV-01）。
type Inventory struct {
	stacks map[ItemKey]int
}

// Add は在庫を増やす。無効な個数・品質なら在庫を変えない（DES-INV-01、DES-WLD-01）。
func (i *Inventory) Add(key ItemKey, n int) error {
	if n <= 0 {
		return fmt.Errorf("追加する個数 %d: %w", n, ErrInvalidCount)
	}
	if key.Quality < 1 || key.Quality > 200 {
		return fmt.Errorf("品質 %d: %w", key.Quality, ErrInvalidQuality)
	}
	if n > int(^uint(0)>>1)-i.Count(key) {
		return fmt.Errorf("追加後の個数が範囲外: %w", ErrInvalidCount)
	}
	if i.stacks == nil {
		i.stacks = make(map[ItemKey]int)
	}
	i.stacks[key] += n
	return nil
}

// Remove は在庫を減らす。不足や無効な個数なら在庫を変えない（DES-INV-01、DES-WLD-01）。
func (i *Inventory) Remove(key ItemKey, n int) error {
	if n <= 0 {
		return fmt.Errorf("取り出す個数 %d: %w", n, ErrInvalidCount)
	}
	if i.Count(key) < n {
		return fmt.Errorf("%s 品質 %d を %d 個: %w", key.ItemID, key.Quality, n, ErrNotEnoughItems)
	}
	i.stacks[key] -= n
	if i.stacks[key] == 0 {
		delete(i.stacks, key)
	}
	return nil
}

// Count は指定した種類・品質の個数を返す（DES-INV-01）。
func (i *Inventory) Count(key ItemKey) int { return i.stacks[key] }

// Stacks は在庫単位を種類・品質順に返す（DES-INV-01）。
func (i *Inventory) Stacks() []Stack {
	if len(i.stacks) == 0 {
		return nil
	}
	stacks := make([]Stack, 0, len(i.stacks))
	for key, count := range i.stacks {
		stacks = append(stacks, Stack{key, count})
	}
	sort.Slice(stacks, func(a, b int) bool {
		if stacks[a].Key.ItemID != stacks[b].Key.ItemID {
			return stacks[a].Key.ItemID < stacks[b].Key.ItemID
		}
		return stacks[a].Key.Quality < stacks[b].Key.Quality
	})
	return stacks
}
