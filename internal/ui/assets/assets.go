// Package assets は画面で使う画像素材を埋め込み、PNG として読み込む（DES-UI-14）。
package assets

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
)

//go:embed tiny_town.png
var tinyTownPNG []byte

//go:embed tiny_dungeon.png
var tinyDungeonPNG []byte

//go:embed kenney_tiny_town_License.txt
var tinyTownLicense []byte

//go:embed kenney_tiny_dungeon_License.txt
var tinyDungeonLicense []byte

// TinyTown は埋め込まれた Tiny Town の PNG を読む（DES-UI-14）。
func TinyTown() (image.Image, error) {
	img, err := png.Decode(bytes.NewReader(tinyTownPNG))
	if err != nil {
		return nil, fmt.Errorf("Tiny Town の画像: %w", err)
	}
	return img, nil
}

// TinyDungeon は埋め込まれた Tiny Dungeon の PNG を読む（DES-UI-14）。
func TinyDungeon() (image.Image, error) {
	img, err := png.Decode(bytes.NewReader(tinyDungeonPNG))
	if err != nil {
		return nil, fmt.Errorf("Tiny Dungeon の画像: %w", err)
	}
	return img, nil
}
