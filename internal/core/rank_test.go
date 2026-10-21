package core

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func rankCatalog() defs.Catalog {
	c := golemCatalog()
	c.Game.RankReputation = []int{3, 8, 15}
	return c
}

func TestRankBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		reputation, rank, next int
		hasNext                bool
	}{
		{"TC-FR-PRG-04-01 評判 0", 0, 1, 3, true},
		{"TC-FR-PRG-04-01 評判 2", 2, 1, 3, true},
		{"TC-FR-PRG-04-02 評判 3", 3, 2, 8, true},
		{"TC-FR-PRG-04-02 評判 7", 7, 2, 8, true},
		{"TC-FR-PRG-04-03 評判 8", 8, 3, 15, true},
		{"TC-FR-PRG-04-03 評判 14", 14, 3, 15, true},
		{"TC-FR-PRG-04-04 評判 15", 15, 4, 0, false},
		{"TC-FR-PRG-04-04 評判 99", 99, 4, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(rankCatalog(), 1)
			w.Reputation = tc.reputation
			if got := w.Rank(); got != tc.rank {
				t.Fatalf("Rank() = %d、期待 %d", got, tc.rank)
			}
			if got, ok := w.NextRankReputation(); got != tc.next || ok != tc.hasNext {
				t.Fatalf("NextRankReputation() = %d, %t", got, ok)
			}
		})
	}
	t.Run("TC-FR-PRG-04-05 閾値がない目録", func(t *testing.T) {
		catalog := golemCatalog()
		recipe := catalog.Recipes["potion"]
		recipe.Rank = 1
		catalog.Recipes["potion"] = recipe
		w := NewWorld(catalog, 1)
		w.Reputation = 99
		if w.Rank() != 1 {
			t.Fatal("閾値なしでランクが上がった")
		}
		if _, ok := w.NextRankReputation(); ok {
			t.Fatal("閾値なしで次のランクがある")
		}
		if _, err := w.PreviewCraft("potion", Selection{}); !errors.Is(err, ErrInvalidCount) {
			t.Fatalf("ランク 1 のレシピが使えない: %v", err)
		}
	})
}

func TestRankUnlocked(t *testing.T) {
	c := rankCatalog()
	c.Recipes = map[string]defs.RecipeDef{"c": {ID: "c", Rank: 2}, "a": {ID: "a", Rank: 2}, "b": {ID: "b", Rank: 2}, "first": {ID: "first", Rank: 1}}
	c.Gathers = map[string]defs.GatherDef{"c": {ID: "c", Rank: 2}, "a": {ID: "a", Rank: 2}, "b": {ID: "b", Rank: 2}, "mountain": {ID: "mountain", Rank: 3}}
	c.Quests = map[string]defs.QuestDef{"c": {ID: "c", Rank: 2}, "a": {ID: "a", Rank: 2}, "b": {ID: "b", Rank: 2}, "q1": {ID: "q1", Rank: 1}}
	w := NewWorld(c, 1)
	t.Run("TC-FR-PRG-04-06 新しく解放する ID の順", func(t *testing.T) {
		got := w.Unlocked(2)
		want := []string{"a", "b", "c"}
		if !reflect.DeepEqual(got.Recipes, want) || !reflect.DeepEqual(got.Gathers, want) || !reflect.DeepEqual(got.Quests, want) {
			t.Fatalf("Unlocked(2) = %+v", got)
		}
	})
	t.Run("TC-FR-PRG-04-07 他ランクの ID を含めない", func(t *testing.T) {
		got := w.Unlocked(3)
		if len(got.Recipes) != 0 || !reflect.DeepEqual(got.Gathers, []string{"mountain"}) || len(got.Quests) != 0 {
			t.Fatalf("Unlocked(3) = %+v", got)
		}
	})
}

func TestRankLockedOperations(t *testing.T) {
	c := golemCatalog()
	c.Game.RankReputation = []int{3, 8, 15}
	r := c.Recipes["potion"]
	r.Rank = 2
	c.Recipes["potion"] = r
	item := c.Items["potion"]
	item.Name = "回復薬"
	c.Items["potion"] = item
	c.Gathers = map[string]defs.GatherDef{"forest": {ID: "forest", Name: "古い森", Rank: 2, Rolls: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 20, QualityMax: 20}}}}
	w := NewWorld(c, 1)
	buildWorker(t, w)
	selection := craftSelection([]Stack{craftStack("herb", 40, 2)}, []Stack{craftStack("water", 31, 1)})
	for _, tc := range []struct {
		name, display string
		run           func() error
	}{
		{"TC-FR-ALC-07-01 Craft", "回復薬", func() error { _, err := w.Craft("potion", selection); return err }},
		{"TC-FR-ALC-07-02 PreviewCraft", "回復薬", func() error { _, err := w.PreviewCraft("potion", selection); return err }},
		{"TC-FR-ALC-07-03 AssignGolem レシピ", "回復薬", func() error { return w.AssignGolem(0, Assignment{RecipeID: "potion"}) }},
		{"TC-FR-PRG-04-08 AssignGolem 採取地", "古い森", func() error { return w.AssignGolem(0, Assignment{GatherID: "forest"}) }},
		{"TC-FR-PRG-04-09 Gather", "古い森", func() error { _, err := w.Gather("forest"); return err }},
		{"TC-FR-PRG-04-10 StartBattle", "古い森", func() error { return w.StartBattle("forest") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			beforeStock, beforeGolem := w.Inventory.Stacks(), w.Golems[0]
			err := tc.run()
			if !errors.Is(err, ErrLocked) || !strings.Contains(err.Error(), tc.display) || !strings.Contains(err.Error(), "ランク 2") {
				t.Fatalf("ErrLocked と表示名・ランク = %v", err)
			}
			if !reflect.DeepEqual(w.Inventory.Stacks(), beforeStock) || !reflect.DeepEqual(w.Golems[0], beforeGolem) || w.Battle != nil || len(w.gatherReady) != 0 {
				t.Fatal("失敗で状態が変化")
			}
		})
	}
}

func TestRankCheckOrderAndUnlock(t *testing.T) {
	c := golemCatalog()
	c.Game.RankReputation = []int{3, 8, 15}
	r := c.Recipes["potion"]
	r.Rank = 2
	c.Recipes["potion"] = r
	c.Gathers = map[string]defs.GatherDef{"forest": {ID: "forest", Name: "古い森", Rank: 2, Rolls: 1, Drops: []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 20, QualityMax: 20}}}}
	w := NewWorld(c, 1)
	buildWorker(t, w)
	for _, tc := range []struct {
		name string
		run  func() error
		want error
	}{
		{"TC-FR-ALC-07-04 未知のレシピが先 Craft", func() error { _, err := w.Craft("missing", Selection{}); return err }, ErrUnknownID},
		{"TC-FR-ALC-07-04 未知のレシピが先 PreviewCraft", func() error { _, err := w.PreviewCraft("missing", Selection{}); return err }, ErrUnknownID},
		{"TC-FR-ALC-07-04 未知のレシピが先 AssignGolem", func() error { return w.AssignGolem(0, Assignment{RecipeID: "missing"}) }, ErrUnknownID},
		{"TC-FR-PRG-04-11 未知の採取地が先 AssignGolem", func() error { return w.AssignGolem(0, Assignment{GatherID: "missing"}) }, ErrUnknownID},
		{"TC-FR-PRG-04-11 未知の採取地が先 Gather", func() error { _, err := w.Gather("missing"); return err }, ErrUnknownID},
		{"TC-FR-PRG-04-11 未知の採取地が先 StartBattle", func() error { return w.StartBattle("missing") }, ErrUnknownID},
		{"TC-FR-ALC-07-05 入力枠数より解放検査が先", func() error { _, err := w.Craft("potion", Selection{}); return err }, ErrLocked},
		{"TC-FR-ALC-07-05 入力枠数より解放検査が先 PreviewCraft", func() error { _, err := w.PreviewCraft("potion", Selection{}); return err }, ErrLocked},
		{"TC-FR-ALC-07-05 入力枠数より解放検査が先 AssignGolem", func() error { return w.AssignGolem(0, Assignment{RecipeID: "potion"}) }, ErrLocked},
		{"TC-FR-PRG-04-12 探索不可より解放検査が先", func() error { return w.StartBattle("forest") }, ErrLocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !errors.Is(err, tc.want) {
				t.Fatalf("誤り = %v、期待 %v", err, tc.want)
			}
		})
	}
	t.Run("TC-FR-ALC-07-06 ランク到達でレシピを使用", func(t *testing.T) {
		w.Reputation = 3
		if _, err := w.PreviewCraft("potion", Selection{}); !errors.Is(err, ErrInvalidCount) {
			t.Fatalf("誤り = %v", err)
		}
		if err := w.AssignGolem(0, Assignment{RecipeID: "potion", SlotItems: []string{"herb", "water"}}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("TC-FR-ALC-07-07 未解放レシピの入力候補", func(t *testing.T) {
		w.Reputation = 0
		stock := []Stack{craftStack("herb", 40, 2), craftStack("water", 31, 1)}
		for _, stack := range stock {
			if err := w.Inventory.Add(stack.Key, stack.Count); err != nil {
				t.Fatal(err)
			}
		}
		before := w.Inventory.Stacks()
		got, err := w.SlotCandidates("potion", 0)
		if err != nil || !reflect.DeepEqual(got, []Stack{stock[0]}) {
			t.Fatalf("候補 = %v、誤り = %v", got, err)
		}
		if !reflect.DeepEqual(w.Inventory.Stacks(), before) {
			t.Fatal("候補照会で在庫が変化")
		}
	})
	t.Run("TC-FR-PRG-04-13 ランク到達で採取地を使用", func(t *testing.T) {
		w.Reputation = 3
		if _, err := w.Gather("forest"); err != nil {
			t.Fatal(err)
		}
		if err := w.StartBattle("forest"); !errors.Is(err, ErrNoBattle) {
			t.Fatalf("誤り = %v", err)
		}
	})
}

func TestRankReputationOverflow(t *testing.T) {
	c := questCatalog("q")
	c.Game.RankReputation = []int{3, 8, 15}
	q := c.Quests["q"]
	q.Reputation = 7
	c.Quests["q"] = q
	max := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name              string
		money, reputation int
		want              error
	}{
		{"TC-FR-ECO-05-32 評判が上限ちょうど", 0, max - 7, nil},
		{"TC-FR-ECO-05-33 評判があふれる", 0, max - 6, ErrInvalidCount},
		{"TC-FR-ECO-05-34 お金が先にあふれる", max - 119, max - 6, ErrInvalidCount},
		{"TC-FR-ECO-05-35 在庫不足が先", max - 119, max - 6, ErrNotEnoughItems},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := acceptedQuestWorld(c, 1)
			w.Money, w.Reputation = tc.money, tc.reputation
			w.rankUps = []int{2}
			if tc.want != ErrNotEnoughItems {
				addQuestItems(t, w, 20, 20)
			}
			beforeState := questState(w)
			beforeItems, beforeQuests := w.Inventory.Stacks(), w.Quests
			err := w.DeliverQuest(0)
			if !errors.Is(err, tc.want) {
				t.Fatalf("誤り = %v、期待 %v", err, tc.want)
			}
			if tc.want == nil {
				if w.Reputation != max {
					t.Fatalf("評判 = %d", w.Reputation)
				}
				return
			}
			if !reflect.DeepEqual(questState(w), beforeState) || w.Reputation != tc.reputation || w.Money != tc.money || !reflect.DeepEqual(w.Inventory.Stacks(), beforeItems) || w.Quests != beforeQuests || !reflect.DeepEqual(w.TakeRankUps(), []int{2}) {
				t.Fatal("失敗で状態が変化")
			}
			if tc.money == 0 && !strings.Contains(err.Error(), "評判") {
				t.Fatalf("評判の誤り文 = %v", err)
			}
			if tc.money != 0 && strings.Contains(err.Error(), "評判") {
				t.Fatalf("お金より評判を先に検査した: %v", err)
			}
		})
	}
}

func TestRankQuestAndReputation(t *testing.T) {
	c := questCatalog("a", "b", "c")
	c.Game.RankReputation = []int{3, 8, 15}
	for id, q := range c.Quests {
		q.Reputation = 7
		c.Quests[id] = q
	}
	c.Quests["d"] = defs.QuestDef{ID: "d", Rank: 2, Item: "burger", MinQuality: 20, Count: 2}
	w := NewWorld(c, 42)
	t.Run("TC-FR-ECO-05-30 初期候補は解放済みだけ", func(t *testing.T) {
		for _, id := range w.Quests {
			if id == "d" || id == "" {
				t.Fatalf("初期依頼板 = %v", w.Quests)
			}
		}
	})
	t.Run("TC-FR-ECO-05-31 計算例の二段階上昇", func(t *testing.T) {
		w.Reputation = 2
		if err := w.AcceptQuest(0); err != nil {
			t.Fatal(err)
		}
		addQuestItems(t, w, 20, 20)
		if err := w.DeliverQuest(0); err != nil {
			t.Fatal(err)
		}
		if w.Reputation != 9 || w.Rank() != 3 || !reflect.DeepEqual(w.TakeRankUps(), []int{2, 3}) || len(w.TakeRankUps()) != 0 {
			t.Fatalf("評判 = %d、ランク = %d、次の知らせ = %v", w.Reputation, w.Rank(), w.TakeRankUps())
		}
	})
	t.Run("TC-NFR-06-40 解放後の日替わりも ID 順と同じ乱数列", func(t *testing.T) {
		ids := []string{"a", "b", "c"}
		rng := rand.New(rand.NewPCG(42, 1))
		var shown [3]string
		for i := range shown {
			n := rng.IntN(len(ids))
			shown[i] = ids[n]
			ids = append(ids[:n], ids[n+1:]...)
		}
		if w.Quests != [3]string{"", shown[1], shown[2]} {
			t.Fatalf("納品で補充された: %v", w.Quests)
		}
		ids = []string{"a", "b", "c", "d"}
		var want [3]string
		for i := range want {
			n := rng.IntN(len(ids))
			want[i] = ids[n]
			ids = append(ids[:n], ids[n+1:]...)
		}
		w.catalog.Game.DaySeconds = 1
		w.Now = 59
		w.Tick()
		if w.Quests != want {
			t.Fatalf("日替わり候補=%v 期待=%v", w.Quests, want)
		}
	})
}

func TestRankSingleAndNoRankUp(t *testing.T) {
	for _, tc := range []struct {
		name               string
		before, reputation int
		wantRank           int
		wantUps            []int
	}{
		{"TC-FR-ECO-05-36 評判 0 から 1 はランク据え置き", 0, 1, 1, nil},
		{"TC-FR-ECO-05-37 評判 2 から 3 で一段上昇", 2, 3, 2, []int{2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := questCatalog("q")
			c.Game.RankReputation = []int{3, 8, 15}
			q := c.Quests["q"]
			q.Reputation = 1
			c.Quests["q"] = q
			w := acceptedQuestWorld(c, 1)
			w.Reputation = tc.before
			addQuestItems(t, w, 20, 20)
			if err := w.DeliverQuest(0); err != nil {
				t.Fatal(err)
			}
			ups := w.TakeRankUps()
			if w.Reputation != tc.reputation || w.Rank() != tc.wantRank || !reflect.DeepEqual(ups, tc.wantUps) || len(w.TakeRankUps()) != 0 {
				t.Fatalf("評判 = %d、ランク = %d、知らせ = %v", w.Reputation, w.Rank(), ups)
			}
		})
	}
}
