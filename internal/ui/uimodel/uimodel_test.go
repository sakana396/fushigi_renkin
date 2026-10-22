package uimodel

import (
	"testing"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"

	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func TestWindowTarget(t *testing.T) {
	for _, tc := range []struct {
		name             string
		target           core.Target
		window, facility string
	}{
		{"TC-IF-08-61 工房", core.Target{Kind: core.TargetFixture, ID: defs.FixtureWorkshopDoor}, defs.FixtureWorkshopDoor, ""},
		{"TC-IF-08-62 区画", core.Target{Kind: core.TargetLot, ID: "market"}, "lot", "market"},
		{"TC-IF-08-63 対象外", core.Target{Kind: core.TargetEnemy, ID: "meadow"}, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, f := WindowForTarget(tc.target)
			if w != tc.window || f != tc.facility {
				t.Fatalf("%q %q", w, f)
			}
		})
	}
}

func TestWindowTextAndGeometry(t *testing.T) {
	t.Run("TC-IF-06-40 タブの初期値と保持", func(t *testing.T) {
		for _, tab := range []string{"錬金", "ゴーレム", "在庫"} {
			if WorkshopTab(tab) != tab {
				t.Fatal(tab)
			}
		}
		if WorkshopTab("") != "錬金" {
			t.Fatal("初期値")
		}
	})
	t.Run("TC-FR-TWN-07-60 区画の名前", func(t *testing.T) {
		if LotName("市場", 0) != "市場（空き地）" || LotName("市場", 2) != "市場 Lv2" {
			t.Fatal("名前")
		}
	})
	t.Run("TC-FR-TWN-07-61 区画の段階とタイル", func(t *testing.T) {
		catalog, err := defload.Load(data.Base)
		if err != nil {
			t.Fatal(err)
		}
		// 期待値は DES-DAT-04 5.1 節の全施設・全段階から取る。
		for id, looks := range map[string][3]string{
			"market":     {"", "stall", "house"},
			"workshop":   {"", "cottage", "house"},
			"golem_shed": {"", "cottage", "manor"},
			"inn":        {"", "cottage", "manor"},
		} {
			for level, want := range looks {
				if got := LotLook(catalog.Facilities[id], level); got != want {
					t.Errorf("%s 段階%d: got=%q want=%q", id, level, got, want)
				}
			}
		}
		// 期待値は DES-IT06 7章の区画タイル表。固定長の全要素を照合する。
		for look, want := range map[string][9]int{
			"":        {44, 45, 46, 56, 0, 58, 68, 83, 70},
			"stall":   {0, 0, 0, 0, 57, 0, 0, 0, 0},
			"cottage": {52, 53, 54, 64, 65, 66, 72, 86, 75},
			"house":   {48, 49, 50, 60, 61, 62, 84, 86, 87},
			"manor":   {48, 51, 50, 60, 63, 62, 84, 86, 87},
		} {
			if got := LotTiles(look); got != want {
				t.Errorf("%q: got=%v want=%v", look, got, want)
			}
		}
	})
	t.Run("TC-IF-08-64 向きの印", func(t *testing.T) {
		up := DirectionMarker(100, 100, defs.DirUp)
		down := DirectionMarker(100, 100, defs.DirDown)
		for i := 1; i < 6; i += 2 {
			if up[i] >= 68 || down[i] <= 116 {
				t.Fatal(up, down)
			}
		}
	})
	t.Run("TC-FR-PRG-05-60 日の文字列", func(t *testing.T) {
		if DayText(2) != "2 日目" || DayNotice(2) != "2 日目になりました。依頼板の依頼が入れ替わりました" {
			t.Fatal("日の表示")
		}
	})
}

func TestQuestLines(t *testing.T) {
	t.Run("TC-IF-06-41 依頼の条件と報酬", func(t *testing.T) {
		a, b := QuestLines(defs.QuestDef{MinQuality: 20, Count: 3, Reward: 45, Reputation: 1}, "小麦粉")
		if a != "小麦粉 Q20 以上 ×3" || b != "報酬: 45 円 / 評判 +1" {
			t.Fatal(a, b)
		}
	})
	t.Run("TC-IF-06-42 戻るボタンの位置", func(t *testing.T) {
		if ReturnX != 10 || ReturnY != 76 || ReturnWidth != 120 || ReturnHeight != 28 || GatherTop != 112 || ReturnY+ReturnHeight >= GatherTop {
			t.Fatal("見出しと一覧が重なる")
		}
	})
}

// 日替わりは過去の操作結果を置き換え、同じ更新のランクだけを併記する（DES-UI-27）。
func TestNoticeMessage(t *testing.T) {
	t.Run("TC-FR-PRG-05-60 知らせの置き換えと同時通知", func(t *testing.T) {
		for _, tc := range []struct {
			name, current string
			ranks         []string
			days          []int
			want          string
		}{
			{"知らせなし", "錬金しました", nil, nil, "錬金しました"},
			{"日だけ", "錬金しました", nil, []int{2}, "2 日目になりました。依頼板の依頼が入れ替わりました"},
			{"ランクだけ", "納品しました", []string{"工房ランク 2"}, nil, "工房ランク 2"},
			{"同じ更新", "納品しました", []string{"工房ランク 2", "工房ランク 3"}, []int{2}, "工房ランク 2／工房ランク 3／2 日目になりました。依頼板の依頼が入れ替わりました"},
			{"最新の日だけ", "前の知らせ", nil, []int{2, 3}, "3 日目になりました。依頼板の依頼が入れ替わりました"},
		} {
			if got := NoticeMessage(tc.current, tc.ranks, tc.days); got != tc.want {
				t.Errorf("%s: got=%q want=%q", tc.name, got, tc.want)
			}
		}
		message := "錬金しました"
		for day := 2; day <= 100; day++ {
			message = NoticeMessage(message, nil, []int{day})
			if message != DayNotice(day) {
				t.Fatalf("%d日目に過去の知らせが残る: %q", day, message)
			}
			if got := NoticeMessage(message, nil, nil); got != message {
				t.Fatalf("次の更新で変わる: %q", got)
			}
		}
	})
}
