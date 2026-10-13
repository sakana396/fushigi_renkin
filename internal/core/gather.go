package core

import "fmt"

// Gather は重み付きで素材を抽選し、品質を一様に決める（DES-GAT-01、DES-GAT-02、DES-WLD-01）。
func (w *World) Gather(gatherID string) ([]Stack, error) {
	site, ok := w.catalog.Gather(gatherID)
	if !ok {
		return nil, fmt.Errorf("採取地 %s: %w", gatherID, ErrUnknownID)
	}
	if ready, used := w.gatherReady[gatherID]; used && w.Now < ready {
		return nil, ErrGatherCooldown
	}
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
	cooldown := uint64(site.CooldownSeconds) * 60
	if cooldown/60 != uint64(site.CooldownSeconds) || w.Now > ^uint64(0)-cooldown {
		return nil, ErrInvalidCount
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
	if w.gatherReady == nil {
		w.gatherReady = make(map[string]uint64)
	}
	w.gatherReady[gatherID] = w.Now + cooldown
	return stacks, nil
}
