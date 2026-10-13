@AGENTS.md

## Claude Code 固有の設定

- 共通ルールは上の AGENTS.md に書く。このファイルには Claude Code に固有のことだけを書く
- 要求 ID を指定した実装は、スキル `/implement-req`（`.claude/skills/implement-req/`）で行う
- `.go` ファイルを編集すると、フックが自動で `gofmt` をかける（`.claude/hooks/gofmt-on-edit.sh`）
- Go のビルド・テスト、Git の通常操作、`gh pr create` は許可済み（`.claude/settings.json`）。force push と `gh pr merge` は禁止している
