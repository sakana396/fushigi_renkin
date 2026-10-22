package ui

import (
	"fmt"
	"strings"

	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// windowForTarget は工房・店・依頼板・区画の窓だけを選ぶ（DES-UI-22・23・25）。
func windowForTarget(target core.Target) (string, string) {
	if target.Kind == core.TargetLot {
		return "lot", target.ID
	}
	if target.Kind == core.TargetFixture {
		switch target.ID {
		case defs.FixtureWorkshopDoor, defs.FixtureShop, defs.FixtureQuestBoard:
			return target.ID, ""
		}
	}
	return "", ""
}

// workshopTab は未選択なら錬金、選択済みなら前のタブを返す（DES-UI-22）。
func workshopTab(tab string) string {
	if tab == "" {
		return "錬金"
	}
	return tab
}

// lotName は区画の施設名と現在の段階を組み立てる（DES-UI-24）。
func lotName(name string, level int) string {
	if level == 0 {
		return name + "（空き地）"
	}
	return fmt.Sprintf("%s Lv%d", name, level)
}

// lotLook は施設の現在の段階から絵を選び、段階0は空き地にする（DES-UI-24）。
func lotLook(facility defs.FacilityDef, level int) string {
	if level > 0 && level <= len(facility.Levels) {
		return facility.Levels[level-1].Look
	}
	return ""
}

// lotTiles は区画の左上から行順に Tiny Town の番号を返す（DES-UI-24）。
func lotTiles(look string) [9]int {
	switch look {
	case "stall":
		return [9]int{0, 0, 0, 0, 57, 0, 0, 0, 0}
	case "cottage":
		return [9]int{52, 53, 54, 64, 65, 66, 72, 86, 75}
	case "house":
		return [9]int{48, 49, 50, 60, 61, 62, 84, 86, 87}
	case "manor":
		return [9]int{48, 51, 50, 60, 63, 62, 84, 86, 87}
	default:
		return [9]int{44, 45, 46, 56, 0, 58, 68, 83, 70}
	}
}

// directionMarker は主人公の絵（y-32〜y+16）の外に上下の印を置く（DES-UI-28 U-4）。
func directionMarker(x, y int, dir defs.Dir) [6]float32 {
	switch dir {
	case defs.DirUp:
		return [6]float32{float32(x), float32(y - 44), float32(x - 6), float32(y - 36), float32(x + 6), float32(y - 36)}
	case defs.DirDown:
		return [6]float32{float32(x), float32(y + 28), float32(x - 6), float32(y + 20), float32(x + 6), float32(y + 20)}
	case defs.DirLeft:
		return [6]float32{float32(x - 17), float32(y), float32(x - 9), float32(y - 6), float32(x - 9), float32(y + 6)}
	default:
		return [6]float32{float32(x + 17), float32(y), float32(x + 9), float32(y - 6), float32(x + 9), float32(y + 6)}
	}
}

// dayText と dayNotice は日の表示と通知を組み立てる（DES-UI-27）。
func dayText(day int) string { return fmt.Sprintf("%d 日目", day) }
func dayNotice(day int) string {
	return dayText(day) + "になりました。依頼板の依頼が入れ替わりました"
}

// noticeMessage は新しい知らせで操作結果を置き換え、同じ更新のランクと日を並べる（DES-UI-11・27）。
func noticeMessage(current string, ranks []string, days []int) string {
	notices := append([]string(nil), ranks...)
	if len(days) > 0 {
		notices = append(notices, dayNotice(days[len(days)-1]))
	}
	if len(notices) == 0 {
		return current
	}
	return strings.Join(notices, "／")
}

// questLines は候補と受注済みで共通の依頼条件・報酬を組み立てる（DES-UI-25）。
func questLines(quest defs.QuestDef, name string) (string, string) {
	return fmt.Sprintf("%s Q%d 以上 ×%d", name, quest.MinQuality, quest.Count), fmt.Sprintf("報酬: %d 円 / 評判 +%d", quest.Reward, quest.Reputation)
}

// 窓の見出しと一覧の開始位置をそろえる（DES-UI-28 U-5）。
const returnX, returnY, returnWidth, returnHeight, gatherTop = 10, 76, 120, 28, 112
