package uimodel

import (
	"fmt"
	"strings"

	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// GatherEntry は解放前の案内と、行くボタンの有無を返す（DES-UI-26）。
func GatherEntry(required, rank int) (string, bool) {
	if rank < required {
		return fmt.Sprintf("ランク %d で解放", required), false
	}
	return "", true
}

// Seconds は残りtickを秒に切り上げる（DES-PCK-02、DES-UI-26）。
func Seconds(ticks uint64) uint64 {
	if ticks == 0 {
		return 0
	}
	return 1 + (ticks-1)/60
}

// InjuryText は負傷中だけ一覧の案内を出す（DES-UI-26）。
func InjuryText(remaining uint64) string {
	if remaining == 0 {
		return ""
	}
	return fmt.Sprintf("負傷中（残り %d 秒）", Seconds(remaining))
}

// GatherAction は決定キーを採取地の対象の操作に対応付ける（DES-UI-26）。
func GatherAction(target core.Target) string {
	switch target.Kind {
	case core.TargetPick:
		return "pick"
	case core.TargetEnemy:
		return "battle"
	}
	return ""
}

// PickDisplay は空の場所の名前を残り秒数に置き換え、絵を薄くする（DES-UI-26）。
func PickDisplay(remaining uint64) (string, float32) {
	if remaining == 0 {
		return "拾う場所", 1
	}
	return fmt.Sprintf("あと %d 秒", Seconds(remaining)), 0.4
}

// FacingCell は同じ種類の設備が並んでも向いたマスだけを強調する（DES-UI-15・26）。
func FacingCell(target core.Target, facing bool, x, y int) bool {
	return facing && target.X == x && target.Y == y
}

// EnemyName は最初の敵の名前に、複数なら「たち」を付ける（DES-UI-26）。
func EnemyName(c defs.Catalog, site defs.GatherDef) string {
	if site.Battle == nil || len(site.Battle.Enemies) == 0 {
		return ""
	}
	name := c.Units[site.Battle.Enemies[0].Unit].Name
	if len(site.Battle.Enemies) > 1 {
		name += " たち"
	}
	return name
}

// StackText は採取結果と戦闘報酬を日本語名・品質・個数で並べる（DES-UI-26・29）。
func StackText(stacks []core.Stack, name func(string) string) string {
	parts := make([]string, 0, len(stacks))
	for _, stack := range stacks {
		parts = append(parts, fmt.Sprintf("%s Q%d ×%d", name(stack.Key.ItemID), stack.Key.Quality, stack.Count))
	}
	if len(parts) == 0 {
		return "なし"
	}
	return strings.Join(parts, "、")
}

// PickedText は拾った物を操作結果の文にする（DES-UI-26）。
func PickedText(stacks []core.Stack, name func(string) string) string {
	return "拾いました: " + StackText(stacks, name)
}

// BattleResultWindow は出来事の再生後、終了した戦闘の結果だけを開く（DES-UI-29）。
func BattleResultWindow(active bool, report *core.BattleReport) string {
	if !active && report != nil {
		return "battle_result"
	}
	return ""
}

// BattleResultLines は勝敗・報酬・現在の負傷を表示する（DES-UI-10・29）。
func BattleResultLines(report core.BattleReport, site string, now uint64, name func(string) string) []string {
	if report.Won {
		return []string{site + ": 勝利", "報酬: " + StackText(report.Rewards, name)}
	}
	injury := "負傷: 治った"
	if now < report.InjuredUntil {
		injury = fmt.Sprintf("負傷: 残り %d 秒", Seconds(report.InjuredUntil-now))
	}
	return []string{site + ": 敗北", injury}
}

// GatherExitMarker は街の採取地行き出口に矢印と明るい名前を付ける（DES-UI-28 U-3）。
func GatherExitMarker(to string) bool { return to == defs.ExitGather }
