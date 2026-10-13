# トレーサビリティマトリクス

- 文書 ID: TRC-001
- 対象: 反復 1（IT01）の範囲。反復ごとに行を追加する

要求から設計・テスト・実装までの対応を表す。実装の担当者（エージェントを含む）は、自分が担当した要求の行だけを更新する。

| 要求 | Issue | 設計 | テストケース | 実装 | 状態 |
|---|---|---|---|---|---|
| FR-ALC-01 | #5 | DES-ALC-01, DES-ALC-02 | TC-FR-ALC-01-01〜04 | `internal/core/craft.go` | 実装済み |
| FR-ALC-02 | #5, #21 | DES-ALC-01, DES-ALC-02, DES-WLD-01 | TC-FR-ALC-02-01〜05 | `internal/core/craft.go` | 実装済み（入力枠番号を 1 始まりに修正） |
| FR-ALC-03 | #5, #8 | defs.RecipeDef | TC-FR-ALC-03-01〜04 | `internal/defs/defs.go`, `internal/core/craft.go`, `data/base/recipes.toml` | 実装済み |
| FR-ALC-04 | #5 | DES-QLT-02, DES-QLT-03 | TC-FR-ALC-04-01〜05 | `internal/core/craft.go`, `internal/core/quality.go` | 実装済み |
| FR-ALC-06 | #5, #8 | defs.RecipeDef | TC-FR-ALC-06-01〜04 | `internal/core/craft.go`, `data/base/recipes.toml` | 実装済み |
| FR-ITM-01 | #4, #8 | DES-DEF-02, DES-VAL-03 | TC-FR-ITM-01-01 | `internal/core/inventory.go`, `internal/defs/defs.go`, `internal/defload/load.go`, `data/base/items.toml` | 実装済み |
| FR-ITM-02 | #4 | DES-VAL-05, DES-QLT-02 | TC-FR-ITM-02-01〜04 | `internal/core/inventory.go`, `internal/core/quality.go` | 一部実装（定義データの検証は #3） |
| FR-ITM-03 | #4 | DES-INV-01 | TC-FR-ITM-03-01〜07 | `internal/core/inventory.go` | 実装済み |
| FR-ITM-04 | #4, #7 | DES-QLT-01, DES-QLT-04 | TC-FR-ITM-04-01〜06 | `internal/core/quality.go`, `internal/core/golem.go` | 実装済み |
| FR-ITM-05 | #5 | DES-QLT-02, DES-QLT-04 | TC-FR-ITM-05-01〜03 | `internal/core/craft.go`, `internal/core/quality.go` | 一部実装（手作りの製品生成まで。自動生産は #7） |
| FR-MAT-01 | #6, #21 | DES-GAT-01, DES-GAT-02, DES-WLD-01 | TC-FR-MAT-01-01〜07, TC-IF-01-01（手動） | `internal/core/gather.go`, `internal/ui/game.go` | 実装済み |
| FR-MAT-02 | #6, #8, #21 | DES-ECO-02 | TC-FR-MAT-02-01〜04, TC-IF-01-02（手動） | `internal/core/economy.go`, `internal/ui/game.go`, `data/base/game.toml` | 実装済み |
| FR-ECO-01 | #6 | DES-ECO-01 | TC-FR-ECO-01-01〜03 | `internal/core/economy.go` | 実装済み |
| FR-ECO-02 | #6 | DES-ECO-01 | TC-FR-ECO-02-01〜03 | `internal/core/economy.go` | 実装済み |
| FR-ECO-04 | #6 | World.Buy, World.BuildFacility | TC-FR-ECO-04-01〜03 | `internal/core/economy.go` | 実装済み |
| FR-TWN-02 | #6 | DES-TWN-01, DES-TWN-02 | TC-FR-TWN-02-01〜03 | `internal/core/economy.go`, `internal/core/world.go` | 実装済み |
| FR-GOL-01 | #7 | World.BuildGolem, DES-GOL-06 | TC-FR-GOL-01-01〜03 | `internal/core/golem.go` | 実装済み |
| FR-GOL-02 | #7, #21 | DES-GOL-01, DES-GOL-03, DES-GOL-04, DES-WLD-01 | TC-FR-GOL-02-01〜04 | `internal/core/golem.go`, `internal/core/world.go` | 実装済み（入力枠番号を 1 始まりに修正） |
| FR-GOL-03 | #7 | DES-GOL-01, DES-GOL-03, DES-GOL-04 | TC-FR-GOL-03-01〜03 | `internal/core/golem.go` | 実装済み（反復 1 は在庫との受け渡し） |
| FR-GOL-06 | #7 | DES-TIME-01, DES-GOL-05 | TC-FR-GOL-06-01〜03 | `internal/core/world.go`, `internal/core/golem.go` | 実装済み |
| FR-MOD-01 | #3 | DES-PKG-03, DES-PKG-06 | TC-FR-MOD-01-01〜03 | `internal/defload/load.go` | 実装済み |
| FR-MOD-02 | #3, #8 | DES-PKG-03, DES-PKG-06 | TC-FR-MOD-02-01〜05 | `internal/defload/load.go`, `data/base.go`, `data/base/*.toml` | 実装済み |
| FR-MOD-04 | #3, #9, #21 | DES-VAL-01〜08, DES-PKG-01 | TC-FR-MOD-04-01〜10 | `internal/defload/load.go`, `cmd/fushigi/main.go`, `cmd/fushigi/error_windows.go`, `cmd/fushigi/error_other.go` | 実装済み |
| NFR-03 | #21 | DES-PKG-01 | TC-FR-MOD-04-10（手動） | `cmd/fushigi/error_windows.go`, `cmd/fushigi/error_other.go` | 実装済み（Windows 11 で MessageBox を確認） |
| NFR-04 | #2 | DES-PKG-04 | TC-NFR-04-01〜03 | `internal/core/dependency_test.go`, `.github/workflows/ci.yml` | 実装済み |
| NFR-06 | #4, #7 | DES-TIME-01, DES-GOL-05 | TC-NFR-06-01〜04 | `internal/core/world.go`, `internal/core/golem.go` | 実装済み（反復 1 の自動生産） |
| IF-01 | #9, #21 | DES-PKG-05、DES-IT01 4 章、DES-TIME-01 | TC-IF-01-01〜09（手動） | `internal/ui/game.go`, `cmd/fushigi/main.go` | マウス操作は実装済み（TC-IF-01-06〜09 の成功の流れは確認待ち）。キーボード操作は反復 1 の対象外 |
| IF-02 | #9 | DES-PKG-05、DES-IT01 4 章 | TC-IF-02-01（手動） | `internal/ui/game.go` | 実装済み |
| QA-02 | — | tools/tccheck | TC-QA-02-01〜04 | `tools/tccheck`、CI | 実装済み |
