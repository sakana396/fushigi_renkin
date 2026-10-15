# ドキュメント一覧

本プロジェクトの成果物とその置き場所。工程が進むたびにこの表を更新する。開発は反復型で進める（[PLAN-001](00_plan/project-plan.md) 3 章）。

| 工程 | 成果物 | ファイル | 状態 |
|---|---|---|---|
| 計画 | プロジェクト計画書 | [00_plan/project-plan.md](00_plan/project-plan.md) | v0.2 |
| 企画 | 企画書 | [01_concept/concept.md](01_concept/concept.md) | ドラフト |
| 企画 | 企画案の比較（旧版） | [01_concept/concept-proposals.md](01_concept/concept-proposals.md) | 廃止（CON-002 に置き換え） |
| 企画 | 要求抽出記録（ヒアリング） | [01_concept/elicitation-notes.md](01_concept/elicitation-notes.md) | 追記中 |
| 企画 | 技術選定 ADR | [adr/0001-tech-stack.md](adr/0001-tech-stack.md) | 承認 |
| 要求 | 要求仕様書 (SRS) | [02_requirements/srs.md](02_requirements/srs.md) | v0.4 確定 |
| 要求 | ユースケース記述・図 | [02_requirements/use-cases.md](02_requirements/use-cases.md) | v0.4 承認（反復 4 の範囲を追加） |
| 反復 | 反復 1（プロトタイプ）計画書 | [iterations/it01/plan.md](iterations/it01/plan.md) | 承認 |
| 反復 | 反復 1 振り返り | [iterations/it01/retrospective.md](iterations/it01/retrospective.md) | v1.0 確定 |
| 反復 | 反復 2 計画書 | [iterations/it02/plan.md](iterations/it02/plan.md) | v1.0 確定 |
| 反復 | 反復 2 振り返り | [iterations/it02/retrospective.md](iterations/it02/retrospective.md) | v1.0 確定 |
| 反復 | 反復 3 計画書 | [iterations/it03/plan.md](iterations/it03/plan.md) | v1.0 確定 |
| 反復 | 反復 3 振り返り | [iterations/it03/retrospective.md](iterations/it03/retrospective.md) | v1.0 確定 |
| 反復 | 反復 4 計画書 | [iterations/it04/plan.md](iterations/it04/plan.md) | v1.0 確定 |
| 反復 | 反復 4 振り返り | [iterations/it04/retrospective.md](iterations/it04/retrospective.md) | v0.1 案 |
| 設計 | 設計書（反復 1）: アーキテクチャ・クラス図・状態遷移図・シーケンス図 | [03_design/design-it01.md](03_design/design-it01.md) | v0.3 承認 |
| 設計 | 設計書（反復 2）: 自動採取・錬金の素材選び・画面 | [03_design/design-it02.md](03_design/design-it02.md) | v0.2 承認 |
| 設計 | 設計書（反復 3）: 品質の効果・戦闘・依頼・場面ごとの画面 | [03_design/design-it03.md](03_design/design-it03.md) | v0.2 承認 |
| 設計 | 設計書（反復 4）: 工房のランクと解放・戦闘の出来事の表示 | [03_design/design-it04.md](03_design/design-it04.md) | v0.2 承認 |
| 設計 | 定義データ形式 ADR | [adr/0002-definition-data-format.md](adr/0002-definition-data-format.md) | 承認 |
| 試験 | テスト仕様書（反復 1） | [04_test/test-spec-it01.md](04_test/test-spec-it01.md) | 雛形 |
| 試験 | テスト仕様書（反復 2） | [04_test/test-spec-it02.md](04_test/test-spec-it02.md) | 反復 2 の範囲を記入済み |
| 試験 | テスト仕様書（反復 3） | [04_test/test-spec-it03.md](04_test/test-spec-it03.md) | 雛形 |
| 試験 | テスト仕様書（反復 4） | [04_test/test-spec-it04.md](04_test/test-spec-it04.md) | 雛形 |
| 横断 | トレーサビリティマトリクス | [traceability.md](traceability.md) | 反復 1〜4 の行を用意 |
| 横断 | レビュー記録 | [reviews/](reviews/) | 記録中 |

## エージェント向けのルール

コーディングエージェントの共通ルールはリポジトリ直下の [AGENTS.md](../AGENTS.md) にある。

## ID 規約

トレーサビリティのため、要素には次の ID を振る。

| 接頭辞 | 対象 | 例 |
|---|---|---|
| `FR-<区分>-` | 機能要求（区分: ALC, ITM, MAT, ECO, TWN, GOL, RD, BTL, PRG, GFT, CHR, SYS, MOD） | `FR-ALC-04` |
| `NFR-` | 非機能要求 | `NFR-06` |
| `IF-` | 外部インタフェース要求 | `IF-03` |
| `CST-` | 制約 | `CST-02` |
| `QA-` | 品質保証の要求 | `QA-01` |
| `TBD-` | 要求仕様書の未決事項 | `TBD-03` |
| `UC-` | ユースケース | `UC-02` |
| `DES-` | 設計要素（モジュール・クラス） | `DES-ALC-01` |
| `TC-` | テストケース | `TC-FR-ALC-04-01` |
| `ADR-` | 設計判断記録 | `ADR-0001` |
| `E-` | 要求抽出記録の発言 | `E-44` |
| `Q-` | 企画書の未決事項 | `Q-01` |

## 図

UML は [Mermaid](https://mermaid.js.org/) で Markdown に直接書く（GitHub 上でそのまま描画され、差分も追える）。
