# テスト仕様書（反復 6）

- 文書 ID: TST-IT06
- 対象: [反復 6 計画書](../iterations/it06/plan.md) の範囲
- 根拠: 設計書 DES-IT06

## 連番の割当

連番は要求 ID ごと。過去の反復で使った番号は再利用しない。

| Issue | 担当と予約範囲 |
|---|---|
| #102 | テストケースなし（反復 5 の exe を残すだけ） |
| #103 | defs・defload: FR-FLD-01・04 は 50〜79、FR-TWN-02・07、FR-FLD-05、FR-ECO-07、FR-PRG-05 は 30〜49、FR-MOD-04 は 80〜89、NFR-08 は 10〜29 |
| #104 | core の日・依頼: FR-PRG-05 は 01〜07、FR-ECO-05 は 14〜19、FR-ECO-06 は 01〜11、FR-ECO-07 は 01〜06（実装済み） |
| #105 | core の地図・採取・敵: FR-FLD-01〜04 は 80〜89、FR-FLD-05 は 01〜29、FR-MAT-01・FR-BTL-10 は 50〜69、IF-08 は 70〜89 |
| #106 | 本体データ: FR-MOD-02 は 60〜79、FR-FLD-01 は 44〜49、FR-FLD-04 は 43〜49、FR-TWN-02・07 は 50〜59 |
| #107 | 画面: IF-06 は 40〜69、IF-08 は 61〜69、FR-TWN-07・FR-FLD-05・FR-PRG-05 は 60〜79 |
| #108 | 画像: CST-04 は 10〜29 |
| #109 | 結合・手動: 関連要求の未使用の 90〜99（FR-FLD-01〜04 も 90〜99。過去に使用済みの ID は再利用しない） |

上表の #104 の実使用番号と #103 の 30 番以降は重複しない。#105〜#109 の予約範囲も test-spec-it01〜05 の使用済み番号と照合済み（#109 は要求ごとに未使用の 90〜99 のみ）。#104 の節と行は同担当だけが更新する。

## #103 定義と検証

前提は `internal/defload/testdata/it06/maps.toml`（街、3×3 の区画、採取地、施設の見た目、日・再出現時間）。各行で指定した値だけを変更して Load する。異常系ではファイル名・キー経路（マスなら行・列）・理由を持つ診断を返す。

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-FLD-04-50 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | town.start=(5,5)、区画 a は 3×3、meadow は拾う場所 3 個・敵なし | maps.toml を変更せず Load | DaySeconds=300、meadow.RespawnSeconds=2、market の見た目 stall/house、区画 a の左上 (2,1) を保持 | 正常・境界 |
| TC-FR-FLD-04-51 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 正常定義の対象値: `id = "meadow" / rows` | `id = "meadow" / rows` を `id = "unknown" / rows` に置換して Load | `map[id=unknown].id`: 地図の ID は town または採取地の ID（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-52 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 正常定義の対象値: `id = "meadow" / rows` | `id = "meadow" / rows` を `id = "town" / rows` に置換して Load | `map[id=town].id`: ID が重複（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-53 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 正常定義の対象値: `[[map]] / id = "meadow"` | `[[map]] / id = "meadow"` を `[[map]] / id = "missing"` に置換して Load | `map[id=meadow]`: 地図がない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-54 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 採取地 meadow と同 ID の地図がある | 採取地と地図の ID meadow を両方 gather に変更 | `gather[id=gather].id`: 採取地の ID に town・gather は使えない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-55 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 採取地 meadow と地図 town/meadow がある | 採取地と地図の ID meadow を両方 town に変更 | `gather[id=town].id`: 採取地の ID に town・gather は使えない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-56 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | game.day_seconds=300、town と meadow の地図がある | [game] と day_seconds の行を削除 | `game.day_seconds`: 書かれていない（地図を使うときは必須）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-04-57 | FR-FLD-04 / DES-DEF-09〜11、DES-VAL-17・25・29 | 採取地 meadow は lots を省略 | meadow の map 節に lots=[] を追加 | `map[id=meadow].lots`: lots は街にだけ書く（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-01-50 | FR-FLD-01 / DES-VAL-20・21 | 正常定義の対象値: `start = {x=5,y=5}` | `start = {x=5,y=5}` を `` に置換して Load | `map[id=town].start`: 開始位置が書かれていない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-01-51 | FR-FLD-01 / DES-VAL-20・21 | 正常定義の対象値: `id = "meadow" / rows` | `id = "meadow" / rows` を `id = "meadow" / start={x=1,y=1} / rows` に置換して Load | `map[id=meadow].start`: 採取地の地図には開始位置を書かない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-01-52 | FR-FLD-01 / DES-VAL-20・21 | 正常定義の対象値: `T,aaa,,DT` | `T,aaa,,DT` を `T,aaa,,,T` に置換して Load | `map[id=town]`: 工房の扉（D）がない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-01-53 | FR-FLD-01 / DES-VAL-20・21 | 正常定義の対象値: `T*,,,,T` | `T*,,,,T` を `Ta,,,,T` に置換して Load | `map[id=meadow].rows[1][1]`: 区画（a）は街にしか置けない（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-01-54 | FR-FLD-01 / DES-VAL-20・21 | 正常定義の対象値: `T,,,,,,MT` | `T,,,,,,MT` を `T*,,,,,MT` に置換して Load | `map[id=town].rows[4][1]`: 拾う場所（*）は採取地にしか置けない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-02-30 | FR-TWN-02 / DES-DEF-12、DES-VAL-26 | market は stall/house の 2 段階、town の区画 a に対応 | maps.toml を変更せず Load | market の見た目 stall/house と区画の左上 (2,1) を保持 | 正常・境界 |
| TC-FR-TWN-02-31 | FR-TWN-02 / DES-DEF-12、DES-VAL-26 | market の第 1 段階は stall、第 2 段階は house | 第 1 段階を house に変更 | market の両段階の Look=house を保持して成功 | 正常・境界 |
| TC-FR-TWN-02-32 | FR-TWN-02 / DES-DEF-12、DES-VAL-26 | market の第 2 段階は house | 第 2 段階の look を castle に変更 | `facility[id=market].levels[1].look`: 知らない見た目 "castle"（stall・cottage・house・manor が必要）（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-30 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `T,aaa,,,T` | 対象値を `T,aa,,,,T` に置換して Load（空欄は省略） | `map[id=town].lots[0]`: 区画 "a" が 3 × 3 の四角になっていない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-31 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `mark="a",facility="market"` | 対象値を `mark="b",facility="market"` に置換して Load（空欄は省略） | `map[id=town].rows[1][2]`: 区画の定義がない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-32 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `facility="market"` | 対象値を `facility="missing"` に置換して Load（空欄は省略） | `map[id=town].lots[0].facility`: 存在しない施設: missing（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-33 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `mark="a",facility="market"` | 対象値を `mark="j",facility="market"` に置換して Load（空欄は省略） | `map[id=town].lots[0].mark`: mark は a〜i の 1 文字（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-34 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `lots = [{mark="a",facility="market"}]` | 対象値を `lots = [{mark="a",facility="market"},{mark="a",facility="market"}]` に置換して Load（空欄は省略） | `map[id=town].lots[1].mark`: mark が重複（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-35 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `look="stall"` | 対象値を `look="castle"` に置換して Load（空欄は省略） | `facility[id=market].levels[0].look`: 知らない見た目 "castle"（stall・cottage・house・manor が必要）（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-36 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 正常定義の対象値: `look="stall"` | 対象値を `money=1` に置換して Load（空欄は省略） | `facility[id=market].levels[0].look`: 書かれていない（地図を使うときは必須）（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-37 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | town.lots=[{mark="a",facility="market"}]、market が目録にある | town.lots=[] にする（rows の a は残す） | `map[id=town].lots`: 施設 market の区画がない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-38 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | market の区画 a が 1 つある | 街の行 4 を T,bbb,,MT にし、区画 b も facility=market として追加 | `map[id=town].lots[1].facility`: 同じ施設の区画が重複（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-39 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 街の区画 a は 9 マスで 3×3 | 行 1 を T,aaa,aQT、行 2 を T,aa,,,,T にして 9 マスを離す | `map[id=town].lots[0]`: 区画 "a" が 3 × 3 の四角になっていない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-40 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 街の区画 a は 9 マス | rows の a をすべて # にし、lots の定義は残す | `map[id=town].lots[0]`: 区画 "a" が 3 × 3 の四角になっていない（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-41 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | town の区画 a の周りは床 | 行 1 を T#aaa#,QT、行 2 を T#aaa#,,T、行 3 を T#aaa#,DT、行 4 を T,###,,MT にして区画を壁で囲む | `map[id=town].lots[0]`: 開始位置から行けない: 区画（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-42 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | map 節なし、market.levels[0].look=stall | 第 1 段階の look を manor に変更して Load | 読み込み成功。Maps は空で、指定した manor を保持 | 正常・境界 |
| TC-FR-TWN-07-43 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | map 節なし、market.levels[0].look=stall | 第 1 段階の look を castle に変更して Load | `facility[id=market].levels[0].look`: 知らない見た目 "castle"（stall・cottage・house・manor が必要）（ファイル: maps.toml） | 異常 |
| TC-FR-TWN-07-44 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | map 節なし、market は stall/house の 2 段階 | 両段階の look を省略（money=1/2 に置換）して Load | 読み込み成功。Maps は空で、各段階の Look は空文字 | 正常・境界 |
| TC-FR-TWN-07-45 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 地図あり、market.levels[0].look=stall | 第 1 段階の look を cottage に変更して Load | market.levels[0].Look=cottage を保持して成功 | 正常・境界 |
| TC-FR-TWN-07-46 | FR-TWN-07 / DES-DEF-11・12、DES-VAL-23・25・26 | 地図あり、market.levels[0].look=stall | 第 1 段階の look に空文字を明記して Load | `facility[id=market].levels[0].look`: 知らない見た目 ""（stall・cottage・house・manor が必要）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-30 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `T*,,,,T` | 対象値を `T,,,,,T` に置換して Load（空欄は省略） | meadow の拾う場所が 2 マス、RespawnSeconds=2 を保持して成功 | 正常・境界 |
| TC-FR-FLD-05-31 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `T*,,,,T` | 対象値を `TE,,,,T` に置換して Load（空欄は省略） | `map[id=meadow].rows`: 敵（E）が 1 マスある（必要なマス数は 0）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-32 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `respawn_seconds = 2` | 対象値を `respawn_seconds = 0` に置換して Load（空欄は省略） | `gather[id=meadow].respawn_seconds`: 1 以上 86400 以下にする（0）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-33 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `respawn_seconds = 2` | 対象値を `respawn_seconds = 86400` に置換して Load（空欄は省略） | meadow.RespawnSeconds=86400 を保持して成功 | 正常・境界 |
| TC-FR-FLD-05-34 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `respawn_seconds = 2` | 対象値を `respawn_seconds = 86401` に置換して Load（空欄は省略） | `gather[id=meadow].respawn_seconds`: 1 以上 86400 以下にする（86401）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-35 | FR-FLD-05 / DES-VAL-23・27・28 | 正常定義の対象値: `respawn_seconds = 2` | 対象値を 省略して Load | `gather[id=meadow].respawn_seconds`: 書かれていない（地図を使うときは必須）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-36 | FR-FLD-05 / DES-VAL-23・27・28 | meadow には * が 2 マス、o が 1 マス | * と o のすべてを床に変更 | `map[id=meadow].rows`: 拾う場所は 1 マス以上必要（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-37 | FR-FLD-05 / DES-VAL-23・27・28 | 地図あり、meadow.respawn_seconds=2 | respawn_seconds=1 に変更 | meadow.RespawnSeconds=1 を保持して成功 | 正常・境界 |
| TC-FR-FLD-05-38 | FR-FLD-05 / DES-VAL-23・27・28 | map 節なし、meadow.respawn_seconds=2 | respawn_seconds=-1 に変更 | `gather[id=meadow].respawn_seconds`: 1 以上 86400 以下にする（-1）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-39 | FR-FLD-05 / DES-VAL-23・27・28 | map 節なし、meadow.respawn_seconds=2 | respawn_seconds の行を省略 | 読み込み成功。Maps は空で、RespawnSeconds は 0 | 正常・境界 |
| TC-FR-FLD-05-40 | FR-FLD-05 / DES-VAL-23・27・28 | battle.toml と同じ hero/foe の戦闘定義、meadow に敵 E が 1 マス | 戦闘ありの定義を変更せず Load | meadow に戦闘定義と敵 E がちょうど 1 マスある目録を返す | 正常・境界 |
| TC-FR-FLD-05-41 | FR-FLD-05 / DES-VAL-23・27・28 | hero/foe の戦闘定義あり、meadow の行 1 は T*,,,ET | 行 1 を T*,,,,T にし、敵を 0 マスにする | `map[id=meadow].rows`: 敵（E）が 0 マスある（必要なマス数は 1）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-42 | FR-FLD-05 / DES-VAL-23・27・28 | hero/foe の戦闘定義あり、meadow の行 1 は T*,,,ET | 行 1 を T*,,EET にし、敵を 2 マスにする | `map[id=meadow].rows`: 敵（E）が 2 マスある（必要なマス数は 1）（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-43 | FR-FLD-05 / DES-VAL-23・27・28 | 敵なしの meadow。拾う場所 (1,1) の前は床 | 行 1・3 を T*#,,,T、行 2 を T#,#,,T にし、拾う場所を囲む | `map[id=meadow].rows[1][1]`: 開始位置から行けない: 拾う場所（ファイル: maps.toml） | 異常 |
| TC-FR-FLD-05-44 | FR-FLD-05 / DES-VAL-23・27・28 | meadow は戦闘なし・敵なし、拾う場所は 3 個 | 行 1・3 を T*#,,ET、行 2 を T#,#,,T にし、敵 2 マスと拾う場所の到達不能を併発 | `map[id=meadow].rows`: 敵（E）が 2 マスある（必要なマス数は 0）。`map[id=meadow].rows[1][1]`: 開始位置から行けない: 拾う場所。敵数不正でも到達性を検査する | 異常 |
| TC-FR-PRG-05-30 | FR-PRG-05 / DES-VAL-29 | 正常定義の対象値: `day_seconds = 300` | 対象値を `day_seconds = 1` に置換して Load（空欄は省略） | Game.DaySeconds=1 を保持して成功 | 正常・境界 |
| TC-FR-PRG-05-31 | FR-PRG-05 / DES-VAL-29 | 正常定義の対象値: `day_seconds = 300` | 対象値を `day_seconds = 86400` に置換して Load（空欄は省略） | Game.DaySeconds=86400 を保持して成功 | 正常・境界 |
| TC-FR-PRG-05-32 | FR-PRG-05 / DES-VAL-29 | 正常定義の対象値: `day_seconds = 300` | 対象値を `day_seconds = 0` に置換して Load（空欄は省略） | `game.day_seconds`: 1 以上 86400 以下にする（0）（ファイル: maps.toml） | 異常 |
| TC-FR-PRG-05-33 | FR-PRG-05 / DES-VAL-29 | 正常定義の対象値: `day_seconds = 300` | 対象値を `day_seconds = 86401` に置換して Load（空欄は省略） | `game.day_seconds`: 1 以上 86400 以下にする（86401）（ファイル: maps.toml） | 異常 |
| TC-FR-PRG-05-34 | FR-PRG-05 / DES-VAL-29 | 正常定義の対象値: `day_seconds = 300` | 対象値を 省略して Load | `game.day_seconds`: 書かれていない（地図を使うときは必須）（ファイル: maps.toml） | 異常 |
| TC-FR-PRG-05-35 | FR-PRG-05 / DES-VAL-29 | map 節なし、day_seconds=300 | day_seconds の行を省略 | 読み込み成功。Maps は空で、DaySeconds は 0 | 正常・境界 |
| TC-FR-PRG-05-36 | FR-PRG-05 / DES-VAL-29 | map 節なし、day_seconds=300 | day_seconds=-1 に変更 | `game.day_seconds`: 1 以上 86400 以下にする（-1）（ファイル: maps.toml） | 異常 |
| TC-FR-PRG-05-37 | FR-PRG-05 / DES-VAL-29 | map 節なし、day_seconds=300 | day_seconds=0 を明記 | `game.day_seconds`: 1 以上 86400 以下にする（0）（ファイル: maps.toml） | 異常 |
| TC-FR-ECO-07-30 | FR-ECO-07 / DES-VAL-15 | it04_test.go の rankData から依頼を置換。ランク 2・3 には評判 1 の依頼を各 1 件用意 | ランク 1 の依頼が 0 件を Load する | `quest`: ランク 1 の依頼が 0 件しかない（6 件以上が必要）（maps.toml） | 異常 |
| TC-FR-ECO-07-31 | FR-ECO-07 / DES-VAL-15 | it04_test.go の rankData から依頼を置換。ランク 2・3 には評判 1 の依頼を各 1 件用意 | ランク 1 の依頼が 5 件を Load する | `quest`: ランク 1 の依頼が 5 件しかない（6 件以上が必要）（maps.toml） | 異常 |
| TC-FR-ECO-07-32 | FR-ECO-07 / DES-VAL-15 | it04_test.go の rankData から依頼を置換。ランク 2・3 には評判 1 の依頼を各 1 件用意 | ランク 1 の依頼が 6 件を Load する | 読み込み成功 | 正常・境界 |
| TC-FR-ECO-07-33 | FR-ECO-07 / DES-VAL-15 | it04_test.go の rankData から依頼を置換。ランク 2・3 には評判 1 の依頼を各 1 件用意 | ランク 1 の依頼が 7 件を Load する | 読み込み成功 | 正常・境界 |
| TC-FR-MOD-04-80 | FR-MOD-04 / DES-VAL-26〜29 | 施設・採取地・game は definitions.toml、地図は maps.toml | definitions.toml の未知の lookを Load する | 診断が 1 件。File=definitions.toml、Path=`facility[id=market].levels[0].look`、Reason=知らない見た目 "castle"（stall・cottage・house・manor が必要） | 異常 |
| TC-FR-MOD-04-81 | FR-MOD-04 / DES-VAL-26〜29 | 施設・採取地・game は definitions.toml、地図は maps.toml | definitions.toml の再出現時間 0を Load する | 診断が 1 件。File=definitions.toml、Path=`gather[id=meadow].respawn_seconds`、Reason=1 以上 86400 以下にする（0） | 異常 |
| TC-FR-MOD-04-82 | FR-MOD-04 / DES-VAL-26〜29 | 施設・採取地・game は definitions.toml、地図は maps.toml | definitions.toml の日の長さ 0を Load する | 診断が 1 件。File=definitions.toml、Path=`game.day_seconds`、Reason=1 以上 86400 以下にする（0） | 異常 |
| TC-NFR-08-10 | NFR-08 / DES-VAL-24 | validMaps、testdata/it06/maps.toml・battle.toml、および各規則を壊した種 | 区画・拾う場所・敵と各規則の異常種、任意バイト列を FuzzLoadMaps に渡す | 異常終了しない。成功時は採取地との地図の対応、文字、区画、敵数、秒数、出口と到達性を満たす | 正常・境界 |
| TC-FR-FLD-04-58 | FR-FLD-04 / DES-VAL-17・22 | 本体の街と採取地 4 枚 | 全採取地の地図と出口を調べる | 工房の地図なし、全採取地に town 行き出口が 1 つ | 正常 |
| TC-FR-FLD-04-59 | FR-FLD-04 / DES-VAL-21・22・25 | 本体の街と施設 2 件 | 街の設備・区画・出口を調べる | D/M/Q があり、全施設に区画、gather 行き出口が 1 つ | 正常 |
| TC-FR-FLD-04-60 | FR-FLD-04 / DES-VAL-17 | 正常定義の town と meadow | town の map 節だけを除去 | `map[id=town]`: 地図がない | 異常 |
| TC-FR-FLD-04-61 | FR-FLD-04 / DES-VAL-22 | 正常定義の街。gather 行き出口 2 がある | 街に mark=3、to=meadow、dir=up、look=road の出口と対応するマスを追加 | `map[id=town].exits[1].to`: 行き先がない: meadow（必要な行き先は gather） | 異常 |
| TC-FR-FLD-01-58 | FR-FLD-01 / DES-VAL-21 | 街の (1,4) は床 | 街の (1,4) を E に変更 | `map[id=town].rows[4][1]`: 敵（E）は採取地にしか置けない | 異常 |
| TC-FR-FLD-01-59 | FR-FLD-01 / DES-VAL-21 | 街の (1,4) は床 | 街の (1,4) を o に変更 | `map[id=town].rows[4][1]`: 拾う場所（o）は採取地にしか置けない | 異常 |
| TC-FR-FLD-01-60 | FR-FLD-01 / DES-VAL-21 | 街の店 M は (7,4) の 1 個 | M を床に変更 | `map[id=town]`: 店（M）がない | 異常 |
| TC-FR-FLD-01-61 | FR-FLD-01 / DES-VAL-21・23 | 敵なしの meadow。拾う場所は 3 個 | 行 1 を T*DMQ,T、行 2 を T,,,,,T にする | 成功。meadow.Rows[1]=T*DMQ,T を保持（全設備の前の床に到達できる） | 正常 |
| TC-FR-FLD-04-62 | FR-FLD-04 / DES-VAL-25・23 | 街の区画 a は 3×3、依頼板 Q は (7,1) | 行 1 を T,aaa,#QT、行 2 を T,aa,,,#T にし、区画を欠けさせて Q も囲む | `map[id=town].lots[0]`: 区画 "a" が 3 × 3 の四角になっていない。開始位置から行けない／到着のマスの診断は出ない | 異常 |
| TC-FR-FLD-04-63 | FR-FLD-04 / DES-VAL-23 | battle.toml。meadow に敵 E が 1 マス | 行 1 を T*,,#ET、行 2 を T,,o,#T にして E を囲む | `map[id=meadow].rows[1][5]`: 開始位置から行けない: 敵 | 異常 |
| TC-FR-FLD-04-64 | FR-FLD-04 / DES-VAL-23 | 敵なしの meadow | 行 1・3 を T*,,#DT、行 2 を T,,o,#T にする | `map[id=meadow].rows[1][5]`: 開始位置から行けない: 工房の扉 | 異常 |
| TC-FR-FLD-01-62 | FR-FLD-01 / DES-VAL-20 | 街の start=(5,5) | start.x=-1 にする | `map[id=town].start`: 開始位置 (-1, 5) が地図の外。到達性診断は出ない | 境界・異常 |

## 既存ケースの置き換え（反復 6 で変更）

工房の地図・設備 K/S/G/F・戻り道をなくし、街と採取地、設備 D/M/Q、区画に置き換える。

| 既存のケース | 反復 6 で変更する内容 |
|---|---|
| TC-FR-FLD-01-03〜04 | 固定文字表と設備名を D/M/Q、区画、拾う場所、敵、水に変更 |
| TC-FR-FLD-01-05〜09 | 街の start の座標省略・地図外と、診断順序（20 回の DeepEqual と先頭キー経路）を検証 |
| TC-FR-FLD-04-01〜39 | 正常定義を街・採取地に変更。必須設備、start、行き先と到達性は DES-IT06 6 章を適用 |
| TC-FR-FLD-01-30〜31、TC-FR-FLD-04-40〜42 | 工房地図を廃止。04-41・42 の距離試験は廃止し、最小移行データの地図集合・設備・出入り口は 04-58・59 で確認（街の距離と完成データは #106） |
| TC-FR-MOD-04-54 | ランク 1 の依頼の最小件数を 3 から 6 に変更（5 件は誤り） |
| TC-NFR-08-01〜09 | 新しい文字表、街から一覧・採取地から街への出口、採取地の到着位置からの到達性に変更 |

core の移動・設備テストは、同じマス位置の設備文字だけを D/M/Q に置換し、当たり判定と選択の期待値を維持する。

## 移行中の扱い

- `GameDef.DaySeconds` は #103 で追加した（読み込み・DES-VAL-29）。日の処理は #104 で実装済み。
- #105 のあとは街から始まり、歩ける。工房の扉と区画では中身のない窓が開き、採取地の地図では拾えず戦えない。採取地の一覧には「探索」が残る。これらの画面の移行は #107・#108 で行う。REV-056 F-01 で開発者が承認したのは #103 のあとの状態（起動直後に操作できない）で、#105 で解消した。#105 のあとの上の状態は、Issue の分け方で想定した途中の状態で、#107・#108 で解消する。
- `defs.MapWorkshop` は設計どおり削除した。#105 で core の起動処理は街に移行済み。ui に残る工房の場面は #107・#108 で移行する。
- 本体は街と採取地 4 枚、既存施設 2 区画、設計 5.2 の q_flour・q_potion_easy と q_rope 代用の仮依頼 1 件を持つ最小の移行データ。施設 4 件・依頼 17 件・景色の完成は #106。
- core の設備テスト（TC-IF-08-01〜03・06・17〜18・20〜21）は、設備位置と選択規則を維持して D/M/Q の期待値へ変更した。
- 本体件数の TC-FR-MOD-02-04・42 は、依頼 12 件、地図 5 枚という最小移行データに更新した。

## FR-PRG-05

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-PRG-05-01 | FR-PRG-05、DES-DAY-01・02 | D-1、DaySeconds=1、Now=0 | Day() | 1 | 正常 |
| TC-FR-PRG-05-02 | FR-PRG-05、DES-DAY-01・02 | D-2、DaySeconds=1、Now=0 | 59 Tick | Now=59、1日目、知らせなし | 境界 |
| TC-FR-PRG-05-03 | FR-PRG-05、DES-DAY-01・02 | D-3、DaySeconds=1、Now=59 | 1 Tick、知らせを2回読む | Now=60、2日目、[2]、次は空 | 境界 |
| TC-FR-PRG-05-04 | FR-PRG-05、DES-DAY-01・02 | D-4、DaySeconds=1 | 120 Tick | 3日目、知らせ[2,3] | 正常 |
| TC-FR-PRG-05-05 | FR-PRG-05、DES-DAY-01・02 | D-5、DaySeconds=0 | 1000 Tick | 1日目、知らせなし、候補不変 | 境界 |
| TC-FR-PRG-05-06 | FR-PRG-05、DES-DAY-01・02 | DaySeconds=300、Now=17999 | 1 Tick | Now=18000、2日目、知らせ[2] | 境界 |
| TC-FR-PRG-05-07 | FR-PRG-05、DES-DAY-01・02、DES-TIME-01 | 生産中のゴーレム、DaySeconds=1 | 120 Tickを1回ずつ／10回ずつ進める | 時刻・日・候補・受注・在庫・ゴーレム・知らせが一致、3日目、知らせ[2,3] | 正常 |

## FR-ECO-06

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-ECO-06-01 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-1、q1〜q6の目録 | NewWorld | 候補3件が非空・重複なし、受注は空 | 正常 |
| TC-FR-ECO-06-02 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-2、候補[A,B,C] | AcceptQuest(1) | 受注[B]、候補[A,空,C] | 正常 |
| TC-FR-ECO-06-03 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-3、Q-2の受注後 | AcceptQuest(1) | ErrNoQuest、状態不変 | 異常 |
| TC-FR-ECO-06-04 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-4、受注[B,A,C]、候補すべて空 | AcceptQuest(0) | ErrNoQuestを上限より先に返す、状態不変 | 境界 |
| TC-FR-ECO-06-05 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-5、受注[B,A,C]、候補[D,空,空] | AcceptQuest(0) | ErrQuestLimit、状態不変 | 境界 |
| TC-FR-ECO-06-06 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-6、候補[A,B,C] | AcceptQuest(3) | ErrNoQuest、状態不変 | 異常 |
| TC-FR-ECO-06-07 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-10、受注[B,A]、お金100 | AbandonQuest(0) | 受注[A]、お金100、評判・在庫・候補不変 | 正常 |
| TC-FR-ECO-06-08 | FR-ECO-06、DES-QST-04・05、DES-WLD-01 | Q-11、受注なし | AbandonQuest(0) | ErrNoQuest、状態不変 | 異常 |
| TC-FR-ECO-06-09 | FR-ECO-06、DES-QST-05、DES-WLD-01 | 候補あり、既存の知らせあり | AcceptQuest(-1/3) | ErrNoQuest、状態・乱数不変 | 異常 |
| TC-FR-ECO-06-10 | FR-ECO-06、DES-QST-05、DES-WLD-01 | 受注1件、既存の知らせあり | AbandonQuest(-1/1/3) | ErrNoQuest、状態・乱数不変 | 異常 |
| TC-FR-ECO-06-11 | FR-ECO-06、DES-QST-05 | 受注[q1,q2,q3] | 中央／末尾をやめる | 受注[q1,q3]／[q1,q2]、候補・資産不変 | 境界 |

## FR-ECO-05

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-ECO-05-14 | FR-ECO-05、DES-QST-04・05、DES-WLD-01 | Q-7、受注[B,A]、Bの必要在庫あり | DeliverQuest(0) | 受注[A]、在庫を消費、報酬・評判加算、候補不変 | 正常 |
| TC-FR-ECO-05-15 | FR-ECO-05、DES-QST-04・05、DES-WLD-01 | Q-8、受注[B,A] | DeliverQuest(2) | ErrNoQuest、状態不変 | 異常 |
| TC-FR-ECO-05-16 | FR-ECO-05、DES-QST-04・05、DES-WLD-01 | Q-9、受注[B,A]、Aの在庫不足 | DeliverQuest(1) | ErrNotEnoughItems、状態不変 | 異常 |
| TC-FR-ECO-05-17 | FR-ECO-05、DES-QST-05、DES-WLD-01 | 受注1件、在庫・既存の知らせあり | DeliverQuest(-1/1/3) | ErrNoQuest、状態と乱数不変 | 異常 |
| TC-FR-ECO-05-18 | FR-ECO-05、DES-QST-05、DES-WLD-01 | 候補あり、受注なし、在庫あり | DeliverQuest(0) | ErrNoQuest、候補への直接納品は不可、状態・乱数不変 | 異常 |
| TC-FR-ECO-05-19 | FR-ECO-05、DES-QST-05 | 受注[q1,q2,q3]、在庫あり | 中央／末尾を納品 | 受注[q1,q3]／[q1,q2]、候補不変 | 境界 |

## FR-ECO-07

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-ECO-07-01 | FR-ECO-07、DES-QST-04・05、DES-WLD-01 | Q-12、q1〜q6、受注[A,B,C]、Now=59 | Tick | 候補はD,E,F各1件、受注不変 | 正常 |
| TC-FR-ECO-07-02 | FR-ECO-07、DES-QST-04・05、DES-WLD-01 | Q-13、q1〜q4、受注[A,B,C]、Now=59 | Tick | 候補[D,空,空]、受注不変 | 境界 |
| TC-FR-ECO-07-03 | FR-ECO-07、DES-QST-04・05、DES-WLD-01 | Q-14、q1〜q6、候補[A,空,空]、Now=59 | Tick | 初期抽選3回後の `rand.NewPCG(42, 1)` で、ID昇順のq1〜q6から重複なしに3回引いた並びと一致（全位置を引き直し、Aの再登場可） | 正常 |
| TC-FR-ECO-07-04 | FR-ECO-07、DES-QST-04・05、DES-WLD-01 | Q-15、同じ目録・シードのWorld2つ | 同じ受注・中止・日替わり操作 | 候補と受注が一致（NFR-06） | 正常 |
| TC-FR-ECO-07-05 | FR-ECO-07、DES-QST-04・05 | 依頼1件、DaySeconds=1 | 受注して納品／中止、59 Tick、さらに1 Tick | 同日中は候補が空、翌日は同じ依頼が候補に戻る | 正常 |
| TC-FR-ECO-07-06 | FR-ECO-07、DES-QST-04、NFR-06 | 同じ目録・シード | 一方だけ本体乱数を100回進め、両方を120 Tick | 候補は一致、日替わりで本体乱数は進まない | 正常 |

## 既存の試験の置き換え（反復 6 で変更）

ID の重複定義を避け、既存の定義行は元の仕様書で更新する。納品を扱う既存単体テストは、先に受注して `DeliverQuest` を呼ぶ前提に変更する。品質・在庫・報酬・評判の従来の検証は維持する。

| 既存 ID | 変更内容 |
|---|---|
| TC-FR-ECO-05-03 | 納品後は受注から削除し、候補は補充しない |
| TC-FR-ECO-05-04 | 範囲外の誤りを ErrUnknownID から ErrNoQuest へ変更 |
| TC-FR-ECO-05-05〜07、TC-FR-ECO-05-09、TC-FR-ECO-05-11〜13 | 納品先を受注一覧へ変更。品質・在庫・上限・検査順序・未知IDの期待値は維持 |
| TC-FR-ECO-05-08 | 納品後も候補は空のまま、再抽選しない |
| TC-FR-ECO-05-10 | 納品後に候補と依頼乱数が変わらないことへ置き換え |
| TC-FR-ECO-05-31〜37 | 先に受注して評判・ランク・知らせ・検査順序を確認 |
| TC-NFR-06-30〜32 | 先に受注して納品する形に変更。30は目録の並び順によらず受注一覧と納品後の状態が一致、31は本体乱数の独立性、32は失敗時に依頼乱数が進まず納品後の状態が一致することを確認 |
| TC-NFR-06-40 | ランク上昇の直後は補充せず、翌日の全候補を ID 順・同じ依頼乱数列で引くことへ置き換え |

画面の既存手動テスト（TC-FR-ECO-05-20〜24）は過去の実施記録として保持する。反復 6 の画面・手動試験は後続の UI Issue で扱う。

## #105 core の向いている物・拾う場所・敵・場面

`internal/core/pick_test.go` の目録は設計 3.7 節の街（9×7、開始 (5,5)、区画 a=market）と草原（meadow、7×5）をそのまま組み立てる。本体の `data/` は使用しない。草原は Rolls=3、RespawnSeconds=2、薬草 Q40 のみの採取表と戦闘定義を持つ。位置は 1 マス=96 単位。異常系は誤りの種類・文の完全一致と、操作前後の World（在庫・戻る tick・主人公・戦闘を含む）および本体乱数の保持を確認する。

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-FLD-03-80 | FR-FLD-03 / DES-FLD-05 | F-1、街の開始マス (5,5) | NewWorld | 街、(528,512)、上向き（E-110） | 正常 |
| TC-IF-08-70 | IF-08 / DES-FLD-07 | F-2、街 (512,240)、左 | FacingTarget | TargetLot、market、マス (4,2)、true、位置不変 | 正常 |
| TC-IF-08-71 | IF-08 / DES-FLD-07 | F-3、街 (640,336)、右 | FacingTarget | TargetFixture、workshop_door、(7,3)、true、位置不変 | 正常 |
| TC-FR-FLD-03-81 | FR-FLD-03 / DES-FLD-05 | F-4、採取地の一覧 | EnterGather("meadow") | 草原 (336,336)、上向き | 正常 |
| TC-FR-FLD-03-82 | FR-FLD-03 / DES-FLD-05、DES-WLD-01 | F-5、街にいる | EnterGather("meadow") | ErrNotInGather、主人公不変 | 異常 |
| TC-FR-FLD-03-83 | FR-FLD-03 / DES-FLD-05、DES-PRG-02 | F-6、一覧、ランク1、草原はランク2 | EnterGather("meadow") | ErrLocked、主人公不変 | 異常 |
| TC-FR-FLD-02-80 | FR-FLD-02 / DES-FLD-02・03 | F-7、草原 (336,336)、上 | MovePlayer(0,-1) を3回 | 岩に止まり、草原 (336,320)、上向き | 境界 |
| TC-IF-08-72 | IF-08 / DES-FLD-07 | F-8、草原 (336,320)、上 | FacingTarget | TargetPick、IDは空、(3,2)、true、位置不変 | 正常 |
| TC-FR-MAT-01-50 | FR-MAT-01 / DES-PCK-01・02 | F-9、草原 (336,320)、上、Now=0、空の在庫 | PickFacing | 戻り値と在庫が薬草 Q40×3、(meadow,3,2) の戻る tick=120 | 正常 |
| TC-FR-MAT-01-51 | FR-MAT-01 / DES-PCK-01・02 | F-10、F-9後、Now=50 | PickFacing | ErrPickNotReady、全文「あと 2 秒で拾えます: まだ拾えません」、在庫と戻る tick=120 は不変 | 境界・異常 |
| TC-FR-MAT-01-52 | FR-MAT-01 / DES-PCK-01・02 | F-11、F-9後、Now=120 | PickFacing | 薬草 Q40×3、合計6個、戻る tick=240 | 境界 |
| TC-FR-MAT-01-53 | FR-MAT-01 / DES-PCK-01 | F-12、草原 (336,336)、上 | PickFacing | 床を向いているため ErrNothingToPick | 異常 |
| TC-IF-08-73 | IF-08 / DES-FLD-07 | F-13、草原 (528,224)、上 | FacingTarget | TargetEnemy、IDは空、(5,1)、true、位置不変 | 正常 |
| TC-FR-BTL-10-50 | FR-BTL-10 / DES-BTL-07、DES-FLD-05 | F-14、草原 (528,224)、上 | StartBattleFacing | Battle.GatherID=meadow、主人公の場面・座標・向きは不変 | 正常 |
| TC-FR-BTL-10-51 | FR-BTL-10 / DES-BTL-07 | F-15、F-14の敵全員のHPを0にして勝利 | 戦闘終了後 StartBattleFacing | 同じ位置・場面で meadow の戦闘が再開 | 正常 |
| TC-FR-MAT-01-54 | FR-MAT-01 / DES-PCK-01 | F-16、敵に向いて戦闘中 | PickFacing | 向いている物の種類より戦闘検査が先で ErrInBattle | 異常 |
| TC-FR-FLD-03-84 | FR-FLD-03 / DES-FLD-05 | F-17、草原 (336,336)、下 | MovePlayer(0,1) を7回 | 1〜6回目は (336,343)〜(336,378)、7回目は街 (432,528)、上向き | 境界 |
| TC-FR-FLD-05-01 | FR-FLD-05 / DES-PCK-02、DES-FLD-05 | F-18、(meadow,3,2) を拾い戻る tick=120 | 草原の出口から街へ戻り、一覧から再入場、上へ3回移動、Now=50で拾う | 草原 (336,320)、上、残り70 tick、ErrPickNotReady | 正常 |
| TC-FR-MAT-01-55 | FR-MAT-01 / DES-PCK-01、DES-WLD-01 | 一覧かつ戦闘中／一覧かつ採取地定義なし／空の拾う場所かつ採取地定義なし | PickFacing | 順に ErrInBattle／ErrNothingToPick／ErrPickNotReady（あと2秒）、全状態と乱数不変 | 異常 |
| TC-FR-MAT-01-56 | FR-MAT-01 / DES-PCK-01、DES-WLD-01 | 草原 (336,320)、上、meadow の採取地定義を削除 | PickFacing | ErrUnknownID、全文「採取地 meadow: ID が見つかりません」、全状態と乱数不変 | 異常 |
| TC-FR-MAT-01-57 | FR-MAT-01 / DES-PCK-01、DES-WLD-01 | 拾える岩、RespawnSeconds=0／-1／int最大値、または Now=uint64最大値-119（秒数2） | PickFacing | 4条件とも ErrInvalidCount、全文「個数が正しくありません」、全状態と乱数不変 | 境界・異常 |
| TC-FR-MAT-01-58 | FR-MAT-01 / DES-PCK-01、DES-WLD-01 | 薬草Q40をint最大値-2個持ち、3個拾える岩 | PickFacing | ErrInvalidCount、全文「追加後の個数が範囲外: 個数が正しくありません」、在庫・乱数・戻る tick を含む全状態不変 | 境界・異常 |
| TC-FR-MAT-01-59 | FR-MAT-01 / DES-PCK-01、DES-WLD-01 | 草原の採取表の QualityMax=201 | PickFacing | ErrInvalidQuality、全文「品質は 1〜200 で指定してください」、全状態と乱数不変 | 異常 |
| TC-FR-FLD-03-85 | FR-FLD-03 / DES-FLD-05、DES-WLD-01 | 街かつ戦闘中／一覧かつ戦闘中、どちらも未知IDを指定 | EnterGather("missing") | 順に ErrNotInGather／ErrInBattle、全状態と乱数不変 | 異常 |
| TC-FR-FLD-03-86 | FR-FLD-03 / DES-FLD-05、DES-PRG-02、DES-WLD-01 | 一覧、未知ID／ランク1でランク2の草原の地図なし | EnterGather("missing")／EnterGather("meadow") | 順に ErrUnknownID（IDを文に残す）／ErrLocked（草原・ランク2を文に残す）、全状態と乱数不変 | 異常 |
| TC-FR-FLD-03-87 | FR-FLD-03 / DES-FLD-05、DES-WLD-01 | 一覧、meadow の採取地定義あり、地図なし／未検証の地図で街への出口なし | EnterGather("meadow") | ErrNoGatherMap、全文「この採取地には地図がありません」、全状態と乱数不変 | 異常 |
| TC-FR-BTL-10-52 | FR-BTL-10 / DES-BTL-07、DES-WLD-01 | 一覧かつ戦闘中／草原の拾う岩を向く | StartBattleFacing | 順に ErrInBattle／ErrNoEnemy（「戦う相手がいません」）、全状態と乱数不変 | 異常 |
| TC-FR-BTL-10-53 | FR-BTL-10 / DES-BTL-07、DES-PRG-02、DES-WLD-01 | 敵の前で負傷中、戦闘定義なし。草原は未解放／解放済み | StartBattleFacing | 順に ErrLocked／ErrNoBattle、負傷検査より先、全状態と乱数不変 | 異常 |
| TC-FR-BTL-10-54 | FR-BTL-10 / DES-BTL-07、DES-WLD-01 | 草原の敵を向き、Now=0、InjuredUntil=100 | StartBattleFacing | ErrInjured、全文「負傷中です」、全状態と乱数不変 | 異常 |
| TC-FR-BTL-10-55 | FR-BTL-10 / DES-BTL-07 | 草原 (528,240)、上 | 上に移動して Tick を10回 | 敵の前 (528,224)、上で止まる。Battle=nil | 正常 |
| TC-FR-BTL-10-56 | FR-BTL-10、IF-06 / DES-BTL-07、DES-FLD-02・05 | 草原 (528,224)、上、戦闘開始済み | 右へ移動を試み、味方HPを0にして敗北、下へ移動 | 戦闘中は ErrInBattle、全状態と乱数不変。敗北後は同じ場面・位置、InjuredUntil=3600、続く移動で (528,231)、下 | 正常・異常 |
| TC-FR-FLD-05-02 | FR-FLD-05 / DES-PCK-02 | Now=0、空の在庫、草原と同じ配置の other 地図と採取地 | meadow (3,2)、meadow (1,1)、other (3,2) で順に拾う | 3操作とも成功、薬草Q40が9個、戻る tick 表はこの3キーだけ各120 | 正常 |
| TC-FR-FLD-05-03 | FR-FLD-05 / DES-PCK-02 | DaySeconds=1、(meadow,3,2) の戻る tick=120、Now=59。敵 (meadow,5,1)・壁 (meadow,0,0)・未知の地図 (missing,3,2) にも照会前に戻る tick=120 を設定 | Tick、PickRemaining、Nowを120・121にして再照会 | 2日目、残り60 tick、未知の地図・壁・敵は表に未来の tick があっても0、拾う場所は120以降に0。表の戻る tick=120 は照会・日替わりで不変 | 境界 |
| TC-FR-FLD-03-88 | FR-FLD-03、FR-MAT-01 / DES-FLD-05、DES-PCK-01 | 一覧、Now=0、InjuredUntil=100 | 草原に入る、上へ3回、岩で拾う | 入場と採取が成功、薬草Q40×3、負傷期限は100のまま | 正常 |
| TC-FR-MAT-01-60 | FR-MAT-01、NFR-06 / DES-PCK-01、DES-GAT-01・03 | 同一シード42の2世界、herb重み3（Q1〜2）、potion重み1（Q50）、Rolls=3 | 同じ拾う場所で PickFacing | 両方の戻り値・在庫が herb Q2×1、potion Q50×2 と完全一致し、乱数状態も一致 | 正常 |
| TC-IF-08-74 | IF-08 / DES-FLD-07 | mixed 地図 (2,1)=拾う場所、(3,1)=敵、主人公 (288,224)／(290,224)、上 | FacingTarget | 同面積では TargetPick (2,1)、面積差では TargetEnemy (3,1)。IDは空、true、位置不変 | 境界 |
| TC-IF-08-75 | IF-08 / DES-FLD-07 | mixed 地図 (3,1)=敵、(3,2)=依頼板、主人公 (256,192)／(256,194)、右 | FacingTarget | 同面積では上の TargetEnemy (3,1)、面積差では TargetFixture quest_board (3,2)、true、位置不変 | 境界 |
| TC-IF-08-76 | IF-08 / DES-FLD-07 | mixed 地図 (2,2)=区画 a、施設market、主人公 (240,320)、上 | FacingTarget | TargetLot market (2,2)、true、位置不変 | 正常 |
| TC-IF-08-77 | IF-08 / DES-FLD-07 | 未知の地図／一覧／向きbad／左の壁だけを向く | FacingTarget | すべて Target{}、false、主人公不変 | 異常 |
| TC-FR-FLD-02-81 | FR-FLD-02 / DES-FLD-02・03 | 草原 (144,224)、上で草を向く／(528,224)、上で敵を向く | MovePlayer(0,-1)、施設建設済みでも繰り返す | それぞれ草・敵に遮られ座標・向き不変 | 境界 |
| TC-FR-FLD-02-82 | FR-FLD-02 / DES-FLD-02・03 | 街 (512,240)、左で区画marketを向く | MovePlayer(-1,0)、market段階1にして再実行 | 建設前後とも街 (512,240)、左のまま | 境界 |
| TC-IF-08-78 | IF-08 / DES-FLD-07、DES-PCK-01・02、DES-BTL-07、DES-WLD-01 | mixed 地図の (2,1)=拾う場所、(3,1)=敵、主人公は上向き | (290,224) で FacingTarget・PickFacing、(288,224) で FacingTarget・StartBattleFacing・PickFacing | (290,224) は敵 (3,1)、PickFacing は ErrNothingToPick で全状態と乱数不変。(288,224) は拾う場所 (2,1)、StartBattleFacing は ErrNoEnemy で全状態と乱数不変、PickFacing は薬草Q40×3、戻る tick の表は (mixed,2,1)=120 だけ | 境界 |
| TC-FR-FLD-03-89 | FR-FLD-03 / DES-FLD-05、DES-IT05 1.4 節、DES-IT06 3.6 節 | 街への出口を持つ採取地／通常地図。街に元の地図への戻り道あり／なし。戻り道がある場合は gather 行きとの並び順を入れ替える | 出口の手前 (336,378) から MovePlayer(0,1) | 並び順にかかわらず戻り道の到着マス (144,144)、上を優先。戻り道がない採取地だけ街の gather 行きの到着マス (432,528)、上へ移る。通常地図に戻り道がなければ元の地図 (336,385)、下のまま | 境界 |
| TC-IF-08-80 | IF-08 / DES-FLD-05・07 | data.Base を defload.Load で読んだ本体目録 | NewWorld 直後に FacingTarget | TargetFixture、workshop_door、(10,14)、true、主人公不変（E-110） | 正常 |

### #105 による既存の試験の置き換え（反復 6 で変更）

元の仕様書の該当行にも「（反復 6 で変更）」を記す。抽選の回数・品質・在庫あふれ・乱数の保持・当たり判定の検証は維持し、抽選の既知シードの個数と待ち時間の全文も厳密に検証する。

| 既存 ID | 変更内容 |
|---|---|
| TC-FR-MAT-01-01〜02 | GatherをPickFacingへ移行。シード42・1000回は herb Q1×353、Q2×397、potion Q50×250。シード1・100回のQ1〜2はQ1×54、Q2×46。反復 5 の Gather と同じ結果（10ef801 の Gather で同じシード・同じ表から照合した）。herb が potion より多いこと、品質の両端が出ることも検査する |
| TC-FR-MAT-01-03〜07 | 採取地の再使用待ちを、同じ拾うマスの再出現待ちへ変更。未知の採取地・在庫あふれの検証を維持。待ちの誤りはErrPickNotReady、切り上げ秒数を含む全文一致 |
| TC-FR-GOL-04-05〜06、TC-FR-GOL-04-08 | プレイヤーをPickFacingへ移行。06はRespawnSeconds=12として待ちが820 tickまで続き、ゴーレム採取720 tick・合計6個・待ち不干渉を維持。08は拾うマスの待ちを設定しない |
| TC-IF-02-14〜15 | 廃止したGatherからEnterGatherへ変更。14は未解放の誤りで採取地の表示名を検証、15は未知IDを文に残す検証を維持 |
| TC-FR-PRG-04-09、TC-FR-PRG-04-11、TC-FR-PRG-04-13 | Gatherの解放検査をEnterGatherへ移行。未知ID・未解放の検査順序を維持し、ランク到達後は採取地の入口に入れる |
| TC-FR-FLD-02-01 | NewWorldの初期位置を既存テスト用の街の開始位置 (336,224)、上向きへ変更（Y を当たり判定の上端に合わせる。E-110） |
| TC-IF-08-01〜06、TC-IF-08-11、TC-IF-08-17〜21 | FacingFixtureをFacingTargetへ変更。従来の設備ID・選択規則・不変条件を維持。地図外はTarget{}、false |
| TC-IF-08-13 | 街から始まるため、右3回の期待位置を (357,224) に変更（E-110）。Tickは位置を変えず、移動はNow=10を変えない |

ui は API 置換と一覧の「行く」への接続だけを行う。画面の完成と手動試験は #107・#108・#109 が担当する。

#105 の検証補足: 再出現の境界・戻る tick・Rolls・面積の比較・開始時の向き・戦闘時の場面・在庫あふれ時の乱数復元・解放検査に、それぞれ意図的な誤りを入れた8種類すべてでテストが失敗することを確認し、元に戻した。設計が前提とする検証済みの地図以外への防御として、街への出口のない採取地の地図も EnterGather で ErrNoGatherMap とする（状態不変）。この未検証データの誤りの種類は設計に明記されていない。SRS・設計書は変更していない。

## #106 本体データ（DES-DAT-04）

組み込みの `data.Base` を `defload.Load` で読み、検証が成功した目録を照合する。単体テストは `data/base_it06_test.go`。施設の全段階・製品・レシピ・追加依頼は構造体全体を比較する。

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-FR-MOD-02-60 | FR-MOD-02、FR-FLD-01・04 / DES-DAT-04、DES-VAL-17〜29 | 反復6の本体TOML一式がある | 読み込みと件数の照合 | 検証成功。アイテム19、レシピ15、施設4、依頼17、地図5、採取地4 | 正常 |
| TC-FR-MOD-02-61 | FR-MOD-02、FR-TWN-02 / DES-DAT-04 5.1 | 本体に新製品2件と各レシピがある | 製品・レシピの全項目を照合 | iron_ingot=鉄の延べ棒、属性[金属,硬質]、売値30、使用効果なし。ランク1、鉄鉱石3→1個、5秒、硬質3個で+10。herb_rope=薬草の縄、属性[繊維,しなやか]、売値20、使用効果なし。ランク1、薬草2・小麦1→1個、4秒、草2個で+10 | 正常 |
| TC-FR-MOD-02-62 | FR-ECO-07 / DES-DAT-04 5.2、DES-VAL-15 | 既存9件に正式な8件を加えた依頼 | 追加8件のID・ランク・評判・製品・品質・個数・報酬とランク別件数を照合 | 下表の8件と完全一致。ランク1/2/3/4の件数は6/3/4/4。q_cheese_extraは存在しない | 正常・境界 |
| TC-FR-MOD-02-63 | FR-PRG-05、FR-FLD-05 / DES-DAT-04 5.3、DES-DAY-01、DES-PCK-01・02 | 本体の日と採取の設定 | DaySeconds、各採取地のRespawnSecondsとRollsを読む | 日=300秒。風の草原45秒、泉の洞窟45秒、陽だまりの丘60秒、水晶の森75秒。各Rolls=3 | 正常 |
| TC-FR-TWN-02-50 | FR-TWN-02・07 / DES-DAT-04 5.1、DES-DEF-12 | workshop・market・golem_shed・innが各2段階ある | 施設のID・名前・全段階の費用・材料・効果・見た目を比較 | 下表の全8段階と完全一致 | 正常 |
| TC-FR-TWN-07-50 | FR-TWN-07 / DES-DAT-04 5.4、DES-DEF-11 | 街にa〜dの3×3区画がある | 区画の文字・施設・左上座標を照合 | a=market(3,2)、b=inn(24,2)、c=workshop(3,8)、d=golem_shed(24,8)の順で4件 | 正常 |
| TC-FR-FLD-01-44 | FR-FLD-01・04 / DES-DAT-04 5.4 | 本体に街と採取地4枚がある | 街の名前・開始位置・全20行・出口、工房地図の有無を照合 | 名前は街、30×20、開始(10,15)。rowsは設計5.4の全行に完全一致。出口は2、位置(14,19)、行き先gather、方向up、絵roadの1件。workshop地図はない | 正常 |
| TC-FR-FLD-01-45 | FR-FLD-01・04・05 / DES-DAT-04 5.4、DES-DEF-10 | 4枚の採取地地図がある | 各地図の名前・大きさ・開始位置・出口・区画・拾う場所・敵・景色の文字数を照合 | 各24×16、名前は風の草原/泉の洞窟/陽だまりの丘/水晶の森。区画なし、到着(12,14)、出口は1、位置(12,15)、行き先town、方向up、絵roadの1件。各敵は(19,8)だけ、拾う場所と景色は下表と完全一致 | 正常・境界 |
| TC-FR-FLD-04-43 | FR-FLD-04 / DES-DAT-04 5.4、DES-FLD-06 | 街の工房の扉(10,14)、店(19,15)、依頼板(13,16)、出口(14,19) | 各対象の隣接床から他の対象の隣接床まで、床だけを上下左右に歩く最短歩数を算出 | 工房→店8、工房→依頼板3、工房→出口7、店→依頼板5、店→出口7、依頼板→出口2。設計の計算例と一致 | 正常・境界 |
| TC-FR-FLD-04-44 | FR-FLD-04・05 / DES-DAT-04 5.4 | 各採取地の出口の到着マス(12,14)から床を歩く | 拾う場所・敵の隣接床までの最短歩数を算出 | 下表の各地点に到達し、(5,5)=15、(10,4)=11、(17,6)=12、(5,10)=10、(17,11)=7、(10,9)=6、(8,7)=10、(15,3)=13マス。各地図に存在する地点だけを照合。全4枚の最遠は15マス、敵までは12マス | 正常・境界 |

施設の照合値（上段からレベル1・2）:

| ID・名前 | 費用 | 材料 | 効果 | 見た目 |
|---|---|---|---|---|
| workshop 錬金工房 1 | 60 | bread×2 | golem_slots +1 | cottage |
| workshop 錬金工房 2 | 140 | iron_ingot×2、lantern×1 | golem_slots +1 | house |
| market 街の市場 1 | 70 | burger×2 | sell_rate +10 | stall |
| market 街の市場 2 | 150 | herb_rope×3、cheese×2 | sell_rate +10 | house |
| golem_shed ゴーレム小屋 1 | 100 | iron_ingot×3 | golem_slots +1 | cottage |
| golem_shed ゴーレム小屋 2 | 220 | iron_ingot×4、glow_crystal×2 | golem_slots +1 | manor |
| inn 旅人の宿 1 | 90 | herb_rope×2、bread×3 | sell_rate +5 | cottage |
| inn 旅人の宿 2 | 200 | burger×3、potion×2 | sell_rate +10 | manor |

追加依頼の照合値:

| ID | ランク | 評判 | 製品 | 品質 | 個数 | 報酬 |
|---|---|---|---|---|---|---|
| q_flour | 1 | 1 | flour | 20 | 3 | 45 |
| q_potion_easy | 1 | 1 | potion | 20 | 2 | 50 |
| q_rope | 1 | 1 | herb_rope | 20 | 2 | 55 |
| q_ingot | 2 | 3 | iron_ingot | 50 | 2 | 90 |
| q_rope_fine | 3 | 4 | herb_rope | 100 | 3 | 120 |
| q_ingot_fine | 3 | 4 | iron_ingot | 100 | 2 | 140 |
| q_bomb_master | 4 | 0 | bomb | 150 | 2 | 200 |
| q_lantern_master | 4 | 0 | lantern | 150 | 1 | 220 |

採取地の配置（座標は0始まり。記載順ではなく座標と文字の集合で比較）:

| 地図 | 拾う場所 |
|---|---|
| meadow | *(5,5)、*(10,4)、o(17,6)、*(5,10)、o(17,11)、*(10,9) |
| spring_cave | o(5,5)、o(10,4)、o(17,6)、*(5,10) |
| sunny_hill | *(5,5)、*(10,4)、o(17,6)、*(5,10)、o(17,11) |
| crystal_forest | o(5,5)、*(10,4)、o(17,6)、*(5,10)、o(17,11)、*(10,9)、o(8,7)、o(15,3) |

景色は草原の草地と木、洞窟の石床・壁・泉、丘の草地・岩（`#` 石の壁）・木、水晶の森の密な木と水で分ける。文字数の期待値:

| 地図 | 文字と個数 |
|---|---|
| meadow | `*`=4、`,`=295、`1`=1、`E`=1、`T`=81、`o`=2 |
| spring_cave | `#`=81、`*`=1、`.`=291、`1`=1、`E`=1、`o`=3、`~`=6 |
| sunny_hill | `*`=3、`,`=289、`.`=2、`1`=1、`E`=1、`#`=6、`T`=80、`o`=2 |
| crystal_forest | `*`=3、`,`=239、`1`=1、`E`=1、`T`=133、`o`=5、`~`=2 |

### #106 既存の試験の置き換え（反復 6 で変更）

元のテスト仕様書の行に「（反復 6 で変更）」を記し、従来の検証を維持して完成データの期待値へ更新する。

| 既存 ID | 変更内容 |
|---|---|
| TC-FR-MOD-02-04 | アイテム19件（新製品2件を追加）、レシピ15件、施設4件、依頼17件、地図5枚。他の件数・ID検証は維持 |
| TC-FR-MOD-02-25 | 市場の費用の完全一覧を[70]から[70,150]へ更新。工房費用と品質20の製品売値は維持 |
| TC-FR-MOD-02-41 | レシピ15件、ランク1は10件。iron_ingot・herb_ropeをランク1として追加し、既存13件のランクは維持 |
| TC-FR-MOD-02-42 | 仮依頼q_cheese_extraをq_ropeへ置換し、ランク2〜4の正式5件を追加。全17件の完全比較を維持し、ランク別件数を6/3/4/4へ変更 |

新レシピの手作りボーナスは設計5.1がデータ担当に委ねた範囲で、鉄の延べ棒を硬質3個で+10、薬草の縄を草2個で+10とした。設計・SRSの変更なし。異常データの拒否は #103 の既存defload試験で扱い、#106は完成データの正常値と距離の境界を検証する。

F-03 の検証補足: TC-FR-FLD-05-03 は、拾う場所でないマスにも未来の戻る tick を設定した上で 0 を確認する。PickRemaining から拾う場所の種類の判定を外す変異で、壁 (meadow,0,0) が残り60 tickを返して試験が失敗することを確認し、実装を元に戻した。

## #107 工房・区画・依頼板・日の表示

描画を伴わない計算は、Ebitengine に依存しないパッケージ `internal/ui/uimodel` に置き、`go test ./internal/ui/uimodel` で実行する。ui パッケージにテストを置くと、テストの実行時に Ebitengine が画面を初期化して、画面のない CI と WSL で落ちるため（REV-059 F-09）。

| ID | 対象（要求・設計） | 前提 | 入力・操作 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-IF-06-40 | IF-06 / DES-UI-22 | タブが未選択または選択済み | タブの解決 | 未選択は錬金、錬金・ゴーレム・在庫は選択を保持 | 単体・正常・境界 |
| TC-IF-06-41 | FR-ECO-06・07 / DES-UI-25 | 小麦粉、品質20、3個、45円、評判1 | 依頼の文字列を組み立てる | 製品名、品質の下限、個数、報酬、評判を欠かさず表示 | 単体・正常 |
| TC-IF-06-42 | IF-06 / DES-UI-28 U-5 | 採取地の一覧 | 配置定数を確認 | 戻るボタンは(10,76)、120×28、一覧はy112、両者は重ならない | 単体・境界 |
| TC-IF-08-61 | IF-08 / DES-UI-22 | 工房の扉を向く | 開く窓を選ぶ | 工房の窓 | 単体・正常 |
| TC-IF-08-62 | IF-08 / DES-UI-23 | 市場の区画を向く | 開く窓を選ぶ | 区画の窓、施設IDはmarket | 単体・正常 |
| TC-IF-08-63 | IF-08 / DES-UI-22・23 | 敵を向く | #107で開く窓を選ぶ | 工房や区画の窓を開かない | 単体・異常 |
| TC-IF-08-64 | IF-08 / DES-UI-28 U-4 | 主人公の絵はy-32〜y+16 | 上下の向きの印を計算 | 上は頭の上、下は足の下、全頂点が絵の外 | 単体・境界 |
| TC-FR-TWN-07-60 | FR-TWN-07 / DES-UI-24 | 市場の段階0・2 | 名前を組み立てる | 市場（空き地）、市場 Lv2 | 単体・正常・境界 |
| TC-FR-TWN-07-61 | FR-TWN-07 / DES-UI-24 | 本体目録の4施設・段階0〜2 | lotLookで絵を選び、lotTilesの全要素を照合 | DES-DAT-04 5.1節どおり、marketは空き地→stall→house、workshopは空き地→cottage→house、golem_shed・innは空き地→cottage→manor。空き地のTTは上段44,45,46／中段56,0,58／下段68,83,70。他の4種も設計書7章のタイル表の9枚すべてと一致 | 単体・正常 |
| TC-FR-PRG-05-60 | FR-PRG-05 / DES-UI-27 | 操作結果あり。日替わり・ランク上昇あり／なし | 日の表示と通知を組み立て、2〜100日目を繰り返す | 「2 日目」「2 日目になりました。依頼板の依頼が入れ替わりました」。新しい知らせは操作結果を置き換え、日だけなら常に最新1件。同じ更新のランク上昇と日の知らせだけを「／」で連結。知らせなしなら表示を保持 | 単体・正常 |

Windows 手動試験は別担当が実行する。2026-10-22の結果と修正後の再試験予定を末尾に記録する。期待結果は今回の修正を反映したものであり、変更した箇所の合格は再試験で確認する。新規開始、必要に応じてF12で費用・素材を補充する。依頼の在庫不足は補充前に確認する。

| ID | 対象（要求・設計） | 前提 | 手順 | 期待結果 | 種別 |
|---|---|---|---|---|---|
| TC-IF-06-43 | IF-06、IF-08 / DES-UI-22 | 新規開始 | 起動直後の操作結果を確認。Eで工房を開く。錬金、ゴーレム、在庫を順にクリック。Escで閉じ、スペースで再び開く。「閉じる」でも閉じる | 起動時は「WASD・矢印キーで歩き、工房の扉や店の前で E かスペースを押してください」。見出しは工房。「閉じる」は(10,76)。初回は錬金。各タブに従来の欄が出る（錬金: 在庫・レシピ・素材候補・手作り錬金、ゴーレム: ゴーレム・割り当て、在庫: 在庫だけ）。欄はy112以降、下端y935以下で隠れる行はページ送りできる。再度開くと在庫。開いている間は移動キーやE・スペースで位置・向き・操作結果が変わらない。閉じても主人公の位置は変わらない | 手動・正常 |
| TC-IF-06-44 | IF-06 / DES-UI-22 | 工房の錬金タブ | レシピ、入力枠、素材を選び錬金。ゴーレムタブで作成、採取・錬金の割当、解除。在庫タブで確認。窓を閉じて再び開き、選択の保持を確認。窓を開いた間と閉じた間で約4秒ずつ待ち進捗を確認 | 素材選び・予想品質・錬金・作成・割当・解除が従来どおり動く。在庫タブから売れない。ボタンの描画位置とクリック位置が一致。錬金成功時は帯に「完成: …」。レシピ・ページ・割り当ての選択は閉じて開いても残る。ゴーレムの進捗は窓を開いた間も閉じた間も時間とともに進む | 手動・正常 |
| TC-IF-06-45 | FR-TWN-02、IF-08 / DES-UI-23 | 未建設の区画 | 各区画の前でE。費用不足で建設、F12後に建設・強化、最大まで繰り返す | 見出しが対象の施設名。在庫は売れず、対象施設だけの現在段階・費用・材料・効果・建設ボタンが出る。不足時は理由を表示して状態不変。最大時は「最大の段階です」、建設ボタンなし | 手動・正常・異常・境界 |
| TC-FR-TWN-07-62 | FR-TWN-07 / DES-UI-24 | 街の4区画 | 未建設、建設後、強化後に窓を閉じて地図を確認。区画の各辺から向く | 未建設は3×3の草地を柵（TT上段44,45,46／中段56,0,58／下段68,83,70）で囲み、下の中央に立て札83。DES-DAT-04 5.1節どおり、街の市場は段階1で屋台stall（57）、2で灰色屋根のhouse。錬金工房は1で赤い屋根のcottage、2でhouse。ゴーレム小屋・旅人の宿は1でcottage、2で灰色屋根に飾りのあるmanor。家はいずれも下段に壁と扉86。施設名と空き地またはLvを1か所に表示。向いている区画だけ「E: 街の市場 Lv1」のように接頭辞が付き名前が明るい | 手動・正常 |
| TC-IF-06-46 | FR-ECO-06・07 / DES-UI-25 | 今日の依頼3件、受注0件 | 依頼板を開き、3件受ける。受けた依頼の中央をやめる | 今日の依頼と受けた依頼（n/3）が分かれる。受けると候補位置は「（明日また来てください）」。中止は確認なし、罰なし、後ろが詰まる。やめても候補は補充されない | 手動・正常 |
| TC-IF-06-47 | FR-ECO-05・06 / DES-UI-25 | 受注済みで納品在庫不足 | 受注一覧の先頭の納品を押す。必要な製品を作り、再び納品 | 不足時は「操作失敗: 依頼番号 1: アイテムが足りません」（受注一覧の先頭を納品した場合）。在庫・お金・評判・候補・受注一覧は不変。充足後は受注一覧から消え、在庫減少・お金・評判が更新される。候補は補充されない。ランク上昇時は知らせが出る | 手動・正常・異常 |
| TC-IF-06-48 | FR-ECO-06 / DES-UI-25 | 受注3件 | F11で翌日まで進め、新しい候補の「受ける」を押す | 「操作失敗: 受けられる依頼は 3 件までです」。候補・受注一覧は変わらない。やめて空けると受けられる | 手動・境界 |
| TC-FR-PRG-05-61 | FR-PRG-05、FR-ECO-07 / DES-UI-27 | 街、受注あり | 上の帯を確認し、工房・依頼板の窓を開いたままF11で翌日まで待ち、操作せずさらに数日進める | 経過時間の右が1日目から2日目へ更新。操作結果は「n 日目になりました。依頼板の依頼が入れ替わりました」に置き換わる。何日たっても日の知らせは最新の1件だけで窓の見出しに重ならない。同じUpdateでランクも上がったときは、そのランクの知らせと日の知らせを「／」でつなぐ。候補を引き直し、受注は残る。同じ候補の再選出は許容。知らせがない次の更新では表示を保持し、重複しない | 手動・正常・境界 |
| TC-IF-08-65 | IF-08 / DES-UI-28 U-4・U-5 | 街・採取地の地図と店の窓 | 両地図で帯の1行目を確認。上下を向き、店で売買数を増減し、購入・在庫選択・売却を行う。街の出口から一覧へ進み「街へ戻る」を押す | 街・採取地の地図ではランクと評判「ランク n / 評判 p / 次の必要評判」（最高時は「ランク n（最高）/ 評判 p」）が(700,10)に出て、操作結果と重ならない。印が顔に重ならず絵の外。売買数と±が店の見出し行内、1〜99で止まり、購入・売却数に反映。店の窓には在庫と店の欄があり、在庫のクリックだけでは売れず、「指定数を売る」で売却する。戻るボタンは(10,76)、120×28、一覧はy112以降。押すと街に戻る | 手動・正常・境界 |

手動試験の結果記録:

| ID | 結果 |
|---|---|
| TC-IF-06-43 | 合格（2026-10-22、Windows、4e63bcb）。再試験も合格（2026-10-22、Windows、df8a4d7。起動文・開いている間のキー操作・「閉じる」を追加で確認） |
| TC-IF-06-44 | 合格（2026-10-22、Windows、4e63bcb） |
| TC-IF-06-45 | 合格（2026-10-22、Windows、4e63bcb） |
| TC-FR-TWN-07-62 | 不合格（2026-10-22、Windows、4e63bcb）。空き地の外周のTT 32が木の切れ端に見え、柵になっていない。そのほかの見た目・名前・向いた区画の強調は合格。F-02の修正後に再試験する（F-04の「E: 」も併せて確認）。再試験は合格（2026-10-22、Windows、df8a4d7）。空き地の柵（44,45,46／56,0,58／68,83,70）と立て札、4 施設の段階 1・2 の見た目、各辺から向いたときの「E: 」付きの明るい名前を確認 |
| TC-IF-06-46 | 合格（2026-10-22、Windows、4e63bcb） |
| TC-IF-06-47 | 合格（2026-10-22、Windows、4e63bcb）。再試験も合格（2026-10-22、Windows、df8a4d7。誤りの全文「操作失敗: 依頼番号 1: アイテムが足りません」を確認） |
| TC-IF-06-48 | 合格（2026-10-22、Windows、4e63bcb） |
| TC-FR-PRG-05-61 | 合格（2026-10-22、Windows、4e63bcb）。F-01の修正後に再試験する。再試験は合格（2026-10-22、Windows、df8a4d7）。依頼板・工房の窓を開いたまま 1〜6 日目まで操作せずに進め、日の知らせは最新の 1 件だけに置き換わることを確認。ランクと日が同じ更新で重なる場合は単体試験 TC-FR-PRG-05-60 で確認 |
| TC-IF-08-65 | 合格（2026-10-22、Windows、4e63bcb）。F-03・F-04の修正後に再試験する。再試験は合格（2026-10-22、Windows、df8a4d7）。ランク・評判が街・採取地の地図で (700,10) に出て、操作結果と重ならないことと、起動したときの文を確認。向きの印・店・戻るボタンは 4e63bcb の結果を引き継ぐ（該当コードの変更なし）。最高ランクの表示は未確認 |

#107 は採取地の一覧の内容・採取地の描画・拾う・戦う・戦闘結果の窓を変更しない。それらは #108 の確認対象。反復5の工房の設備別の窓、施設の看板、依頼板から直接納品する画面、向きの印と戻るボタンの旧位置は、#107 の上記試験に置き換える。既存のcoreの検証は維持する。

区画の絵の選択はDES-UI-24が実装担当に委ねた範囲。使用番号は設計書7章のタイル表と `internal/ui/assets/README.md` に記録し、要求・設計の決定事項は変更しない。

### #107 による既存の手動試験の置き換え（反復 6 で変更）

元の仕様書（test-spec-it05.md）の該当行に「（反復 6 で変更）」を付け、過去の結果は当時の記録として残す。

| 既存 ID | 置き換え先・変更内容 |
|---|---|
| TC-FR-FLD-03-42 | TC-IF-08-65。採取地一覧の「街へ戻る」を(270,76)から(10,76)へ変更 |
| TC-IF-08-42・43・44・58 | TC-IF-06-43・44。錬金釜・棚の窓を工房の錬金・在庫タブへ変更。閉じ方・移動抑止・欄の範囲・ページ送りの検証を維持 |
| TC-IF-08-45 | TC-IF-06-44。工房の錬金タブで素材・品質・錬金・選択の保持を確認 |
| TC-IF-08-46・47・57 | TC-IF-06-43・44。作業台の窓を工房のゴーレムタブへ変更。作成・割り当て・解除・進捗・選択の保持を確認 |
| TC-IF-08-48 | TC-IF-08-65。売買数をy137付近から店の見出し行へ移動。店の在庫・購入・売却は従来どおり確認 |
| TC-IF-08-50 | TC-IF-06-46・47・48。今日の依頼と受注一覧を分離し、受注済みだけ納品。ランク上昇の知らせの確認を維持 |
| TC-IF-08-51 | TC-IF-06-45、TC-FR-TWN-07-62。施設の看板を区画の窓に変更し、施設ごとに建設・強化する |
| TC-IF-08-53 | TC-IF-06-43、TC-FR-TWN-07-62。起動時の操作結果を工房の扉・店の案内へ変更し、区画にも「E: 」を表示 |
| TC-IF-08-60 | TC-IF-08-65。ランク・評判は街と採取地の地図の1行目に表示 |

修正後の単体試験では、TC-FR-TWN-07-61の段階選択を1つずらす変異で全4施設の段階1・2の照合が失敗することを確認し、実装を元に戻した。追加・変更した期待結果を含むWindowsの手動再試験は未実施。
