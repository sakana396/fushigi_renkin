# 画像素材の出典

ADR-0003 の決定 3 に従い、配布物の画像と同梱の利用許諾を保存する。利用許諾のファイルは原本から変更していない。

| 素材名と版 | 作者 | 配布元 | 利用許諾 | 確かめた日 | 取り込んだファイル（元のファイル名） |
|---|---|---|---|---|---|
| Tiny Town 1.1 | Kenney | <https://kenney.nl/assets/tiny-town> | CC0 1.0（`kenney_tiny_town_License.txt`） | 2026-10-15 | `tiny_town.png`（`Tilemap/tilemap_packed.png`）、`kenney_tiny_town_License.txt`（`License.txt`） |
| Tiny Dungeon 1.0 | Kenney | <https://kenney.nl/assets/tiny-dungeon> | CC0 1.0（`kenney_tiny_dungeon_License.txt`） | 2026-10-15 | `tiny_dungeon.png`（`Tilemap/tilemap_packed.png`）、`kenney_tiny_dungeon_License.txt`（`License.txt`） |

## #107 区画のタイル（DES-UI-24）

既存のTiny Townを使用。番号は左上0、横12枚で数える。各行は区画の上・中・下の3マス。主人公の向きの印は図形で描く。

| 見た目 | 上段 | 中段 | 下段 |
|---|---|---|---|
| 空き地 | 44,45,46 | 56,0,58 | 68,83,70 |
| stall | 0,0,0 | 0,57,0 | 0,0,0 |
| cottage | 52,53,54 | 64,65,66 | 72,86,75 |
| house | 48,49,50 | 60,61,62 | 84,86,87 |
| manor | 48,51,50 | 60,63,62 | 84,86,87 |

空き地は草地に柵と立て札（83）。屋台は57。家は上2段が屋根、下段が壁と扉（86）。cottageとhouseは屋根の色、manorは屋根の飾りで区別する。新しい画像ファイルの追加はない。
