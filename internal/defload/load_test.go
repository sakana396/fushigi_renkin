package defload

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func files(entries map[string]string) fs.FS {
	m := fstest.MapFS{}
	for name, body := range entries {
		m[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return m
}

type readFailFS struct{ fs.FS }

func (r readFailFS) ReadFile(name string) ([]byte, error) {
	return nil, errors.New("読み取り不可")
}

// 読み込み経路と基本的な失敗を確認する（DES-PKG-03）。
func TestLoad(t *testing.T) {
	cases := []struct {
		name    string
		fs      fs.FS
		want    string
		items   int
		recipes int
	}{
		{"TC-FR-MOD-01-01 アイテムとレシピ", files(map[string]string{"items.toml": "[[item]]\nid='herb'\n", "recipes.toml": "[[recipe]]\nid='tea'\noutput='herb'\noutput_count=1\nseconds=1\ninputs=[{item='herb',count=1}]\n"}), "", 1, 1},
		{"TC-FR-MOD-01-02 空の定義", files(nil), "", 0, 0},
		{"TC-FR-MOD-01-03 文法誤り", files(map[string]string{"broken.toml": "[[item]\n"}), "broken.toml", 0, 0},
		{"TC-FR-MOD-02-01 組み込み可能な FS", mustSub(t), "", 3, 1},
		{"TC-FR-MOD-02-02 複数ファイル", files(map[string]string{"a.toml": "[[item]]\nid='a'\n", "dir/b.toml": "[[item]]\nid='b'\n"}), "", 2, 0},
		{"TC-FR-MOD-02-03 読めないファイル", readFailFS{files(map[string]string{"bad.toml": ""})}, "bad.toml", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := Load(tc.fs)
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("誤り = %v; %q を含むべき", err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(catalog.Items) != tc.items || len(catalog.Recipes) != tc.recipes {
				t.Fatalf("目録 = %+v", catalog)
			}
		})
	}
}

func mustSub(t *testing.T) fs.FS {
	t.Helper()
	sub, err := fs.Sub(os.DirFS("testdata"), "base")
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

// 検証規則と誤りの集約を確認する（DES-VAL-01〜08）。
func TestValidation(t *testing.T) {
	cases := []struct {
		name, body string
		wants      []string
	}{
		{"TC-FR-MOD-04-01 正常", "[[item]]\nid='herb'\nattributes=['草']\n[[recipe]]\nid='tea'\noutput='herb'\noutput_count=1\nseconds=1\ninputs=[{attribute='草',count=1}]\n", nil},
		{"TC-FR-MOD-04-02 境界", "[[item]]\nid='herb'\nattributes=['草','癒し','緑']\n[[gather]]\nid='field'\nrolls=1\ncooldown_seconds=1\ndrops=[{item='herb',weight=1,quality_min=1,quality_max=200}]\n", nil},
		{"TC-FR-MOD-04-03 重複と参照", "[[item]]\nid='herb'\n[[item]]\nid='herb'\n[[recipe]]\nid='tea'\noutput='missing'\noutput_count=1\nseconds=1\ninputs=[{item='missing',count=1}]\n", []string{"item[id=herb].id", "recipe[id=tea].output", "recipe[id=tea].inputs[0].item"}},
		{"TC-FR-MOD-04-04 壊れた値", "[[item]]\nid='herb'\nattributes=['草','草','A','B']\n[[recipe]]\nid='tea'\noutput='herb'\noutput_count=0\nseconds=0\ninputs=[{item='herb',attribute='X',count=0}]\nbonuses=[{attribute='X',min_count=0}]\n[[gather]]\nid='field'\nrolls=0\ndrops=[{item='herb',weight=0,quality_min=201,quality_max=0}]\n[[facility]]\nid='hut'\nlevels=[{effects=[{kind='unknown',value=1}]}]\n", []string{"item[id=herb].attributes", "item[id=herb].attributes[1]", "recipe[id=tea].output_count", "recipe[id=tea].seconds", "recipe[id=tea].inputs[0]:", "recipe[id=tea].inputs[0].attribute", "recipe[id=tea].inputs[0].count", "recipe[id=tea].bonuses[0].attribute", "recipe[id=tea].bonuses[0].min_count", "gather[id=field].rolls", "gather[id=field].drops[0].weight", "gather[id=field].drops[0].quality_min", "gather[id=field].drops[0].quality_max", "gather[id=field].drops[0]:", "facility[id=hut].levels[0].effects[0].kind"}},
		{"TC-FR-MOD-04-08 空の ID と空の一覧", "[[item]]\nid=''\n[[item]]\nid='herb'\n[[recipe]]\nid='tea'\noutput='herb'\noutput_count=1\nseconds=1\n[[gather]]\nid='field'\nrolls=1\ncooldown_seconds=1\n[[facility]]\nid='hut'\n", []string{"item[id=].id", "recipe[id=tea].inputs", "gather[id=field].drops", "facility[id=hut].levels"}},
		{"TC-FR-MOD-04-09 ゴーレムの速度と材料", "[[item]]\nid='clay'\n[[golem]]\nid='g0'\nspeed_percent=0\n[[golem]]\nid='g1'\nspeed_percent=100\nmaterials=[{item='clay',count=1}]\n", []string{"golem[id=g0].speed_percent", "golem[id=g0].materials"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(files(map[string]string{"defs.toml": tc.body}))
			if len(tc.wants) == 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("誤りがない")
			}
			for _, want := range tc.wants {
				if !strings.Contains(err.Error(), "defs.toml: "+want) {
					t.Errorf("%q がない: %v", want, err)
				}
			}
		})
	}
}

// ファイル間での誤りの集約と全定義型の変換を確かめる（DES-PKG-03、DES-VAL-02、04、05、08）。
func TestMultipleFiles(t *testing.T) {
	cases := []struct {
		name     string
		contents map[string]string
		wants    []string
	}{
		{
			name: "TC-FR-MOD-04-05 未知キーと型の誤り",
			contents: map[string]string{
				"unknown.toml": "[[item]]\nid='herb'\nunknown=1\n",
				"type.toml":    "[[item]]\nid=123\n",
				"valid.toml":   "[[recipe]]\nid='tea'\noutput='missing'\noutput_count=1\nseconds=1\n",
			},
			wants: []string{"unknown.toml: item.unknown", "type.toml: <toml>", "valid.toml: recipe[id=tea].output"},
		},
		{
			name: "TC-FR-MOD-04-06 複数ファイルの検証",
			contents: map[string]string{
				"a.toml": "[[gather]]\nid='field'\nrolls=0\ncooldown_seconds=1\ndrops=[{item='missing',weight=0,quality_min=201,quality_max=2}]\n",
				"b.toml": "[[golem]]\nid='helper'\nmaterials=[{item='missing',count=0}]\n",
			},
			wants: []string{"a.toml: gather[id=field].rolls", "a.toml: gather[id=field].drops[0].item", "a.toml: gather[id=field].drops[0].weight", "a.toml: gather[id=field].drops[0].quality_min", "b.toml: golem[id=helper].materials[0].item", "b.toml: golem[id=helper].materials[0].count"},
		},
		{
			name: "TC-FR-MOD-04-07 全定義型",
			contents: map[string]string{
				"items.toml": "[[item]]\nid='herb'\n",
				"all.toml":   "[game]\nstart_money=10\nbase_golem_slots=1\nshop=[{item='herb',quality=200,price=5}]\n[[gather]]\nid='field'\nrolls=1\ncooldown_seconds=2\ndrops=[{item='herb',weight=1,quality_min=1,quality_max=200}]\n[[golem]]\nid='helper'\nmaterials=[{item='herb',count=1}]\nspeed_percent=100\n[[facility]]\nid='hut'\nlevels=[{money=3,materials=[{item='herb',count=1}],effects=[{kind='golem_slots',value=1},{kind='sell_rate',value=5}]}]\n",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := Load(files(tc.contents))
			if len(tc.wants) > 0 {
				if err == nil {
					t.Fatal("誤りがない")
				}
				for _, want := range tc.wants {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("%q がない: %v", want, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if catalog.Game.StartMoney != 10 || len(catalog.Game.Shop) != 1 || len(catalog.Gathers) != 1 || len(catalog.Golems) != 1 || len(catalog.Facilities) != 1 || len(catalog.Facilities["hut"].Levels[0].Effects) != 2 {
				t.Fatalf("目録 = %+v", catalog)
			}
		})
	}
}
