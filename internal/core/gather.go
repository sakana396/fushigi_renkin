package core

import (
	"fmt"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// PickKey は拾う場所の地図とマスを識別する（DES-PCK-02）。
type PickKey struct {
	MapID string
	X, Y  int
}

// PickRemaining は拾う場所が戻るまでの tick を状態を変えずに返す（DES-PCK-02）。
func (w *World) PickRemaining(mapID string, x, y int) uint64 {
	m, ok := w.catalog.Map(mapID)
	if !ok || defs.TileInfo(m.Cell(x, y)).Kind != defs.CellPick {
		return 0
	}
	ready := w.pickReady[PickKey{mapID, x, y}]
	if ready <= w.Now {
		return 0
	}
	return ready - w.Now
}

// PickFacing は向いている場所で抽選し、再出現 tick を記録する（DES-PCK-01・02、DES-WLD-01）。
func (w *World) PickFacing() ([]Stack, error) {
	if w.Battle != nil {
		return nil, ErrInBattle
	}
	target, ok := w.FacingTarget()
	if !ok || target.Kind != TargetPick {
		return nil, ErrNothingToPick
	}
	if remaining := w.PickRemaining(w.Player.Scene, target.X, target.Y); remaining > 0 {
		return nil, fmt.Errorf("あと %d 秒で拾えます: %w", 1+(remaining-1)/60, ErrPickNotReady)
	}
	site, ok := w.catalog.Gather(w.Player.Scene)
	if !ok {
		return nil, fmt.Errorf("採取地 %s: %w", w.Player.Scene, ErrUnknownID)
	}
	duration := uint64(site.RespawnSeconds) * 60
	if site.RespawnSeconds < 1 || duration/60 != uint64(site.RespawnSeconds) || w.Now > ^uint64(0)-duration {
		return nil, ErrInvalidCount
	}
	stacks, err := w.gatherIntoInventory(site)
	if err != nil {
		return nil, err
	}
	if w.pickReady == nil {
		w.pickReady = make(map[PickKey]uint64)
	}
	w.pickReady[PickKey{w.Player.Scene, target.X, target.Y}] = w.Now + duration
	return stacks, nil
}

// gatherIntoInventory は共通の抽選を行い、失敗時は在庫と乱数を保つ（DES-GAT-01、DES-GAT-03、DES-WLD-01）。
func (w *World) gatherIntoInventory(site defs.GatherDef) ([]Stack, error) {
	if site.Rolls <= 0 || site.CooldownSeconds < 0 || len(site.Drops) == 0 {
		return nil, ErrInvalidCount
	}
	rngState, err := w.rngSource.MarshalBinary()
	if err != nil {
		return nil, err
	}
	stacks, err := w.drawDrops(site.Rolls, site.Drops)
	if err != nil {
		return nil, err
	}
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

// drawDrops は採取と戦闘報酬に共通の重み・品質抽選を行う（DES-GAT-01、DES-BTL-02）。
func (w *World) drawDrops(rolls int, drops []defs.Drop) ([]Stack, error) {
	if rolls <= 0 || len(drops) == 0 {
		return nil, ErrInvalidCount
	}
	total := 0
	max := int(^uint(0) >> 1)
	for _, drop := range drops {
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
	results := make(map[ItemKey]int)
	for i := 0; i < rolls; i++ {
		draw := w.rng.IntN(total)
		var selected int
		for j, drop := range drops {
			if draw < drop.Weight {
				selected = j
				break
			}
			draw -= drop.Weight
		}
		drop := drops[selected]
		quality := drop.QualityMin + w.rng.IntN(drop.QualityMax-drop.QualityMin+1)
		results[ItemKey{drop.Item, quality}]++
	}
	return gatherStacks(results), nil
}
