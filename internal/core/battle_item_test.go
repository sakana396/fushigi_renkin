package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// itemBattle は敵が先に行動しない戦闘を用意する（DES-BTL-02）。
func itemBattle(t *testing.T) *World {
	t.Helper()
	c := battleCatalog()
	bomb := testBombUse()
	bomb.Name = "爆弾"
	c.Items["bomb"] = bomb
	burger := testBurgerUse()
	burger.Name = "ハンバーガー"
	c.Items["burger"] = burger
	for _, id := range []string{"slime", "wolf"} {
		u := c.Units[id]
		u.Speed, u.Move, u.Attack = 1, 0, 0
		c.Units[id] = u
	}
	c.Gathers["field"] = defs.GatherDef{ID: "field", Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "slime", X: 2, Y: 3}, {Unit: "wolf", X: 3, Y: 3}}, Rolls: 2, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}}}
	w := NewWorld(c, 1)
	if err := w.StartBattle("field"); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestBattleItemBombQuality(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		quality, radius, damage int
	}{
		{"TC-FR-BTL-05-01 品質20の爆弾", 20, 0, 8},
		{"TC-FR-BTL-05-02 品質60の爆弾", 60, 1, 24},
		{"TC-FR-BTL-05-03 品質120の爆弾", 120, 1, 72},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := itemBattle(t)
			w.Battle.Units[0].X, w.Battle.Units[0].Y = 1, 3 // 範囲内の仲間
			w.Battle.Units[2].HP, w.Battle.Units[3].HP = 100, 100
			key := ItemKey{"bomb", tc.quality}
			if err := w.Inventory.Add(key, 2); err != nil {
				t.Fatal(err)
			}
			battle := w.Battle
			if err := w.BattleUseItem(key, 2, 3); err != nil {
				t.Fatal(err)
			}
			if got := w.Inventory.Count(key); got != 1 {
				t.Fatalf("在庫=%d", got)
			}
			if battle.Units[0].HP != 30 {
				t.Fatalf("仲間への巻き込み: %+v", battle.Units[0])
			}
			if got := battle.Units[2].HP; got != 100-tc.damage {
				t.Fatalf("中心のHP=%d", got)
			}
			want := 100
			if tc.radius == 1 {
				want = 100 - tc.damage
			}
			if got := battle.Units[3].HP; got != want {
				t.Fatalf("周囲のHP=%d、期待=%d", got, want)
			}
			wantLog := []string{fmt.Sprintf("術士は爆弾（品質%d）を使用", tc.quality), fmt.Sprintf("スライムに%dのダメージ", tc.damage)}
			if tc.radius == 1 {
				wantLog = append(wantLog, fmt.Sprintf("オオカミに%dのダメージ", tc.damage))
			}
			if !reflect.DeepEqual(battle.Log, wantLog) {
				t.Fatalf("記録=%v、期待=%v", battle.Log, wantLog)
			}
			if w.Battle == nil || w.Battle.Turn != 1 {
				t.Fatalf("番が進まない: %+v", w.Battle)
			}
		})
	}
}

func TestBattleItemBombVictory(t *testing.T) {
	t.Run("TC-FR-BTL-04-14 爆弾で敵が全滅すると即時勝利", func(t *testing.T) {
		w := itemBattle(t)
		key := ItemKey{"bomb", 60}
		if err := w.Inventory.Add(key, 1); err != nil {
			t.Fatal(err)
		}
		battle := w.Battle
		if err := w.BattleUseItem(key, 2, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle != nil || w.LastBattle == nil || !w.LastBattle.Won || battle.Units[2].HP != 0 || battle.Units[3].HP != 0 || w.Inventory.Count(key) != 0 {
			t.Fatalf("結果=%+v 戦闘=%+v", w.LastBattle, battle)
		}
		wantRewards := []Stack{{Key: ItemKey{"herb", 40}, Count: 2}}
		if !reflect.DeepEqual(w.LastBattle.Rewards, wantRewards) || w.Inventory.Count(ItemKey{"herb", 40}) != 2 {
			t.Fatalf("報酬=%v 在庫=%v", w.LastBattle.Rewards, w.Inventory.Stacks())
		}
		if got := w.LastBattle.Log[len(w.LastBattle.Log)-1]; got != "勝利" {
			t.Fatalf("記録=%v", w.LastBattle.Log)
		}
	})
}

func TestBattleItemBurgerQuality(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		quality, hp, attack, speed int
	}{
		{"TC-FR-BTL-05-04 品質30の回復", 30, 14, 0, 0},
		{"TC-FR-BTL-05-05 品質100の回復と攻撃", 100, 18, 5, 0},
		{"TC-FR-BTL-05-06 品質160の回復と攻撃と素早さ", 160, 18, 8, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := itemBattle(t)
			w.Battle.Units[1].HP = 5
			key := ItemKey{"burger", tc.quality}
			if err := w.Inventory.Add(key, 1); err != nil {
				t.Fatal(err)
			}
			if err := w.BattleUseItem(key, 0, 3); err != nil {
				t.Fatal(err)
			}
			u := w.Battle.Units[1]
			if u.HP != tc.hp || u.AttackBonus != tc.attack || u.SpeedBonus != tc.speed {
				t.Fatalf("効果=%+v", u)
			}
			wantLog := []string{fmt.Sprintf("術士はハンバーガー（品質%d）を使用", tc.quality), fmt.Sprintf("術士の体力が%d回復", tc.hp-5)}
			if tc.attack > 0 {
				wantLog = append(wantLog, fmt.Sprintf("術士の攻撃が%d上がった", tc.attack))
			}
			if tc.speed > 0 {
				wantLog = append(wantLog, fmt.Sprintf("術士の素早さが%d上がった", tc.speed))
			}
			if !reflect.DeepEqual(w.Battle.Log, wantLog) {
				t.Fatalf("記録=%v、期待=%v", w.Battle.Log, wantLog)
			}
			if w.Inventory.Count(key) != 0 || w.Battle.Turn != 1 {
				t.Fatalf("在庫・番: %d %+v", w.Inventory.Count(key), w.Battle)
			}
		})
	}
}

func TestBattleItemEffects(t *testing.T) {
	t.Run("TC-FR-BTL-04-01 回復上限と倒れた仲間", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[1].HP = 17
		w.Battle.Units[0].HP = 0
		w.Battle.Units[0].X, w.Battle.Units[0].Y = 0, 2
		w.Battle.Units[2].X, w.Battle.Units[2].Y = 0, 4 // 範囲内の敵
		w.Battle.Units[2].HP = 10
		burger := testBurgerUse()
		burger.Use.Stages[0].Radius = 1
		w.catalog.Items["burger"] = burger
		key := ItemKey{"burger", 100}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[1].HP != 18 || w.Battle.Units[0].HP != 0 || w.Battle.Units[0].AttackBonus != 0 || w.Battle.Units[2].HP != 10 {
			t.Fatalf("効果=%+v", w.Battle.Units)
		}
	})
	t.Run("TC-FR-BTL-04-02 攻撃強化は次の攻撃に加算", func(t *testing.T) {
		w := itemBattle(t)
		key := ItemKey{"burger", 100}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleAttack(2, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[2].HP != 4 {
			t.Fatalf("攻撃後HP=%d", w.Battle.Units[2].HP)
		}
	})
	t.Run("TC-FR-BTL-04-03 強化は重なり戦闘中続く", func(t *testing.T) {
		w := itemBattle(t)
		key := ItemKey{"burger", 160}
		_ = w.Inventory.Add(key, 2)
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[1].AttackBonus != 8 || w.Battle.Units[1].SpeedBonus != 4 {
			t.Fatalf("巡後=%+v", w.Battle.Units[1])
		}
		if err := w.BattleUseItem(key, 0, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[1].AttackBonus != 16 || w.Battle.Units[1].SpeedBonus != 8 {
			t.Fatalf("重ね掛け=%+v", w.Battle.Units[1])
		}
	})
	t.Run("TC-FR-BTL-04-04 素早さは次の巡から", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[0].X, w.Battle.Units[0].Y = 1, 3
		key := ItemKey{"burger", 160}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 1, 3); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{1, 0, 2, 3}) || w.Battle.Turn != 1 {
			t.Fatalf("巡中=%v/%d", w.Battle.Order, w.Battle.Turn)
		}
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{0, 1, 2, 3}) || w.Battle.Turn != 0 {
			t.Fatalf("次巡=%v/%d", w.Battle.Order, w.Battle.Turn)
		}
	})
	t.Run("TC-FR-BTL-04-05 倒れた敵と範囲外の敵は効果を受けない", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[3].HP = 0
		w.Battle.Units = append(w.Battle.Units, BattleUnit{DefID: "wolf", Ally: false, HP: 18, X: 7, Y: 5})
		key := ItemKey{"bomb", 60}
		_ = w.Inventory.Add(key, 1)
		if err := w.BattleUseItem(key, 2, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[3].HP != 0 || w.Battle.Units[3].AttackBonus != 0 || w.Battle.Units[3].SpeedBonus != 0 || w.Battle.Units[4].HP != 18 {
			t.Fatalf("対象外=%+v", w.Battle.Units)
		}
	})
	t.Run("TC-FR-BTL-04-15 爆弾の射程ちょうど三マス", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[2].X = 3
		w.Battle.Units[3].X = 4
		key := ItemKey{"bomb", 20}
		if err := w.Inventory.Add(key, 1); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleUseItem(key, 3, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[2].HP != 6 || w.Battle.Units[3].HP != 18 || w.Inventory.Count(key) != 0 || w.Battle.Turn != 1 {
			t.Fatalf("射程境界の効果=%+v", w.Battle)
		}
	})
	t.Run("TC-FR-BTL-04-16 移動後の使用と敵の番", func(t *testing.T) {
		w := itemBattle(t)
		slime := w.catalog.Units["slime"]
		slime.Attack = 3
		w.catalog.Units["slime"] = slime
		if err := w.BattleWait(); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleMove(1, 3); err != nil {
			t.Fatal(err)
		}
		key := ItemKey{"burger", 160}
		if err := w.Inventory.Add(key, 1); err != nil {
			t.Fatal(err)
		}
		logStart := len(w.Battle.Log)
		if err := w.BattleUseItem(key, 1, 3); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[0].HP != 27 || !reflect.DeepEqual(w.Battle.Log[logStart:], []string{"剣士はハンバーガー（品質160）を使用", "剣士の体力が0回復", "剣士の攻撃が8上がった", "剣士の素早さが4上がった", "スライムは剣士に3のダメージ"}) {
			t.Fatalf("敵の行動=%+v 記録=%v", w.Battle.Units[0], w.Battle.Log[logStart:])
		}
		if !reflect.DeepEqual(w.Battle.Order, []int{0, 1, 2, 3}) || w.Battle.Turn != 0 || w.Battle.Moved {
			t.Fatalf("次の巡=%+v", w.Battle)
		}
	})
	t.Run("TC-FR-BTL-04-17 剣士の最大体力で回復を止める", func(t *testing.T) {
		w := itemBattle(t)
		w.Battle.Units[0].HP = 25
		w.Battle.Units[1].Y = 2
		key := ItemKey{"burger", 100}
		if err := w.Inventory.Add(key, 1); err != nil {
			t.Fatal(err)
		}
		if err := w.BattleUseItem(key, 1, 2); err != nil {
			t.Fatal(err)
		}
		if w.Battle.Units[0].HP != 30 {
			t.Fatalf("剣士のHP=%d", w.Battle.Units[0].HP)
		}
	})
}

func TestBattleItemErrors(t *testing.T) {
	t.Run("TC-FR-BTL-04-06 戦闘外が最優先", func(t *testing.T) {
		w := NewWorld(battleCatalog(), 1)
		if err := w.BattleUseItem(ItemKey{"missing", 0}, -1, -1); !errors.Is(err, ErrNotInBattle) {
			t.Fatal(err)
		}
	})
	for _, tc := range []struct {
		name string
		key  ItemKey
		x, y int
		want error
		add  bool
	}{
		{"TC-FR-BTL-04-07 未定義の種類", ItemKey{"missing", 100}, -1, -1, ErrNotUsable, false},
		{"TC-FR-BTL-04-08 使用定義がない種類", ItemKey{"herb", 100}, -1, -1, ErrNotUsable, false},
		{"TC-FR-BTL-04-09 在庫不足が射程より先", ItemKey{"bomb", 100}, -1, -1, ErrNotEnoughItems, false},
		{"TC-FR-BTL-04-10 盤面外", ItemKey{"bomb", 100}, -1, 3, ErrOutOfRange, true},
		{"TC-FR-BTL-04-11 射程外", ItemKey{"bomb", 100}, 4, 3, ErrOutOfRange, true},
		{"TC-FR-BTL-04-12 対象の陣営が違う", ItemKey{"bomb", 100}, 1, 2, ErrNoTarget, true},
		{"TC-FR-BTL-04-13 対象が倒れている", ItemKey{"bomb", 100}, 2, 3, ErrNoTarget, true},
		{"TC-FR-BTL-04-18 x座標が同じ空きマス", ItemKey{"bomb", 100}, 2, 2, ErrNoTarget, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := itemBattle(t)
			if tc.add {
				_ = w.Inventory.Add(tc.key, 1)
			}
			if tc.x == 2 && tc.y == 3 && errors.Is(tc.want, ErrNoTarget) {
				w.Battle.Units[2].HP = 0
			}
			if tc.x == 4 {
				w.Battle.Units[3].X = 4
			}
			beforeBattle := copyBattle(w.Battle)
			beforeInventory := w.Inventory.Stacks()
			beforeReport := w.LastBattle
			if err := w.BattleUseItem(tc.key, tc.x, tc.y); !errors.Is(err, tc.want) {
				t.Fatalf("誤り=%v、期待=%v", err, tc.want)
			}
			if !reflect.DeepEqual(*w.Battle, beforeBattle) || !reflect.DeepEqual(w.Inventory.Stacks(), beforeInventory) || !reflect.DeepEqual(w.LastBattle, beforeReport) {
				t.Fatalf("失敗後に状態が変化: %+v", w.Battle)
			}
		})
	}
}
