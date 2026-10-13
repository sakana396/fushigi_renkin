package core

import (
	"fmt"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// Gather は重み付きで素材を抽選し、品質を一様に決める（DES-GAT-01、DES-GAT-02、DES-WLD-01）。
func (w *World) Gather(gatherID string) ([]Stack, error) {
	site, ok := w.catalog.Gather(gatherID)
	if !ok {
		return nil, fmt.Errorf("採取地 %s: %w", gatherID, ErrUnknownID)
	}
	if ready, used := w.gatherReady[gatherID]; used && w.Now < ready {
		remaining := ready - w.Now
		return nil, fmt.Errorf("残り %d 秒: %w", 1+(remaining-1)/60, ErrGatherCooldown)
	}
	cooldown := uint64(site.CooldownSeconds) * 60
	if site.CooldownSeconds < 0 || cooldown/60 != uint64(site.CooldownSeconds) || w.Now > ^uint64(0)-cooldown {
		return nil, ErrInvalidCount
	}
	stacks, err := w.gatherIntoInventory(site)
	if err != nil {
		return nil, err
	}
	if w.gatherReady == nil {
		w.gatherReady = make(map[string]uint64)
	}
	w.gatherReady[gatherID] = w.Now + cooldown
	return stacks, nil
}

// gatherIntoInventory は共通の抽選を行い、失敗時は在庫と乱数を保つ（DES-GAT-01、DES-GAT-03、DES-WLD-01）。
func (w *World) gatherIntoInventory(site defs.GatherDef) ([]Stack, error) {
	if site.Rolls <= 0 || site.CooldownSeconds < 0 || len(site.Drops) == 0 {
		return nil, ErrInvalidCount
	}
	total := 0
	max := int(^uint(0) >> 1)
	for _, drop := range site.Drops {
		if _, ok := w.catalog.Item(drop.Item); !ok {
			return nil, fmt.Errorf("採取品 %s: %w", drop.Item, ErrUnknownID)
		}
		if drop.Weight <= 0 || total > max-drop.Weight {
			return nil, ErrInvalidCount
		}
		if drop.QualityMin < 1 || drop.QualityMax > 200 || drop.QualityMin > drop.QualityMax {
			return nil, ErrInvalidQuality
		}
		total += drop.Weight
	}
	rngState, err := w.rngSource.MarshalBinary()
	if err != nil {
		return nil, err
	}
	results := make(map[ItemKey]int)
	for i := 0; i < site.Rolls; i++ {
		draw := w.rng.IntN(total)
		var selected int
		for j, drop := range site.Drops {
			if draw < drop.Weight {
				selected = j
				break
			}
			draw -= drop.Weight
		}
		drop := site.Drops[selected]
		quality := drop.QualityMin + w.rng.IntN(drop.QualityMax-drop.QualityMin+1)
		results[ItemKey{drop.Item, quality}]++
	}
	stacks := gatherStacks(results)
	trial := cloneInventory(w.Inventory)
	for _, stack := range stacks {
		if err := trial.Add(stack.Key, stack.Count); err != nil {
			if restoreErr := w.rngSource.UnmarshalBinary(rngState); restoreErr != nil {
				return nil, restoreErr
			}
			return nil, err
		}
	}
	w.Inventory = trial
	return stacks, nil
}
