package core

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func dayQuestCatalog() defs.Catalog {
	c := questCatalog("q1", "q2", "q3", "q4", "q5", "q6")
	c.Game.DaySeconds = 1
	for id, q := range c.Quests {
		q.Rank = 1
		q.Reputation = 7
		c.Quests[id] = q
	}
	return c
}

func TestDayExamples(t *testing.T) {
	for _, tc := range []struct {
		name    string
		seconds int
		now     uint64
		ticks   int
		wantNow uint64
		day     int
		days    []int
	}{
		{"TC-FR-PRG-05-01 D-1 初日", 1, 0, 0, 0, 1, nil},
		{"TC-FR-PRG-05-02 D-2 日替わり直前", 1, 0, 59, 59, 1, nil},
		{"TC-FR-PRG-05-03 D-3 日替わり", 1, 59, 1, 60, 2, []int{2}},
		{"TC-FR-PRG-05-04 D-4 複数日の知らせ", 1, 0, 120, 120, 3, []int{2, 3}},
		{"TC-FR-PRG-05-05 D-5 未設定", 0, 0, 1000, 1000, 1, nil},
		{"TC-FR-PRG-05-06 300秒の日の境界", 300, 17999, 1, 18000, 2, []int{2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := dayQuestCatalog()
			c.Game.DaySeconds = tc.seconds
			w := NewWorld(c, 42)
			w.Now = tc.now
			before := w.Quests
			for i := 0; i < tc.ticks; i++ {
				w.Tick()
			}
			if w.Now != tc.wantNow || w.Day() != tc.day || !reflect.DeepEqual(w.TakeNewDays(), tc.days) || len(w.TakeNewDays()) != 0 {
				t.Fatalf("Now=%d 日=%d 知らせの期待=%v", w.Now, w.Day(), tc.days)
			}
			if tc.seconds == 0 && w.Quests != before {
				t.Fatal("未設定で依頼板が変化")
			}
		})
	}
}

func TestQuestExamples(t *testing.T) {
	for _, tc := range []struct {
		name              string
		setup             func(*World)
		operation         func(*World) error
		want              error
		board             [3]string
		accepted          []string
		money, reputation int
	}{
		{name: "TC-FR-ECO-06-01 Q-1 初期候補"},
		{name: "TC-FR-ECO-06-02 Q-2 受注", operation: func(w *World) error { return w.AcceptQuest(1) }, board: [3]string{"q1", "", "q3"}, accepted: []string{"q2"}},
		{name: "TC-FR-ECO-06-03 Q-3 同じ位置を再受注", setup: func(w *World) {
			if err := w.AcceptQuest(1); err != nil {
				t.Fatal(err)
			}
		}, operation: func(w *World) error { return w.AcceptQuest(1) }, want: ErrNoQuest},
		{name: "TC-FR-ECO-06-04 Q-4 空きの検査が先", setup: func(w *World) { w.Accepted = []string{"q2", "q1", "q3"}; w.Quests = [3]string{} }, operation: func(w *World) error { return w.AcceptQuest(0) }, want: ErrNoQuest},
		{name: "TC-FR-ECO-06-05 Q-5 上限", setup: func(w *World) { w.Accepted = []string{"q2", "q1", "q3"}; w.Quests = [3]string{"q4", "", ""} }, operation: func(w *World) error { return w.AcceptQuest(0) }, want: ErrQuestLimit},
		{name: "TC-FR-ECO-06-06 Q-6 範囲外", operation: func(w *World) error { return w.AcceptQuest(3) }, want: ErrNoQuest},
		{name: "TC-FR-ECO-05-14 Q-7 受注順の納品", setup: func(w *World) {
			w.Accepted = []string{"q2", "q1"}
			w.Quests = [3]string{"q4", "q5", "q6"}
			addQuestItems(t, w, 20, 40)
		}, operation: func(w *World) error { return w.DeliverQuest(0) }, board: [3]string{"q4", "q5", "q6"}, accepted: []string{"q1"}, money: 220, reputation: 7},
		{name: "TC-FR-ECO-05-15 Q-8 範囲外の納品", setup: func(w *World) { w.Accepted = []string{"q2", "q1"} }, operation: func(w *World) error { return w.DeliverQuest(2) }, want: ErrNoQuest},
		{name: "TC-FR-ECO-05-16 Q-9 在庫不足", setup: func(w *World) { w.Accepted = []string{"q2", "q1"} }, operation: func(w *World) error { return w.DeliverQuest(1) }, want: ErrNotEnoughItems},
		{name: "TC-FR-ECO-06-07 Q-10 やめる", setup: func(w *World) { w.Accepted = []string{"q2", "q1"} }, operation: func(w *World) error { return w.AbandonQuest(0) }, board: [3]string{"q1", "q2", "q3"}, accepted: []string{"q1"}, money: 100},
		{name: "TC-FR-ECO-06-08 Q-11 未受注をやめる", operation: func(w *World) error { return w.AbandonQuest(0) }, want: ErrNoQuest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(dayQuestCatalog(), 42)
			control := NewWorld(dayQuestCatalog(), 42)
			if tc.operation == nil {
				seen := map[string]bool{}
				for _, id := range w.Quests {
					if id == "" || seen[id] {
						t.Fatalf("候補=%v", w.Quests)
					}
					seen[id] = true
				}
				if len(w.Accepted) != 0 {
					t.Fatal(w.Accepted)
				}
				return
			}
			w.Quests = [3]string{"q1", "q2", "q3"}
			w.Money = 100
			if tc.setup != nil {
				tc.setup(w)
			}
			before := questState(w)
			err := tc.operation(w)
			if w.questRNG.Uint64() != control.questRNG.Uint64() {
				t.Fatal("受注・中止・納品で依頼乱数が進んだ")
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("誤り=%v 期待=%v", err, tc.want)
			}
			if tc.want != nil {
				if !reflect.DeepEqual(questState(w), before) {
					t.Fatal("失敗で状態が変化")
				}
				return
			}
			if w.Quests != tc.board || !reflect.DeepEqual(w.Accepted, tc.accepted) {
				t.Fatalf("候補=%v 受注=%v", w.Quests, w.Accepted)
			}
			if tc.money != 0 && (w.Money != tc.money || w.Reputation != tc.reputation) {
				t.Fatalf("お金=%d 評判=%d", w.Money, w.Reputation)
			}
			if tc.money == 100 && !reflect.DeepEqual(w.Inventory.Stacks(), before.items) {
				t.Fatal("やめる操作で在庫が変化")
			}
		})
	}
}

// questState は失敗時に維持する値と乱数の状態を写す（DES-WLD-01）。
type questSnapshot struct {
	board             [3]string
	accepted          []string
	items             []Stack
	money, reputation int
	now               uint64
	rankUps, days     []int
	rng               []byte
}

func questState(w *World) questSnapshot {
	main, _ := w.rngSource.MarshalBinary()
	// Rand からソースは取り出せないので、依頼乱数は別途同じ操作列との比較で確かめる。
	return questSnapshot{board: w.Quests, accepted: append([]string(nil), w.Accepted...), items: w.Inventory.Stacks(), money: w.Money, reputation: w.Reputation, now: w.Now, rankUps: append([]int(nil), w.rankUps...), days: append([]int(nil), w.newDays...), rng: main}
}

func TestDailyQuestExamples(t *testing.T) {
	for _, tc := range []struct {
		name      string
		count     int
		accepted  []string
		board     [3]string
		wantCount int
	}{
		{"TC-FR-ECO-07-01 Q-12 受注3件を維持", 6, []string{"q1", "q2", "q3"}, [3]string{}, 3},
		{"TC-FR-ECO-07-02 Q-13 残り1件", 4, []string{"q1", "q2", "q3"}, [3]string{}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := dayQuestCatalog()
			if tc.count == 4 {
				delete(c.Quests, "q5")
				delete(c.Quests, "q6")
			}
			w := NewWorld(c, 42)
			w.Accepted = append([]string(nil), tc.accepted...)
			w.Quests = tc.board
			w.Now = 59
			w.Tick()
			seen := map[string]bool{}
			for _, id := range w.Accepted {
				seen[id] = true
			}
			for i, id := range w.Quests {
				if i >= tc.wantCount {
					if id != "" {
						t.Fatal(w.Quests)
					}
					continue
				}
				if id == "" || seen[id] {
					t.Fatalf("候補=%v 受注=%v", w.Quests, w.Accepted)
				}
				seen[id] = true
			}
			if !reflect.DeepEqual(w.Accepted, tc.accepted) {
				t.Fatal("受注が変化")
			}
		})
	}
	t.Run("TC-FR-ECO-07-03 Q-14 全位置を引き直す", func(t *testing.T) {
		w := NewWorld(dayQuestCatalog(), 42)
		w.Quests = [3]string{"q1", "", ""}
		w.Now = 59

		rng := rand.New(rand.NewPCG(42, 1))
		// NewWorld の初期抽選 3 回分を進める（DES-QST-04）。
		for count := 6; count > 3; count-- {
			rng.IntN(count)
		}
		ids := []string{"q1", "q2", "q3", "q4", "q5", "q6"}
		var want [3]string
		for i := range want {
			n := rng.IntN(len(ids))
			want[i] = ids[n]
			ids = append(ids[:n], ids[n+1:]...)
		}
		w.Tick()
		if w.Quests != want {
			t.Fatalf("候補 = %v、全位置を引き直した期待 = %v", w.Quests, want)
		}
		if len(w.Accepted) != 0 {
			t.Fatalf("受注が変化: %v", w.Accepted)
		}
	})
	t.Run("TC-FR-ECO-07-04 Q-15 同じシードと入力", func(t *testing.T) {
		a, b := NewWorld(dayQuestCatalog(), 42), NewWorld(dayQuestCatalog(), 42)
		for day := 0; day < 5; day++ {
			for _, w := range []*World{a, b} {
				if len(w.Accepted) < MaxAcceptedQuests {
					if err := w.AcceptQuest(0); err != nil {
						t.Fatal(err)
					}
				}
				if day%2 == 1 {
					if err := w.AbandonQuest(0); err != nil {
						t.Fatal(err)
					}
				}
				for i := 0; i < 60; i++ {
					w.Tick()
				}
			}
			if a.Quests != b.Quests || !reflect.DeepEqual(a.Accepted, b.Accepted) {
				t.Fatal("再現しない")
			}
		}
	})
}

func TestQuestInvalidIndices(t *testing.T) {
	for _, tc := range []struct {
		name      string
		operation func(*World, int) error
		accepted  []string
		indices   []int
	}{
		{"TC-FR-ECO-06-09 受注の範囲外", (*World).AcceptQuest, nil, []int{-1, 3}},
		{"TC-FR-ECO-05-17 納品の範囲外", (*World).DeliverQuest, []string{"q1"}, []int{-1, 1, 3}},
		{"TC-FR-ECO-06-10 中止の範囲外", (*World).AbandonQuest, []string{"q1"}, []int{-1, 1, 3}},
		{"TC-FR-ECO-05-18 候補へ直接納品できない", (*World).DeliverQuest, nil, []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, index := range tc.indices {
				w, control := NewWorld(dayQuestCatalog(), 42), NewWorld(dayQuestCatalog(), 42)
				w.Accepted = append([]string(nil), tc.accepted...)
				w.rankUps = []int{2}
				w.newDays = []int{2}
				addQuestItems(t, w, 20, 40)
				before := questState(w)
				if err := tc.operation(w, index); !errors.Is(err, ErrNoQuest) {
					t.Fatalf("位置%d: %v", index, err)
				}
				if !reflect.DeepEqual(questState(w), before) || w.questRNG.Uint64() != control.questRNG.Uint64() {
					t.Fatal("失敗で状態または依頼乱数が変化")
				}
			}
		})
	}
}

func TestQuestRemoval(t *testing.T) {
	for _, tc := range []struct {
		name      string
		operation func(*World, int) error
	}{
		{"TC-FR-ECO-05-19 中央と末尾の納品", (*World).DeliverQuest},
		{"TC-FR-ECO-06-11 中央と末尾の中止", (*World).AbandonQuest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, index := range []int{1, 2} {
				w := NewWorld(dayQuestCatalog(), 42)
				w.Quests = [3]string{"q4", "q5", "q6"}
				w.Accepted = []string{"q1", "q2", "q3"}
				addQuestItems(t, w, 20, 40)
				before := questState(w)
				if err := tc.operation(w, index); err != nil {
					t.Fatal(err)
				}
				want := []string{"q1", "q3"}
				if index == 2 {
					want = []string{"q1", "q2"}
				}
				if !reflect.DeepEqual(w.Accepted, want) || w.Quests != before.board {
					t.Fatalf("受注=%v 候補=%v", w.Accepted, w.Quests)
				}
				if tc.name == "TC-FR-ECO-06-11 中央と末尾の中止" && (w.Money != before.money || w.Reputation != before.reputation || !reflect.DeepEqual(w.Inventory.Stacks(), before.items)) {
					t.Fatal("中止で資産が変化")
				}
			}
		})
	}
}

func TestDailyQuestReturn(t *testing.T) {
	t.Run("TC-FR-ECO-07-05 納品・中止した依頼は翌日から候補", func(t *testing.T) {
		for _, operation := range []func(*World, int) error{(*World).DeliverQuest, (*World).AbandonQuest} {
			c := questCatalog("q1")
			c.Game.DaySeconds = 1
			w := NewWorld(c, 42)
			if err := w.AcceptQuest(0); err != nil {
				t.Fatal(err)
			}
			addQuestItems(t, w, 20, 40)
			if err := operation(w, 0); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 59; i++ {
				w.Tick()
			}
			if w.Quests != [3]string{} {
				t.Fatal("同日中に補充")
			}
			w.Tick()
			if w.Quests != [3]string{"q1", "", ""} {
				t.Fatal(w.Quests)
			}
		}
	})
	t.Run("TC-FR-ECO-07-06 日替わりと本体乱数の独立", func(t *testing.T) {
		a, b := NewWorld(dayQuestCatalog(), 42), NewWorld(dayQuestCatalog(), 42)
		control := NewWorld(dayQuestCatalog(), 42)
		for i := 0; i < 100; i++ {
			a.rng.Uint64()
		}
		for i := 0; i < 120; i++ {
			a.Tick()
			b.Tick()
		}
		if a.Quests != b.Quests {
			t.Fatal("本体乱数が依頼に影響")
		}
		if b.rng.Uint64() != control.rng.Uint64() {
			t.Fatal("日替わりで本体乱数が進んだ")
		}
	})
}

func TestDayTickBatch(t *testing.T) {
	t.Run("TC-FR-PRG-05-07 Tickを10回ずつ呼んでも同じ", func(t *testing.T) {
		c := golemCatalog()
		c.Game.DaySeconds = 1
		c.Quests = dayQuestCatalog().Quests
		a, b := NewWorld(c, 42), NewWorld(c, 42)
		for _, w := range []*World{a, b} {
			buildWorker(t, w)
			addGolemStock(t, w, craftStack("herb", 40, 20), craftStack("water", 31, 10))
			assignPotion(t, w, 0)
		}
		for i := 0; i < 120; i++ {
			a.Tick()
		}
		for update := 0; update < 12; update++ {
			for i := 0; i < 10; i++ {
				b.Tick()
			}
		}
		if !reflect.DeepEqual(questState(a), questState(b)) || !reflect.DeepEqual(a.Golems, b.Golems) || a.Day() != 3 || !reflect.DeepEqual(a.TakeNewDays(), []int{2, 3}) {
			t.Fatal("まとめたTickの結果が違う")
		}
		if a.Inventory.Count(ItemKey{"potion", 37}) == 0 {
			t.Fatal("日替わりをまたぐ生産が行われていない")
		}
	})
}
