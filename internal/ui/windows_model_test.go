package ui

import (
	"reflect"
	"testing"

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
			w, f := windowForTarget(tc.target)
			if w != tc.window || f != tc.facility {
				t.Fatalf("%q %q", w, f)
			}
		})
	}
}

func TestWindowTextAndGeometry(t *testing.T) {
	t.Run("TC-IF-06-40 タブの初期値と保持", func(t *testing.T) {
		for _, tab := range []string{"錬金", "ゴーレム", "在庫"} {
			if workshopTab(tab) != tab {
				t.Fatal(tab)
			}
		}
		if workshopTab("") != "錬金" {
			t.Fatal("初期値")
		}
	})
	t.Run("TC-FR-TWN-07-60 区画の名前", func(t *testing.T) {
		if lotName("市場", 0) != "市場（空き地）" || lotName("市場", 2) != "市場 Lv2" {
			t.Fatal("名前")
		}
	})
	t.Run("TC-FR-TWN-07-61 区画のタイル", func(t *testing.T) {
		for _, look := range []string{"", "stall", "cottage", "house", "manor"} {
			tiles := lotTiles(look)
			if len(tiles) != 9 {
				t.Fatal(look)
			}
			if look == "" && tiles[7] != 83 {
				t.Fatal("立て札")
			}
			if look == "stall" && tiles[4] != 57 {
				t.Fatal("屋台")
			}
			if look != "" && look != "stall" && tiles[7] != 86 {
				t.Fatal("扉")
			}
		}
		if reflect.DeepEqual(lotTiles("cottage"), lotTiles("house")) || reflect.DeepEqual(lotTiles("house"), lotTiles("manor")) {
			t.Fatal("見た目が同じ")
		}
	})
	t.Run("TC-IF-08-64 向きの印", func(t *testing.T) {
		up := directionMarker(100, 100, defs.DirUp)
		down := directionMarker(100, 100, defs.DirDown)
		for i := 1; i < 6; i += 2 {
			if up[i] >= 68 || down[i] <= 116 {
				t.Fatal(up, down)
			}
		}
	})
	t.Run("TC-FR-PRG-05-60 日の文字列", func(t *testing.T) {
		if dayText(2) != "2 日目" || dayNotice(2) != "2 日目になりました。依頼板の依頼が入れ替わりました" {
			t.Fatal("日の表示")
		}
	})
}

func TestQuestLines(t *testing.T) {
	t.Run("TC-IF-06-41 依頼の条件と報酬", func(t *testing.T) {
		a, b := questLines(defs.QuestDef{MinQuality: 20, Count: 3, Reward: 45, Reputation: 1}, "小麦粉")
		if a != "小麦粉 Q20 以上 ×3" || b != "報酬: 45 円 / 評判 +1" {
			t.Fatal(a, b)
		}
	})
	t.Run("TC-IF-06-42 戻るボタンの位置", func(t *testing.T) {
		if returnX != 10 || returnY != 76 || returnWidth != 120 || returnHeight != 28 || gatherTop != 112 || returnY+returnHeight >= gatherTop {
			t.Fatal("見出しと一覧が重なる")
		}
	})
}
