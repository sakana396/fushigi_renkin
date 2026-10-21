package core

import (
	"errors"
	"fmt"
	"sort"
)

// ErrNoQuest は指定した候補または受けた依頼がないときの誤り（DES-QST-05）。
var ErrNoQuest = errors.New("依頼がありません")

// MaxAcceptedQuests は同時に受けられる依頼の上限（DES-QST-04）。
const MaxAcceptedQuests = 3

// ErrQuestLimit は受けた依頼が上限に達したときの誤り（DES-QST-05）。
var ErrQuestLimit = errors.New("受けられる依頼は 3 件までです")

// drawQuest は候補にも受けた依頼にもない依頼を ID 順から抽選する（DES-QST-04）。
func (w *World) drawQuest() string {
	var candidates []string
	for id, quest := range w.catalog.Quests {
		if w.locked(quest.Rank) {
			continue
		}
		shown := false
		for _, current := range w.Quests {
			if current == id {
				shown = true
				break
			}
		}
		for _, accepted := range w.Accepted {
			if accepted == id {
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

// AcceptQuest は候補を受けた依頼の末尾へ移す（DES-QST-05、DES-WLD-01）。
func (w *World) AcceptQuest(index int) error {
	if index < 0 || index >= len(w.Quests) || w.Quests[index] == "" {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrNoQuest)
	}
	if len(w.Accepted) >= MaxAcceptedQuests {
		return ErrQuestLimit
	}
	w.Accepted = append(w.Accepted, w.Quests[index])
	w.Quests[index] = ""
	return nil
}

// AbandonQuest は受けた依頼をやめ、後ろの依頼を前に詰める（DES-QST-05、DES-WLD-01）。
func (w *World) AbandonQuest(index int) error {
	if index < 0 || index >= len(w.Accepted) {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrNoQuest)
	}
	w.Accepted = append(w.Accepted[:index], w.Accepted[index+1:]...)
	return nil
}

// DeliverQuest は受けた依頼へ低品質から納品し、報酬と評判を得る（DES-QST-02・03・05、DES-WLD-01）。
func (w *World) DeliverQuest(index int) error {
	if index < 0 || index >= len(w.Accepted) {
		return fmt.Errorf("依頼番号 %d: %w", index+1, ErrNoQuest)
	}
	id := w.Accepted[index]
	quest, ok := w.catalog.Quest(id)
	if !ok {
		return fmt.Errorf("依頼 ID %s: %w", id, ErrUnknownID)
	}
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
	if quest.Reputation < 0 || w.Reputation > max-quest.Reputation {
		return fmt.Errorf("依頼番号 %d の評判: %w", index+1, ErrInvalidCount)
	}
	oldRank := w.Rank()
	w.Inventory = trial
	w.Money += quest.Reward
	w.Reputation += quest.Reputation
	for rank := oldRank + 1; rank <= w.Rank(); rank++ {
		w.rankUps = append(w.rankUps, rank)
	}
	w.Accepted = append(w.Accepted[:index], w.Accepted[index+1:]...)
	return nil
}
