package core

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"sort"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func questCatalog(ids ...string) defs.Catalog {
	quests := make(map[string]defs.QuestDef, len(ids))
	for _, id := range ids {
		quests[id] = defs.QuestDef{ID: id, Item: "burger", MinQuality: 20, Count: 2, Reward: 120}
	}
	return defs.Catalog{Quests: quests}
}

func addQuestItems(t *testing.T, w *World, qualities ...int) {
	t.Helper()
	for _, quality := range qualities {
		if err := w.Inventory.Add(ItemKey{"burger", quality}, 1); err != nil {
			t.Fatal(err)
		}
	}
}

func TestQuestBoard(t *testing.T) {
	t.Run("TC-FR-ECO-05-01 ID順の抽選", func(t *testing.T) {
		ids := []string{"q_d", "q_b", "q_a", "q_c"}
		w := NewWorld(questCatalog(ids...), 42)
		sort.Strings(ids)
		rng := rand.New(rand.NewPCG(42, 1))
		var want [3]string
		for i := range want {
			choice := rng.IntN(len(ids))
			want[i] = ids[choice]
			ids = append(ids[:choice], ids[choice+1:]...)
		}
		if w.Quests != want {
			t.Fatalf("依頼板 = %v、期待 = %v", w.Quests, want)
		}
	})
	for _, tc := range []struct {
		name string
		ids  []string
		want int
	}{
		{"0件", nil, 0}, {"1件", []string{"q_a"}, 1}, {"2件", []string{"q_b", "q_a"}, 2},
	} {
		t.Run("TC-FR-ECO-05-02 "+tc.name, func(t *testing.T) {
			w := NewWorld(questCatalog(tc.ids...), 42)
			seen := map[string]bool{}
			for i, id := range w.Quests {
				if i < tc.want {
					if id == "" || seen[id] {
						t.Fatalf("位置 %d: %v", i, w.Quests)
					}
					seen[id] = true
				} else if id != "" {
					t.Fatalf("余分な依頼: %v", w.Quests)
				}
			}
		})
	}
	t.Run("TC-FR-ECO-05-10 納品後の候補順と乱数位置", func(t *testing.T) {
		w := NewWorld(questCatalog("q_d", "q_b", "q_a", "q_c"), 7)
		rng := rand.New(rand.NewPCG(7, 1))
		ids := []string{"q_a", "q_b", "q_c", "q_d"}
		var initial [3]string
		for i := range initial {
			choice := rng.IntN(len(ids))
			initial[i] = ids[choice]
			ids = append(ids[:choice], ids[choice+1:]...)
		}
		candidates := []string{initial[1], ids[0]}
		sort.Strings(candidates)
		want := candidates[rng.IntN(len(candidates))]
		if want == initial[1] {
			t.Fatalf("テストの前提違反: 補充 ID %s が納品前と同じ", want)
		}
		if w.Quests != initial {
			t.Fatalf("初期依頼板 = %v、期待 = %v", w.Quests, initial)
		}
		addQuestItems(t, w, 20, 40)
		if err := w.Deliver(1); err != nil || w.Quests[1] != want || w.Quests[0] != initial[0] || w.Quests[2] != initial[2] {
			t.Fatalf("納品後の依頼板 = %v、期待する位置 1 = %s、誤り = %v", w.Quests, want, err)
		}
	})
}

func TestDeliver(t *testing.T) {
	t.Run("TC-FR-ECO-05-03 設計書の品質計算例", func(t *testing.T) {
		w := NewWorld(questCatalog("q_a"), 3)
		w.Money = 7
		for quality, count := range map[int]int{15: 3, 25: 1, 40: 2} {
			if err := w.Inventory.Add(ItemKey{"burger", quality}, count); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Deliver(0); err != nil {
			t.Fatal(err)
		}
		want := []Stack{{ItemKey{"burger", 15}, 3}, {ItemKey{"burger", 40}, 1}}
		if !reflect.DeepEqual(w.Inventory.Stacks(), want) || w.Money != 127 || w.Quests[0] != "q_a" {
			t.Fatalf("在庫 = %v、お金 = %d、依頼板 = %v", w.Inventory.Stacks(), w.Money, w.Quests)
		}
	})
	t.Run("TC-FR-ECO-05-08 納品した依頼を再抽選", func(t *testing.T) {
		w := NewWorld(questCatalog("q_a", "q_b", "q_c"), 4)
		before := w.Quests
		addQuestItems(t, w, 20, 20)
		if err := w.Deliver(1); err != nil || w.Quests != before {
			t.Fatalf("結果 = %v、依頼板 = %v", err, w.Quests)
		}
	})
	t.Run("TC-FR-ECO-05-09 最低品質の境界", func(t *testing.T) {
		w := NewWorld(questCatalog("q_a"), 1)
		addQuestItems(t, w, 20, 20, 40)
		if err := w.Deliver(0); err != nil || w.Inventory.Count(ItemKey{"burger", 20}) != 0 || w.Inventory.Count(ItemKey{"burger", 40}) != 1 {
			t.Fatalf("結果 = %v、在庫 = %v", err, w.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-ECO-05-11 別のアイテムは納品しない", func(t *testing.T) {
		w := NewWorld(questCatalog("q_a"), 1)
		addQuestItems(t, w, 20, 40)
		if err := w.Inventory.Add(ItemKey{"potion", 200}, 3); err != nil {
			t.Fatal(err)
		}
		if err := w.Deliver(0); err != nil {
			t.Fatal(err)
		}
		if got := w.Inventory.Stacks(); !reflect.DeepEqual(got, []Stack{{ItemKey{"potion", 200}, 3}}) {
			t.Fatalf("納品後の在庫 = %v", got)
		}
	})
	max := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name      string
		money     int
		wantError error
	}{
		{"TC-FR-ECO-05-07 ちょうど上限", max - 120, nil},
		{"TC-FR-ECO-05-07 上限を 1 超える", max - 119, ErrInvalidCount},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(questCatalog("q_a"), 1)
			w.Money = tc.money
			addQuestItems(t, w, 20, 40)
			beforeItems, beforeQuests := w.Inventory.Stacks(), w.Quests
			err := w.Deliver(0)
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("誤り = %v、期待 = %v", err, tc.wantError)
			}
			if tc.wantError != nil {
				if !reflect.DeepEqual(w.Inventory.Stacks(), beforeItems) || w.Money != tc.money || w.Quests != beforeQuests {
					t.Fatalf("失敗後に状態が変化: %v、%d、%v", w.Inventory.Stacks(), w.Money, w.Quests)
				}
			} else if w.Money != max || len(w.Inventory.Stacks()) != 0 {
				t.Fatalf("納品後のお金 = %d、在庫 = %v", w.Money, w.Inventory.Stacks())
			}
		})
	}
	for _, tc := range []struct {
		name  string
		index int
		setup func(*World)
		want  error
	}{
		{"TC-FR-ECO-05-04 負の位置", -1, func(w *World) { addQuestItems(t, w, 20, 40) }, ErrUnknownID},
		{"TC-FR-ECO-05-04 大きすぎる位置", 3, func(w *World) { addQuestItems(t, w, 20, 40) }, ErrUnknownID},
		{"TC-FR-ECO-05-05 空の位置", 1, func(w *World) {
			w.Quests[1] = ""
			w.Money = int(^uint(0) >> 1)
		}, ErrNoQuest},
		{"TC-FR-ECO-05-06 品質条件と個数が不足", 0, func(w *World) {
			w.Money = int(^uint(0) >> 1)
			addQuestItems(t, w, 15, 15, 20)
			if err := w.Inventory.Add(ItemKey{"potion", 200}, 3); err != nil {
				t.Fatal(err)
			}
		}, ErrNotEnoughItems},
		{"TC-FR-ECO-05-12 別のアイテムを不足数に含めない", 0, func(w *World) {
			addQuestItems(t, w, 15, 20)
			if err := w.Inventory.Add(ItemKey{"potion", 200}, 3); err != nil {
				t.Fatal(err)
			}
		}, ErrNotEnoughItems},
		{"TC-FR-ECO-05-13 目録にない依頼 ID", 0, func(w *World) {
			w.Quests[0] = "q_missing"
			addQuestItems(t, w, 20, 40)
		}, ErrUnknownID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(questCatalog("q_a", "q_b", "q_c", "q_d"), 19)
			if tc.setup != nil {
				tc.setup(w)
			}
			beforeItems, beforeMoney, beforeQuests := w.Inventory.Stacks(), w.Money, w.Quests
			if err := w.Deliver(tc.index); !errors.Is(err, tc.want) {
				t.Fatalf("誤り = %v、期待 = %v", err, tc.want)
			}
			if !reflect.DeepEqual(w.Inventory.Stacks(), beforeItems) || w.Money != beforeMoney || w.Quests != beforeQuests {
				t.Fatalf("失敗後に状態が変化: %v、%d、%v", w.Inventory.Stacks(), w.Money, w.Quests)
			}
		})
	}
}

func TestQuestRandomIndependence(t *testing.T) {
	t.Run("TC-NFR-06-30 同じシードの依頼板", func(t *testing.T) {
		a, b := NewWorld(questCatalog("q_a", "q_b", "q_c", "q_d"), 77), NewWorld(questCatalog("q_d", "q_c", "q_b", "q_a"), 77)
		if a.Quests != b.Quests {
			t.Fatalf("初期依頼板 = %v、%v", a.Quests, b.Quests)
		}
		addQuestItems(t, a, 20, 40)
		addQuestItems(t, b, 20, 40)
		if err := a.Deliver(1); err != nil {
			t.Fatal(err)
		}
		if err := b.Deliver(1); err != nil || a.Quests != b.Quests {
			t.Fatalf("納品後の依頼板 = %v、%v、%v", a.Quests, b.Quests, err)
		}
	})
	t.Run("TC-NFR-06-31 本体乱数を進めない", func(t *testing.T) {
		w := NewWorld(questCatalog("q_a", "q_b", "q_c", "q_d"), 77)
		control := NewWorld(questCatalog(), 77)
		if w.rng.Uint64() != control.rng.Uint64() {
			t.Fatal("作成時に本体乱数が進んだ")
		}
		addQuestItems(t, w, 20, 40)
		if err := w.Deliver(0); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			if w.rng.Uint64() != control.rng.Uint64() {
				t.Fatalf("納品後の本体乱数が違う: %d 番目", i)
			}
		}
	})
	t.Run("TC-NFR-06-32 失敗時は依頼乱数を進めない", func(t *testing.T) {
		a, b := NewWorld(questCatalog("q_a", "q_b", "q_c", "q_d"), 77), NewWorld(questCatalog("q_a", "q_b", "q_c", "q_d"), 77)
		if err := a.Deliver(0); !errors.Is(err, ErrNotEnoughItems) {
			t.Fatalf("誤り = %v", err)
		}
		if a.questRNG.Uint64() != b.questRNG.Uint64() {
			t.Fatal("失敗した納品が依頼乱数を進めた")
		}
		addQuestItems(t, a, 20, 40)
		addQuestItems(t, b, 20, 40)
		if err := a.Deliver(0); err != nil {
			t.Fatal(err)
		}
		if err := b.Deliver(0); err != nil || a.Quests != b.Quests {
			t.Fatalf("納品後の依頼板 = %v、%v、%v", a.Quests, b.Quests, err)
		}
	})
}
