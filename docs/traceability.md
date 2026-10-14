# トレーサビリティマトリクス

- 文書 ID: TRC-001
- 対象: 反復 1（IT01）〜反復 3（IT03）の範囲。反復ごとに行を追加する

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
| FR-GOL-06 | #7, #29 | DES-TIME-01, DES-GOL-05, DES-GOL-08 | TC-FR-GOL-06-01〜07 | `internal/core/world.go`, `internal/core/golem.go` | 実装済み（採取も tick で進む） |
| FR-MOD-01 | #3 | DES-PKG-03, DES-PKG-06 | TC-FR-MOD-01-01〜03 | `internal/defload/load.go` | 実装済み |
| FR-MOD-02 | #3, #8, #44, #49 | DES-PKG-03, DES-PKG-06, DES-DEF-05〜07 | TC-FR-MOD-02-01〜05、TC-FR-MOD-02-10〜12、TC-FR-MOD-02-20〜27 | `internal/defload/load.go`, `internal/defs/defs.go`, `data/base.go`, `data/base/*.toml`, `data/base_it03_test.go` | 実装済み（反復 3 の本体データを追加） |
| FR-MOD-04 | #3, #9, #21, #44 | DES-VAL-01〜13, DES-PKG-01 | TC-FR-MOD-04-01〜10、TC-FR-MOD-04-20〜47 | `internal/defload/load.go`, `cmd/fushigi/main.go`, `cmd/fushigi/error_windows.go`, `cmd/fushigi/error_other.go` | 実装済み（反復 3 の検証を追加） |
| NFR-03 | #21 | DES-PKG-01 | TC-FR-MOD-04-10（手動） | `cmd/fushigi/error_windows.go`, `cmd/fushigi/error_other.go` | 実装済み（Windows 11 で MessageBox を確認） |
| NFR-04 | #2 | DES-PKG-04 | TC-NFR-04-01〜03 | `internal/core/dependency_test.go`, `.github/workflows/ci.yml` | 実装済み |
| NFR-06 | #4, #7, #29, #46, #48 | DES-TIME-01, DES-GOL-05, DES-GAT-03, DES-BTL-03, DES-QST-01 | TC-NFR-06-01〜06, TC-NFR-06-30〜32 | `internal/core/world.go`, `internal/core/golem.go`, `internal/core/gather.go`, `internal/core/quest.go` | 実装済み（採取を含む自動生産・依頼。戦闘は #46） |
| IF-01 | #9, #21, #24, #33 | DES-PKG-05、DES-IT01 4 章、DES-TIME-01、DES-UI-04、DES-UI-05 | TC-IF-01-01〜10、TC-IF-01-20〜27（手動） | `internal/ui/game.go`, `internal/ui/cheat.go`, `cmd/fushigi/main.go` | 実装済み。マウスで UC-01〜08 を操作できる。キーボードはテスト用のチート（F11、F12）だけ。#33 で割り当て先の一覧とページ送りを追加（手動確認は合格） |
| IF-02 | #9, #31 | DES-PKG-05、DES-IT01 4 章、DES-WLD-02 | TC-IF-02-01（手動）、TC-IF-02-10〜21 | `internal/ui/game.go`, `internal/core/display_name.go`, `internal/core/craft.go`, `internal/core/golem.go`, `internal/core/gather.go`, `internal/core/economy.go`, `internal/core/inventory.go` | core の誤り文は実装済み。画面の表示は #32、#33 |
| QA-02 | — | tools/tccheck | TC-QA-02-01〜04 | `tools/tccheck`、CI | 実装済み |
| FR-GOL-04 | #29, #31, #33 | DES-GOL-07, DES-GOL-08, DES-GOL-09, DES-GAT-03, DES-UI-04 | TC-FR-GOL-04-01〜08、TC-FR-GOL-04-20〜25、TC-FR-GOL-04-40〜45（手動） | `internal/core/golem.go`, `internal/core/gather.go`, `internal/ui/game.go` | 実装済み（手動確認は合格） |
| FR-ALC-08 | #30, #32 | DES-ALC-04, DES-UI-01, DES-UI-03 | TC-FR-ALC-08-01〜07、TC-FR-ALC-08-20〜28（手動） | `internal/core/craft.go`, `internal/core/golem.go`, `internal/ui/game.go` | 実装済み（手動確認は合格） |
| FR-ITM-07 | #32 | DES-ITM-01, DES-UI-02 | TC-FR-ITM-07-01〜04（手動） | `internal/ui/game.go` | 実装済み（手動確認は合格） |
| FR-ITM-08 | #45, #47, #49, #50 | DES-QLT-05, DES-UI-09 | TC-FR-ITM-08-01〜10、TC-FR-MOD-02-20、TC-FR-MOD-02-27 | `internal/core/use_effect.go`, `data/base/items.toml` | core の効果計算を実装済み（戦闘・表示は後続 Issue）。使用効果の基本値を追加（効果の計算は #45） |
| FR-ITM-09 | #44, #45, #47, #49, #50 | DES-DEF-05, DES-QLT-05, DES-VAL-09, DES-VAL-10, DES-UI-09 | TC-FR-MOD-04-20〜25、TC-FR-MOD-04-39〜40、TC-FR-MOD-04-43、TC-FR-ITM-09-01〜08、TC-FR-MOD-02-20、TC-FR-MOD-02-27 | `internal/defs/defs.go`, `internal/defload/load.go`, `internal/core/use_effect.go`, `data/base/items.toml` | 定義・検証・core の効果計算を実装済み（戦闘・表示は後続 Issue）。定義・検証・本体データを実装済み（効果の計算は #45） |
| FR-BTL-01 | #44, #46, #49, #51 | DES-BTL-01, DES-BTL-02, DES-BTL-03, DES-VAL-12, DES-UI-08 | TC-FR-MOD-04-29〜31、TC-FR-MOD-04-35〜37、TC-FR-MOD-04-41〜42、TC-FR-MOD-04-44〜47、TC-FR-MOD-02-21〜22 | `internal/defs/defs.go`, `internal/defload/load.go`, `data/base/game.toml`, `data/base/gathers.toml`, `data/base/units.toml` | 定義と検証を実装済み（戦闘本体は #46）。定義・検証・本体データを実装済み（戦闘本体は #46） |
| FR-BTL-02 | #44, #46, #49 | DES-DEF-06, DES-VAL-11 | TC-FR-MOD-04-26〜28、TC-FR-MOD-02-21 | `internal/defs/defs.go`, `internal/defload/load.go`, `data/base/units.toml` | 定義と検証を実装済み（戦闘は #46）。定義・検証・本体データを実装済み（戦闘は #46） |
| FR-BTL-04 | #47, #51 | DES-BTL-02 |  |  | 未着手 |
| FR-BTL-05 | #45, #47 | DES-QLT-05 | TC-FR-ITM-08-01〜10、TC-FR-ITM-09-01〜08 | `internal/core/use_effect.go` | core の効果計算を実装済み（戦闘中の適用は #47） |
| FR-BTL-07 | #46, #51 | DES-BTL-02, DES-BTL-04 |  |  | 未着手 |
| FR-BTL-10 | #46, #50, #51 | DES-BTL-02, DES-UI-07 |  |  | 未着手 |
| FR-MAT-03 | #46 | DES-BTL-02, DES-DEF-06 |  |  | 未着手（一部） |
| FR-ECO-05 | #44, #48, #49, #50 | DES-DEF-07, DES-QST-01, DES-QST-02, DES-VAL-13 | TC-FR-MOD-04-32〜34、TC-FR-MOD-02-23〜24、TC-FR-MOD-04-32〜34, TC-FR-ECO-05-01〜13 | `internal/defs/defs.go`, `internal/defload/load.go`, `data/base/quests.toml`, `data/base/game.toml`, `internal/core/world.go`, `internal/core/quest.go` | 定義・検証・本体データを実装済み（依頼板と納品は #48）。定義・検証・依頼板・納品を実装済み（画面は #50） |
| IF-06 | #50, #51 | DES-UI-06, DES-UI-07, DES-UI-10 |  |  | 未着手 |
