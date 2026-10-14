package core

import "fmt"

// enemyAction は最も近い仲間を決定的に追い、届けば攻撃する（DES-BTL-03、NFR-06）。
func (w *World) enemyAction(index int) {
	b := w.Battle
	enemy := &b.Units[index]
	def, _ := w.catalog.Unit(enemy.DefID)
	target := -1
	nearest := int(^uint(0) >> 1)
	for i, ally := range b.Units {
		if !ally.Ally || ally.HP <= 0 {
			continue
		}
		d := battleDistance(enemy.X, enemy.Y, ally.X, ally.Y)
		if d < nearest {
			nearest, target = d, i
		}
	}
	if target < 0 {
		return
	}
	ally := &b.Units[target]
	if nearest > def.Range {
		bestX, bestY, bestDistance := enemy.X, enemy.Y, nearest
		for y := 0; y < w.catalog.Game.BattleHeight; y++ {
			for x := 0; x < w.catalog.Game.BattleWidth; x++ {
				if battleDistance(enemy.X, enemy.Y, x, y) > def.Move || b.occupied(x, y) {
					continue
				}
				d := battleDistance(x, y, ally.X, ally.Y)
				if d < bestDistance {
					bestX, bestY, bestDistance = x, y, d
				}
			}
		}
		if bestDistance < nearest {
			enemy.X, enemy.Y = bestX, bestY
			b.Log = append(b.Log, fmt.Sprintf("%sは(%d, %d)へ移動", enemy.Label, bestX, bestY))
		}
	}
	if battleDistance(enemy.X, enemy.Y, ally.X, ally.Y) <= def.Range {
		ally.HP -= def.Attack
		if ally.HP < 0 {
			ally.HP = 0
		}
		b.Log = append(b.Log, fmt.Sprintf("%sは%sに%dのダメージ", enemy.Label, ally.Label, def.Attack))
	}
}
