package core

// BattleEvent は行動直後の盤面を持つ戦闘の出来事（DES-BTL-05）。
type BattleEvent struct {
	Kind   string
	Actor  int
	Target int
	FromX  int
	FromY  int
	ToX    int
	ToY    int
	Amount int
	Text   string
	Units  []BattleUnit
}

// addEvent は記録と同じ文の出来事に盤面の写しを残す（DES-BTL-05）。
func (b *Battle) addEvent(e BattleEvent, text string) {
	e.Text = text
	e.Units = append([]BattleUnit(nil), b.Units...)
	b.Events = append(b.Events, e)
	if e.Kind != "start" {
		b.Log = append(b.Log, text)
	}
}

// AttackTargets は今の仲間が攻撃できる敵のマスを返す（DES-BTL-06）。
func (w *World) AttackTargets() [][2]int {
	if w.Battle == nil {
		return nil
	}
	b := w.Battle
	actor := b.Units[b.Order[b.Turn]]
	if !actor.Ally || actor.HP <= 0 {
		return nil
	}
	def, _ := w.catalog.Unit(actor.DefID)
	var targets [][2]int
	for _, u := range b.Units {
		if !u.Ally && u.HP > 0 && w.inBattleBounds(u.X, u.Y) && battleDistance(actor.X, actor.Y, u.X, u.Y) <= def.Range {
			targets = append(targets, [2]int{u.X, u.Y})
		}
	}
	return targets
}

// UseTargets は所持する使用可能なアイテムで狙えるマスを返す（DES-BTL-06）。
func (w *World) UseTargets(key ItemKey) [][2]int {
	if w.Battle == nil || w.Inventory.Count(key) < 1 {
		return nil
	}
	item, ok := w.catalog.Item(key.ItemID)
	if !ok || item.Use == nil {
		return nil
	}
	b := w.Battle
	actor := b.Units[b.Order[b.Turn]]
	if !actor.Ally || actor.HP <= 0 {
		return nil
	}
	ally := item.Use.Target == "ally"
	var targets [][2]int
	for _, u := range b.Units {
		if u.Ally == ally && u.HP > 0 && w.inBattleBounds(u.X, u.Y) && battleDistance(actor.X, actor.Y, u.X, u.Y) <= item.Use.Range {
			targets = append(targets, [2]int{u.X, u.Y})
		}
	}
	return targets
}
