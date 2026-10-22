package defload

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestMapReview は開始位置の境界と診断形式を確かめる（DES-VAL-20）。
func TestMapReview(t *testing.T) {
	for _, tc := range []struct{ id, value, reason string }{
		{"TC-FR-FLD-01-06", "y = 2", "開始位置が書かれていない"},
		{"TC-FR-FLD-01-07", "x = 1", "開始位置が書かれていない"},
		{"TC-FR-FLD-01-08", "x = 7, y = 2", "開始位置 (7, 2) が地図の外"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			_, err := loadMapText(strings.Replace(validMaps, "x = 1, y = 2", tc.value, 1))
			mapDiagnostics(t, err, "map[id=town].start: "+tc.reason)
		})
	}
	t.Run("TC-FR-FLD-01-09", func(t *testing.T) {
		broken := strings.NewReplacer("###1###", "#321###", "###2#####", "#321#####").Replace(validMaps)
		var first Errors
		for i := 0; i < 20; i++ {
			_, err := loadMapText(broken)
			var ds Errors
			if !errors.As(err, &ds) {
				t.Fatal(err)
			}
			if i > 0 && !reflect.DeepEqual(ds, first) {
				t.Fatalf("診断順が変わった: %v / %v", first, ds)
			}
			want := []string{"map[id=meadow].rows[4][3]", "map[id=meadow].rows[4][1]", "map[id=town].rows[4][2]", "map[id=town].rows[4][1]"}
			if len(ds) < len(want) {
				t.Fatal(ds)
			}
			for j, p := range want {
				if ds[j].Path != p {
					t.Fatalf("診断順: %v", ds)
				}
			}
			first = ds
		}
	})
}

// TestMapReviewSuppression は形・開始位置・設備・出口の誤りが後続の到達性診断を抑えることを確かめる（DES-VAL-18〜23）。
func TestMapReviewSuppression(t *testing.T) {
	for _, tc := range []struct {
		name, old, new, path, reason string
		absent                       []string
	}{
		{"設備不足", "#DMQ..#", "#.MQ#.#", "map[id=town]", "工房の扉（D）がない", []string{"開始位置から行けない", "開始位置から到着のマス"}},
		{"出口不正", `mark = "1"`, `mark = "12"`, "map[id=town].exits[0].mark", "mark は 1〜9 の 1 文字", []string{"開始位置から行けない", "開始位置から到着のマス"}},
		{"行長不正", "#DMQ..#", "#DMQ.#", "map[id=town].rows[1]", "行の長さが rows[0] と違う（6 文字。rows[0] は 7 文字）", []string{".start", "依頼板（Q）がない", "gather 行き", "開始位置から"}},
		{"未知文字", "#DMQ..#", "#DXQ..#", "map[id=town].rows[1][2]", `知らない文字 "X"`, []string{".start", "店（M）がない", "開始位置から"}},
	} {
		t.Run("TC-FR-FLD-04-38 "+tc.name, func(t *testing.T) {
			input := replaceOne(validMaps, `"#.....#", "#.....#"`, `"#..#..#", "#.#...#"`)
			if tc.name == "行長不正" || tc.name == "未知文字" {
				input = replaceOne(input, "x = 1, y = 2", "x = 0, y = 0")
			}
			_, err := loadMapText(replaceOne(input, tc.old, tc.new))
			ds := mapDiagnostics(t, err, tc.path+": "+tc.reason)
			for _, d := range ds {
				for _, absent := range tc.absent {
					if strings.Contains(d.Path+": "+d.Reason, absent) {
						t.Fatalf("不要な診断 %q: %v", absent, ds)
					}
				}
			}
		})
	}
	t.Run("TC-FR-FLD-04-37 到着先が別の出口", func(t *testing.T) {
		_, err := loadMapText(replaceOne(validMaps, `"#.....#", "#.....#"`, `"#.....#", "#..2..#"`))
		mapDiagnostics(t, err, "map[id=town].exits[0].dir: 到着のマス (3, 3) が床ではない")
	})
}
