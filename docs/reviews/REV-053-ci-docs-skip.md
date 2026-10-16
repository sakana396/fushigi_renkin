# REV-053: CI で文書だけの変更のときに Go の検査を省く（PR #100）

- 対象: ブランチ `chore/ci-skip-go-on-docs`（`.github/workflows/ci.yml`）
- 日付: 2026-10-16
- 作成: Claude Code
- レビュー: spec-reviewer（仕様適合レビュー）
- 承認: **要**（F-03。文書だけのプルリクエストで gofmt・go vet・go test を省くことが QA-01 を満たすという解釈）。要求・設計・企画の文書と ADR は変更していない。開発者の承認: **承認**（F-03 の解釈を開発者が承認した。「おk」）
- 結果: F-01・F-02・F-04 を修正した。F-03 は開発者が承認した

## 検証

- `gofmt -l .` は出力なし。`GOOS=windows go vet ./...` は成功
- `go test` は core・defload・defs・data・tools/tccheck が成功（ui と cmd は Linux に X11 のヘッダーがないためビルドできない）
- `go run ./tools/tccheck` は整合 OK（警告 164 件は手動テスト）。空のモジュールキャッシュと `GOPROXY=off` でも動くので、Ebitengine 用パッケージを省いても tccheck は動く
- PR #100 の CI は成功（ci.yml を変えるので、すべての段階が走った）
- 判定のシェルを GNU grep 3.11 で模擬した。空の差分 → code=true（安全側）、`docs/` と `.md` だけ → false、`data/base/*.toml` や `.github/` を含む → true、`tools/tccheck/testdata/**/*.md` だけ → false（F-01）、`.go` から `docs/*.md` への名前変更 → false（F-02）、日本語のファイル名（引用符付きで出る）→ true（安全側）
- `pull_request` の HEAD はマージコミットなので、`origin/<base>...HEAD` はプルリクエストの変更だけを並べる。`fetch-depth: 0` で比較元のブランチも取れる
- Go のコードで docs/ を読むのは tccheck だけ（`docs/04_test/*.md`、`docs/traceability.md`）。go:embed は `data/base/*.toml` と ui の画像・ライセンスだけ

## 指摘と対応

| # | 指摘 | 区分 | 対応 |
|---|---|---|---|
| F-01 | `\.md$` が `tools/tccheck/testdata/**/*.md`（tccheck の単体テストの入力）にも当たり、そこだけを変えると `go test` が省かれる | 設計どおりへの修正 | パスに `/testdata/` を含む変更があれば code=true にした |
| F-02 | `git diff --name-only` は名前変更で新しい名前しか出さないので、`.go` を docs/ の `.md` へ移すと検査が省かれる | 設計どおりへの修正 | `--no-renames` を付けた |
| F-03 | QA-01「プッシュごとに…実行すること」に対し、文書だけのプルリクエストへの push では gofmt・go vet・go test が走らない。Go の入力が変わらず、main への push では必ず走るので目的は保たれるが、文面どおりではない | 人の確認が必要 | 開発者が承認した。文書だけの変更は Go の入力を変えず、main への push で必ず 3 つとも走るので、QA-01 を満たすと解釈する |
| F-04 | `${{ github.base_ref }}` を run に直接埋め込んでいる | 設計の範囲内（任意） | `env` の `EVENT_NAME`・`BASE_REF` を経由するようにした |
| F-05 | 文書だけの経路は、このプルリクエストの CI では走っていない | 情報 | 上の模擬で確かめた。次の文書だけのプルリクエストで、残りの段階が skipped になり job が成功することを確かめる |

## 開発者に確認してほしい点

| # | 点 |
|---|---|
| Q-1 | 文書（`docs/` と `.md`。testdata を除く）だけのプルリクエストで gofmt・go vet・go test を省いても、main への push で必ず走るので QA-01 を満たすと解釈してよいか（F-03） |
