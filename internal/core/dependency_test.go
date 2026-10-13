package core

import (
	"os/exec"
	"strings"
	"testing"
)

// core の推移的な依存を検査する（DES-PKG-04、NFR-04）。
func TestCoreDependencies(t *testing.T) {
	t.Run("TC-NFR-04-01 core の依存に Ebitengine がない", func(t *testing.T) {
		cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", ".")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("go list -deps に失敗: %v\n%s", err, output)
		}
		if dependency := ebitengineDependency(strings.Fields(string(output))); dependency != "" {
			t.Errorf("core が Ebitengine に依存している: %s", dependency)
		}
	})

	cases := []struct {
		name string
		deps []string
		want string
	}{
		{
			name: "TC-NFR-04-02 類似名は違反としない",
			deps: []string{"fmt", "example.com/ebiten/v2", "github.com/hajimehoshi/ebiten/v2extra"},
		},
		{
			name: "TC-NFR-04-03 推移的なサブパッケージ依存を検出する",
			deps: []string{"fmt", "example.com/wrapper", "github.com/hajimehoshi/ebiten/v2/inpututil"},
			want: "github.com/hajimehoshi/ebiten/v2/inpututil",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ebitengineDependency(tc.deps); got != tc.want {
				t.Errorf("違反パッケージ = %q、期待値 = %q", got, tc.want)
			}
		})
	}
}

// ebitengineDependency は推移的な依存一覧から Ebitengine のパッケージを探す（DES-PKG-04、NFR-04）。
func ebitengineDependency(deps []string) string {
	const module = "github.com/hajimehoshi/ebiten/v2"
	for _, dep := range deps {
		if dep == module || strings.HasPrefix(dep, module+"/") {
			return dep
		}
	}
	return ""
}
