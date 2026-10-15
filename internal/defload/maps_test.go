package defload

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const validMaps = `[[map]]
id = "workshop"
name = "工房"
start = { x = 1, y = 2 }
rows = ["#######", "#KSG..#", "#.....#", "#.....#", "###1###"]
exits = [{ mark = "1", to = "town", dir = "up", look = "stone_door" }]
[[map]]
id = "town"
name = "街"
start = { x = 1, y = 2 }
rows = ["#########", "#M.Q.F..#", "#.......#", "#.......#", "###1#2###"]
exits = [{ mark = "1", to = "workshop", dir = "up", look = "wood_door" }, { mark = "2", to = "gather", dir = "up", look = "road" }]
`

func loadMapText(s string) (defs.Catalog, error) {
	return Load(fstest.MapFS{"maps.toml": &fstest.MapFile{Data: []byte(s)}})
}
func replaceOne(s, old, new string) string { return strings.Replace(s, old, new, 1) }

// mapDiagnostics はファイル名・経路・理由を個別に検査する（DES-VAL-17〜23）。
func mapDiagnostics(t *testing.T, err error, want string) Errors {
	t.Helper()
	var diagnostics Errors
	if !errors.As(err, &diagnostics) {
		t.Fatalf("Errors がない: %v", err)
	}
	found := false
	for _, d := range diagnostics {
		if d.File == "maps.toml" && strings.Contains(d.Path+": "+d.Reason, want) {
			found = true
		}
	}
	if !found {
		t.Fatalf("want %q; got %v", want, diagnostics)
	}
	return diagnostics
}

// TestMaps は地図の読み込みと各検証規則の診断を確かめる（DES-DEF-09、DES-VAL-17〜23）。
func TestMaps(t *testing.T) {
	valid := validMaps
	cases := []struct {
		id   string
		edit func(string) string
		want string
	}{
		{"TC-FR-FLD-04-01 地図なし", func(string) string { return "" }, ""},
		{"TC-FR-FLD-04-02 正常な二枚", func(s string) string { return s }, ""},
		{"TC-FR-FLD-04-03 地図IDが未知", func(s string) string { return replaceOne(s, "id = \"town\"", "id = \"market\"") }, "map[id=market].id: 地図の ID"},
		{"TC-FR-FLD-04-04 同じID", func(s string) string {
			return s + strings.Replace(s[:strings.LastIndex(s, "[[map]]")], "workshop", "workshop", 1)
		}, "map[id=workshop].id: ID が重複"},
		{"TC-FR-FLD-04-05 片方不足", func(s string) string { return s[:strings.LastIndex(s, "[[map]]")] }, "map[id=town]: 地図がない"},
		{"TC-FR-FLD-04-06 空の行", func(s string) string { return replaceOne(s, "#KSG..#", "") }, "rows[1]: 行の長さは 1〜100"},
		{"TC-FR-FLD-04-07 長さが違う", func(s string) string { return replaceOne(s, "#KSG..#", "#KSG.#") }, "rows[1]: 行の長さが rows[0] と違う（6 文字。rows[0] は 7 文字）"},
		{"TC-FR-FLD-04-08 未知の文字", func(s string) string { return replaceOne(s, "#KSG..#", "#KXG..#") }, "rows[1][2]: 知らない文字 \"X\""},
		{"TC-FR-FLD-04-09 多バイト文字の位置", func(s string) string { return replaceOne(s, "#KSG..#", "#KあX..#") }, "rows[1][2]: 知らない文字 \"あ\""},
		{"TC-FR-FLD-04-10 start省略", func(s string) string { return replaceOne(s, "start = { x = 1, y = 2 }\n", "") }, "map[id=workshop].start: 開始位置が書かれていない"},
		{"TC-FR-FLD-04-11 startは壁", func(s string) string { return replaceOne(s, "x = 1, y = 2", "x = 0, y = 0") }, "map[id=workshop].start: 開始位置 (0, 0) が床ではない"},
		{"TC-FR-FLD-04-12 必須設備不足", func(s string) string { return replaceOne(s, "#KSG..#", "#KS...#") }, "map[id=workshop]: ゴーレムの作業台（G）がない"},
		{"TC-FR-FLD-04-13 出口markは0不可", func(s string) string { return replaceOne(s, "mark = \"1\"", "mark = \"0\"") }, "exits[0].mark: mark は 1〜9"},
		{"TC-FR-FLD-04-14 出口の定義不足", func(s string) string { return replaceOne(s, "###1###", "###2###") }, "rows[4][3]: 出入り口 \"2\" の定義がない"},
		{"TC-FR-FLD-04-15 行き先がない", func(s string) string { return replaceOne(s, "to = \"town\"", "to = \"market\"") }, "exits[0].to: 行き先がない: market"},
		{"TC-FR-FLD-04-16 向きが未知", func(s string) string { return replaceOne(s, "dir = \"up\"", "dir = \"north\"") }, "exits[0].dir: 向きは"},
		{"TC-FR-FLD-04-17 絵が未知", func(s string) string { return replaceOne(s, "look = \"stone_door\"", "look = \"none\"") }, "exits[0].look: 絵は"},
		{"TC-FR-FLD-04-18 到着先が壁", func(s string) string { return strings.Replace(s, "#.....#", "#..#..#", 2) }, "exits[0].dir: 到着のマス (3, 3) が床ではない"},
		{"TC-FR-FLD-04-19 同じ行き先", func(s string) string { return replaceOne(s, "to = \"gather\"", "to = \"workshop\"") }, "exits[1].to: 同じ行き先が重複"},
		{"TC-FR-FLD-04-20 設備に行けない", func(s string) string {
			return replaceOne(replaceOne(s, "#M.Q.F..#", "#M.Q#F#.#"), "#.......#", "#....#..#")
		}, "rows[1][5]: 開始位置から行けない: 施設の看板"},
		{"TC-FR-FLD-04-21 到着マスへ行けない", func(s string) string { return replaceOne(replaceOne(s, "#.....#", "#..#..#"), "#.....#", "#.#...#") }, "開始位置から到着のマス (3, 3) へ行けない"},
		{"TC-FR-FLD-04-22 出口が二マス", func(s string) string { return replaceOne(s, "###1###", "##11###") }, "出入り口 \"1\" が 2 マスある"},
		{"TC-FR-FLD-04-23 自分への出口", func(s string) string { return replaceOne(s, "to = \"town\"", "to = \"workshop\"") }, "exits[0].to: 自分の地図には行けない"},
		{"TC-FR-FLD-04-24 工房に採取地出口", func(s string) string { return replaceOne(s, "to = \"town\"", "to = \"gather\"") }, "gather 行きは town にだけ置く"},
		{"TC-FR-FLD-04-25 形が誤りなら到達性は検査しない", func(s string) string {
			return replaceOne(replaceOne(s, "#M.Q.F..#", "#M.Q#F#.#"), "#.......#", "#....#..")
		}, "rows[2]: 行の長さが rows[0] と違う"},
		{"TC-FR-FLD-04-26 開始位置が誤りなら到達性は検査しない", func(s string) string {
			return replaceOne(replaceOne(replaceOne(s, "#M.Q.F..#", "#M.Q#F#.#"), "#.......#", "#....#..#"), "x = 1, y = 2", "x = 0, y = 0")
		}, "start: 開始位置 (0, 0) が床ではない"},
		{"TC-FR-FLD-04-27 行がない", func(s string) string {
			return replaceOne(s, "rows = [\"#######\", \"#KSG..#\", \"#.....#\", \"#.....#\", \"###1###\"]", "rows = []")
		}, "rows: 行数は 1〜100"},
		{"TC-FR-FLD-04-28 幅が101", func(s string) string { return replaceOne(s, "#KSG..#", strings.Repeat(".", 101)) }, "rows[1]: 行の長さは 1〜100 文字"},
		{"TC-FR-FLD-04-29 高さが101", func(s string) string {
			return replaceOne(s, "###1###\"]", "###1###\", "+strings.Repeat("\"#.....#\", ", 96)+"\"#.....#\"]")
		}, "rows: 行数は 1〜100"},
		{"TC-FR-FLD-04-30 設計例の行長診断", func(string) string {
			rows := []string{}
			for i := 0; i < 4; i++ {
				rows = append(rows, fmt.Sprintf("%q", strings.Repeat(".", 30)))
			}
			rows = append(rows, fmt.Sprintf("%q", strings.Repeat(".", 29)))
			return "[[map]]\nid=\"town\"\nrows=[" + strings.Join(rows, ",") + "]\n"
		}, "map[id=town].rows[4]: 行の長さが rows[0] と違う（29 文字。rows[0] は 30 文字）"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			c, err := loadMapText(tc.edit(valid))
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.id == "TC-FR-FLD-04-02 正常な二枚" {
					want := map[string]defs.MapDef{
						defs.MapWorkshop: {ID: defs.MapWorkshop, Name: "工房", StartX: 1, StartY: 2, Rows: []string{"#######", "#KSG..#", "#.....#", "#.....#", "###1###"}, Exits: []defs.ExitDef{{Mark: '1', To: defs.MapTown, Dir: defs.DirUp, Look: "stone_door", X: 3, Y: 4}}},
						defs.MapTown:     {ID: defs.MapTown, Name: "街", StartX: 1, StartY: 2, Rows: []string{"#########", "#M.Q.F..#", "#.......#", "#.......#", "###1#2###"}, Exits: []defs.ExitDef{{Mark: '1', To: defs.MapWorkshop, Dir: defs.DirUp, Look: "wood_door", X: 3, Y: 4}, {Mark: '2', To: defs.ExitGather, Dir: defs.DirUp, Look: "road", X: 5, Y: 4}}},
					}
					if !reflect.DeepEqual(c.Maps, want) {
						t.Fatalf("maps: %+v", c.Maps)
					}
				}
				return
			}
			diagnostics := mapDiagnostics(t, err, tc.want)
			if strings.Contains(tc.id, "検査しない") {
				for _, d := range diagnostics {
					if strings.Contains(d.Reason, "開始位置から行けない") || strings.Contains(d.Reason, "開始位置から到着のマス") {
						t.Fatalf("到達性を検査した: %v", diagnostics)
					}
				}
			}
		})
	}
}

// FuzzLoadMaps は任意の TOML を読み、成功した地図の基本不変条件を調べる（DES-VAL-24）。
func FuzzLoadMaps(f *testing.F) {
	const seedCase = "TC-NFR-08-01 地図TOMLのシードと任意バイト列"
	f.Log(seedCase)
	valid := validMaps
	for _, s := range []string{valid, "", "[[map]]\nid=\"town\"\nrows=[]", strings.Replace(valid, "#KSG..#", "#KあG..#", 1), strings.Replace(valid, "start = { x = 1, y = 2 }\n", "", 1), strings.Replace(valid, "to = \"town\"", "to = \"market\"", 1), strings.Replace(valid, "#KSG..#", "#KSG.#", 1), strings.Replace(valid, "#KSG..#", "#KS...#", 1), strings.Replace(valid, "#M.Q.F..#", "#M.Q#F#.#", 1)} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		c, err := Load(fstest.MapFS{
			"maps.toml":    &fstest.MapFile{Data: data},
			"items.toml":   &fstest.MapFile{Data: []byte("[[item]]\nid=\"herb\"\n")},
			"recipes.toml": &fstest.MapFile{Data: []byte("[[recipe]]\nid=\"tea\"\noutput=\"herb\"\noutput_count=1\nseconds=1\ninputs=[{item=\"herb\",count=1}]\n")},
		})
		if err != nil {
			return
		}
		for id, m := range c.Maps {
			if id != defs.MapWorkshop && id != defs.MapTown {
				t.Fatalf("ID: %s", id)
			}
			if m.Width() < 1 || m.Width() > 100 || m.Height() < 1 || m.Height() > 100 {
				t.Fatal("size")
			}
			for _, row := range m.Rows {
				if len(row) != m.Width() {
					t.Fatal("width")
				}
				for i := 0; i < len(row); i++ {
					if defs.TileInfo(row[i]).Kind == defs.CellUnknown {
						t.Fatal("tile")
					}
				}
			}
			if defs.TileInfo(m.Cell(m.StartX, m.StartY)).Kind != defs.CellFloor {
				t.Fatal("start")
			}
			for _, e := range m.Exits {
				if e.Mark < '1' || e.Mark > '9' || m.Cell(e.X, e.Y) != e.Mark {
					t.Fatal("exit")
				}
			}
			checkMapSemantics(t, m)
			checkReachableMap(t, m)
		}
		if len(c.Maps) != 0 && len(c.Maps) != 2 {
			t.Fatal(fmt.Sprint("maps: ", len(c.Maps)))
		}
	})
}

// checkMapSemantics は読み込み結果から設備と出入り口の規則を独立に確認する（DES-VAL-21・22・24）。
func checkMapSemantics(t *testing.T, m defs.MapDef) {
	t.Helper()
	required := "KSG"
	if m.ID == defs.MapTown {
		required = "MQF"
	}
	for _, ch := range required {
		found := false
		for _, row := range m.Rows {
			found = found || strings.ContainsRune(row, ch)
		}
		if !found {
			t.Fatalf("TC-NFR-08-02 必須設備がない: %c", ch)
		}
	}
	marks, destinations := map[byte]bool{}, map[string]bool{}
	for _, e := range m.Exits {
		if marks[e.Mark] || strings.Count(strings.Join(m.Rows, ""), string(e.Mark)) != 1 {
			t.Fatalf("TC-NFR-08-03 mark の重複またはマス数: %c", e.Mark)
		}
		marks[e.Mark] = true
		if e.To == m.ID || e.To != defs.MapWorkshop && e.To != defs.MapTown && e.To != defs.ExitGather {
			t.Fatalf("TC-NFR-08-04 行き先が不正: %s", e.To)
		}
		if e.Dir != defs.DirUp && e.Dir != defs.DirDown && e.Dir != defs.DirLeft && e.Dir != defs.DirRight {
			t.Fatalf("TC-NFR-08-05 向きが不正: %s", e.Dir)
		}
		if e.Look != "stone_door" && e.Look != "wood_door" && e.Look != "road" {
			t.Fatalf("TC-NFR-08-06 絵が不正: %s", e.Look)
		}
		if destinations[e.To] {
			t.Fatalf("TC-NFR-08-07 同じ行き先が重複: %s", e.To)
		}
		destinations[e.To] = true
		if e.To == defs.ExitGather && m.ID != defs.MapTown {
			t.Fatal("TC-NFR-08-08 gather 行きが街の外にある")
		}
	}
	for _, row := range m.Rows {
		for i := 0; i < len(row); i++ {
			if defs.TileInfo(row[i]).Kind == defs.CellExit && !marks[row[i]] {
				t.Fatalf("TC-NFR-08-03 定義のない出口: %c", row[i])
			}
		}
	}
	if !destinations[map[string]string{defs.MapWorkshop: defs.MapTown, defs.MapTown: defs.MapWorkshop}[m.ID]] || m.ID == defs.MapTown && !destinations[defs.ExitGather] {
		t.Fatal("TC-NFR-08-09 必須の出口がない")
	}
}

// checkReachableMap は読み込み結果を独立に探索する（DES-VAL-23、DES-VAL-24）。
func checkReachableMap(t *testing.T, m defs.MapDef) {
	t.Helper()
	w, h := m.Width(), m.Height()
	visited := map[[2]int]bool{{m.StartX, m.StartY}: true}
	queue := [][2]int{{m.StartX, m.StartY}}
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			x, y := v[0]+d[0], v[1]+d[1]
			p := [2]int{x, y}
			if x >= 0 && x < w && y >= 0 && y < h && !visited[p] && defs.TileInfo(m.Cell(x, y)).Kind == defs.CellFloor {
				visited[p] = true
				queue = append(queue, p)
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if defs.TileInfo(m.Cell(x, y)).Kind != defs.CellFixture {
				continue
			}
			ok := false
			for _, d := range dirs {
				if visited[[2]int{x + d[0], y + d[1]}] {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("unreachable fixture (%d,%d)", x, y)
			}
		}
	}
	for _, e := range m.Exits {
		x, y := e.X, e.Y
		switch e.Dir {
		case defs.DirUp:
			y--
		case defs.DirDown:
			y++
		case defs.DirLeft:
			x--
		case defs.DirRight:
			x++
		default:
			t.Fatal("invalid direction")
		}
		if !visited[[2]int{x, y}] {
			t.Fatalf("unreachable exit %c", e.Mark)
		}
	}
}
