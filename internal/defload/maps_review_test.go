package defload

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestMapReview は REV-047 の不足した境界と検証順を確かめる（DES-VAL-18〜23）。
func TestMapReview(t *testing.T) {
	cases := []struct {
		id, want string
		edit     func(string) string
		absent   []string
	}{
		{"TC-FR-FLD-01-06 開始位置の x 省略", "map[id=workshop].start: 開始位置が書かれていない", func(s string) string { return replaceOne(s, "x = 1, y = 2", "y = 2") }, nil},
		{"TC-FR-FLD-01-07 開始位置の y 省略", "map[id=workshop].start: 開始位置が書かれていない", func(s string) string { return replaceOne(s, "x = 1, y = 2", "x = 1") }, nil},
		{"TC-FR-FLD-01-08 開始位置が地図外", "map[id=workshop].start: 開始位置 (7, 2) が地図の外", func(s string) string { return replaceOne(s, "x = 1, y = 2", "x = 7, y = 2") }, nil},
		{"TC-FR-FLD-04-31 街の依頼板不足", "map[id=town]: 依頼板（Q）がない", func(s string) string { return replaceOne(s, "#M.Q.F..#", "#M...F..#") }, nil},
		{"TC-FR-FLD-04-32 工房の必須出口不足", "map[id=workshop].exits: town 行きがない", func(s string) string { return replaceOne(s, "to = \"town\"", "to = \"gather\"") }, nil},
		{"TC-FR-FLD-04-33 街の工房出口不足", "map[id=town].exits: workshop 行きがない", func(s string) string { return replaceOne(s, "to = \"workshop\"", "to = \"gather\"") }, nil},
		{"TC-FR-FLD-04-34 街の採取地出口不足", "map[id=town].exits: gather 行きがない", func(s string) string { return replaceOne(s, "to = \"gather\"", "to = \"town\"") }, nil},
		{"TC-FR-FLD-04-35 出口mark重複", "map[id=town].exits[1].mark: mark が重複", func(s string) string { return replaceOne(s, "mark = \"2\"", "mark = \"1\"") }, nil},
		{"TC-FR-FLD-04-36 出口markが二文字", "map[id=town].exits[1].mark: mark は 1〜9 の 1 文字", func(s string) string { return replaceOne(s, "mark = \"2\"", "mark = \"12\"") }, nil},
		{"TC-FR-FLD-04-37 到着先が設備", "map[id=workshop].exits[0].dir: 到着のマス (3, 3) が床ではない", func(s string) string { return replaceOne(s, "\"#.....#\", \"#.....#\"", "\"#.....#\", \"#..K..#\"") }, nil},
		{"TC-FR-FLD-04-37 到着先が別の出口", "map[id=workshop].exits[0].dir: 到着のマス (3, 3) が床ではない", func(s string) string { return replaceOne(s, "\"#.....#\", \"#.....#\"", "\"#.....#\", \"#..2..#\"") }, nil},
		{"TC-FR-FLD-04-38 設備不足なら塗りつぶさない", "map[id=town]: 依頼板（Q）がない", func(s string) string {
			return replaceOne(replaceOne(s, "#M.Q.F..#", "#M..#F#.#"), "#.......#", "#....#..#")
		}, []string{"開始位置から行けない", "開始位置から到着のマス"}},
		{"TC-FR-FLD-04-38 出口不正なら塗りつぶさない", "map[id=town].exits[1].mark: mark は 1〜9 の 1 文字", func(s string) string {
			return replaceOne(replaceOne(replaceOne(s, "#M.Q.F..#", "#M.Q#F#.#"), "#.......#", "#....#..#"), "mark = \"2\"", "mark = \"12\"")
		}, []string{"開始位置から行けない", "開始位置から到着のマス"}},
		{"TC-FR-FLD-04-38 行長不正なら後続検査をしない", "map[id=workshop].rows[1]: 行の長さが rows[0] と違う", func(s string) string {
			return replaceOne(replaceOne(s, "#KSG..#", "#KSG.#"), "x = 1, y = 2", "x = 0, y = 0")
		}, []string{".start", "錬金釜（K）がない", "town 行きがない", "開始位置から"}},
		{"TC-FR-FLD-04-38 未知文字なら後続検査をしない", "map[id=workshop].rows[1][1]: 知らない文字", func(s string) string {
			return replaceOne(replaceOne(s, "#KSG..#", "#XSG..#"), "x = 1, y = 2", "x = 0, y = 0")
		}, []string{".start", "錬金釜（K）がない", "開始位置から"}},
		{"TC-FR-FLD-04-39 内側の出口を通って到着先へ行けない", "map[id=workshop].exits[0]: 開始位置から到着のマス (3, 4) へ行けない", func(s string) string {
			return replaceOne(replaceOne(s, "rows = [\"#######\", \"#KSG..#\", \"#.....#\", \"#.....#\", \"###1###\"]", "rows = [\"#######\", \"#KSG..#\", \"#.#...#\", \"#.#...#\", \"#.1...#\", \"#.#...#\", \"#.#...#\"]"), "dir = \"up\"", "dir = \"right\"")
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			_, err := loadMapText(tc.edit(validMaps))
			diagnostics := mapDiagnostics(t, err, tc.want)
			for _, d := range diagnostics {
				if d.File != "maps.toml" || !strings.Contains(d.Path, "map[id=") {
					t.Fatalf("位置が不正: %+v", d)
				}
				for _, absent := range tc.absent {
					if strings.Contains(d.Path+": "+d.Reason, absent) {
						t.Fatalf("不要な診断 %q: %+v", absent, d)
					}
				}
			}
		})
	}
	t.Run("TC-FR-FLD-01-09 診断順は地図IDと数字順", func(t *testing.T) {
		broken := replaceOne(replaceOne(validMaps, "###1###", "#321###"), "###1#2###", "#321#2###")
		var first Errors
		for i := 0; i < 20; i++ {
			_, err := loadMapText(broken)
			var ds Errors
			if !errors.As(err, &ds) {
				t.Fatalf("Errors がない: %v", err)
			}
			if i > 0 && !reflect.DeepEqual(ds, first) {
				t.Fatalf("診断順が変わった: %v / %v", first, ds)
			}
			if len(ds) < 5 || ds[0].Path != "map[id=town].exits[1].mark" || ds[1].Path != "map[id=town].rows[4][5]" || ds[2].Path != "map[id=town].rows[4][1]" || ds[3].Path != "map[id=workshop].rows[4][2]" || ds[4].Path != "map[id=workshop].rows[4][1]" {
				t.Fatalf("地図 ID と数字の順が不正: %v", ds)
			}
			first = ds
		}
	})
}
