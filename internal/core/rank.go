package core

import (
	"errors"
	"fmt"
	"sort"
)

// ErrLocked はランクによってまだ使えない定義を表す（DES-PRG-02）。
var ErrLocked = errors.New("まだ解放されていません")

// Unlocks は指定ランクで新しく解放される定義の ID を保持する（DES-PRG-02）。
type Unlocks struct {
	Recipes []string
	Gathers []string
	Quests  []string
}

// Rank は評判が到達した閾値の数から工房のランクを求める（DES-PRG-01）。
func (w *World) Rank() int {
	rank := 1
	for _, threshold := range w.catalog.Game.RankReputation {
		if w.Reputation >= threshold {
			rank++
		}
	}
	return rank
}

// NextRankReputation は次のランクに必要な評判の合計を返す（DES-PRG-01）。
func (w *World) NextRankReputation() (int, bool) {
	rank := w.Rank()
	if rank > len(w.catalog.Game.RankReputation) {
		return 0, false
	}
	return w.catalog.Game.RankReputation[rank-1], true
}

// Unlocked は指定ランクで新しく解放される ID を昇順で返す（DES-PRG-02）。
func (w *World) Unlocked(rank int) Unlocks {
	var result Unlocks
	for id, recipe := range w.catalog.Recipes {
		if effectiveRank(recipe.Rank) == rank {
			result.Recipes = append(result.Recipes, id)
		}
	}
	for id, site := range w.catalog.Gathers {
		if effectiveRank(site.Rank) == rank {
			result.Gathers = append(result.Gathers, id)
		}
	}
	for id, quest := range w.catalog.Quests {
		if effectiveRank(quest.Rank) == rank {
			result.Quests = append(result.Quests, id)
		}
	}
	sort.Strings(result.Recipes)
	sort.Strings(result.Gathers)
	sort.Strings(result.Quests)
	return result
}

// TakeRankUps は未読のランク上昇を読み出して空にする（DES-QST-03）。
func (w *World) TakeRankUps() []int {
	ups := w.rankUps
	w.rankUps = nil
	return ups
}

func effectiveRank(rank int) int {
	if rank == 0 {
		return 1
	}
	return rank
}

func (w *World) locked(rank int) bool {
	return len(w.catalog.Game.RankReputation) > 0 && w.Rank() < effectiveRank(rank)
}

func lockedError(name string, rank int) error {
	return fmt.Errorf("%sはランク %d で解放されます: %w", name, effectiveRank(rank), ErrLocked)
}
