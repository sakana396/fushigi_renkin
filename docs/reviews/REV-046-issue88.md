# REV-046: Issue #88（画像素材と assets パッケージ）のレビュー

- 対象: ブランチ `it05/88-assets`（Issue #88）
- 日付: 2026-10-16
- 実装: Codex（素材の 4 ファイルは、Claude Code が配布元から取得した zip から置いた。指摘の修正は Claude Code）
- レビュー: spec-reviewer（仕様適合レビュー）
- 承認: 不要（要求・設計・企画の文書と ADR は変更していない）
- 結果: 指摘を修正して承認

## 検証

- `gofmt -l .` は出力なし。`GOOS=windows go vet ./...` は成功（Linux では ui と cmd が X11 の不足でビルドできない。assets・core・defload・data・tools は Linux でも vet と test が成功）
- `go test ./internal/ui/assets/...` は成功。assets は X11 なしでビルドできる
- `go run ./tools/tccheck` は整合 OK（増えた警告は手動の TC-CST-04-05）
- assets の import は `bytes`・`embed`・`fmt`・`image`・`image/png` だけで、Ebitengine を使っていない（DES-UI-14、DES-PKG）
- 置いた 4 ファイルを、配布 zip（Tiny Town 1.1、Tiny Dungeon 1.0）の `Tilemap/tilemap_packed.png` と `License.txt` と `cmp` で比べ、作業ツリーとコミットされた blob の両方がバイト単位で一致することを確かめた。License.txt は CRLF のまま、手を加えていない（ADR-0003 決定 3）
- README の出典表に、ADR-0003 決定 3 の列がそろっていることを確かめた
- テストを変異で試した。大きさの違う画像・壊れた PNG・nil を返す・誤りを捨てる・License の入れ替えは検出した。修正前は、2 枚の画像の取り違えと License の改行の変換を検出できなかった（指摘 1、2）。修正後、Claude Code が「`TinyDungeon()` が tiny_town.png を読む」変異を入れ、TC-CST-04-07 が失敗することを確かめた

## 指摘と対応

| # | 指摘 | 区分 | 対応 |
|---|---|---|---|
| 1 | TC-CST-04-01・02 は大きさ（192×176）だけを確かめる。2 つの素材は同じ大きさなので、関数や PNG のファイルの取り違えが通る（中） | 設計どおりへの修正 | 埋め込んだ 4 つのバイト列の SHA-256 を配布物の値と照合する TC-CST-04-06 と、2 つの関数がそれぞれ自分の PNG の画像を返し、互いに異なることを確かめる TC-CST-04-07 を足した（Claude Code） |
| 2 | License.txt を「手を加えずに」置くことをテストで守れていない（軽微） | 設計どおりへの修正 | TC-CST-04-06 の照合に License を含めた（Claude Code） |

## 実装者の判断事項

| # | 判断 | 判定 |
|---|---|---|
| 1 | License.txt も `go:embed` し、テストで版と CC0 の文言を確かめる | 設計の範囲内（ADR-0003 決定 3 は置き場だけを決めている） |
| 2 | 読み込みの誤りに素材名を付けて包む | 設計の範囲内（DES-UI-14 の起動中止のときに原因が分かる） |
| 3 | README の照合を手動テスト TC-CST-04-05 にし、結果をテスト仕様書に書く | 設計の範囲内 |
