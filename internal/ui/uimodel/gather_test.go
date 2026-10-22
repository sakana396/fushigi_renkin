package uimodel

import (
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func TestGatherEntry(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rank  int
		want  string
		enter bool
	}{
		{"TC-FR-FLD-03-30 解放前", 1, "ランク 2 で解放", false},
		{"TC-FR-FLD-03-30 解放ちょうど", 2, "", true},
		{"TC-FR-FLD-03-30 解放後", 3, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, enter := GatherEntry(2, tc.rank)
			if text != tc.want || enter != tc.enter {
				t.Fatalf("%q %v", text, enter)
			}
		})
	}
	t.Run("TC-FR-FLD-03-31 負傷の秒数", func(t *testing.T) {
		for _, tc := range []struct {
			ticks uint64
			want  string
		}{{0, ""}, {1, "負傷中（残り 1 秒）"}, {60, "負傷中（残り 1 秒）"}, {61, "負傷中（残り 2 秒）"}} {
			if got := InjuryText(tc.ticks); got != tc.want {
				t.Errorf("%d: %q", tc.ticks, got)
			}
		}
	})
}

func TestGatherInteraction(t *testing.T) {
	t.Run("TC-IF-08-22 決定時の操作", func(t *testing.T) {
		for _, tc := range []struct {
			kind core.TargetKind
			want string
		}{{core.TargetPick, "pick"}, {core.TargetEnemy, "battle"}, {core.TargetFixture, ""}, {core.TargetLot, ""}} {
			if got := GatherAction(core.Target{Kind: tc.kind}); got != tc.want {
				t.Errorf("%v: %q", tc.kind, got)
			}
		}
	})
	t.Run("TC-FR-FLD-05-80 拾う場所の表示", func(t *testing.T) {
		for _, tc := range []struct {
			ticks uint64
			name  string
			alpha float32
		}{{0, "拾う場所", 1}, {1, "あと 1 秒", 0.4}, {60, "あと 1 秒", 0.4}, {61, "あと 2 秒", 0.4}} {
			name, alpha := PickDisplay(tc.ticks)
			if name != tc.name || alpha != tc.alpha {
				t.Errorf("%d: %q %v", tc.ticks, name, alpha)
			}
		}
	})
	t.Run("TC-FR-FLD-05-81 向いたマスだけ強調", func(t *testing.T) {
		target := core.Target{Kind: core.TargetPick, X: 3, Y: 2}
		for _, tc := range []struct {
			x, y     int
			ok, want bool
		}{{3, 2, true, true}, {4, 2, true, false}, {3, 3, true, false}, {3, 2, false, false}} {
			if got := FacingCell(target, tc.ok, tc.x, tc.y); got != tc.want {
				t.Errorf("%+v", tc)
			}
		}
	})
	t.Run("TC-FR-FLD-05-82 敵の名前", func(t *testing.T) {
		c := defs.Catalog{Units: map[string]defs.UnitDef{"slime": {Name: "スライム"}}}
		for _, tc := range []struct {
			enemies []defs.Placement
			want    string
		}{{nil, ""}, {[]defs.Placement{{Unit: "slime"}}, "スライム"}, {[]defs.Placement{{Unit: "slime"}, {Unit: "slime"}}, "スライム たち"}} {
			site := defs.GatherDef{Battle: &defs.BattleDef{Enemies: tc.enemies}}
			if got := EnemyName(c, site); got != tc.want {
				t.Errorf("%q", got)
			}
		}
		if EnemyName(c, defs.GatherDef{}) != "" {
			t.Fatal("戦闘なし")
		}
	})
	t.Run("TC-FR-MAT-01-70 拾った物の全文", func(t *testing.T) {
		items := []core.Stack{{Key: core.ItemKey{ItemID: "herb", Quality: 20}, Count: 2}, {Key: core.ItemKey{ItemID: "wheat", Quality: 30}, Count: 1}}
		name := func(id string) string {
			if id == "herb" {
				return "薬草"
			}
			return "小麦"
		}
		if got := PickedText(items, name); got != "拾いました: 薬草 Q20 ×2、小麦 Q30 ×1" {
			t.Fatal(got)
		}
	})
}

func TestGatherBattleResult(t *testing.T) {
	t.Run("TC-IF-07-50 結果の窓の選択", func(t *testing.T) {
		r := &core.BattleReport{}
		if BattleResultWindow(true, r) != "" || BattleResultWindow(false, nil) != "" || BattleResultWindow(false, r) != "battle_result" {
			t.Fatal("戦闘終了時だけ開く")
		}
	})
	t.Run("TC-IF-07-51 勝敗と報酬と負傷の全文", func(t *testing.T) {
		for _, tc := range []struct {
			report core.BattleReport
			now    uint64
			want   []string
		}{{core.BattleReport{Won: true}, 0, []string{"風の草原: 勝利", "報酬: なし"}}, {core.BattleReport{Won: true, Rewards: []core.Stack{{Key: core.ItemKey{ItemID: "herb", Quality: 40}, Count: 2}}}, 0, []string{"風の草原: 勝利", "報酬: 薬草 Q40 ×2"}}, {core.BattleReport{InjuredUntil: 61}, 0, []string{"風の草原: 敗北", "負傷: 残り 2 秒"}}, {core.BattleReport{InjuredUntil: 61}, 61, []string{"風の草原: 敗北", "負傷: 治った"}}} {
			if got := BattleResultLines(tc.report, "風の草原", tc.now, func(string) string { return "薬草" }); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%v", got)
			}
		}
	})
	t.Run("TC-FR-FLD-03-32 街の出口の目印", func(t *testing.T) {
		if !GatherExitMarker(defs.ExitGather) || GatherExitMarker(defs.MapTown) {
			t.Fatal("採取地行きだけ矢印と強調")
		}
	})
}

func TestGatherMap(t *testing.T) {
	t.Run("TC-FR-FLD-01-10 地面と採取地の絵", func(t *testing.T) {
		for _, tc := range []struct {
			ch   byte
			town bool
			n    int
		}{{',', true, 0}, {'.', false, 48}, {'*', true, 5}, {'o', false, 6}, {'T', true, 4}, {'#', false, 40}} {
			town, n := MapTile(defs.MapDef{}, 0, 0, tc.ch)
			if town != tc.town || n != tc.n {
				t.Errorf("%c: %v %d", tc.ch, town, n)
			}
		}
	})
	t.Run("TC-FR-FLD-01-11 カメラと見える範囲", func(t *testing.T) {
		for _, tc := range []struct{ size, viewport, player, want int }{{1152, 1600, 0, 224}, {960, 868, 0, 0}, {960, 868, 480, -46}, {960, 868, 960, -92}} {
			if got := CameraOffset(tc.size, tc.viewport, tc.player); got != tc.want {
				t.Errorf("%+v: %d", tc, got)
			}
		}
		for _, tc := range []struct{ offset, count, viewport, first, last int }{{224, 24, 1600, 0, 24}, {-92, 20, 868, 1, 20}, {-1, 20, 48, 0, 2}} {
			a, b := VisibleRange(tc.offset, tc.count, tc.viewport)
			if a != tc.first || b != tc.last {
				t.Errorf("%+v: %d %d", tc, a, b)
			}
		}
	})
	t.Run("TC-FR-FLD-01-12 名前の位置と出口の矢印", func(t *testing.T) {
		for _, tc := range []struct{ x, y, w, h, wantX, wantY int }{{-8, 0, 100, 24, 0, 0}, {1590, 40, 100, 24, 1500, 16}, {224, 48, 100, 24, 224, 24}} {
			x, y := MapNamePosition(tc.x, tc.y, tc.w, tc.h, 1600)
			if x != tc.wantX || y != tc.wantY {
				t.Errorf("%+v: %d %d", tc, x, y)
			}
		}
		if got := ExitArrow(10, 20); got != [6]float32{34, 29, 34, 57, 22, 45} {
			t.Fatal(got)
		}
	})
}

func TestEnemyAppearance(t *testing.T) {
	for _, tc := range []struct {
		name, unit, unitName string
		tile                 int
		initial              string
	}{
		{"TC-FR-FLD-05-84 スライム", "slime", "スライム", 108, ""},
		{"TC-FR-FLD-05-84 甲殻虫", "beetle", "甲殻虫", 110, ""},
		{"TC-FR-FLD-05-84 オオカミ", "wolf", "オオカミ", 123, ""},
		{"TC-FR-FLD-05-84 石ゴーレム", "stone_golem", "石ゴーレム", 109, ""},
		{"TC-FR-FLD-05-85 MODの日本語名", "mod_slime", "青スライム", -1, "青"},
		{"TC-FR-FLD-05-85 MODの英語名", "mod_beast", "Beast", -1, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := defs.Catalog{Units: map[string]defs.UnitDef{"slime": {Name: "スライム"}}}
			c.Units[tc.unit] = defs.UnitDef{Name: tc.unitName}
			site := defs.GatherDef{Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: tc.unit}, {Unit: "slime"}}}}
			tile, initial := EnemyAppearance(c, site)
			if tile != tc.tile || initial != tc.initial {
				t.Fatalf("絵=%d 頭文字=%q、期待=%d %q", tile, initial, tc.tile, tc.initial)
			}
			if got := EnemyName(c, site); got != tc.unitName+" たち" {
				t.Fatalf("最初の敵の名前: %q", got)
			}
		})
	}
	t.Run("TC-FR-FLD-05-85 戦闘なし・敵なし・空の名前", func(t *testing.T) {
		for _, site := range []defs.GatherDef{{}, {Battle: &defs.BattleDef{}}, {Battle: &defs.BattleDef{Enemies: []defs.Placement{{Unit: "mod_empty"}}}}} {
			if tile, initial := EnemyAppearance(defs.Catalog{}, site); tile != -1 || initial != "" {
				t.Fatalf("絵=%d 頭文字=%q", tile, initial)
			}
		}
	})
}
