# トレーサビリティマトリクス

- 文書 ID: TRC-001
- 対象: 反復 1（IT01）の範囲。反復ごとに行を追加する

要求から設計・テスト・実装までの対応を表す。実装の担当者（エージェントを含む）は、自分が担当した要求の行だけを更新する。

| 要求 | Issue | 設計 | テストケース | 実装 | 状態 |
|---|---|---|---|---|---|
| FR-ALC-01 | #5 | DES-ALC-01, DES-ALC-02 | TC-FR-ALC-01-01〜04 | `internal/core/craft.go` | 実装済み |
| FR-ALC-02 | #5 | DES-ALC-01, DES-ALC-02 | TC-FR-ALC-02-01〜04 | `internal/core/craft.go` | 実装済み |
| FR-ALC-03 | #5 | defs.RecipeDef | TC-FR-ALC-03-01〜03 | `internal/defs/defs.go`, `internal/core/craft.go` | 実装済み |
| FR-ALC-04 | #5 | DES-QLT-02, DES-QLT-03 | TC-FR-ALC-04-01〜05 | `internal/core/craft.go`, `internal/core/quality.go` | 実装済み |
| FR-ALC-06 | #5 | defs.RecipeDef | TC-FR-ALC-06-01〜03 | `internal/core/craft.go` | 実装済み |
| FR-ITM-01 | #4 | DES-DEF-02, DES-VAL-03 | | `internal/core/inventory.go`（種類・品質のキー） | 一部実装（属性の検証は #3） |
| FR-ITM-02 | #4 | DES-VAL-05, DES-QLT-02 | TC-FR-ITM-02-01〜04 | `internal/core/inventory.go`, `internal/core/quality.go` | 一部実装（定義データの検証は #3） |
| FR-ITM-03 | #4 | DES-INV-01 | TC-FR-ITM-03-01〜07 | `internal/core/inventory.go` | 実装済み |
| FR-ITM-04 | #4, #7 | DES-QLT-01, DES-QLT-04 | TC-FR-ITM-04-01〜05 | `internal/core/quality.go` | 一部実装（自動生産への接続は #7） |
| FR-ITM-05 | #5 | DES-QLT-02, DES-QLT-04 | TC-FR-ITM-05-01〜03 | `internal/core/craft.go`, `internal/core/quality.go` | 一部実装（手作りの製品生成まで。自動生産は #7） |
| FR-MAT-01 | #6 | DES-GAT-01, DES-GAT-02 | | | 未着手 |
| FR-MAT-02 | #6 | DES-ECO-02 | | | 未着手 |
| FR-ECO-01 | #6 | DES-ECO-01 | | | 未着手 |
| FR-ECO-02 | #6 | DES-ECO-01 | | | 未着手 |
| FR-ECO-04 | #6 | World.Buy, World.BuildFacility | | | 未着手 |
| FR-TWN-02 | #6 | DES-TWN-01, DES-TWN-02 | | | 未着手 |
| FR-GOL-01 | #7 | World.BuildGolem, DES-GOL-06 | | | 未着手 |
| FR-GOL-02 | #7 | DES-GOL-01, DES-GOL-03, DES-GOL-04 | | | 未着手 |
| FR-GOL-03 | #7 | DES-GOL-01, DES-GOL-03, DES-GOL-04 | | | 未着手 |
| FR-GOL-06 | #7 | DES-TIME-01, DES-GOL-05 | | | 未着手 |
| FR-MOD-01 | #3 | DES-PKG-03, DES-PKG-06 | | | 未着手 |
| FR-MOD-02 | #3, #8 | DES-PKG-03, DES-PKG-06 | | | 未着手 |
| FR-MOD-04 | #3 | DES-VAL-01〜08 | | | 未着手 |
| NFR-04 | #2 | DES-PKG-04 | TC-NFR-04-01〜03 | `internal/core/dependency_test.go`, `.github/workflows/ci.yml` | 実装済み |
| NFR-06 | #4, #7 | DES-TIME-01, DES-GOL-05 | TC-NFR-06-01〜03 | `internal/core/world.go` | 一部実装（自動生産の決定性は #7） |
