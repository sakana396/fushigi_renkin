# REV-037: Issue #69（core: 戦闘の出来事と、届く相手の判定）のレビュー

- 対象: ブランチ `it04/69-battle-events`（Issue #69）
- 日付: 2026-10-15
- 実装: Codex
- レビュー: spec-reviewer（仕様適合レビュー）。ミューテーションの再確認: Claude Code
- 承認: 不要（要求・設計・企画の文書と ADR は変更していない）
- 結果: 指摘を修正して承認

## 検証

- `gofmt -l .` は出力なし。`GOOS=windows go vet ./...` と `GOOS=windows go build -buildvcs=false ./cmd/fushigi` は成功
- `go test` は core・defload・data・tools/tccheck が成功。既存の戦闘のテストもそのまま通る（規則は変わっていない）
- `go run ./tools/tccheck` は整合 OK
- Log に 1 行足す場所を洗い出した: 仲間と敵の移動・攻撃、待機、use、damage、heal、attack_up、speed_up、fall、勝ち、負け、報酬の破棄、抽選の失敗。すべてが `addEvent` を通り、出来事が 1 つずつ足される。項目の値は DES-BTL-05 の表と一致する
- 盤面の写しと BattleReport の Events・Units は複製で、写しは状態を変えた直後に取っている。start は StartBattle の最初にあり、Log に足さない
- AttackTargets・UseTargets の判定は BattleAttack・BattleUseItem の検査と一致し、状態を変えない
- 設計書 2.2 節の計算例（Events 7、Log 6、剣士の攻撃のあとの添字 4〜6）をテストにしている
- ミューテーション: 修正前は 58 通りのうち 14 通りが通り抜けた。修正後、Codex が 14 通りすべての検出を確かめ、Claude Code も「AttackTargets の射程を Range+1 にする」変異を TC-IF-07-07 が検出することを再確認した

## 指摘と対応

| # | 指摘 | 区分 | 対応 |
|---|---|---|---|
| R-01 | 写しが「直後」かを、仲間の移動と爆弾の damage で見分けられない | 設計どおりへの修正 | 各出来事の写しの位置と体力を確かめるようにした（Codex） |
| R-02 | AttackTargets・UseTargets の射程の境界を確かめていない | 設計どおりへの修正 | Range と Range+1 の両方を、BattleAttack・BattleUseItem の結果と合わせて確かめた（Codex） |
| R-03 | 報酬の抽選に失敗したときの reward_lost のテストがない | 設計どおりへの修正 | 足した（Codex） |
| R-04 | 報酬の破棄が全数を捨てるケースだけ | 設計どおりへの修正 | 一部だけ受け取るケースにした（Codex） |
| R-05 | fall・lose・start の項目を確かめていない | 設計どおりへの修正 | 表の期待値と比べるようにした（Codex） |
| R-06 | heal・speed_up の From と、攻撃の Amount が強化込みであることを確かめていない | 設計どおりへの修正 | 確かめるようにした（Codex） |

## 実装者からの提案と判断

| # | 提案 | 判断 |
|---|---|---|
| P-01 | AttackTargets・UseTargets は、戦闘なし・今の番が仲間でない・倒れているときに nil を返す | 設計の範囲内 |
| P-02 | 返すマスの順は Units の添字の順 | 設計の範囲内 |
| P-03 | `addEvent` で出来事と Log の追加をまとめる（start は Log に足さない） | 設計の範囲内 |
| P-04 | AttackTargets でも盤面の中かを確かめる | 設計の範囲内 |
