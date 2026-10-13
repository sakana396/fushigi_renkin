package core

import "errors"

// 状態変更の失敗理由を表す誤り（DES-WLD-01）。
var (
	ErrNotEnoughItems   = errors.New("アイテムが足りません")
	ErrNotEnoughMoney   = errors.New("お金が足りません")
	ErrUnknownID        = errors.New("ID が見つかりません")
	ErrInvalidCount     = errors.New("個数が正しくありません")
	ErrInvalidQuality   = errors.New("品質は 1〜200 で指定してください")
	ErrGatherCooldown   = errors.New("採取地は再使用待ちです")
	ErrFacilityMaxLevel = errors.New("施設は最大段階です")
)

// ErrInvalidSelection は入力枠の条件に合わない素材を表す（DES-ALC-02）。
var ErrInvalidSelection = errors.New("入力枠の条件に合わない素材です")
