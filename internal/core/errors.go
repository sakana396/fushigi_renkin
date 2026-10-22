package core

import "errors"

// 状態変更の失敗理由を表す誤り（DES-WLD-01）。
var (
	ErrInvalidDirection = errors.New("向きが正しくありません")
	ErrNotOnMap         = errors.New("主人公が地図の上にいません")
	ErrNotInGather      = errors.New("採取地の一覧にいません")
	ErrNoGolemSlots     = errors.New("ゴーレムの所有枠がありません")
	ErrNotEnoughItems   = errors.New("アイテムが足りません")
	ErrNotEnoughMoney   = errors.New("お金が足りません")
	ErrUnknownID        = errors.New("ID が見つかりません")
	ErrInvalidCount     = errors.New("個数が正しくありません")
	ErrInvalidQuality   = errors.New("品質は 1〜200 で指定してください")
	ErrNothingToPick    = errors.New("拾える物がありません")
	ErrPickNotReady     = errors.New("まだ拾えません")
	ErrNoEnemy          = errors.New("戦う相手がいません")
	ErrNoGatherMap      = errors.New("この採取地には地図がありません")
	ErrFacilityMaxLevel = errors.New("施設は最大段階です")
)

// ErrInvalidSelection は入力枠の条件に合わない素材を表す（DES-ALC-02）。
var ErrInvalidSelection = errors.New("入力枠の条件に合わない素材です")
