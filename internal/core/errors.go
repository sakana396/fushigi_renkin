package core

import "errors"

// 状態変更の失敗理由を表す誤り（DES-WLD-01）。
var (
	ErrNotEnoughItems = errors.New("アイテムが足りません")
	ErrNotEnoughMoney = errors.New("お金が足りません")
	ErrUnknownID      = errors.New("ID が見つかりません")
	ErrInvalidCount   = errors.New("個数が正しくありません")
	ErrInvalidQuality = errors.New("品質は 1〜200 で指定してください")
)
