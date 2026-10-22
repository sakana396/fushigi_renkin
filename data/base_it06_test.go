package data_test

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 反復6の完成データを設計値と照合する（DES-DAT-04）。
func TestBaseDataIT06(t *testing.T) {
	c, err := defload.Load(data.Base)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("TC-FR-MOD-02-60 完成データの件数", func(t *testing.T) {
		got := []int{len(c.Items), len(c.Recipes), len(c.Facilities), len(c.Quests), len(c.Maps), len(c.Gathers)}
		if !reflect.DeepEqual(got, []int{19, 15, 4, 17, 5, 4}) {
			t.Fatalf("件数 = %v", got)
		}
	})
	t.Run("TC-FR-MOD-02-61 新製品とレシピ", func(t *testing.T) {
		for _, tc := range []struct {
			item   defs.ItemDef
			recipe defs.RecipeDef
		}{
			{defs.ItemDef{ID: "iron_ingot", Name: "鉄の延べ棒", Attributes: []string{"金属", "硬質"}, BasePrice: 30}, defs.RecipeDef{ID: "iron_ingot", Rank: 1, Output: "iron_ingot", OutputCount: 1, Seconds: 5, Inputs: []defs.InputSlot{{Item: "ore", Count: 3}}, Bonuses: []defs.Bonus{{Attribute: "硬質", MinCount: 3, Add: 10}}}},
			{defs.ItemDef{ID: "herb_rope", Name: "薬草の縄", Attributes: []string{"繊維", "しなやか"}, BasePrice: 20}, defs.RecipeDef{ID: "herb_rope", Rank: 1, Output: "herb_rope", OutputCount: 1, Seconds: 4, Inputs: []defs.InputSlot{{Item: "herb", Count: 2}, {Item: "wheat", Count: 1}}, Bonuses: []defs.Bonus{{Attribute: "草", MinCount: 2, Add: 10}}}},
		} {
			if got := c.Items[tc.item.ID]; !reflect.DeepEqual(got, tc.item) {
				t.Errorf("製品: got=%+v want=%+v", got, tc.item)
			}
			if got := c.Recipes[tc.recipe.ID]; !reflect.DeepEqual(got, tc.recipe) {
				t.Errorf("レシピ: got=%+v want=%+v", got, tc.recipe)
			}
		}
	})
	t.Run("TC-FR-MOD-02-62 追加依頼8件とランク別件数", func(t *testing.T) {
		for _, want := range []defs.QuestDef{
			{ID: "q_flour", Rank: 1, Reputation: 1, Item: "flour", MinQuality: 20, Count: 3, Reward: 45},
			{ID: "q_potion_easy", Rank: 1, Reputation: 1, Item: "potion", MinQuality: 20, Count: 2, Reward: 50},
			{ID: "q_rope", Rank: 1, Reputation: 1, Item: "herb_rope", MinQuality: 20, Count: 2, Reward: 55},
			{ID: "q_ingot", Rank: 2, Reputation: 3, Item: "iron_ingot", MinQuality: 50, Count: 2, Reward: 90},
			{ID: "q_rope_fine", Rank: 3, Reputation: 4, Item: "herb_rope", MinQuality: 100, Count: 3, Reward: 120},
			{ID: "q_ingot_fine", Rank: 3, Reputation: 4, Item: "iron_ingot", MinQuality: 100, Count: 2, Reward: 140},
			{ID: "q_bomb_master", Rank: 4, Item: "bomb", MinQuality: 150, Count: 2, Reward: 200},
			{ID: "q_lantern_master", Rank: 4, Item: "lantern", MinQuality: 150, Count: 1, Reward: 220},
		} {
			if got := c.Quests[want.ID]; got != want {
				t.Errorf("依頼: got=%+v want=%+v", got, want)
			}
		}
		counts := map[int]int{}
		for _, q := range c.Quests {
			counts[q.Rank]++
		}
		if !reflect.DeepEqual(counts, map[int]int{1: 6, 2: 3, 3: 4, 4: 4}) {
			t.Errorf("ランク別件数=%v", counts)
		}
		if _, ok := c.Quests["q_cheese_extra"]; ok {
			t.Error("仮依頼が残っている")
		}
	})
	t.Run("TC-FR-MOD-02-63 日と再出現時間", func(t *testing.T) {
		if c.Game.DaySeconds != 300 {
			t.Errorf("日の秒数=%d", c.Game.DaySeconds)
		}
		for id, want := range map[string]int{"meadow": 45, "spring_cave": 45, "sunny_hill": 60, "crystal_forest": 75} {
			g := c.Gathers[id]
			if g.RespawnSeconds != want || g.Rolls != 3 {
				t.Errorf("%s: 秒数=%d Rolls=%d", id, g.RespawnSeconds, g.Rolls)
			}
		}
	})
	t.Run("TC-FR-TWN-02-50 施設4件の全段階", func(t *testing.T) {
		for _, want := range []defs.FacilityDef{
			{ID: "workshop", Name: "錬金工房", Levels: []defs.FacilityLevel{
				{Look: "cottage", Money: 60, Materials: []defs.ItemAmount{{Item: "bread", Count: 2}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 1}}},
				{Look: "house", Money: 140, Materials: []defs.ItemAmount{{Item: "iron_ingot", Count: 2}, {Item: "lantern", Count: 1}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 1}}},
			}},
			{ID: "market", Name: "街の市場", Levels: []defs.FacilityLevel{
				{Look: "stall", Money: 70, Materials: []defs.ItemAmount{{Item: "burger", Count: 2}}, Effects: []defs.Effect{{Kind: "sell_rate", Value: 10}}},
				{Look: "house", Money: 150, Materials: []defs.ItemAmount{{Item: "herb_rope", Count: 3}, {Item: "cheese", Count: 2}}, Effects: []defs.Effect{{Kind: "sell_rate", Value: 10}}},
			}},
			{ID: "golem_shed", Name: "ゴーレム小屋", Levels: []defs.FacilityLevel{
				{Look: "cottage", Money: 100, Materials: []defs.ItemAmount{{Item: "iron_ingot", Count: 3}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 1}}},
				{Look: "manor", Money: 220, Materials: []defs.ItemAmount{{Item: "iron_ingot", Count: 4}, {Item: "glow_crystal", Count: 2}}, Effects: []defs.Effect{{Kind: "golem_slots", Value: 1}}},
			}},
			{ID: "inn", Name: "旅人の宿", Levels: []defs.FacilityLevel{
				{Look: "cottage", Money: 90, Materials: []defs.ItemAmount{{Item: "herb_rope", Count: 2}, {Item: "bread", Count: 3}}, Effects: []defs.Effect{{Kind: "sell_rate", Value: 5}}},
				{Look: "manor", Money: 200, Materials: []defs.ItemAmount{{Item: "burger", Count: 3}, {Item: "potion", Count: 2}}, Effects: []defs.Effect{{Kind: "sell_rate", Value: 10}}},
			}},
		} {
			if got := c.Facilities[want.ID]; !reflect.DeepEqual(got, want) {
				t.Errorf("施設: got=%+v want=%+v", got, want)
			}
		}
	})
	t.Run("TC-FR-TWN-07-50 区画と施設の対応", func(t *testing.T) {
		want := []defs.LotDef{{Mark: 'a', Facility: "market", X: 3, Y: 2}, {Mark: 'b', Facility: "inn", X: 24, Y: 2}, {Mark: 'c', Facility: "workshop", X: 3, Y: 8}, {Mark: 'd', Facility: "golem_shed", X: 24, Y: 8}}
		if got := c.Maps["town"].Lots; !reflect.DeepEqual(got, want) {
			t.Errorf("区画: got=%+v want=%+v", got, want)
		}
	})
}

type tilePoint struct{ x, y int }

// 上下左右の床だけを歩く最短距離を求める（DES-DAT-04、DES-FLD-06）。
func floorDistances(m defs.MapDef, starts []tilePoint) map[tilePoint]int {
	distances := map[tilePoint]int{}
	queue := []tilePoint{}
	for _, p := range starts {
		if defs.TileInfo(m.Cell(p.x, p.y)).Kind == defs.CellFloor {
			distances[p] = 0
			queue = append(queue, p)
		}
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, n := range neighbors(p) {
			if _, ok := distances[n]; ok {
				continue
			}
			if defs.TileInfo(m.Cell(n.x, n.y)).Kind != defs.CellFloor {
				continue
			}
			distances[n] = distances[p] + 1
			queue = append(queue, n)
		}
	}
	return distances
}

func neighbors(p tilePoint) []tilePoint {
	return []tilePoint{{p.x - 1, p.y}, {p.x + 1, p.y}, {p.x, p.y - 1}, {p.x, p.y + 1}}
}

func distanceToTarget(distances map[tilePoint]int, p tilePoint) int {
	best := -1
	for _, n := range neighbors(p) {
		if d, ok := distances[n]; ok && (best < 0 || d < best) {
			best = d
		}
	}
	return best
}

// 本体の地図の配置と歩く距離を確認する（DES-DAT-04、DES-FLD-06）。
func TestBaseMapsIT06(t *testing.T) {
	c, err := defload.Load(data.Base)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("TC-FR-FLD-01-44 街の地図", func(t *testing.T) {
		m := c.Maps["town"]
		wantRows := []string{
			"TTTTTTTTTTTTTTTTTTTTTTTTTTTTTT",
			"T,,,,,,,,,,,,,,,,,,,,,,,,,,,,T",
			"T,,aaa,,,,,,,,,,,,,,,,,,bbb,,T",
			"T,,aaa,,,,,,,,::,,,,,,,,bbb,,T",
			"T,,aaa,,,,,,,,::,,,,,,,,bbb,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,::::::::::::::::::::,,,,T",
			"T,,T,,,,,,,,,,::,,,,,,,,,,T,,T",
			"T,,ccc,,,,,,,,::,,,,,,,,ddd,,T",
			"T,,ccc,,,,,,,,::,,,,,,,,ddd,,T",
			"T,,ccc,,,,,,,,::,,,,,,,,ddd,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"T,,,,,,,^^^^^,::,^^^^^,,,,,,,T",
			"T,,,,,,,^^^^^,::,^^^^^,,,,,,,T",
			"T,,,,,,,WWDWW,::,WWWWW,,,,,,,T",
			"T,,,,,,,,::::::::::M,,,,,,,,,T",
			"T,,,,,,,,::::Q:::::,,,,,,,,,,T",
			"T,,,,,,,,::::::::::,,,,,,,,,,T",
			"T,,,,,,,,,,,,,::,,,,,,,,,,,,,T",
			"TTTTTTTTTTTTTT2TTTTTTTTTTTTTTT",
		}
		if m.Name != "街" || m.StartX != 10 || m.StartY != 15 || !reflect.DeepEqual(m.Rows, wantRows) {
			t.Errorf("街=%+v", m)
		}
		wantExits := []defs.ExitDef{{Mark: '2', To: "gather", Dir: defs.DirUp, Look: "road", X: 14, Y: 19}}
		if !reflect.DeepEqual(m.Exits, wantExits) {
			t.Errorf("出口=%+v", m.Exits)
		}
		if _, ok := c.Maps["workshop"]; ok {
			t.Error("工房地図が残っている")
		}
	})
	t.Run("TC-FR-FLD-04-43 街の6組の距離", func(t *testing.T) {
		m := c.Maps["town"]
		points := []tilePoint{{10, 14}, {19, 15}, {13, 16}, {14, 19}}
		for _, tc := range []struct{ a, b, want int }{{0, 1, 8}, {0, 2, 3}, {0, 3, 7}, {1, 2, 5}, {1, 3, 7}, {2, 3, 2}} {
			got := distanceToTarget(floorDistances(m, neighbors(points[tc.a])), points[tc.b])
			if got != tc.want {
				t.Errorf("%v → %v: got=%d want=%d", points[tc.a], points[tc.b], got, tc.want)
			}
		}
	})
	for _, tc := range []struct {
		id, name string
		picks    map[tilePoint]byte
		scenery  map[byte]int
	}{
		{"meadow", "風の草原", map[tilePoint]byte{{5, 5}: '*', {10, 4}: '*', {17, 6}: 'o', {5, 10}: '*', {17, 11}: 'o', {10, 9}: '*'}, map[byte]int{'*': 4, ',': 295, '1': 1, 'E': 1, 'T': 81, 'o': 2}},
		{"spring_cave", "泉の洞窟", map[tilePoint]byte{{5, 5}: 'o', {10, 4}: 'o', {17, 6}: 'o', {5, 10}: '*'}, map[byte]int{'#': 81, '*': 1, '.': 291, '1': 1, 'E': 1, 'o': 3, '~': 6}},
		{"sunny_hill", "陽だまりの丘", map[tilePoint]byte{{5, 5}: '*', {10, 4}: '*', {17, 6}: 'o', {5, 10}: '*', {17, 11}: 'o'}, map[byte]int{'*': 3, ',': 289, '.': 2, '1': 1, 'E': 1, 'T': 80, '^': 6, 'o': 2}},
		{"crystal_forest", "水晶の森", map[tilePoint]byte{{5, 5}: 'o', {10, 4}: '*', {17, 6}: 'o', {5, 10}: '*', {17, 11}: 'o', {10, 9}: '*', {8, 7}: 'o', {15, 3}: 'o'}, map[byte]int{'*': 3, ',': 239, '1': 1, 'E': 1, 'T': 133, 'o': 5, '~': 2}},
	} {
		t.Run("TC-FR-FLD-01-45 "+tc.id, func(t *testing.T) {
			m := c.Maps[tc.id]
			if m.Width() != 24 || m.Height() != 16 || m.Name != tc.name || m.StartX != 12 || m.StartY != 14 || len(m.Lots) != 0 {
				t.Fatalf("地図=%+v", m)
			}
			wantExits := []defs.ExitDef{{Mark: '1', To: "town", Dir: defs.DirUp, Look: "road", X: 12, Y: 15}}
			if !reflect.DeepEqual(m.Exits, wantExits) {
				t.Errorf("出口=%+v", m.Exits)
			}
			picks := map[tilePoint]byte{}
			enemies := []tilePoint{}
			scenery := map[byte]int{}
			for y, row := range m.Rows {
				for x := range row {
					ch := row[x]
					scenery[ch]++
					switch defs.TileInfo(ch).Kind {
					case defs.CellPick:
						picks[tilePoint{x, y}] = ch
					case defs.CellEnemy:
						enemies = append(enemies, tilePoint{x, y})
					}
				}
			}
			if !reflect.DeepEqual(picks, tc.picks) || !reflect.DeepEqual(enemies, []tilePoint{{19, 8}}) || !reflect.DeepEqual(scenery, tc.scenery) {
				t.Errorf("拾う場所=%v 敵=%v 景色=%v", picks, enemies, scenery)
			}
		})
		t.Run("TC-FR-FLD-04-44 "+tc.id, func(t *testing.T) {
			m := c.Maps[tc.id]
			if len(m.Exits) != 1 {
				t.Fatal("出口は1つ必要")
			}
			e := m.Exits[0]
			if e.Dir != defs.DirUp {
				t.Fatal("到着方向は上")
			}
			distances := floorDistances(m, []tilePoint{{e.X, e.Y - 1}})
			// 配置ごとの期待歩数を固定し、上限だけを確かめる試験にしない。
			wants := map[tilePoint]int{{5, 5}: 15, {10, 4}: 11, {17, 6}: 12, {5, 10}: 10, {17, 11}: 7, {10, 9}: 6, {8, 7}: 10, {15, 3}: 13}
			wantDistances := map[tilePoint]int{}
			for p := range tc.picks {
				wantDistances[p] = wants[p]
			}
			gotDistances := map[tilePoint]int{}
			farthest := -1
			for y, row := range m.Rows {
				for x := range row {
					if defs.TileInfo(row[x]).Kind != defs.CellPick {
						continue
					}
					p := tilePoint{x, y}
					got := distanceToTarget(distances, p)
					gotDistances[p] = got
					if got < 0 || got > 15 {
						t.Errorf("%v: 距離=%d、0〜15が必要", p, got)
					}
					farthest = max(farthest, got)
				}
			}
			if !reflect.DeepEqual(gotDistances, wantDistances) {
				t.Errorf("拾う場所までの距離: got=%v want=%v", gotDistances, wantDistances)
			}
			if farthest != 15 {
				t.Errorf("最遠距離=%d、期待15", farthest)
			}
			if got := distanceToTarget(distances, tilePoint{19, 8}); got != 12 {
				t.Errorf("敵まで=%d、期待12", got)
			}
		})
	}
}
