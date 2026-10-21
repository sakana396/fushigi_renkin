# テスト仕様書（反復 6）

- 対象: [反復 6 計画書](../iterations/it06/plan.md) の範囲
- 根拠: 設計書 DES-IT06

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
