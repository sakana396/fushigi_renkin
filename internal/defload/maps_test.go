package defload

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

const validMaps = `[game]
day_seconds=1
[[item]]
id="herb"
[[gather]]
id="meadow"
respawn_seconds=1
rolls=1
cooldown_seconds=1
drops=[{item="herb",weight=1,quality_min=1,quality_max=1}]
[[map]]
id = "town"
name = "街"
start = { x = 1, y = 2 }
rows = ["#######", "#DMQ..#", "#.....#", "#.....#", "###1###"]
exits = [{ mark = "1", to = "gather", dir = "up", look = "stone_door" }]
[[map]]
id = "meadow"
name = "草原"
rows = ["#########", "#*...o..#", "#.......#", "#.......#", "###2#####"]
exits = [{ mark = "2", to = "town", dir = "up", look = "road" }]
`

func loadMapText(s string) (defs.Catalog, error) {
	return Load(fstest.MapFS{"maps.toml": &fstest.MapFile{Data: []byte(s)}})
}
func replaceOne(s, old, new string) string { return strings.Replace(s, old, new, 1) }

// mapDiagnostics は診断のファイル・位置・理由を確かめる（DES-VAL-17〜29）。
func mapDiagnostics(t *testing.T, err error, want string) Errors {
	t.Helper()
	var ds Errors
	if !errors.As(err, &ds) {
		t.Fatalf("診断がない: %v", err)
	}
	path, reason, ok := strings.Cut(want, ": ")
	if !ok {
		t.Fatalf("期待診断には完全なキー経路と理由が必要: %q", want)
	}
	found := false
	for _, d := range ds {
		if d.File == "" || d.Path == "" || d.Reason == "" {
			t.Fatalf("不完全な診断: %+v", d)
		}
		if d.File == "maps.toml" && d.Path == path && d.Reason == reason {
			found = true
		}
	}
	if !found {
		t.Fatalf("期待 %q; 診断 %v", want, err)
	}
	return ds
}

// TestMaps は反復 6 の地図の形・出口・到達性を確かめる（DES-VAL-17〜24）。
func TestMaps(t *testing.T) {
	cases := []struct{ id, old, new, want string }{
		{"TC-FR-FLD-04-01", "", "", ""},
		{"TC-FR-FLD-04-02", "", "", ""},
		{"TC-FR-FLD-04-03", "id = \"town\"", "id = \"market\"", "map[id=market].id: 地図の ID は town または採取地の ID"},
		{"TC-FR-FLD-04-04", "", "", "map[id=town].id: ID が重複"},
		{"TC-FR-FLD-04-05", "", "", "map[id=meadow]: 地図がない"},
		{"TC-FR-FLD-04-06", "#DMQ..#", "", "map[id=town].rows[1]: 行の長さは 1〜100 文字"},
		{"TC-FR-FLD-04-07", "#DMQ..#", "#DMQ.#", "map[id=town].rows[1]: 行の長さが rows[0] と違う（6 文字。rows[0] は 7 文字）"},
		{"TC-FR-FLD-04-08", "#DMQ..#", "#DXQ..#", "map[id=town].rows[1][2]: 知らない文字 \"X\""},
		{"TC-FR-FLD-04-09", "#DMQ..#", "#DあX..#", "map[id=town].rows[1][2]: 知らない文字 \"あ\""},
		{"TC-FR-FLD-04-10", "start = { x = 1, y = 2 }", "", "map[id=town].start: 開始位置が書かれていない"},
		{"TC-FR-FLD-04-11", "x = 1, y = 2", "x = 0, y = 0", "map[id=town].start: 開始位置 (0, 0) が床ではない（\"#\"）"},
		{"TC-FR-FLD-04-12", "#DMQ..#", "#DM...#", "map[id=town]: 依頼板（Q）がない"},
		{"TC-FR-FLD-04-13", "mark = \"1\"", "mark = \"0\"", "map[id=town].exits[0].mark: mark は 1〜9 の 1 文字"},
		{"TC-FR-FLD-04-14", "###1###", "###3###", "map[id=town].rows[4][3]: 出入り口 \"3\" の定義がない"},
		{"TC-FR-FLD-04-15", "to = \"gather\"", "to = \"missing\"", "map[id=town].exits[0].to: 行き先がない: missing（必要な行き先は gather）"},
		{"TC-FR-FLD-04-16", "dir = \"up\"", "dir = \"north\"", "map[id=town].exits[0].dir: 向きは up・down・left・right"},
		{"TC-FR-FLD-04-17", "look = \"stone_door\"", "look = \"none\"", "map[id=town].exits[0].look: 絵は stone_door・wood_door・road"},
		{"TC-FR-FLD-04-18", "\"#.....#\", \"#.....#\"", "\"#.....#\", \"#..#..#\"", "map[id=town].exits[0].dir: 到着のマス (3, 3) が床ではない"},
		{"TC-FR-FLD-04-19", "", "", "map[id=town].exits[1].to: 同じ行き先が重複"},
		{"TC-FR-FLD-04-20", "\"#DMQ..#\", \"#.....#\"", "\"#DMQ#.#\", \"#..#..#\"", "map[id=town].rows[1][3]: 開始位置から行けない: 依頼板"},
		{"TC-FR-FLD-04-21", "\"#.....#\", \"#.....#\"", "\"#..#..#\", \"#.#...#\"", "map[id=town].exits[0]: 開始位置から到着のマス (3, 3) へ行けない"},
		{"TC-FR-FLD-04-22", "###1###", "##11###", "map[id=town].rows[4][3]: 出入り口 \"1\" が 2 マスある"},
		{"TC-FR-FLD-04-23", "to = \"gather\"", "to = \"town\"", "map[id=town].exits[0].to: 自分の地図には行けない"},
		{"TC-FR-FLD-04-24", "to = \"town\"", "to = \"gather\"", "map[id=meadow].exits[0].to: 行き先がない: gather（必要な行き先は town）"},
		{"TC-FR-FLD-04-25", "#DMQ..#", "#DMQ.#", "map[id=town].rows[1]: 行の長さが rows[0] と違う（6 文字。rows[0] は 7 文字）"},
		{"TC-FR-FLD-04-26", "x = 1, y = 2", "x = 0, y = 0", "map[id=town].start: 開始位置 (0, 0) が床ではない（\"#\"）"},
		{"TC-FR-FLD-04-27", "", "", "map[id=town].rows: 行数は 1〜100"},
		{"TC-FR-FLD-04-28", "", "", "map[id=town].rows[1]: 行の長さは 1〜100 文字"},
		{"TC-FR-FLD-04-29", "", "", "map[id=town].rows: 行数は 1〜100"},
		{"TC-FR-FLD-04-30", "", "", "map[id=town].rows[1]: 行の長さが rows[0] と違う（1 文字。rows[0] は 2 文字）"},
		{"TC-FR-FLD-04-31", "#DMQ..#", "#DM...#", "map[id=town]: 依頼板（Q）がない"},
		{"TC-FR-FLD-04-32", "to = \"town\"", "to = \"gather\"", "map[id=meadow].exits: town 行きがちょうど 1 つ必要"},
		{"TC-FR-FLD-04-33", "to = \"gather\"", "to = \"town\"", "map[id=town].exits: gather 行きがちょうど 1 つ必要"},
		{"TC-FR-FLD-04-34", "to = \"gather\"", "to = \"meadow\"", "map[id=town].exits: gather 行きがちょうど 1 つ必要"},
		{"TC-FR-FLD-04-35", "", "", "map[id=town].exits[1].mark: mark が重複"},
		{"TC-FR-FLD-04-36", "mark = \"2\"", "mark = \"12\"", "map[id=meadow].exits[0].mark: mark は 1〜9 の 1 文字"},
		{"TC-FR-FLD-04-37", "\"#.....#\", \"#.....#\"", "\"#.....#\", \"#..D..#\"", "map[id=town].exits[0].dir: 到着のマス (3, 3) が床ではない"},
		{"TC-FR-FLD-04-38", "#DMQ..#", "#DXQ..#", "map[id=town].rows[1][2]: 知らない文字 \"X\""},
		{"TC-FR-FLD-04-39", "", "", "map[id=town].exits[0]: 開始位置から到着のマス (3, 4) へ行けない"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			input := replaceOne(validMaps, tc.old, tc.new)
			switch tc.id {
			case "TC-FR-FLD-04-01":
				input = ""
			case "TC-FR-FLD-04-04":
				input = validMaps + validMaps[strings.Index(validMaps, "[[map]]"):strings.LastIndex(validMaps, "[[map]]")]
			case "TC-FR-FLD-04-05":
				input = validMaps[:strings.LastIndex(validMaps, "[[map]]")]
			case "TC-FR-FLD-04-19":
				input = replaceOne(validMaps, `look = "stone_door" }]`, `look = "stone_door" }, {mark="3",to="gather",dir="up",look="road"}]`)
			case "TC-FR-FLD-04-35":
				input = replaceOne(validMaps, `look = "stone_door" }]`, `look = "stone_door" }, {mark="1",to="gather",dir="up",look="road"}]`)
			case "TC-FR-FLD-04-27":
				input = replaceOne(validMaps, `rows = ["#######", "#DMQ..#", "#.....#", "#.....#", "###1###"]`, `rows=[]`)
			case "TC-FR-FLD-04-28":
				input = replaceOne(validMaps, "#DMQ..#", strings.Repeat(".", 101))
			case "TC-FR-FLD-04-29":
				input = replaceOne(validMaps, `"###1###"]`, strings.Repeat(`"#.....#",`, 97)+`"###1###"]`)
			case "TC-FR-FLD-04-30":
				input = `[[map]]
 id="town"
 rows=["..", "."]`
			case "TC-FR-FLD-04-39":
				input = replaceOne(validMaps, `rows = ["#######", "#DMQ..#", "#.....#", "#.....#", "###1###"]`, `rows=["#######","#DMQ..#","#.#...#","#.#...#","#.1...#","#.#...#","#.#...#"]`)
				input = replaceOne(input, `dir = "up"`, `dir = "right"`)
			}
			c, err := loadMapText(input)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				if tc.id == "TC-FR-FLD-04-02" {
					want := map[string]defs.MapDef{
						"town":   {ID: "town", Name: "街", StartX: 1, StartY: 2, Rows: []string{"#######", "#DMQ..#", "#.....#", "#.....#", "###1###"}, Exits: []defs.ExitDef{{Mark: '1', To: "gather", Dir: defs.DirUp, Look: "stone_door", X: 3, Y: 4}}},
						"meadow": {ID: "meadow", Name: "草原", StartX: 3, StartY: 3, Rows: []string{"#########", "#*...o..#", "#.......#", "#.......#", "###2#####"}, Exits: []defs.ExitDef{{Mark: '2', To: "town", Dir: defs.DirUp, Look: "road", X: 3, Y: 4}}},
					}
					if !reflect.DeepEqual(c.Maps, want) {
						t.Fatalf("地図全項目: got=%+v want=%+v", c.Maps, want)
					}
				}
				return
			}
			ds := mapDiagnostics(t, err, tc.want)
			// REV-047: キー経路の完全一致に加え、前段が不正なら到達性診断を出さない。
			if tc.id >= "TC-FR-FLD-04-31" && tc.id <= "TC-FR-FLD-04-38" {
				absent := []string{"開始位置から行けない", "開始位置から到着のマス"}
				if tc.id == "TC-FR-FLD-04-38" {
					absent = append(absent, ".start", "店（M）がない")
				}
				for _, d := range ds {
					for _, a := range absent {
						if strings.Contains(d.Path+": "+d.Reason, a) {
							t.Fatalf("不要な診断 %q: %v", a, ds)
						}
					}
				}
			}

			if tc.id == "TC-FR-FLD-04-25" || tc.id == "TC-FR-FLD-04-26" || tc.id == "TC-FR-FLD-04-38" {
				for _, d := range ds {
					if strings.Contains(d.Reason, "開始位置から") {
						t.Fatal(ds)
					}
				}
			}
		})
	}
}

// FuzzLoadMaps は任意の TOML を読み、成功した地図の基本不変条件を調べる（DES-VAL-24）。
func FuzzLoadMaps(f *testing.F) {
	const seedCase = "TC-NFR-08-10 / TC-NFR-08-01 地図TOMLのシードと任意バイト列"
	f.Log(seedCase)
	valid := validMaps
	battleSeed, err := os.ReadFile("testdata/it06/battle.toml")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(battleSeed)
	seed, err := os.ReadFile("testdata/it06/maps.toml")
	if err != nil {
		f.Fatal(err)
	}
	for _, s := range []string{valid, string(seed), strings.ReplaceAll(valid, "meadow", "cave"), "", "[[map]]\nid=\"town\"\nrows=[]"} {
		f.Add([]byte(s))
	}
	// 正常な種から、各規則の異常を一つずつ作る（DES-VAL-24）。
	for _, tc := range []struct{ old, new string }{
		{`id = "meadow"` + "\nrows", `id = "unknown"` + "\nrows"},
		{"T*,,,,T", "T*,,,あ"}, {"T*,,,,T", "T*,,,T"},
		{"start = {x=5,y=5}", ""}, {`T,aaa,,DT`, `T,aaa,,,T`},
		{`to="town"`, `to="gather"`}, {`T*,,,,T`, `T*#,,,T`},
		{"aaa", "aa,"}, {`mark="a"`, `mark="j"`}, {`facility="market"`, `facility="missing"`},
		{`look="stall"`, `look="castle"`}, {"T*,,,,T", "TE,,,,T"},
		{"respawn_seconds = 2", "respawn_seconds = 86401"}, {"day_seconds = 300", "day_seconds = 0"},
	} {
		if !strings.Contains(string(seed), tc.old) {
			f.Fatalf("種の置換対象がない: %s", tc.old)
		}
		f.Add([]byte(strings.Replace(string(seed), tc.old, tc.new, 1)))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		c, err := loadMapText(string(data))
		if err != nil {
			return
		}
		for id, m := range c.Maps {
			if _, ok := c.Gathers[id]; id != defs.MapTown && !ok {
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
			checkIT06Catalog(t, c, m)
			checkMapSemantics(t, m)
			checkReachableMap(t, m)
		}
		if len(c.Maps) != 0 && len(c.Maps) != len(c.Gathers)+1 {
			t.Fatal(fmt.Sprint("maps: ", len(c.Maps)))
		}
	})
}

// checkMapSemantics は読み込み結果から設備と出入り口の規則を独立に確認する（DES-VAL-21・22・24）。
func checkMapSemantics(t *testing.T, m defs.MapDef) {
	t.Helper()
	required := ""
	if m.ID == defs.MapTown {
		required = "DMQ"
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
		if e.To == m.ID || e.To != "meadow" && e.To != defs.MapTown && e.To != defs.ExitGather {
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
	wantTo := defs.MapTown
	if m.ID == defs.MapTown {
		wantTo = defs.ExitGather
	}
	if len(m.Exits) != 1 || !destinations[wantTo] {
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
			kind := defs.TileInfo(m.Cell(x, y)).Kind
			if kind != defs.CellFixture && kind != defs.CellPick && kind != defs.CellEnemy {
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
	for _, l := range m.Lots {
		accessible := false
		for y := l.Y; y < l.Y+3; y++ {
			for x := l.X; x < l.X+3; x++ {
				for _, d := range dirs {
					if visited[[2]int{x + d[0], y + d[1]}] {
						accessible = true
					}
				}
			}
		}
		if !accessible {
			t.Fatal("区画の前へ行けない")
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

// checkIT06Catalog は成功した目録の区画と敵、秒数を独立に検査する（DES-VAL-24〜29）。
func checkIT06Catalog(t *testing.T, c defs.Catalog, m defs.MapDef) {
	t.Helper()
	if c.Game.DaySeconds < 1 || c.Game.DaySeconds > 86400 {
		t.Fatal("日の秒数")
	}
	all := strings.Join(m.Rows, "")
	for i := range all {
		kind := defs.TileInfo(all[i]).Kind
		if m.ID == defs.MapTown {
			if kind == defs.CellPick || kind == defs.CellEnemy {
				t.Fatal("街に拾う場所・敵")
			}
		} else if kind == defs.CellLot {
			t.Fatal("採取地に区画")
		}
	}
	if m.ID != defs.MapTown {
		g := c.Gathers[m.ID]
		want := 0
		if g.Battle != nil {
			want = 1
		}
		if g.RespawnSeconds < 1 || g.RespawnSeconds > 86400 || strings.Count(all, "E") != want || strings.Count(all, "*")+strings.Count(all, "o") < 1 || len(m.Lots) > 0 {
			t.Fatal("採取地の定義")
		}
		return
	}
	if len(m.Lots) != len(c.Facilities) {
		t.Fatal("区画と施設の数")
	}
	marks := map[byte]bool{}
	facilities := map[string]bool{}
	for _, l := range m.Lots {
		f, ok := c.Facilities[l.Facility]
		if !ok || marks[l.Mark] || facilities[l.Facility] || l.Mark < 'a' || l.Mark > 'i' || strings.Count(all, string(l.Mark)) != 9 {
			t.Fatal("区画の対応")
		}
		marks[l.Mark] = true
		facilities[l.Facility] = true
		for y := l.Y; y < l.Y+3; y++ {
			for x := l.X; x < l.X+3; x++ {
				if m.Cell(x, y) != l.Mark {
					t.Fatal("区画の形")
				}
			}
		}
		for _, level := range f.Levels {
			switch level.Look {
			case "stall", "cottage", "house", "manor":
			default:
				t.Fatal("施設の見た目")
			}
		}
	}
}
