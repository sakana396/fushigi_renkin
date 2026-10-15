package assets

import (
	"image"
	"strings"
	"testing"
)

func TestImages(t *testing.T) {
	tests := []struct {
		name string
		load func() (image.Image, error)
	}{
		{"TC-CST-04-01 Tiny Town の画像", TinyTown},
		{"TC-CST-04-02 Tiny Dungeon の画像", TinyDungeon},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := tt.load()
			if err != nil {
				t.Fatalf("PNG の読み込み: %v", err)
			}
			if got := img.Bounds().Size(); got.X != 192 || got.Y != 176 {
				t.Errorf("画像の大きさ = %dx%d、期待値 = 192x176", got.X, got.Y)
			}
		})
	}
}

func TestLicenses(t *testing.T) {
	tests := []struct {
		name    string
		license string
		asset   string
	}{
		{"TC-CST-04-03 Tiny Town の利用許諾", string(tinyTownLicense), "Tiny Town (1.1)"},
		{"TC-CST-04-04 Tiny Dungeon の利用許諾", string(tinyDungeonLicense), "Tiny Dungeon (1.0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range []string{tt.asset, "Creative Commons Zero, CC0", "http://creativecommons.org/publicdomain/zero/1.0/"} {
				if !strings.Contains(tt.license, want) {
					t.Errorf("利用許諾に %q がない", want)
				}
			}
		})
	}
}
