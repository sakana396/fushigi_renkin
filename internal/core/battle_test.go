package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func battleCatalog() defs.Catalog {
	return defs.Catalog{
		Game: defs.GameDef{Party: []defs.Placement{{Unit: "sword", X: 1, Y: 2}, {Unit: "mage", X: 0, Y: 3}}, BattleWidth: 8, BattleHeight: 6, InjurySeconds: 60},
		Units: map[string]defs.UnitDef{
			"sword": {ID: "sword", Name: "剣士", Role: "front", HP: 30, Attack: 6, Speed: 5, Move: 2, Range: 1},
			"mage":  {ID: "mage", Name: "術士", Role: "back", HP: 18, Attack: 5, Speed: 7, Move: 2, Range: 2},
			"slime": {ID: "slime", Name: "スライム", Role: "enemy", HP: 14, Attack: 3, Speed: 4, Move: 1, Range: 1},
			"wolf":  {ID: "wolf", Name: "オオカミ", Role: "enemy", HP: 18, Attack: 5, Speed: 8, Move: 3, Range: 1},
		},
		Items: map[string]defs.ItemDef{"herb": {ID: "herb", Name: "薬草"}},
		Gathers: map[string]defs.GatherDef{
			"field": {ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 3, Y: 2}, {Unit: "wolf", X: 5, Y: 3}}, Rolls: 2, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}},
			"empty": {ID: "empty"},
		},
	}
}

// 戦闘の可変スライスも複製して、誤りの前後を比較する（DES-WLD-01）。
func copyBattle(b *Battle) Battle {
	copy := *b
	copy.Units = append([]BattleUnit(nil), b.Units...)
	copy.Order = append([]int(nil), b.Order...)
	copy.Log = append([]string(nil), b.Log...)
	return copy
}

func assertBattle(t *testing.T, w *World, order []int, turn int, values [][3]int) {
	t.Helper()
	if w.Battle == nil {
		t.Fatal("戦闘が終了した")
	}
	if !reflect.DeepEqual(w.Battle.Order, order) || w.Battle.Turn != turn {
		t.Fatalf("行動順=%v 番=%d、期待=%v %d", w.Battle.Order, w.Battle.Turn, order, turn)
	}
	for i, v := range values {
		u := w.Battle.Units[i]
		if u.X != v[0] || u.Y != v[1] || u.HP != v[2] {
			t.Fatalf("unit %d=(%d,%d) HP%d、期待=%v", i, u.X, u.Y, u.HP, v)
		}
	}
}

// DES-BTL-02、03 の 4 巡の計算例を座標・HP・行動順ごとに確かめる。
func TestBattleFourRounds(t *testing.T) {
	w := NewWorld(battleCatalog(), 1)
	steps := []struct {
		name   string
		action func() error
		order  []int
		turn   int
		units  [][3]int
		log    []string
	}{
		{"TC-FR-BTL-01-01 開始時のオオカミ", func() error { return w.StartBattle("field") }, []int{3, 1, 0, 2}, 1, [][3]int{{1, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 18}}, []string{"オオカミは(2, 3)へ移動"}},
		{"TC-FR-BTL-01-02 術士の射程", func() error { return w.BattleAttack(2, 3) }, []int{3, 1, 0, 2}, 2, [][3]int{{1, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 13}}, []string{"術士はオオカミに5のダメージ"}},
		{"TC-FR-BTL-01-03 剣士の移動", func() error { return w.BattleMove(2, 2) }, []int{3, 1, 0, 2}, 2, [][3]int{{2, 2, 30}, {0, 3, 18}, {3, 2, 14}, {2, 3, 13}}, []string{"剣士は(2, 2)へ移動"}},
		{"TC-FR-BTL-01-04 剣士の攻撃と敵二体", func() error { return w.BattleAttack(2, 3) }, []int{3, 1, 0, 2}, 1, [][3]int{{2, 2, 22}, {0, 3, 18}, {3, 2, 14}, {2, 3, 7}}, []string{"剣士はオオカミに6のダメージ", "スライムは剣士に3のダメージ", "オオカミは剣士に5のダメージ"}},
		{"TC-FR-BTL-01-06 二巡目の術士", func() error { return w.BattleAttack(2, 3) }, []int{3, 1, 0, 2}, 2, [][3]int{{2, 2, 22}, {0, 3, 18}, {3, 2, 14}, {2, 3, 2}}, []string{"術士はオオカミに5のダメージ"}},
		{"TC-FR-BTL-01-07 オオカミが倒れる", func() error { return w.BattleAttack(2, 3) }, []int{1, 0, 2}, 0, [][3]int{{2, 2, 19}, {0, 3, 18}, {3, 2, 14}, {2, 3, 0}}, []string{"剣士はオオカミに6のダメージ", "スライムは剣士に3のダメージ"}},
		{"TC-FR-BTL-01-08 倒れたマスへ移動", func() error { return w.BattleMove(2, 3) }, []int{1, 0, 2}, 0, [][3]int{{2, 2, 19}, {2, 3, 18}, {3, 2, 14}, {2, 3, 0}}, []string{"術士は(2, 3)へ移動"}},
		{"TC-FR-BTL-01-09 三巡目の術士", func() error { return w.BattleAttack(3, 2) }, []int{1, 0, 2}, 1, [][3]int{{2, 2, 19}, {2, 3, 18}, {3, 2, 9}, {2, 3, 0}}, []string{"術士はスライムに5のダメージ"}},
		{"TC-FR-BTL-01-10 三巡目の剣士", func() error { return w.BattleAttack(3, 2) }, []int{1, 0, 2}, 0, [][3]int{{2, 2, 16}, {2, 3, 18}, {3, 2, 3}, {2, 3, 0}}, []string{"剣士はスライムに6のダメージ", "スライムは剣士に3のダメージ"}},
	}
	for _, s := range steps {
		if strings.HasPrefix(s.name, "TC-FR-BTL-01-08") {
			t.Run("TC-FR-BTL-01-05 三巡目の射程外は状態不変", func(t *testing.T) {
				before := copyBattle(w.Battle)
				if err := w.BattleAttack(3, 2); !errors.Is(err, ErrOutOfRange) || !reflect.DeepEqual(*w.Battle, before) {
					t.Fatalf("error=%v battle=%+v", err, w.Battle)
				}
			})
			t.Run("TC-FR-BTL-01-22 倒れた敵は狙えない", func(t *testing.T) {
				before := copyBattle(w.Battle)
				if err := w.BattleAttack(2, 3); !errors.Is(err, ErrNoTarget) || !reflect.DeepEqual(*w.Battle, before) {
					t.Fatalf("error=%v battle=%+v", err, w.Battle)
				}
			})
		}
		t.Run(s.name, func(t *testing.T) {
			logStart := 0
			if w.Battle != nil {
				logStart = len(w.Battle.Log)
			}
			if err := s.action(); err != nil {
				t.Fatal(err)
			}
			assertBattle(t, w, s.order, s.turn, s.units)
			if got := w.Battle.Log[logStart:]; !reflect.DeepEqual(got, s.log) {
				t.Fatalf("記録=%v、期待=%v", got, s.log)
			}
		})
	}
	t.Run("TC-FR-BTL-01-11 四巡目の途中で勝つ", func(t *testing.T) {
		if err := w.BattleAttack(3, 2); err != nil {
			t.Fatal(err)
		}
		if w.Battle != nil || w.LastBattle == nil || !w.LastBattle.Won || w.LastBattle.GatherID != "field" || !reflect.DeepEqual(w.LastBattle.Rewards, []Stack{{Key: ItemKey{"herb", 40}, Count: 2}}) {
			t.Fatalf("結果=%+v、戦闘=%+v", w.LastBattle, w.Battle)
		}
		if got := w.LastBattle.Log[len(w.LastBattle.Log)-2:]; !reflect.DeepEqual(got, []string{"術士はスライムに5のダメージ", "勝利"}) {
			t.Fatalf("最後の行動=%v", got)
		}
	})
}

func TestBattleErrors(t *testing.T) {
	t.Run("TC-FR-BTL-10-01 開始の誤り順と状態保持", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		previous := &BattleReport{GatherID: "前回", Won: true, Log: []string{"前回の記録"}}
		w.LastBattle = previous
		w.InjuredUntil = 10
		cases := []struct {
			id   string
			want error
		}{{"unknown", ErrUnknownID}, {"empty", ErrNoBattle}, {"field", ErrInjured}}
		for _, c := range cases {
			if err := w.StartBattle(c.id); !errors.Is(err, c.want) || w.Battle != nil || !reflect.DeepEqual(w.LastBattle, previous) {
				t.Fatalf("%s: %v", c.id, err)
			}
		}
		w.InjuredUntil = 0
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if w.LastBattle != nil {
			t.Fatal("成功した開始で前回の結果が残った")
		}
		w.InjuredUntil = 10
		for _, c := range []struct {
			id   string
			want error
		}{{"unknown", ErrUnknownID}, {"empty", ErrNoBattle}, {"field", ErrInBattle}} {
			before := copyBattle(w.Battle)
			if err := w.StartBattle(c.id); !errors.Is(err, c.want) || !reflect.DeepEqual(*w.Battle, before) || w.LastBattle != nil {
				t.Fatalf("%s: %v battle=%+v", c.id, err, w.Battle)
			}
		}
	})
	t.Run("TC-FR-BTL-01-12 移動の誤り順", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		checks := []struct {
			x, y int
			want error
		}{{8, 2, ErrOutOfRange}, {3, 2, ErrOutOfRange}, {1, 2, ErrCellOccupied}}
		for _, c := range checks {
			before := copyBattle(w.Battle)
			if err := w.BattleMove(c.x, c.y); !errors.Is(err, c.want) || !reflect.DeepEqual(*w.Battle, before) {
				t.Fatalf("(%d,%d): %v", c.x, c.y, err)
			}
		}
		if err := w.BattleMove(1, 3); err != nil {
			t.Fatal(err)
		}
		before := copyBattle(w.Battle)
		if err := w.BattleMove(8, 2); !errors.Is(err, ErrAlreadyMoved) || !reflect.DeepEqual(*w.Battle, before) {
			t.Fatal(err)
		}
	})
	t.Run("TC-FR-BTL-01-13 攻撃の誤り順と状態保持", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			x, y int
			want error
		}{{8, 2, ErrOutOfRange}, {3, 2, ErrOutOfRange}, {0, 3, ErrNoTarget}} {
			before := copyBattle(w.Battle)
			if err := w.BattleAttack(c.x, c.y); !errors.Is(err, c.want) || !reflect.DeepEqual(*w.Battle, before) {
				t.Fatalf("(%d,%d): %v", c.x, c.y, err)
			}
		}
	})
	t.Run("TC-FR-BTL-01-14 戦闘外の操作", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		for _, op := range []func() error{func() error { return w.BattleMove(0, 0) }, func() error { return w.BattleAttack(0, 0) }, w.BattleWait} {
			if err := op(); !errors.Is(err, ErrNotInBattle) {
				t.Fatal(err)
			}
		}
	})
}

func TestBattleOrderAndEnemy(t *testing.T) {
	t.Run("TC-FR-BTL-01-17 倒れた敵の番を飛ばして次の巡を並べ直す", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[2].HP = 1
		w.Battle.Units[1].X, w.Battle.Units[1].Y = 3, 4
		if err := w.BattleAttack(3, 2); err != nil {
			t.Fatal(err)
		}
		if got := w.Battle.Log[len(w.Battle.Log)-1]; got != "術士はスライムに5のダメージ" || w.Battle.Turn != 2 {
			t.Fatalf("記録=%v 番=%d", w.Battle.Log, w.Battle.Turn)
		}
		logStart := len(w.Battle.Log)
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{3, 1, 0}) || w.Battle.Turn != 1 || w.Battle.Units[2].HP != 0 {
			t.Fatalf("戦闘=%+v", w.Battle)
		}
		if got := w.Battle.Log[logStart:]; !reflect.DeepEqual(got, []string{"剣士は待機", "オオカミは(1, 1)へ移動", "オオカミは剣士に5のダメージ"}) {
			t.Fatalf("倒れたスライムの行動が記録にある: %v", got)
		}
	})
	t.Run("TC-FR-BTL-01-18 敵の移動先が同点なら y 次に x", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 1, Y: 2}, {Unit: "mage", X: 7, Y: 5}}
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 2, Y: 3}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if u := w.Battle.Units[2]; u.X != 1 || u.Y != 1 || w.Battle.Units[0].HP != 25 {
			t.Fatalf("units=%+v", w.Battle.Units)
		}
		if !reflect.DeepEqual(w.Battle.Log, []string{"オオカミは(1, 1)へ移動", "オオカミは剣士に5のダメージ"}) {
			t.Fatalf("記録=%v", w.Battle.Log)
		}
	})
	t.Run("TC-FR-BTL-01-19 近い空きマスがなければ動かない", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 2, Y: 0}, {Unit: "mage", X: 7, Y: 5}}
		wolf := c.Units["wolf"]
		wolf.Move = 1
		c.Units["wolf"] = wolf
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 0, Y: 0}, {Unit: "slime", X: 1, Y: 0}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if u := w.Battle.Units[2]; u.X != 0 || u.Y != 0 || len(w.Battle.Log) != 0 {
			t.Fatalf("units=%+v log=%v", w.Battle.Units, w.Battle.Log)
		}
	})
	t.Run("TC-FR-BTL-01-20 同距離なら置いた順の仲間を狙う", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 1, Y: 2}, {Unit: "mage", X: 3, Y: 0}}
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 3, Y: 3}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if u := w.Battle.Units[2]; u.X != 2 || u.Y != 2 || w.Battle.Units[0].HP != 25 || w.Battle.Units[1].HP != 18 {
			t.Fatalf("units=%+v", w.Battle.Units)
		}
	})
	t.Run("TC-FR-BTL-01-21 移動後も最初の狙いだけを攻撃", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 4, Y: 0}, {Unit: "mage", X: 4, Y: 1}}
		wolf := c.Units["wolf"]
		wolf.Range = 2
		c.Units["wolf"] = wolf
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 0, Y: 0}, {Unit: "slime", X: 1, Y: 0}, {Unit: "slime", X: 2, Y: 0}, {Unit: "slime", X: 3, Y: 0}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if u := w.Battle.Units[2]; u.X != 2 || u.Y != 1 || w.Battle.Units[0].HP != 30 || w.Battle.Units[1].HP != 18 {
			t.Fatalf("units=%+v", w.Battle.Units)
		}
		if !reflect.DeepEqual(w.Battle.Log, []string{"オオカミは(2, 1)へ移動"}) {
			t.Fatalf("記録=%v", w.Battle.Log)
		}
	})
	t.Run("TC-FR-BTL-01-16 素早さ変更は次の巡から", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[0].SpeedBonus = 5
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{3, 1, 0, 2}) || w.Battle.Turn != 2 {
			t.Fatalf("巡中=%v 番=%d", w.Battle.Order, w.Battle.Turn)
		}
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{0, 3, 1, 2}) || w.Battle.Turn != 0 {
			t.Fatalf("次巡=%v 番=%d", w.Battle.Order, w.Battle.Turn)
		}
	})
	t.Run("TC-FR-BTL-02-01 同速は仲間と配置順を優先", func(t *testing.T) {
		c := battleCatalog()
		m := c.Units["mage"]
		m.Speed = 5
		c.Units["mage"] = m
		s := c.Units["slime"]
		s.Speed = 5
		c.Units["slime"] = s
		wolf := c.Units["wolf"]
		wolf.Speed = 5
		c.Units["wolf"] = wolf
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{0, 1, 2, 3}) || w.Battle.Turn != 0 {
			t.Fatalf("%v %d", w.Battle.Order, w.Battle.Turn)
		}
	})
	t.Run("TC-FR-BTL-01-15 敵の移動先と同距離の仲間", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 2, Y: 2}, {Unit: "mage", X: 0, Y: 3}}
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 6, Y: 2}}, Rolls: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if u := w.Battle.Units[2]; u.X != 3 || u.Y != 2 || w.Battle.Units[0].HP != 25 {
			t.Fatalf("units=%+v", w.Battle.Units)
		}
	})
	t.Run("TC-FR-BTL-02-02 同名の敵の表示名", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 3, Y: 2}, {Unit: "slime", X: 4, Y: 2}}, Rolls: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[2].Label != "スライム 1" || w.Battle.Units[3].Label != "スライム 2" {
			t.Fatalf("%+v", w.Battle.Units)
		}
	})
	t.Run("TC-FR-BTL-02-03 敵が一体なら番号を付けない", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 3, Y: 2}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if got := w.Battle.Units[2].Label; got != "スライム" {
			t.Fatalf("表示名=%q", got)
		}
	})
	t.Run("TC-FR-BTL-01-23 攻撃力の加算をダメージに反映", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[1].AttackBonus = 4
		if err := w.BattleAttack(2, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[3].HP != 9 || w.Battle.Log[len(w.Battle.Log)-1] != "術士はオオカミに9のダメージ" {
			t.Fatalf("units=%+v log=%v", w.Battle.Units, w.Battle.Log)
		}
	})
	t.Run("TC-FR-BTL-01-24 敵は倒れた仲間を狙わない", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[0].HP = 0
		start := len(w.Battle.Log)
		w.enemyAction(3)
		if w.Battle.Units[0].HP != 0 || w.Battle.Units[1].HP != 13 {
			t.Fatalf("units=%+v", w.Battle.Units)
		}
		if got := w.Battle.Log[start:]; !reflect.DeepEqual(got, []string{"オオカミは(0, 2)へ移動", "オオカミは術士に5のダメージ"}) {
			t.Fatalf("記録=%v", got)
		}
	})
}

func TestBattleLossAndRewards(t *testing.T) {
	t.Run("TC-FR-BTL-07-03 仲間の HP は 0 で止める", func(t *testing.T) {
		c := battleCatalog()
		s := c.Units["sword"]
		s.HP = 4
		c.Units["sword"] = s
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 2, Y: 2}}}}
		w := NewWorld(c, 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if w.Battle == nil || w.Battle.Units[0].HP != 0 || w.Battle.Units[1].HP != 18 || !reflect.DeepEqual(w.Battle.Log, []string{"オオカミは剣士に5のダメージ"}) {
			t.Fatalf("戦闘=%+v", w.Battle)
		}
	})
	t.Run("TC-FR-BTL-07-01 負けた時刻と回復境界", func(t *testing.T) {
		c := battleCatalog()
		c.Game.Party = []defs.Placement{{Unit: "sword", X: 1, Y: 2}}
		s := c.Units["sword"]
		s.HP = 4
		c.Units["sword"] = s
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "wolf", X: 2, Y: 2}}, Rolls: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		w.Now = 1000
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		if w.Battle != nil || w.LastBattle == nil || w.LastBattle.Won || w.InjuredUntil != 4600 || w.LastBattle.InjuredUntil != 4600 || len(w.LastBattle.Rewards) != 0 || !reflect.DeepEqual(w.LastBattle.Log, []string{"オオカミは剣士に5のダメージ", "敗北"}) {
			t.Fatalf("結果=%+v 負傷=%d", w.LastBattle, w.InjuredUntil)
		}
		for _, now := range []uint64{1001, 4599} {
			w.Now = now
			if err := w.StartBattle("field"); !errors.Is(err, ErrInjured) {
				t.Fatalf("Now=%d: %v", now, err)
			}
		}
		w.Now = 4600
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("TC-FR-BTL-07-02 勝利時に倒れた仲間は負傷しない", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[0].HP = 0
		w.Battle.Units[2].HP = 0
		w.Battle.Units[3].HP = 5
		if err := w.BattleAttack(2, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle != nil || !w.LastBattle.Won || w.InjuredUntil != 0 {
			t.Fatalf("結果=%+v 負傷=%d", w.LastBattle, w.InjuredUntil)
		}
	})
	t.Run("TC-FR-MAT-03-01 報酬は採取と同じ抽選で容量超過を破棄", func(t *testing.T) {
		c := battleCatalog()
		c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 1, Y: 3}}, Rolls: 2, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
		w := NewWorld(c, 1)
		key := ItemKey{"herb", 40}
		if err := w.Inventory.Add(key, int(^uint(0)>>1)); err != nil {
			t.Fatal(err)
		}
		if err := w.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		w.Battle.Units[2].HP = 1
		if err := w.BattleAttack(1, 3); err != nil {
			t.Fatal(err)
		}
		if w.Inventory.Count(key) != int(^uint(0)>>1) || len(w.LastBattle.Rewards) != 0 || !strings.Contains(strings.Join(w.LastBattle.Log, " "), "薬草（品質40）の報酬2個を破棄") {
			t.Fatalf("結果=%+v", w.LastBattle)
		}
		partial := NewWorld(c, 1)
		if err := partial.Inventory.Add(key, int(^uint(0)>>1)-1); err != nil {
			t.Fatal(err)
		}
		if err := partial.StartBattle("field"); err != nil {
			t.Fatal(err)
		}
		partial.Battle.Units[2].HP = 1
		if err := partial.BattleAttack(1, 3); err != nil {
			t.Fatal(err)
		}
		if partial.Inventory.Count(key) != int(^uint(0)>>1) || !reflect.DeepEqual(partial.LastBattle.Rewards, []Stack{{Key: key, Count: 1}}) || !strings.Contains(strings.Join(partial.LastBattle.Log, " "), "薬草（品質40）の報酬1個を破棄") {
			t.Fatalf("一部受入=%+v", partial.LastBattle)
		}
	})
	t.Run("TC-NFR-06-20 同じシードの報酬と敵の動き", func(t *testing.T) {
		a, b := NewWorld(battleCatalog(), 42), NewWorld(battleCatalog(), 42)
		for _, w := range []*World{a, b} {
			if err := w.StartBattle("field"); err != nil {
				t.Fatal(err)
			}
			w.Battle.Units[2].HP = 0
			w.Battle.Units[3].HP = 1
			if err := w.BattleAttack(2, 3); err != nil {
				t.Fatal(err)
			}
		}
		if !reflect.DeepEqual(a.LastBattle, b.LastBattle) {
			t.Fatalf("%+v / %+v", a.LastBattle, b.LastBattle)
		}
	})
}
