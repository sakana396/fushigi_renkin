# テスト仕様書（反復 5）

- 文書 ID: TST-IT05
- 版: 0.1（雛形）
- 対象: [反復 5 計画書](../iterations/it05/plan.md) の範囲、[設計書 DES-IT05](../03_design/design-it05.md)

## 1. 方針

[テスト仕様書（反復 3）](test-spec-it03.md) の 1 章の方針をそのまま使う。加えて次のとおりとする。

| 項目 | 内容 |
|---|---|
| 移動のテスト | 設計書 2.8 節の計算例 1〜31 を、そのままテストケースにする（前後の位置・向き・場面と、返る誤り） |
| 地図の検証のテスト | 設計書 3 章の DES-VAL-17〜24 について、通る例と誤りの例を規則ごとに作り、誤りの文にファイル名・位置・理由が入ることを確かめる |
| ファジング | 地図の読み込みにファジングのテストを入れる（NFR-08）。CI では短い時間だけ回し、見つかった入力は回帰のテストに足す |
| 本体の地図のテスト | 設計書 5 章の地図が検証を通り、5.2 節の歩く距離がすべて 8 マス以内であることを確かめる |
| 手動テスト | 期待結果は、画面に出る文・位置・向きを具体的に書く（反復 4 振り返り 5 章）。「初めて遊ぶ人が、操作を画面から読み取れるか」の観点を入れる |
| ID の割り当て | Issue ごとに使う番号の範囲を 2 章の各節に決める。範囲の中で 01 から詰めて使う必要はない |
| トレーサビリティ | 反復 3 と同じ（テスト仕様書（反復 3）1 章） |

## 2. テストケース

要求ごとに節を分ける。担当者は自分の Issue の範囲の行だけを書く。列は反復 1 と同じ（ID、対象、前提、入力・操作、期待結果、種別）。

### FR-FLD-01 工房と街の地図

- #85（地図の型と読み込み）: FR-FLD-01 の連番 01〜29
- #87（本体の地図）: FR-FLD-01 の連番 30〜39
- #89（地図の画面、手動）: FR-FLD-01 の連番 40〜59

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-FLD-01-01 | FR-FLD-01、DES-DEF-09 | 地図の単体テスト用目録 | 地図の寸法と地図外 | 幅・高さは 2、地図外は 0、(0,1) は K | 正常 |
| TC-FR-FLD-01-02 | FR-FLD-01、DES-DEF-09 | 地図の単体テスト用目録 | 空の地図 | 空の地図の幅は 0、Cell は 0 | 正常 |
| TC-FR-FLD-01-03 | FR-FLD-01、DES-DEF-09 | 地図の単体テスト用目録 | 文字の種類 | 床・壁・設備・出入り口・未知の文字が表どおり | 正常 |
| TC-FR-FLD-01-04 | FR-FLD-01、DES-DEF-09 | 地図の単体テスト用目録 | 設備の種類と表示名 | K/S/G/M/Q/F の種類と日本語名が表どおり | 正常 |
| TC-FR-FLD-01-05 | FR-FLD-01、DES-DEF-09 | 地図の単体テスト用目録 | IDで地図を引く | town が見つかり、workshop は見つからない | 正常 |
| TC-FR-FLD-01-06 | FR-FLD-01、DES-VAL-20 | 正常な二枚 | 工房の start.x を省略 | maps.toml の start に「開始位置が書かれていない」 | 異常 |
| TC-FR-FLD-01-07 | FR-FLD-01、DES-VAL-20 | 正常な二枚 | 工房の start.y を省略 | maps.toml の start に「開始位置が書かれていない」 | 異常 |
| TC-FR-FLD-01-08 | FR-FLD-01、DES-VAL-20 | 幅 7 の工房 | start を (7,2) にする | maps.toml の start に「開始位置 (7, 2) が地図の外」 | 境界 |
| TC-FR-FLD-01-09 | FR-FLD-01、DES-VAL-22 | 出口 2・3 に定義がない二枚 | 同じ読み込みを 20 回行う | 診断は毎回同じ順で、各地図の数字順になる | 正常 |

### FR-FLD-02 壁や設備を通り抜けない

- #86（core の移動と当たり判定）: FR-FLD-02 の連番 01〜39
- #89（地図の画面、手動）: FR-FLD-02 の連番 40〜59

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|

### FR-FLD-03 出入り口から場面を移る

- #86（core の出入り口と LeaveGather）: FR-FLD-03 の連番 01〜29
- #89（地図の画面、手動）: FR-FLD-03 の連番 40〜59

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|

### FR-FLD-04 地図の定義データと検証

- #85（地図の検証 DES-VAL-17〜24）: FR-FLD-04 の連番 01〜39（誤り文にファイル名・位置・理由が入ること（FR-MOD-04）の確認を含む）
- #87（本体の地図が検証を通る、歩く距離）: FR-FLD-04 の連番 40〜49

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-FLD-04-01 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 地図なし | 地図なしの目録を読める | 正常 |
| TC-FR-FLD-04-02 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 正常な二枚 | 両地図の ID・Name・Rows・Start と全出口の Mark・To・Dir・Look・X・Y が設計例どおり | 正常 |
| TC-FR-FLD-04-03 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 地図IDが未知 | maps.toml と位置・理由: map[id=market].id に未知の ID を報告 | 異常 |
| TC-FR-FLD-04-04 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 同じID | maps.toml と位置・理由: map[id=workshop].id に重複を報告 | 異常 |
| TC-FR-FLD-04-05 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 片方不足 | maps.toml と位置・理由: map[id=town] に地図不足を報告 | 異常 |
| TC-FR-FLD-04-06 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 空の行 | maps.toml と位置・理由: rows[1] に長さ 0 を報告 | 異常 |
| TC-FR-FLD-04-07 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 長さが違う | maps.toml と位置・理由: rows[1] に 6 文字、rows[0] は 7 文字と報告 | 異常 |
| TC-FR-FLD-04-08 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 未知の文字 | maps.toml と位置・理由: rows[1][2] に未知の X を報告 | 異常 |
| TC-FR-FLD-04-09 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 多バイト文字の位置 | maps.toml と位置・理由: rows[1][2] に未知の「あ」を報告 | 異常 |
| TC-FR-FLD-04-10 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | start省略 | maps.toml と位置・理由: start に書き忘れを報告 | 異常 |
| TC-FR-FLD-04-11 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | startは壁 | maps.toml と位置・理由: start に (0,0) が壁と報告 | 異常 |
| TC-FR-FLD-04-12 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 必須設備不足 | maps.toml と位置・理由: map[id=workshop] に G 不足を報告 | 異常 |
| TC-FR-FLD-04-13 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 出口markは0不可 | maps.toml と位置・理由: exits[0].mark に 0 は使えないと報告 | 異常 |
| TC-FR-FLD-04-14 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 出口の定義不足 | maps.toml と位置・理由: rows[4][3] に数字 2 の定義不足を報告 | 異常 |
| TC-FR-FLD-04-15 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 行き先がない | maps.toml と位置・理由: exits[0].to に market が存在しないと報告 | 異常 |
| TC-FR-FLD-04-16 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 向きが未知 | maps.toml と位置・理由: exits[0].dir に未知の向きを報告 | 異常 |
| TC-FR-FLD-04-17 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 絵が未知 | maps.toml と位置・理由: exits[0].look に未知の絵を報告 | 異常 |
| TC-FR-FLD-04-18 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 到着先が壁 | maps.toml と位置・理由: exits[0].dir に到着のマス (3,3) が壁と報告 | 異常 |
| TC-FR-FLD-04-19 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 同じ行き先 | maps.toml と位置・理由: exits[1].to に行き先の重複を報告 | 異常 |
| TC-FR-FLD-04-20 | FR-FLD-04、DES-VAL-23 | 地図の単体テスト用目録 | 設備に行けない | maps.toml と位置・理由: rows[1][5] に施設の看板へ行けないと報告 | 異常 |
| TC-FR-FLD-04-21 | FR-FLD-04、DES-VAL-23 | 地図の単体テスト用目録 | 到着マスへ行けない | maps.toml と位置・理由: exits[0] に到着のマス (3,3) へ行けないと報告 | 異常 |
| TC-FR-FLD-04-22 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 出口が二マス | maps.toml と位置・理由: rows[4][3] に出入り口 1 が 2 マスと報告 | 異常 |
| TC-FR-FLD-04-23 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 自分への出口 | maps.toml と位置・理由: exits[0].to に自分への出口を報告 | 異常 |
| TC-FR-FLD-04-24 | FR-FLD-04、DES-VAL-17〜22 | 地図の単体テスト用目録 | 工房に採取地出口 | maps.toml と位置・理由: 工房の gather 行きを報告 | 異常 |
| TC-FR-FLD-04-25 | FR-FLD-04、DES-VAL-23 | 地図の単体テスト用目録 | 形が誤りなら到達性は検査しない | maps.toml と位置・理由: 行の長さを報告し、到達性は報告しない | 異常 |
| TC-FR-FLD-04-26 | FR-FLD-04、DES-VAL-23 | 地図の単体テスト用目録 | 開始位置が誤りなら到達性は検査しない | maps.toml と位置・理由: start の壁を報告し、到達性は報告しない | 異常 |
| TC-FR-FLD-04-27 | FR-FLD-04、DES-VAL-18 | 地図の単体テスト用目録 | rows を空にする | maps.toml の rows に行数 1〜100 の誤り | 境界 |
| TC-FR-FLD-04-28 | FR-FLD-04、DES-VAL-18 | 地図の単体テスト用目録 | 1 行を 101 文字にする | maps.toml の rows[1] に長さ 1〜100 文字の誤り | 境界 |
| TC-FR-FLD-04-29 | FR-FLD-04、DES-VAL-18 | 地図の単体テスト用目録 | rows を 101 行にする | maps.toml の rows に行数 1〜100 の誤り | 境界 |
| TC-FR-FLD-04-30 | FR-FLD-04、DES-VAL-18 | 地図の単体テスト用目録 | 設計書 3 章の 30 文字と 29 文字の行長例 | maps.toml の map[id=town].rows[4] に「29 文字。rows[0] は 30 文字」と報告 | 異常 |
| TC-FR-FLD-04-31 | FR-FLD-04、DES-VAL-21 | 正常な二枚 | 街の Q を床にする | maps.toml に `map[id=town]: 依頼板（Q）がない` | 異常 |
| TC-FR-FLD-04-32 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | 工房の town 行きを消す | maps.toml に `map[id=workshop].exits: town 行きがない` | 異常 |
| TC-FR-FLD-04-33 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | 街の workshop 行きを消す | maps.toml に `map[id=town].exits: workshop 行きがない` | 異常 |
| TC-FR-FLD-04-34 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | 街の gather 行きを消す | maps.toml に `map[id=town].exits: gather 行きがない` | 異常 |
| TC-FR-FLD-04-35 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | 街の出口 mark を重複させる | maps.toml の exits[1].mark に重複を報告 | 異常 |
| TC-FR-FLD-04-36 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | mark を `12` にする | maps.toml の exits[1].mark に 1 文字の制約を報告 | 異常 |
| TC-FR-FLD-04-37 | FR-FLD-04、DES-VAL-22 | 正常な二枚 | 工房の出口の到着先を設備または別の出口にする | maps.toml の exits[0].dir に (3,3) は床でないと報告 | 異常 |
| TC-FR-FLD-04-38 | FR-FLD-04、DES-VAL-18〜23 | 誤りを重ねた二枚 | 設備・出口の誤りと到達不能、または行長・未知文字と開始位置の誤りを重ねる | DES-VAL-21・22 の誤りでは到達性を、DES-VAL-18・19 の誤りでは後続を報告しない | 異常 |
| TC-FR-FLD-04-39 | FR-FLD-04、DES-VAL-23 | 内側の出口だけで分かれた工房 | 出口の向きを右にする | 出口を塗りつぶさず、到着のマス (3,4) へ行けないと報告 | 異常 |

### IF-08 主人公を動かし、設備を使う

- #86（core の設備の前の判定、移動の誤り、Tick との分離）: IF-08 の連番 01〜29。Tick との分離（DES-TIME-02、NFR-06）もここで確かめる
- #89（キー操作と窓、手動）: IF-08 の連番 40〜69

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|

### IF-06 場面の行き来（反復 5 の追加分）

- #89（タブをなくし、出入り口と「街へ戻る」で移る、手動）: IF-06 の連番 30〜49

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|

### CST-04 画像素材の利用許諾

- #88（素材の取り込み、assets パッケージ）: CST-04 の連番 01〜09

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-CST-04-01 | CST-04、DES-UI-14 | Tiny Town 1.1 の `tilemap_packed.png` を配置 | `assets.TinyTown()` で読み込む | 誤りなく読め、画像の大きさは 192×176 px | 正常 |
| TC-CST-04-02 | CST-04、DES-UI-14 | Tiny Dungeon 1.0 の `tilemap_packed.png` を配置 | `assets.TinyDungeon()` で読み込む | 誤りなく読め、画像の大きさは 192×176 px | 正常 |
| TC-CST-04-03 | CST-04、ADR-0003 | Tiny Town 1.1 の同梱 `License.txt` を配置 | 埋め込まれた利用許諾を読む | `Tiny Town (1.1)`、`Creative Commons Zero, CC0`、CC0 1.0 の URL を含む | 正常 |
| TC-CST-04-04 | CST-04、ADR-0003 | Tiny Dungeon 1.0 の同梱 `License.txt` を配置 | 埋め込まれた利用許諾を読む | `Tiny Dungeon (1.0)`、`Creative Commons Zero, CC0`、CC0 1.0 の URL を含む | 正常 |
| TC-CST-04-05 | CST-04、ADR-0003 | `internal/ui/assets/README.md` がある | 出典表を目視で照合する | 両素材について、版・作者・配布元 URL・CC0 1.0・利用許諾ファイル名・確認日 2026-10-15・取り込んだファイルと元のファイル名が記録されている | 手動 |
| TC-CST-04-06 | CST-04、ADR-0003 | 4 つのファイルを配置 | 埋め込んだバイト列の SHA-256 を求める | 配布 zip（Tiny Town 1.1、Tiny Dungeon 1.0）の `Tilemap/tilemap_packed.png` と `License.txt` から求めた値と一致する（tiny_town.png 3a54d99e…、tiny_dungeon.png d24e60a4…、kenney_tiny_town_License.txt c483c167…、kenney_tiny_dungeon_License.txt 9f574a2f…）。改行の変換や取り違えがない | 正常 |
| TC-CST-04-07 | DES-UI-14 | 2 つの画像を配置 | `TinyTown()` と `TinyDungeon()` で読む | それぞれ自分の PNG を読んだ画像と画素が一致し、2 つの画像は互いに異なる | 異常 |

TC-CST-04-05 の結果（2026-10-16）: 出典表の各列を ADR-0003 の決定 2・3 と同梱の License.txt に照合し、両素材とも期待結果を満たした。

### NFR-08 定義データで異常終了しない（反復 5 の追加分）

- #85（地図の読み込みのファジング）: NFR-08 の連番 01〜09

| ID | 対象 | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-NFR-08-01 | NFR-08、DES-VAL-24 | 地図以外の定義がそろった MapFS | 地図 TOML のシードと任意バイト列を FuzzLoadMaps に渡す | 異常終了せず、成功時の地図は基本不変条件を満たす | 異常 |
| TC-NFR-08-02 | NFR-08、DES-VAL-21・24 | 読み込み成功時の地図 | 必須設備を確認 | 工房に K/S/G、街に M/Q/F がそれぞれある | 正常 |
| TC-NFR-08-03 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | mark と行の数字を確認 | mark は重複せず、それぞれ 1 マスにあり、未定義の数字がない | 正常 |
| TC-NFR-08-04 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | 出口の To を確認 | 自分以外の地図または gather である | 正常 |
| TC-NFR-08-05 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | 出口の Dir を確認 | 上下左右の 4 値である | 正常 |
| TC-NFR-08-06 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | 出口の Look を確認 | stone_door・wood_door・road のいずれかである | 正常 |
| TC-NFR-08-07 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | 出口の To の重複を確認 | 同じ行き先は 1 つまでである | 正常 |
| TC-NFR-08-08 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | gather 行きの地図を確認 | town にだけ置かれている | 正常 |
| TC-NFR-08-09 | NFR-08、DES-VAL-22・24 | 読み込み成功時の地図 | 必須の出口を確認 | 工房に town 行き、街に workshop・gather 行きがある | 正常 |

## 3. 手動確認の結果

（#89 で記録する）
