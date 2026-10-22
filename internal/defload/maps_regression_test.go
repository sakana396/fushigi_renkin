package defload

import (
	"os"
	"strings"
	"testing"
)

// TestMapRegressions は配置と検証の抑止、採取地からの到達性を確かめる（DES-VAL-17・20〜23・25）。
func TestMapRegressions(t *testing.T) {
	data, err := os.ReadFile("testdata/it06/maps.toml")
	if err != nil {
		t.Fatal(err)
	}
	valid := string(data)
	battle, err := os.ReadFile("testdata/it06/battle.toml")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		id, input, want string
		absent          []string
	}{
		{"TC-FR-FLD-04-60", valid[:strings.Index(valid, "[[map]]")] + valid[strings.LastIndex(valid, "[[map]]"):], "map[id=town]: 地図がない", nil},
		{"TC-FR-FLD-04-61", strings.NewReplacer("T,,,,,,,T", "T3,,,,,,T", `look="road"}]`, `look="road"},{mark="3",to="meadow",dir="up",look="road"}]`).Replace(valid), "map[id=town].exits[1].to: 行き先がない: meadow（必要な行き先は gather）", nil},
		{"TC-FR-FLD-01-58", replaceOne(valid, "T,,,,,,MT", "TE,,,,,MT"), "map[id=town].rows[4][1]: 敵（E）は採取地にしか置けない", nil},
		{"TC-FR-FLD-01-59", replaceOne(valid, "T,,,,,,MT", "To,,,,,MT"), "map[id=town].rows[4][1]: 拾う場所（o）は採取地にしか置けない", nil},
		{"TC-FR-FLD-01-60", replaceOne(valid, "T,,,,,,MT", "T,,,,,,,T"), "map[id=town]: 店（M）がない", nil},
		{"TC-FR-FLD-01-61", replaceOne(replaceOne(valid, "T*,,,,T", "T*DMQ,T"), "T,,o,,T", "T,,,,,T"), "", nil},
		{"TC-FR-FLD-04-62", strings.NewReplacer("T,aaa,,QT", "T,aaa,#QT", "T,aaa,,,T", "T,aa,,,#T").Replace(valid), `map[id=town].lots[0]: 区画 "a" が 3 × 3 の四角になっていない`, []string{"開始位置から行けない", "開始位置から到着のマス"}},
		{"TC-FR-FLD-04-63", strings.NewReplacer("T*,,,ET", "T*,,#ET", "T,,o,,T", "T,,o,#T").Replace(string(battle)), "map[id=meadow].rows[1][5]: 開始位置から行けない: 敵", nil},
		{"TC-FR-FLD-04-64", strings.NewReplacer("T*,,,,T", "T*,,#DT", "T,,o,,T", "T,,o,#T").Replace(valid), "map[id=meadow].rows[1][5]: 開始位置から行けない: 工房の扉", nil},
		{"TC-FR-FLD-01-62", replaceOne(valid, "start = {x=5,y=5}", "start = {x=-1,y=5}"), "map[id=town].start: 開始位置 (-1, 5) が地図の外", []string{"開始位置から"}},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			c, err := loadMapText(tc.input)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				if c.Maps["meadow"].Rows[1] != "T*DMQ,T" {
					t.Fatal(c.Maps["meadow"])
				}
				return
			}
			ds := mapDiagnostics(t, err, tc.want)
			for _, d := range ds {
				for _, a := range tc.absent {
					if strings.Contains(d.Path+": "+d.Reason, a) {
						t.Fatalf("不要な診断 %q: %v", a, ds)
					}
				}
			}
		})
	}
}
