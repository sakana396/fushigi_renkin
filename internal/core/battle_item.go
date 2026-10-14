package core

import (
	"errors"
	"fmt"
)

// ErrNotUsable は戦闘で使えない種類を表す（DES-BTL-02）。
var ErrNotUsable = errors.New("戦闘で使えないアイテムです")

// BattleUseItem は消耗品を使用し、対象の生きているユニットに効果を与えて番を進める（DES-BTL-02、DES-QLT-05）。
func (w *World) BattleUseItem(key ItemKey, x, y int) error {
	if w.Battle == nil {
		return ErrNotInBattle
	}
	item, ok := w.catalog.Item(key.ItemID)
	if !ok || item.Use == nil {
		return ErrNotUsable
	}
	if w.Inventory.Count(key) < 1 {
		return ErrNotEnoughItems
	}
	b := w.Battle
	actor := b.Units[b.Order[b.Turn]]
	if !w.inBattleBounds(x, y) || battleDistance(actor.X, actor.Y, x, y) > item.Use.Range {
		return ErrOutOfRange
	}
	ally := item.Use.Target == "ally"
	found := false
	for _, u := range b.Units {
		if u.Ally == ally && u.HP > 0 && u.X == x && u.Y == y {
			found = true
			break
		}
	}
	if !found {
		return ErrNoTarget
	}

	effect, _ := UseEffectOf(item, key.Quality)
	// 在庫の検査は済んでいるため、この操作は失敗しない。
	if err := w.Inventory.Remove(key, 1); err != nil {
		return err
	}
	b.Log = append(b.Log, fmt.Sprintf("%sは%s（品質%d）を使用", actor.Label, item.Name, key.Quality))
	for i := range b.Units {
		u := &b.Units[i]
		if u.Ally != ally || u.HP <= 0 || battleDistance(x, y, u.X, u.Y) > effect.Radius {
			continue
		}
		for _, e := range effect.Effects {
			switch e.Kind {
			case "heal":
				maxHP := w.catalog.Units[u.DefID].HP
				before := u.HP
				u.HP += e.Value
				if u.HP > maxHP {
					u.HP = maxHP
				}
				// 体力が満タンで回復しなかったときは記録しない（DES-UI-08 の記録を読みやすくする）。
				if u.HP > before {
					b.Log = append(b.Log, fmt.Sprintf("%sの体力が%d回復", u.Label, u.HP-before))
				}
			case "damage":
				u.HP -= e.Value
				if u.HP < 0 {
					u.HP = 0
				}
				b.Log = append(b.Log, fmt.Sprintf("%sに%dのダメージ", u.Label, e.Value))
				if u.HP == 0 {
					b.Log = append(b.Log, u.Label+"は倒れた")
				}
			case "attack_up":
				u.AttackBonus += e.Value
				b.Log = append(b.Log, fmt.Sprintf("%sの攻撃が%d上がった", u.Label, e.Value))
			case "speed_up":
				u.SpeedBonus += e.Value
				b.Log = append(b.Log, fmt.Sprintf("%sの素早さが%d上がった", u.Label, e.Value))
			}
		}
	}
	w.advanceBattle(true)
	return nil
}
