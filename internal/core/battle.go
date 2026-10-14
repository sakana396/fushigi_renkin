package core

import (
	"errors"
	"fmt"
	"sort"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 戦闘操作の失敗理由（DES-BTL-02、DES-WLD-01）。
var (
	ErrNoBattle     = errors.New("この採取地では探索できません")
	ErrInBattle     = errors.New("戦闘中です")
	ErrInjured      = errors.New("負傷中です")
	ErrNotInBattle  = errors.New("戦闘中ではありません")
	ErrAlreadyMoved = errors.New("すでに移動しました")
	ErrOutOfRange   = errors.New("届かないマスです")
	ErrCellOccupied = errors.New("マスがふさがっています")
	ErrNoTarget     = errors.New("狙う相手がいません")
)

// Battle は現在の戦闘と行動順を保持する（DES-BTL-01）。
type Battle struct {
	GatherID string
	Units    []BattleUnit
	Order    []int
	Turn     int
	Moved    bool
	Log      []string
}

// BattleUnit は盤面上の仲間または敵を表す（DES-BTL-01）。
type BattleUnit struct {
	DefID       string
	Label       string
	Ally        bool
	HP          int
	X           int
	Y           int
	AttackBonus int
	SpeedBonus  int
}

// BattleReport は終了した戦闘の結果を保持する（DES-BTL-01）。
type BattleReport struct {
	GatherID     string
	Won          bool
	Rewards      []Stack
	InjuredUntil uint64
	Log          []string
}

// StartBattle は採取地の探索を始め、最初の仲間の番まで敵を動かす（DES-BTL-01、DES-BTL-02）。
func (w *World) StartBattle(gatherID string) error {
	site, ok := w.catalog.Gather(gatherID)
	if !ok {
		return fmt.Errorf("採取地 %s: %w", gatherID, ErrUnknownID)
	}
	if site.Battle == nil {
		return ErrNoBattle
	}
	if w.Battle != nil {
		return ErrInBattle
	}
	if w.Now < w.InjuredUntil {
		return ErrInjured
	}
	b := &Battle{GatherID: gatherID}
	counts := map[string]int{}
	seen := map[string]int{}
	for _, enemy := range site.Battle.Enemies {
		def, _ := w.catalog.Unit(enemy.Unit)
		counts[def.Name]++
	}
	appendUnit := func(p defs.Placement, ally bool) {
		def, _ := w.catalog.Unit(p.Unit)
		label := def.Name
		if !ally && counts[label] > 1 {
			seen[label]++
			label = fmt.Sprintf("%s %d", label, seen[def.Name])
		}
		b.Units = append(b.Units, BattleUnit{DefID: p.Unit, Label: label, Ally: ally, HP: def.HP, X: p.X, Y: p.Y})
	}
	for _, p := range w.catalog.Game.Party {
		appendUnit(p, true)
	}
	for _, p := range site.Battle.Enemies {
		appendUnit(p, false)
	}
	w.Battle, w.LastBattle = b, nil
	w.sortBattleOrder()
	w.advanceBattle(false)
	return nil
}

// BattleMove は仲間を一度だけ到達可能な空きマスへ動かす（DES-BTL-02）。
func (w *World) BattleMove(x, y int) error {
	if w.Battle == nil {
		return ErrNotInBattle
	}
	if w.Battle.Moved {
		return ErrAlreadyMoved
	}
	u := &w.Battle.Units[w.Battle.Order[w.Battle.Turn]]
	def, _ := w.catalog.Unit(u.DefID)
	if !w.inBattleBounds(x, y) || battleDistance(u.X, u.Y, x, y) > def.Move {
		return ErrOutOfRange
	}
	if w.Battle.occupied(x, y) {
		return ErrCellOccupied
	}
	u.X, u.Y = x, y
	w.Battle.Moved = true
	w.Battle.Log = append(w.Battle.Log, fmt.Sprintf("%sは(%d, %d)へ移動", u.Label, x, y))
	return nil
}

// BattleAttack は届く敵を攻撃して番を進める（DES-BTL-02）。
func (w *World) BattleAttack(x, y int) error {
	if w.Battle == nil {
		return ErrNotInBattle
	}
	b := w.Battle
	u := &b.Units[b.Order[b.Turn]]
	def, _ := w.catalog.Unit(u.DefID)
	if !w.inBattleBounds(x, y) || battleDistance(u.X, u.Y, x, y) > def.Range {
		return ErrOutOfRange
	}
	target := -1
	for i := range b.Units {
		v := &b.Units[i]
		if !v.Ally && v.HP > 0 && v.X == x && v.Y == y {
			target = i
			break
		}
	}
	if target < 0 {
		return ErrNoTarget
	}
	damage := def.Attack + u.AttackBonus
	b.Units[target].HP -= damage
	if b.Units[target].HP < 0 {
		b.Units[target].HP = 0
	}
	b.Log = append(b.Log, fmt.Sprintf("%sは%sに%dのダメージ", u.Label, b.Units[target].Label, damage))
	w.advanceBattle(true)
	return nil
}

// BattleWait は仲間の番を終える（DES-BTL-02）。
func (w *World) BattleWait() error {
	if w.Battle == nil {
		return ErrNotInBattle
	}
	u := w.Battle.Units[w.Battle.Order[w.Battle.Turn]]
	w.Battle.Log = append(w.Battle.Log, fmt.Sprintf("%sは待機", u.Label))
	w.advanceBattle(true)
	return nil
}

func battleDistance(ax, ay, bx, by int) int {
	dx, dy := ax-bx, ay-by
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (w *World) inBattleBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < w.catalog.Game.BattleWidth && y < w.catalog.Game.BattleHeight
}

func (b *Battle) occupied(x, y int) bool {
	for _, u := range b.Units {
		if u.HP > 0 && u.X == x && u.Y == y {
			return true
		}
	}
	return false
}

// sortBattleOrder は巡の始めに素早さ、陣営、配置順で並べる（DES-BTL-02）。
func (w *World) sortBattleOrder() {
	b := w.Battle
	b.Order = b.Order[:0]
	for i, u := range b.Units {
		if u.HP > 0 {
			b.Order = append(b.Order, i)
		}
	}
	sort.SliceStable(b.Order, func(a, c int) bool {
		left, right := b.Units[b.Order[a]], b.Units[b.Order[c]]
		ld, _ := w.catalog.Unit(left.DefID)
		rd, _ := w.catalog.Unit(right.DefID)
		ls, rs := ld.Speed+left.SpeedBonus, rd.Speed+right.SpeedBonus
		if ls != rs {
			return ls > rs
		}
		if left.Ally != right.Ally {
			return left.Ally
		}
		return b.Order[a] < b.Order[c]
	})
}

// advanceBattle は行動直後の勝敗を判定し、次の仲間の番まで進める（DES-BTL-02）。
func (w *World) advanceBattle(afterAction bool) {
	for w.Battle != nil {
		b := w.Battle
		if afterAction {
			if w.finishBattle() {
				return
			}
			b.Turn++
		}
		afterAction = false
		if b.Turn >= len(b.Order) {
			w.sortBattleOrder()
			b.Turn = 0
		}
		for b.Turn < len(b.Order) && b.Units[b.Order[b.Turn]].HP <= 0 {
			b.Turn++
		}
		if b.Turn >= len(b.Order) {
			w.sortBattleOrder()
			b.Turn = 0
		}
		b.Moved = false
		if b.Units[b.Order[b.Turn]].Ally {
			return
		}
		w.enemyAction(b.Order[b.Turn])
		afterAction = true
	}
}

// finishBattle は勝敗と報酬または負傷を確定する（DES-BTL-02、DES-BTL-04）。
func (w *World) finishBattle() bool {
	b := w.Battle
	ally, enemy := false, false
	for _, u := range b.Units {
		if u.HP > 0 {
			if u.Ally {
				ally = true
			} else {
				enemy = true
			}
		}
	}
	if ally && enemy {
		return false
	}
	report := &BattleReport{GatherID: b.GatherID, Won: ally}
	if ally {
		site, _ := w.catalog.Gather(b.GatherID)
		stacks, err := w.drawDrops(site.Battle.Rolls, site.Battle.Drops)
		if err != nil {
			b.Log = append(b.Log, fmt.Sprintf("報酬の抽選に失敗: %v", err))
		}
		for _, stack := range stacks {
			room := int(^uint(0)>>1) - w.Inventory.Count(stack.Key)
			accepted := stack.Count
			if accepted > room {
				accepted = room
			}
			if accepted > 0 {
				if err := w.Inventory.Add(stack.Key, accepted); err == nil {
					report.Rewards = append(report.Rewards, Stack{Key: stack.Key, Count: accepted})
				} else {
					accepted = 0
				}
			}
			if accepted < stack.Count {
				b.Log = append(b.Log, fmt.Sprintf("%sの報酬%d個を破棄", stack.Key.ItemID, stack.Count-accepted))
			}
		}
		b.Log = append(b.Log, "勝利")
	} else {
		w.InjuredUntil = w.Now + uint64(w.catalog.Game.InjurySeconds)*60
		report.InjuredUntil = w.InjuredUntil
		b.Log = append(b.Log, "敗北")
	}
	report.Log = append([]string(nil), b.Log...)
	w.LastBattle, w.Battle = report, nil
	return true
}
