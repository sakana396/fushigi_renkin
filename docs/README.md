# ドキュメント一覧

本プロジェクトの成果物とその置き場所。工程が進むたびにこの表を更新する。

| 工程 | 成果物 | ファイル | 状態 |
|---|---|---|---|
| 計画 | プロジェクト計画書 | [00_plan/project-plan.md](00_plan/project-plan.md) | ドラフト |
| 企画 | 企画書 | [01_concept/concept.md](01_concept/concept.md) | ドラフト |
| 企画 | 企画案の比較（旧版） | [01_concept/concept-proposals.md](01_concept/concept-proposals.md) | 廃止（CON-002 に置き換え） |
| 企画 | 要求抽出記録（ヒアリング） | [01_concept/elicitation-notes.md](01_concept/elicitation-notes.md) | 追記中 |
| 企画 | 技術選定 ADR | [adr/0001-tech-stack.md](adr/0001-tech-stack.md) | 承認 |
| 要求 | 要求仕様書 (SRS) | `02_requirements/srs.md` | 未着手 |
| 要求 | ユースケース記述・図 | `02_requirements/use-cases.md` | 未着手 |
| 設計 | アーキテクチャ設計書 | `03_design/architecture.md` | 未着手 |
| 設計 | 詳細設計（クラス図・状態遷移図・シーケンス図） | `03_design/detailed-design.md` | 未着手 |
| 試験 | テスト計画書・テスト仕様書 | `04_test/test-plan.md` | 未着手 |
| 横断 | トレーサビリティマトリクス | `traceability.md` | 未着手 |
| 横断 | レビュー記録 | `reviews/` | 未着手 |

## ID 規約

トレーサビリティのため、要素には次の ID を振る。

| 接頭辞 | 対象 | 例 |
|---|---|---|
| `FR-` | 機能要求 | `FR-010` |
| `NFR-` | 非機能要求 | `NFR-003` |
| `UC-` | ユースケース | `UC-02` |
| `DES-` | 設計要素（モジュール・クラス） | `DES-ALC-01` |
| `TC-` | テストケース | `TC-FR-010-01` |
| `ADR-` | 設計判断記録 | `ADR-0001` |
| `E-` | 要求抽出記録の発言 | `E-44` |
| `Q-` | 企画書の未決事項 | `Q-01` |

## 図

UML は [Mermaid](https://mermaid.js.org/) で Markdown に直接書く（GitHub 上でそのまま描画され、差分も追える）。
