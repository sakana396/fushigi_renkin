package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
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

// 埋め込んだファイルが配布物から変わっていないことを、配布 zip から求めた SHA-256 で確かめる。
func TestEmbeddedHashes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"tiny_town.png", tinyTownPNG, "3a54d99ecde790d4fdea207a3644cf130fc56fa838f1beb1507c185a95b8e902"},
		{"tiny_dungeon.png", tinyDungeonPNG, "d24e60a41e4ac7a745c0304dfde121143688557f40215f23221c29cfe683825f"},
		{"kenney_tiny_town_License.txt", tinyTownLicense, "c483c167cc3bdfe17c15f957ff847d409f64d81b410b28783636cf3b2b7814c5"},
		{"kenney_tiny_dungeon_License.txt", tinyDungeonLicense, "9f574a2f1f636a3db8a0665ba90212f34e2b5e61ecb533d77c05237766374111"},
	}
	for _, tt := range tests {
		t.Run("TC-CST-04-06 "+tt.name, func(t *testing.T) {
			sum := sha256.Sum256(tt.data)
			if got := hex.EncodeToString(sum[:]); got != tt.want {
				t.Errorf("SHA-256 = %s、期待値 = %s", got, tt.want)
			}
		})
	}
}

// 2 つの関数が、それぞれ自分の素材の画像を返すことを確かめる。
func TestImagesAreDistinct(t *testing.T) {
	t.Run("TC-CST-04-07 素材を取り違えない", func(t *testing.T) {
		town, err := TinyTown()
		if err != nil {
			t.Fatal(err)
		}
		dungeon, err := TinyDungeon()
		if err != nil {
			t.Fatal(err)
		}
		wantTown, _ := png.Decode(bytes.NewReader(tinyTownPNG))
		wantDungeon, _ := png.Decode(bytes.NewReader(tinyDungeonPNG))
		if !samePixels(town, wantTown) {
			t.Error("TinyTown() が tiny_town.png の画像を返さない")
		}
		if !samePixels(dungeon, wantDungeon) {
			t.Error("TinyDungeon() が tiny_dungeon.png の画像を返さない")
		}
		if samePixels(town, dungeon) {
			t.Error("2 つの素材の画像が同じ")
		}
	})
}

func samePixels(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	r := a.Bounds()
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if color.RGBAModel.Convert(a.At(x, y)) != color.RGBAModel.Convert(b.At(x, y)) {
				return false
			}
		}
	}
	return true
}
