package core

import (
	"errors"
	"fmt"
	"sort"
)

// ErrNoQuest は指定した依頼板の位置が空のときの誤り（DES-QST-02）。
var ErrNoQuest = errors.New("依頼がありません")

// drawQuest は表示中でない依頼を ID 順の候補から抽選する（DES-QST-01）。
func (w *World) drawQuest() string {
	var candidates []string
	for id := range w.catalog.Quests {
		shown := false
		for _, current := range w.Quests {
			if current == id {
				shown = true
				break
			}
		}
		if !shown {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.Strings(candidates)
	return candidates[w.questRNG.IntN(len(candidates))]
}

// Deliver は条件を満たす在庫を低品質から納品し、報酬と次の依頼を得る（DES-QST-01、DES-QST-02、DES-WLD-01）。
func (w *World) Deliver(index int) error {
	if index < 0 || index >= len(w.Quests) {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrUnknownID)
	}
	id := w.Quests[index]
	if id == "" {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrNoQuest)
	}
	quest := w.catalog.Quests[id]
	trial := cloneInventory(w.Inventory)
	remaining := quest.Count
	for _, stack := range trial.Stacks() {
		if stack.Key.ItemID != quest.Item || stack.Key.Quality < quest.MinQuality {
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
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrNotEnoughItems)
	}
	max := int(^uint(0) >> 1)
	if quest.Reward < 0 || w.Money > max-quest.Reward {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrInvalidCount)
	}
	w.Inventory = trial
	w.Money += quest.Reward
	w.Quests[index] = ""
	w.Quests[index] = w.drawQuest()
	return nil
}
