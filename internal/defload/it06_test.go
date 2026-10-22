package defload

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// TestIT06Definitions は区画・採取地と定義の検証を確かめる（DES-VAL-17〜29）。
func TestIT06Definitions(t *testing.T) {
	data, err := os.ReadFile("testdata/it06/maps.toml")
	if err != nil {
		t.Fatal(err)
	}
	valid := string(data)
	cases := []struct{ id, old, new, want string }{
		{"TC-FR-FLD-04-50 / TC-FR-TWN-02-30", "", "", ""},
		{"TC-FR-FLD-04-51", `id = "meadow"` + "\nrows", `id = "unknown"` + "\nrows", "map[id=unknown].id: 地図の ID は town または採取地の ID"},
		{"TC-FR-FLD-04-52", `id = "meadow"` + "\nrows", `id = "town"` + "\nrows", "map[id=town].id: ID が重複"},
		{"TC-FR-FLD-04-53", `[[map]]` + "\nid = \"meadow\"", `[[map]]` + "\nid = \"missing\"", "map[id=meadow]: 地図がない"},
		{"TC-FR-FLD-01-50", `start = {x=5,y=5}`, "", "map[id=town].start: 開始位置が書かれていない"},
		{"TC-FR-FLD-01-51", `id = "meadow"` + "\nrows", `id = "meadow"` + "\nstart={x=1,y=1}\nrows", "map[id=meadow].start: 採取地の地図には開始位置を書かない"},
		{"TC-FR-FLD-01-52", `T,aaa,,DT`, `T,aaa,,,T`, "map[id=town]: 工房の扉（D）がない"},
		{"TC-FR-FLD-01-53", `T*,,,,T`, `Ta,,,,T`, "map[id=meadow].rows[1][1]: 区画（a）は街にしか置けない"},
		{"TC-FR-FLD-01-54", `T,,,,,,MT`, `T*,,,,,MT`, "map[id=town].rows[4][1]: 拾う場所（*）は採取地にしか置けない"},
		{"TC-FR-TWN-07-30", `T,aaa,,,T`, `T,aa,,,,T`, "map[id=town].lots[0]: 区画 \"a\" が 3 × 3 の四角になっていない"},
		{"TC-FR-TWN-07-31", `mark="a",facility="market"`, `mark="b",facility="market"`, "map[id=town].rows[1][2]: 区画の定義がない"},
		{"TC-FR-TWN-07-32", `facility="market"`, `facility="missing"`, "map[id=town].lots[0].facility: 存在しない施設: missing"},
		{"TC-FR-TWN-07-33", `mark="a",facility="market"`, `mark="j",facility="market"`, "map[id=town].lots[0].mark: mark は a〜i の 1 文字"},
		{"TC-FR-TWN-07-34", `lots = [{mark="a",facility="market"}]`, `lots = [{mark="a",facility="market"},{mark="a",facility="market"}]`, "map[id=town].lots[1].mark: mark が重複"},
		{"TC-FR-TWN-07-35", `look="stall"`, `look="castle"`, "facility[id=market].levels[0].look: 知らない見た目 \"castle\"（stall・cottage・house・manor が必要）"},
		{"TC-FR-TWN-07-36", `look="stall"`, `money=1`, "facility[id=market].levels[0].look: 書かれていない（地図を使うときは必須）"},
		{"TC-FR-FLD-05-30", `T*,,,,T`, `T,,,,,T`, ""},
		{"TC-FR-FLD-05-31", `T*,,,,T`, `TE,,,,T`, "map[id=meadow].rows: 敵（E）が 1 マスある（必要なマス数は 0）"},
		{"TC-FR-FLD-05-32", `respawn_seconds = 2`, `respawn_seconds = 0`, "gather[id=meadow].respawn_seconds: 1 以上 86400 以下にする（0）"},
		{"TC-FR-FLD-05-33", `respawn_seconds = 2`, `respawn_seconds = 86400`, ""},
		{"TC-FR-FLD-05-34", `respawn_seconds = 2`, `respawn_seconds = 86401`, "gather[id=meadow].respawn_seconds: 1 以上 86400 以下にする（86401）"},
		{"TC-FR-FLD-05-35", `respawn_seconds = 2`, ``, "gather[id=meadow].respawn_seconds: 書かれていない（地図を使うときは必須）"},
		{"TC-FR-PRG-05-30", `day_seconds = 300`, `day_seconds = 1`, ""},
		{"TC-FR-PRG-05-31", `day_seconds = 300`, `day_seconds = 86400`, ""},
		{"TC-FR-PRG-05-32", `day_seconds = 300`, `day_seconds = 0`, "game.day_seconds: 1 以上 86400 以下にする（0）"},
		{"TC-FR-PRG-05-33", `day_seconds = 300`, `day_seconds = 86401`, "game.day_seconds: 1 以上 86400 以下にする（86401）"},
		{"TC-FR-PRG-05-34", `day_seconds = 300`, ``, "game.day_seconds: 書かれていない（地図を使うときは必須）"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			input := valid
			if tc.old != "" {
				if !strings.Contains(input, tc.old) {
					t.Fatal("置換対象がない")
				}
				input = strings.Replace(input, tc.old, tc.new, 1)
			}
			c, err := loadMapText(input)
			if tc.want != "" {
				mapDiagnostics(t, err, tc.want)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantDay, wantRespawn := 300, 2
			switch tc.id {
			case "TC-FR-PRG-05-30":
				wantDay = 1
			case "TC-FR-PRG-05-31":
				wantDay = 86400
			case "TC-FR-FLD-05-33":
				wantRespawn = 86400
			}
			if tc.id == "TC-FR-FLD-05-30" && strings.Count(strings.Join(c.Maps["meadow"].Rows, ""), "*")+strings.Count(strings.Join(c.Maps["meadow"].Rows, ""), "o") != 2 {
				t.Fatal(c.Maps["meadow"])
			}
			if c.Game.DaySeconds != wantDay || c.Gathers["meadow"].RespawnSeconds != wantRespawn || len(c.Maps["town"].Lots) != 1 || c.Maps["town"].Lots[0].X != 2 || c.Maps["town"].Lots[0].Y != 1 || c.Facilities["market"].Levels[0].Look != "stall" || c.Facilities["market"].Levels[1].Look != "house" {
				t.Fatalf("読み込み結果: %+v", c)
			}
		})
	}
}

// TestIT06Boundaries は配置、秒数の省略・境界、区画の到達性を確かめる（DES-VAL-17〜29）。
func TestIT06Boundaries(t *testing.T) {
	data, err := os.ReadFile("testdata/it06/maps.toml")
	if err != nil {
		t.Fatal(err)
	}
	valid := string(data)
	noMaps := valid[:strings.Index(valid, "[[map]]")]
	battle := `[game]
day_seconds=300
battle_width=3
battle_height=3
injury_seconds=1
party=[{unit="hero",x=0,y=0}]
[[unit]]
id="hero"
role="front"
hp=1
attack=1
speed=1
move=1
range=1
[[unit]]
id="foe"
role="enemy"
hp=1
attack=1
speed=1
move=1
range=1
`
	withBattle := strings.Replace(valid, "[game]\nday_seconds = 300\n", battle, 1)
	withBattle = strings.Replace(withBattle, "[[map]]", `[gather.battle]
enemies=[{unit="foe",x=2,y=2}]
rolls=1
drops=[{item="herb",weight=1,quality_min=1,quality_max=1}]
[[map]]`, 1)
	withBattle = strings.Replace(withBattle, "T*,,,,T", "T*,,,ET", 1)
	cases := []struct{ id, input, want string }{
		{"TC-FR-FLD-04-54", strings.ReplaceAll(valid, `"meadow"`, `"gather"`), "gather[id=gather].id: 採取地の ID に town・gather は使えない"},
		{"TC-FR-FLD-04-55", strings.ReplaceAll(valid, `"meadow"`, `"town"`), "gather[id=town].id: 採取地の ID に town・gather は使えない"},
		{"TC-FR-FLD-04-56", strings.Replace(valid, "[game]\nday_seconds = 300\n", "", 1), "game.day_seconds: 書かれていない（地図を使うときは必須）"},
		{"TC-FR-FLD-04-57", strings.Replace(valid, `id = "meadow"`+"\nrows", `id = "meadow"`+"\nlots=[]\nrows", 1), "map[id=meadow].lots: lots は街にだけ書く"},
		{"TC-FR-TWN-07-37", strings.Replace(valid, `lots = [{mark="a",facility="market"}]`, `lots=[]`, 1), "map[id=town].lots: 施設 market の区画がない"},
		{"TC-FR-TWN-07-38", strings.Replace(strings.Replace(valid, "T,,,,,,MT", "T,bbb,,MT", 1), `lots = [{mark="a",facility="market"}]`, `lots=[{mark="a",facility="market"},{mark="b",facility="market"}]`, 1), "map[id=town].lots[1].facility: 同じ施設の区画が重複"},
		{"TC-FR-TWN-07-39", strings.ReplaceAll(strings.Replace(valid, "T,aaa,,QT", "T,aaa,aQT", 1), "T,aaa,,,T", "T,aa,,,,T"), "map[id=town].lots[0]: 区画 \"a\" が 3 × 3 の四角になっていない"},
		{"TC-FR-TWN-07-40", strings.ReplaceAll(valid, "aaa", "###"), "map[id=town].lots[0]: 区画 \"a\" が 3 × 3 の四角になっていない"},
		{"TC-FR-TWN-07-41", strings.NewReplacer("T,aaa,,QT", "T#aaa#,QT", "T,aaa,,,T", "T#aaa#,,T", "T,aaa,,DT", "T#aaa#,DT", "T,,,,,,MT", "T,###,,MT").Replace(valid), "map[id=town].lots[0]: 開始位置から行けない: 区画"},
		{"TC-FR-TWN-07-42", strings.ReplaceAll(noMaps, `look="stall"`, `look="manor"`), ""},
		{"TC-FR-TWN-07-43", strings.ReplaceAll(noMaps, `look="stall"`, `look="castle"`), "facility[id=market].levels[0].look: 知らない見た目 \"castle\"（stall・cottage・house・manor が必要）"},
		{"TC-FR-TWN-07-44", strings.ReplaceAll(strings.ReplaceAll(noMaps, `look="stall"`, `money=1`), `look="house"`, `money=2`), ""},
		{"TC-FR-TWN-02-31", strings.Replace(valid, `look="stall"`, `look="house"`, 1), ""},
		{"TC-FR-TWN-02-32", strings.Replace(valid, `look="house"`, `look="castle"`, 1), "facility[id=market].levels[1].look: 知らない見た目 \"castle\"（stall・cottage・house・manor が必要）"},
		{"TC-FR-TWN-07-45", strings.Replace(valid, `look="stall"`, `look="cottage"`, 1), ""},
		{"TC-FR-TWN-07-46", strings.Replace(valid, `look="stall"`, `look=""`, 1), "facility[id=market].levels[0].look: 知らない見た目 \"\"（stall・cottage・house・manor が必要）"},
		{"TC-FR-FLD-05-36", strings.ReplaceAll(strings.ReplaceAll(valid, "T*,,,,T", "T,,,,,T"), "T,,o,,T", "T,,,,,T"), "map[id=meadow].rows: 拾う場所は 1 マス以上必要"},
		{"TC-FR-FLD-05-37", strings.Replace(valid, "respawn_seconds = 2", "respawn_seconds = 1", 1), ""},
		{"TC-FR-FLD-05-38", strings.Replace(noMaps, "respawn_seconds = 2", "respawn_seconds = -1", 1), "gather[id=meadow].respawn_seconds: 1 以上 86400 以下にする（-1）"},
		{"TC-FR-FLD-05-39", strings.Replace(noMaps, "respawn_seconds = 2", "", 1), ""},
		{"TC-FR-FLD-05-40", withBattle, ""},
		{"TC-FR-FLD-05-41", strings.Replace(withBattle, "T*,,,ET", "T*,,,,T", 1), "map[id=meadow].rows: 敵（E）が 0 マスある（必要なマス数は 1）"},
		{"TC-FR-FLD-05-42", strings.Replace(withBattle, "T*,,,ET", "T*,,EET", 1), "map[id=meadow].rows: 敵（E）が 2 マスある（必要なマス数は 1）"},
		{"TC-FR-FLD-05-43", strings.Replace(strings.Replace(valid, "T*,,,,T", "T*#,,,T", 1), "T,,o,,T", "T#,#,,T", 1), "map[id=meadow].rows[1][1]: 開始位置から行けない: 拾う場所"},
		{"TC-FR-PRG-05-35", strings.Replace(noMaps, "day_seconds = 300", "", 1), ""},
		{"TC-FR-PRG-05-36", strings.Replace(noMaps, "day_seconds = 300", "day_seconds = -1", 1), "game.day_seconds: 1 以上 86400 以下にする（-1）"},
		{"TC-FR-PRG-05-37", strings.Replace(noMaps, "day_seconds = 300", "day_seconds = 0", 1), "game.day_seconds: 1 以上 86400 以下にする（0）"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			c, err := loadMapText(tc.input)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				switch tc.id {
				case "TC-FR-TWN-07-42":
					if len(c.Maps) != 0 || c.Facilities["market"].Levels[0].Look != "manor" {
						t.Fatal(c)
					}
				case "TC-FR-TWN-07-44":
					if len(c.Maps) != 0 || c.Facilities["market"].Levels[0].Look != "" || c.Facilities["market"].Levels[1].Look != "" {
						t.Fatal(c)
					}
				case "TC-FR-FLD-05-39":
					if len(c.Maps) != 0 || c.Gathers["meadow"].RespawnSeconds != 0 {
						t.Fatal(c)
					}
				case "TC-FR-PRG-05-35":
					if len(c.Maps) != 0 || c.Game.DaySeconds != 0 {
						t.Fatal(c)
					}
				case "TC-FR-TWN-02-31":
					if c.Facilities["market"].Levels[0].Look != "house" || c.Facilities["market"].Levels[1].Look != "house" {
						t.Fatal(c)
					}
				case "TC-FR-TWN-07-45":
					if c.Facilities["market"].Levels[0].Look != "cottage" {
						t.Fatal(c)
					}
				case "TC-FR-FLD-05-37":
					if c.Gathers["meadow"].RespawnSeconds != 1 {
						t.Fatal(c)
					}
				case "TC-FR-FLD-05-40":
					if c.Gathers["meadow"].Battle == nil || strings.Count(strings.Join(c.Maps["meadow"].Rows, ""), "E") != 1 {
						t.Fatal(c)
					}
				}
				return
			}
			mapDiagnostics(t, err, tc.want)
		})
	}
	t.Run("TC-FR-ECO-07-30 / TC-FR-ECO-07-31 / TC-FR-ECO-07-32 / TC-FR-ECO-07-33", func(t *testing.T) {
		for _, count := range []int{0, 5, 6, 7} {
			input := rankData[:strings.Index(rankData, "[[quest]]")]
			// ランクごとの評判を得る依頼を用意し、ランク 1 の件数だけを変える。
			for i := 0; i < count; i++ {
				input += fmt.Sprintf("[[quest]]\nid=\"r1_%d\"\nitem=\"potion\"\nmin_quality=1\ncount=1\nreward=1\nreputation=1\n", i)
			}
			for r := 2; r <= 3; r++ {
				input += fmt.Sprintf("[[quest]]\nid=\"r%d\"\nrank=%d\nitem=\"potion\"\nmin_quality=1\ncount=1\nreward=1\nreputation=1\n", r, r)
			}
			_, err := loadMapText(input)
			if count < 6 {
				mapDiagnostics(t, err, fmt.Sprintf("quest: ランク 1 の依頼が %d 件しかない（6 件以上が必要）", count))
			} else if err != nil {
				t.Fatal(err)
			}
		}
	})
}

// TestIT06DiagnosticSources は別ファイルの誤りでも元のファイルと位置を示すことを確かめる（DES-VAL-26〜29、FR-MOD-04）。
func TestIT06DiagnosticSources(t *testing.T) {
	data, err := os.ReadFile("testdata/it06/maps.toml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "[[map]]")
	for _, tc := range []struct{ id, old, new, path, reason string }{
		{"TC-FR-MOD-04-80", `look="stall"`, `look="castle"`, "facility[id=market].levels[0].look", `知らない見た目 "castle"（stall・cottage・house・manor が必要）`},
		{"TC-FR-MOD-04-81", "respawn_seconds = 2", "respawn_seconds = 0", "gather[id=meadow].respawn_seconds", "1 以上 86400 以下にする（0）"},
		{"TC-FR-MOD-04-82", "day_seconds = 300", "day_seconds = 0", "game.day_seconds", "1 以上 86400 以下にする（0）"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			_, err := Load(fstest.MapFS{"definitions.toml": &fstest.MapFile{Data: []byte(strings.Replace(text[:start], tc.old, tc.new, 1))}, "maps.toml": &fstest.MapFile{Data: []byte(text[start:])}})
			var ds Errors
			if !errors.As(err, &ds) {
				t.Fatal(err)
			}
			if len(ds) != 1 || ds[0].File != "definitions.toml" || ds[0].Path != tc.path || ds[0].Reason != tc.reason {
				t.Fatalf("診断: %v", err)
			}
		})
	}
}

// TestIT06ValidationOrder は敵の数の誤りでも到達性を検査する（DES-VAL-23・27）。
func TestIT06ValidationOrder(t *testing.T) {
	t.Run("TC-FR-FLD-05-44", func(t *testing.T) {
		data, err := os.ReadFile("testdata/it06/maps.toml")
		if err != nil {
			t.Fatal(err)
		}
		input := strings.NewReplacer("T*,,,,T", "T*#,,ET", "T,,o,,T", "T#,#,,T").Replace(string(data))
		_, err = loadMapText(input)
		mapDiagnostics(t, err, "map[id=meadow].rows: 敵（E）が 2 マスある（必要なマス数は 0）")
		mapDiagnostics(t, err, "map[id=meadow].rows[1][1]: 開始位置から行けない: 拾う場所")
	})
}
