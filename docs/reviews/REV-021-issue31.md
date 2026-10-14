# REV-021: Issue #31（誤り文の表示名と 1 始まり、ゴーレムの所要 tick の参照）のレビュー

- 対象: ブランチ `it02/31-error-messages`（Issue #31）
- 日付: 2026-10-14
- 実装: Codex
- レビュー: spec-reviewer（仕様適合レビュー）
- 承認: 不要（要求・設計・企画の文書と ADR は変更していない）
- 結果: 承認（指摘を修正済み）

## 検証

- `gofmt -l .` は出力なし。`GOOS=windows go vet ./...` と `GOOS=windows go build -buildvcs=false ./...` は成功
- `go test` は core・defload・data・tools/tccheck が成功（ui と cmd/fushigi は Linux 環境に X11 がなくビルドできないため対象外。今回は core だけの変更）
- `go run ./tools/tccheck` は整合 OK（警告は反復 1 の手動のテストケース 12 件）
- core のテスト以外の全ファイルで `fmt.Errorf` と `errors.New` を洗い出した。目録に載っている ID はすべて表示名になり、ID をそのまま書いているのは目録にない ID の分岐だけで、番号はすべて 1 始まりであることを確かめた。`%w` で包む誤りの種類は変わっていない
- ミューテーション確認: 修正前は、`UnassignGolem`・`GolemDuration` の番号、`BuildFacility` の施設材料の名前、`GolemDuration` の速度（採取・レシピ）を壊してもテストが通った。修正後は 5 か所とも検出することを Claude Code が確かめた

## 指摘と対応

| # | 指摘 | 区分 | 対応 |
|---|---|---|---|
| R-01 | `UnassignGolem` の番号、`GolemDuration` の範囲外の番号、`BuildFacility` の施設材料の名前がテストで守られていない | 設計どおりへの修正 | TC-IF-02-19〜21 を追加（Codex） |
| R-02 | TestGolemDuration が速度 100 だけを使っており、速度を無視する実装と見分けられない | 設計どおりへの修正 | 速度 150 でレシピ 1 秒 → 40 tick、採取地 12 秒 → 480 tick の TC-FR-GOL-04-24〜25 を追加（Codex） |
| R-03 | `Inventory.Remove` の誤り文からアイテムの種類が消えた。World の経路では事前の検査で止まり、画面には出ない | 情報 | 対応不要。ui が Remove を直接呼ぶようになったら見直す |
| R-04 | TestWorldErrorMessages のサブテストが 1 つの World を共有し、順序に頼っている | 軽微 | サブテストごとに World を作るようにした（Codex） |

## 実装者からの提案と判断

| # | 提案 | 判断 |
|---|---|---|
| P-01 | `Inventory.Remove` は目録を持たないので、誤り文にアイテムを書かない | 設計の範囲内 |
| P-02 | `takeLowest` は種類を書かない誤りを返し、`BuildGolem` が表示名で包む | 設計の範囲内（`errors.Is` の値は変わらない） |
| P-03 | 採取の再使用待ちの誤り文に採取地の表示名を足す | 設計の範囲内 |
| P-04 | 表示名を `%q` で囲む | 設計の範囲内（見た目だけ） |
| P-05 | `GolemDuration` で、割り当て先が目録にない場合に `ErrUnknownID` を返す | 設計の範囲内（防御のためのコード） |
