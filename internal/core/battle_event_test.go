package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 出来事と記録の対応、および DES-IT04 2.2 節の計算例を確かめる。
func TestBattleEventsExample(t *testing.T) {
	w := NewWorld(battleCatalog(), 1)
	steps := []struct {
		name  string
		do    func() error
		want  []BattleEvent
		units [][4][3]int
	}{
		{"TC-IF-07-01 開始", func() error { return w.StartBattle("field") }, []BattleEvent{{Kind: "start", Actor: -1, Target: -1}, {Kind: "move", Actor: 3, Target: -1, FromX: 5, FromY: 3, ToX: 2, ToY: 3}}, [][4][3]int{{{1, 2, 30}, {0, 3, 18}, {3, 2, 14}, {5, 3, 18}}, {{1, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 18}}}},
		{"TC-IF-07-01 術士の攻撃", func() error { return w.BattleAttack(2, 3) }, []BattleEvent{{Kind: "attack", Actor: 1, Target: 3, FromX: 0, FromY: 3, ToX: 2, ToY: 3, Amount: 5}}, [][4][3]int{{{1, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 13}}}},
		{"TC-IF-07-01 剣士の移動", func() error { return w.BattleMove(2, 2) }, []BattleEvent{{Kind: "move", Actor: 0, Target: -1, FromX: 1, FromY: 2, ToX: 2, ToY: 2}}, [][4][3]int{{{2, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 13}}}},
		{"TC-IF-07-01 剣士の攻撃", func() error { return w.BattleAttack(2, 3) }, []BattleEvent{{Kind: "attack", Actor: 0, Target: 3, FromX: 2, FromY: 2, ToX: 2, ToY: 3, Amount: 6}, {Kind: "attack", Actor: 2, Target: 0, FromX: 3, FromY: 2, ToX: 2, ToY: 2, Amount: 3}, {Kind: "attack", Actor: 3, Target: 0, FromX: 2, FromY: 3, ToX: 2, ToY: 2, Amount: 5}}, [][4][3]int{{{2, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 7}}, {{2, 2, 27}, {0, 3, 18}, {3, 2, 14}, {2, 3, 7}}, {{2, 2, 22}, {0, 3, 18}, {3, 2, 14}, {2, 3, 7}}}},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			before := 0
			if w.Battle != nil {
				before = len(w.Battle.Events)
			}
			if err := step.do(); err != nil {
				t.Fatal(err)
			}
			b := w.Battle
			got := b.Events[before:]
			if len(got) != len(step.want) {
				t.Fatalf("出来事数=%d、期待=%d", len(got), len(step.want))
			}
			for i, want := range step.want {
				e := got[i]
				if e.Kind != want.Kind || e.Actor != want.Actor || e.Target != want.Target || e.FromX != want.FromX || e.FromY != want.FromY || e.ToX != want.ToX || e.ToY != want.ToY || e.Amount != want.Amount {
					t.Errorf("出来事 %d=%+v、期待=%+v", i, e, want)
				}
				if len(e.Units) != 4 {
					t.Fatalf("写しの数=%d", len(e.Units))
				}
				for j, expected := range step.units[i] {
					u := e.Units[j]
					if [3]int{u.X, u.Y, u.HP} != expected {
						t.Errorf("出来事 %d のユニット %d=(%d,%d,HP%d)、期待=%v", i, j, u.X, u.Y, u.HP, expected)
					}
				}
			}
			if len(b.Events) != len(b.Log)+1 {
				t.Fatalf("出来事=%d、記録=%d", len(b.Events), len(b.Log))
			}
			for i, line := range b.Log {
				if b.Events[i+1].Text != line {
					t.Errorf("%d: %q != %q", i, b.Events[i+1].Text, line)
				}
			}
		})
	}
	if len(w.Battle.Events) != 7 || len(w.Battle.Log) != 6 {
		t.Fatalf("出来事=%d、記録=%d", len(w.Battle.Events), len(w.Battle.Log))
	}
	if w.Battle.Events[0].Text != "" {
		t.Errorf("開始の文=%q", w.Battle.Events[0].Text)
	}
	if u := w.Battle.Events[0].Units[3]; u.X != 5 || u.Y != 3 {
		t.Errorf("開始位置=%+v", u)
	}
	if u := w.Battle.Events[1].Units[3]; u.X != 2 || u.Y != 3 {
		t.Errorf("移動後=%+v", u)
	}
	if u := w.Battle.Events[5].Units[0]; u.HP != 27 {
		t.Errorf("反撃後体力=%d", u.HP)
	}
}

func TestBattleTargetCells(t *testing.T) {
	w := NewWorld(battleCatalog(), 1)
	t.Run("TC-IF-07-07 戦闘外", func(t *testing.T) {
		if len(w.AttackTargets()) != 0 || len(w.UseTargets(ItemKey{"herb", 40})) != 0 {
			t.Fatal("戦闘外で相手がいる")
		}
	})
	if err := w.StartBattle("field"); err != nil {
		t.Fatal(err)
	}
	t.Run("TC-IF-07-06 射程内の敵", func(t *testing.T) {
		before := copyBattle(w.Battle)
		if got := w.AttackTargets(); !reflect.DeepEqual(got, [][2]int{{2, 3}}) {
			t.Fatalf("相手=%v", got)
		}
		if !reflect.DeepEqual(*w.Battle, before) {
			t.Fatal("状態が変わった")
		}
	})
	t.Run("TC-IF-07-07 死亡と射程外", func(t *testing.T) {
		w.Battle.Units[3].HP = 0
		if got := w.AttackTargets(); len(got) != 0 {
			t.Fatalf("相手=%v", got)
		}
	})
	item := testBurgerUse()
	w.catalog.Items["burger"] = item
	bomb := testBombUse()
	w.catalog.Items["bomb"] = bomb
	_ = w.Inventory.Add(ItemKey{"burger", 100}, 1)
	_ = w.Inventory.Add(ItemKey{"bomb", 60}, 1)
	t.Run("TC-IF-07-08 対象と射程", func(t *testing.T) {
		before := copyBattle(w.Battle)
		if got := w.UseTargets(ItemKey{"burger", 100}); !reflect.DeepEqual(got, [][2]int{{0, 3}}) {
			t.Errorf("仲間=%v", got)
		}
		w.Battle.Units[3].HP = 18
		if got := w.UseTargets(ItemKey{"bomb", 60}); !reflect.DeepEqual(got, [][2]int{{2, 3}}) {
			t.Errorf("敵=%v", got)
		}
		w.Battle.Units[3].HP = 0
		if !reflect.DeepEqual(*w.Battle, before) {
			t.Fatal("状態が変わった")
		}
	})
	t.Run("TC-IF-07-09 在庫なし・使用不可・死亡", func(t *testing.T) {
		for _, key := range []ItemKey{{"burger", 20}, {"herb", 40}, {"unknown", 1}, {"bomb", 60}} {
			if got := w.UseTargets(key); len(got) != 0 {
				t.Errorf("%+v: %v", key, got)
			}
		}
	})
}

// 射程の境界は、候補の表示と実際の操作で同じ判定になる（DES-BTL-06）。
func TestBattleTargetRangeBoundary(t *testing.T) {
	for _, tc := range []struct {
		name      string
		item      bool
		x, y      int
		reachable bool
	}{
		{"TC-IF-07-06 攻撃の射程ちょうど", false, 2, 3, true},
		{"TC-IF-07-07 攻撃の射程を一つ超える", false, 3, 3, false},
		{"TC-IF-07-08 アイテムの射程ちょうど", true, 3, 3, true},
		{"TC-IF-07-09 アイテムの射程を一つ超える", true, 4, 3, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := battleCatalog()
			slime := c.Units["slime"]
			slime.Speed, slime.Move, slime.Attack = 1, 0, 0
			c.Units["slime"] = slime
			c.Items["bomb"] = testBombUse()
			c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: tc.x, Y: tc.y}}}}
			w := NewWorld(c, 1)
			if err := w.StartBattle("field"); err != nil {
				t.Fatal(err)
			}
			key := ItemKey{"bomb", 20}
			if err := w.Inventory.Add(key, 1); err != nil {
				t.Fatal(err)
			}
			before := copyBattle(w.Battle)
			var got [][2]int
			if tc.item {
				got = w.UseTargets(key)
			} else {
				got = w.AttackTargets()
			}
			var want [][2]int
			if tc.reachable {
				want = [][2]int{{tc.x, tc.y}}
			}
			if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(*w.Battle, before) {
				t.Fatalf("候補=%v、期待=%v、状態=%+v", got, want, w.Battle)
			}
			var err error
			if tc.item {
				err = w.BattleUseItem(key, tc.x, tc.y)
			} else {
				err = w.BattleAttack(tc.x, tc.y)
			}
			if tc.reachable && err != nil || !tc.reachable && !errors.Is(err, ErrOutOfRange) {
				t.Fatalf("実操作の結果=%v", err)
			}
			if !tc.reachable && !reflect.DeepEqual(*w.Battle, before) {
				t.Fatal("射程外の操作で状態が変わった")
			}
		})
	}
}

func TestBattleEventItemEffects(t *testing.T) {
	t.Run("TC-IF-07-03 爆弾の範囲効果", func(t *testing.T) {
		w := itemBattle(t)
		key := ItemKey{"bomb", 60}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 2, 3); err != nil {
			t.Fatal(err)
		}
		b := w.LastBattle
		want := []BattleEvent{{Kind: "use", Actor: 1, Target: -1, FromX: 0, FromY: 3, ToX: 2, ToY: 3}, {Kind: "damage", Actor: 1, Target: 2, FromX: 0, FromY: 3, ToX: 2, ToY: 3, Amount: 24}, {Kind: "fall", Actor: -1, Target: 2}, {Kind: "damage", Actor: 1, Target: 3, FromX: 0, FromY: 3, ToX: 3, ToY: 3, Amount: 24}, {Kind: "fall", Actor: -1, Target: 3}, {Kind: "win", Actor: -1, Target: -1}}
		if len(b.Events) != len(want)+1 {
			t.Fatalf("出来事=%v", b.Events)
		}
		for i, v := range want {
			e := b.Events[i+1]
			if e.Kind != v.Kind || e.Actor != v.Actor || e.Target != v.Target || e.FromX != v.FromX || e.FromY != v.FromY || e.ToX != v.ToX || e.ToY != v.ToY || e.Amount != v.Amount {
				t.Errorf("%d: %+v", i, e)
			}
			if i == 1 && e.Units[2].HP != 0 || i == 3 && (e.Units[2].HP != 0 || e.Units[3].HP != 0) {
				t.Errorf("damage 直後の写し %d: %+v", i, e.Units)
			}
		}
	})
	t.Run("TC-IF-07-03 / TC-IF-07-04 回復と強化", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[1].HP = 5
		key := ItemKey{"burger", 160}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		b := w.Battle
		want := []struct {
			kind                      string
			amount, hp, attack, speed int
		}{{"use", 0, 5, 0, 0}, {"heal", 13, 18, 0, 0}, {"attack_up", 8, 18, 8, 0}, {"speed_up", 4, 18, 8, 4}}
		for i, v := range want {
			e := b.Events[i+1]
			if e.Kind != v.kind || e.Amount != v.amount || e.Actor != 1 || e.Units[1].HP != v.hp || e.Units[1].AttackBonus != v.attack || e.Units[1].SpeedBonus != v.speed {
				t.Errorf("%d: %+v", i, e)
			}
			if i > 0 && (e.Target != 1 || e.ToX != 0 || e.ToY != 3) {
				t.Errorf("相手=%+v", e)
			}
		}
		key = ItemKey{"burger", 100}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		before := len(b.Events)
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		if len(b.Events) != before+2 || b.Events[before].Kind != "use" || b.Events[before+1].Kind != "attack_up" {
			t.Errorf("満タン時の出来事=%v", b.Events[before:])
		}
		if len(b.Events) != len(b.Log)+1 {
			t.Fatalf("出来事=%d 記録=%d", len(b.Events), len(b.Log))
		}
	})
}

// 効果を受けた仲間と使用者が異なるときの座標と、強化後の攻撃を確かめる（DES-BTL-05）。
func TestBattleEventBuffedAlly(t *testing.T) {
	t.Run("TC-IF-07-03 回復と強化の座標・強化後の攻撃", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[0].X, w.Battle.Units[0].Y, w.Battle.Units[0].HP = 1, 3, 5
		key := ItemKey{"burger", 160}
		if err := w.Inventory.Add(key, 1); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleUseItem(key, 1, 3); err != nil {
			t.Fatal(err)
		}
		b := w.Battle
		for i, want := range []struct {
			kind   string
			amount int
			hp     int
		}{
			{"heal", 25, 30},
			{"attack_up", 8, 30},
			{"speed_up", 4, 30},
		} {
			e := b.Events[i+2]
			if e.Kind != want.kind || e.Actor != 1 || e.Target != 0 || e.FromX != 0 || e.FromY != 3 || e.ToX != 1 || e.ToY != 3 || e.Amount != want.amount || e.Units[0].HP != want.hp {
				t.Errorf("%s=%+v、期待=%+v", want.kind, e, want)
			}
		}
		before := len(b.Events)
		if err := w.BattleAttack(2, 3); err != nil {
			t.Fatal(err)
		}
		e := b.Events[before]
		if e.Kind != "attack" || e.Actor != 0 || e.Target != 2 || e.Amount != 14 || e.Text != fmt.Sprintf("剣士はスライムに%dのダメージ", e.Amount) || e.Units[2].HP != 0 {
			t.Errorf("強化後の攻撃=%+v", e)
		}
		if fall := b.Events[before+1]; fall.Kind != "fall" || fall.Actor != -1 || fall.Target != 2 {
			t.Errorf("仲間の攻撃による fall=%+v", fall)
		}
	})
}

func TestBattleEventReport(t *testing.T) {
	t.Run("TC-IF-07-05 敵の攻撃で添字1が倒れる", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 0, Y: 4}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		// 開始時に敵が先に動くので、体力 5 の術士を用意し直して次巡を進める。
		w.Battle.Units[1].HP = 5
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range w.Battle.Events {
			if e.Kind == "fall" {
				found = true
				if e.Actor != -1 || e.Target != 1 {
					t.Errorf("術士の fall=%+v", e)
				}
			}
		}
		if !found {
			t.Fatal("術士の fall がない")
		}
	})
	t.Run("TC-IF-07-02 / TC-IF-07-05 勝利と報酬破棄", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 1, Y: 3}}, Rolls: 2, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		_ = w.Inventory.Add(ItemKey{"herb", 40}, int(^uint(0)>>1)-1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[2].HP = 1
		if err := w.BattleAttack(1, 3); err != nil {
			t.Fatal(err)
		}
		r := w.LastBattle
		want := []string{"start", "attack", "fall", "reward_lost", "win"}
		if len(r.Events) != len(want) {
			t.Fatalf("出来事=%v", r.Events)
		}
		for i, kind := range want {
			if r.Events[i].Kind != kind {
				t.Errorf("%d: %s", i, r.Events[i].Kind)
			}
		}
		if e := r.Events[3]; e.Amount != 1 || e.Actor != -1 || e.Target != -1 || e.FromX != 0 || e.FromY != 0 || e.ToX != 0 || e.ToY != 0 || e.Text != "薬草（品質40）の報酬1個を破棄" {
			t.Errorf("報酬=%+v", e)
		}
		if !reflect.DeepEqual(r.Rewards, []Stack{{Key: ItemKey{"herb", 40}, Count: 1}}) || w.Inventory.Count(ItemKey{"herb", 40}) != int(^uint(0)>>1) {
			t.Errorf("受け取った報酬=%v、在庫=%d", r.Rewards, w.Inventory.Count(ItemKey{"herb", 40}))
		}
		if e := r.Events[2]; e.Kind != "fall" || e.Actor != -1 || e.Target != 2 {
			t.Errorf("仲間の攻撃による fall=%+v", e)
		}
		if e := r.Events[4]; e.Kind != "win" || e.Actor != -1 || e.Target != -1 || e.FromX != 0 || e.FromY != 0 || e.ToX != 0 || e.ToY != 0 || e.Amount != 0 {
			t.Errorf("勝利=%+v", e)
		}
		if !reflect.DeepEqual(r.Events[len(r.Events)-1].Units, r.Units) || r.Units[2].HP != 0 {
			t.Errorf("最後の盤面=%v", r.Units)
		}
		for i, line := range r.Log {
			if r.Events[i+1].Text != line {
				t.Errorf("記録 %d", i)
			}
		}
	})
	t.Run("TC-IF-07-02 / TC-IF-07-05 待機と敗北", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 1, Y: 2}}
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 2, Y: 2}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[0].HP = 1
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		r := w.LastBattle
		if r == nil || r.Won {
			t.Fatalf("報告=%+v", r)
		}
		want := []string{"start", "attack", "wait", "attack", "fall", "lose"}
		if len(r.Events) != len(want) {
			t.Fatalf("出来事=%v", r.Events)
		}
		for i, kind := range want {
			if r.Events[i].Kind != kind {
				t.Errorf("%d: %s", i, r.Events[i].Kind)
			}
		}
		if e := r.Events[2]; e.Actor != 0 || e.Target != -1 || e.FromX != 1 || e.FromY != 2 || e.ToX != 1 || e.ToY != 2 {
			t.Errorf("待機=%+v", e)
		}
		if e := r.Events[4]; e.Kind != "fall" || e.Actor != -1 || e.Target != 0 {
			t.Errorf("敵の攻撃による fall=%+v", e)
		}
		if e := r.Events[5]; e.Kind != "lose" || e.Actor != -1 || e.Target != -1 || e.FromX != 0 || e.FromY != 0 || e.ToX != 0 || e.ToY != 0 || e.Amount != 0 {
			t.Errorf("敗北=%+v", e)
		}
		if !reflect.DeepEqual(r.Events[len(r.Events)-1].Units, r.Units) || r.Units[0].HP != 0 {
			t.Errorf("最後の盤面=%v", r.Units)
		}
	})
	t.Run("TC-IF-07-05 報酬の抽選失敗", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 1, Y: 3}}, Rolls: 0, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[2].HP = 1
		if err := w.BattleAttack(1, 3); err != nil {
			t.Fatal(err)
		}
		r := w.LastBattle
		if r == nil || !r.Won || len(r.Rewards) != 0 || len(r.Events) != 5 || len(r.Log) != 4 {
			t.Fatalf("報告=%+v", r)
		}
		e := r.Events[3]
		if e.Kind != "reward_lost" || e.Actor != -1 || e.Target != -1 || e.FromX != 0 || e.FromY != 0 || e.ToX != 0 || e.ToY != 0 || e.Amount != 0 || e.Text != r.Log[2] || e.Text == "" {
			t.Errorf("抽選失敗=%+v、記録=%v", e, r.Log)
		}
	})
}
