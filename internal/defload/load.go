package defload

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// Diagnostic は定義データの一つの誤りを表す（DES-PKG-03）。
type Diagnostic struct {
	File   string
	Path   string
	Reason string
}

// Error はファイル・キー経路・理由を表示する（DES-PKG-03）。
func (d Diagnostic) Error() string { return fmt.Sprintf("%s: %s: %s", d.File, d.Path, d.Reason) }

// Errors はすべての読み込み・検証エラーを保持する（DES-PKG-03）。
type Errors []Diagnostic

// Error はすべての誤りを改行で連結する（DES-PKG-03）。
func (e Errors) Error() string {
	parts := make([]string, len(e))
	for i, d := range e {
		parts[i] = d.Error()
	}
	return strings.Join(parts, "\n")
}

type document struct {
	Game       *gameDef      `toml:"game"`
	Items      []itemDef     `toml:"item"`
	Recipes    []recipeDef   `toml:"recipe"`
	Gathers    []gatherDef   `toml:"gather"`
	Golems     []golemDef    `toml:"golem"`
	Facilities []facilityDef `toml:"facility"`
}
type gameDef struct {
	StartMoney     int         `toml:"start_money"`
	BaseGolemSlots int         `toml:"base_golem_slots"`
	Shop           []shopOffer `toml:"shop"`
}
type itemDef struct {
	ID         string   `toml:"id"`
	Name       string   `toml:"name"`
	Attributes []string `toml:"attributes"`
	BasePrice  int      `toml:"base_price"`
}
type recipeDef struct {
	ID          string      `toml:"id"`
	Output      string      `toml:"output"`
	OutputCount int         `toml:"output_count"`
	Inputs      []inputSlot `toml:"inputs"`
	Seconds     int         `toml:"seconds"`
	Bonuses     []bonus     `toml:"bonuses"`
}
type inputSlot struct {
	Item      string `toml:"item"`
	Attribute string `toml:"attribute"`
	Count     int    `toml:"count"`
}
type bonus struct {
	Attribute string `toml:"attribute"`
	MinCount  int    `toml:"min_count"`
	Add       int    `toml:"add"`
}
type gatherDef struct {
	ID              string `toml:"id"`
	Name            string `toml:"name"`
	Rolls           int    `toml:"rolls"`
	CooldownSeconds int    `toml:"cooldown_seconds"`
	Drops           []drop `toml:"drops"`
}
type drop struct {
	Item       string `toml:"item"`
	Weight     int    `toml:"weight"`
	QualityMin int    `toml:"quality_min"`
	QualityMax int    `toml:"quality_max"`
}
type shopOffer struct {
	Item    string `toml:"item"`
	Quality int    `toml:"quality"`
	Price   int    `toml:"price"`
}
type golemDef struct {
	ID           string       `toml:"id"`
	Name         string       `toml:"name"`
	Materials    []itemAmount `toml:"materials"`
	SpeedPercent int          `toml:"speed_percent"`
}
type itemAmount struct {
	Item  string `toml:"item"`
	Count int    `toml:"count"`
}
type facilityDef struct {
	ID     string          `toml:"id"`
	Name   string          `toml:"name"`
	Levels []facilityLevel `toml:"levels"`
}
type facilityLevel struct {
	Money     int          `toml:"money"`
	Materials []itemAmount `toml:"materials"`
	Effects   []effect     `toml:"effects"`
}
type effect struct {
	Kind  string `toml:"kind"`
	Value int    `toml:"value"`
}
type source[T any] struct {
	File  string
	Value T
}

// Load は fs.FS 内の TOML をすべて読み、検証済みの目録を返す（DES-PKG-03、DES-VAL-01〜08）。
func Load(fsys fs.FS) (defs.Catalog, error) {
	c := defs.Catalog{Items: map[string]defs.ItemDef{}, Recipes: map[string]defs.RecipeDef{}, Gathers: map[string]defs.GatherDef{}, Golems: map[string]defs.GolemDef{}, Facilities: map[string]defs.FacilityDef{}}
	var errs Errors
	var games []source[gameDef]
	var items []source[itemDef]
	var recipes []source[recipeDef]
	var gathers []source[gatherDef]
	var golems []source[golemDef]
	var facilities []source[facilityDef]
	var names []string
	walkErr := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, Diagnostic{name, "<file>", err.Error()})
			return nil
		}
		if !entry.IsDir() && strings.EqualFold(path.Ext(name), ".toml") {
			names = append(names, name)
		}
		return nil
	})
	if walkErr != nil {
		errs = append(errs, Diagnostic{".", "<file>", walkErr.Error()})
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			errs = append(errs, Diagnostic{name, "<file>", err.Error()})
			continue
		}
		var d document
		meta, err := toml.Decode(string(data), &d)
		if err != nil {
			errs = append(errs, Diagnostic{name, "<toml>", err.Error()})
			continue
		}
		for _, key := range meta.Undecoded() {
			errs = append(errs, Diagnostic{name, key.String(), "未知のキー"})
		}
		if d.Game != nil {
			games = append(games, source[gameDef]{name, *d.Game})
		}
		for _, v := range d.Items {
			items = append(items, source[itemDef]{name, v})
		}
		for _, v := range d.Recipes {
			recipes = append(recipes, source[recipeDef]{name, v})
		}
		for _, v := range d.Gathers {
			gathers = append(gathers, source[gatherDef]{name, v})
		}
		for _, v := range d.Golems {
			golems = append(golems, source[golemDef]{name, v})
		}
		for _, v := range d.Facilities {
			facilities = append(facilities, source[facilityDef]{name, v})
		}
	}
	add := func(file, key, reason string) { errs = append(errs, Diagnostic{file, key, reason}) }
	positive := func(file, key string, n int) {
		if n < 1 {
			add(file, key, "1 以上であること")
		}
	}
	quality := func(file, key string, n int) {
		if n < 1 || n > 200 {
			add(file, key, "1〜200 であること")
		}
	}
	itemRef := func(file, key, id string) {
		if _, ok := c.Items[id]; !ok {
			add(file, key, "存在しないアイテム ID: "+id)
		}
	}
	seen := map[string]map[string]bool{}
	nonEmpty := func(file, key string, n int) {
		if n == 0 {
			add(file, key, "1 個以上であること")
		}
	}
	unique := func(kind, file, id string) {
		if id == "" {
			add(file, kind+"[id=].id", "ID が空")
			return
		}
		if seen[kind] == nil {
			seen[kind] = map[string]bool{}
		}
		if seen[kind][id] {
			add(file, kind+"[id="+id+"].id", "ID が重複")
		}
		seen[kind][id] = true
	}
	attrs := map[string]bool{}
	for _, s := range items {
		v := s.Value
		p := "item[id=" + v.ID + "]"
		unique("item", s.File, v.ID)
		if len(v.Attributes) > 3 {
			add(s.File, p+".attributes", "属性は 3 個以下")
		}
		local := map[string]bool{}
		for i, a := range v.Attributes {
			if local[a] {
				add(s.File, fmt.Sprintf("%s.attributes[%d]", p, i), "属性が重複")
			}
			local[a] = true
			attrs[a] = true
		}
		c.Items[v.ID] = defs.ItemDef{ID: v.ID, Name: v.Name, Attributes: v.Attributes, BasePrice: v.BasePrice}
	}
	attributeRef := func(file, key, attr string) {
		if !attrs[attr] {
			add(file, key, "存在しない属性: "+attr)
		}
	}
	for _, s := range games {
		v := s.Value
		if len(games) > 1 {
			unique("game", s.File, "game")
		}
		c.Game = defs.GameDef{StartMoney: v.StartMoney, BaseGolemSlots: v.BaseGolemSlots}
		for i, o := range v.Shop {
			p := fmt.Sprintf("game.shop[%d]", i)
			itemRef(s.File, p+".item", o.Item)
			quality(s.File, p+".quality", o.Quality)
			c.Game.Shop = append(c.Game.Shop, defs.ShopOffer{Item: o.Item, Quality: o.Quality, Price: o.Price})
		}
	}
	for _, s := range recipes {
		v := s.Value
		p := "recipe[id=" + v.ID + "]"
		unique("recipe", s.File, v.ID)
		itemRef(s.File, p+".output", v.Output)
		positive(s.File, p+".output_count", v.OutputCount)
		positive(s.File, p+".seconds", v.Seconds)
		nonEmpty(s.File, p+".inputs", len(v.Inputs))
		r := defs.RecipeDef{ID: v.ID, Output: v.Output, OutputCount: v.OutputCount, Seconds: v.Seconds}
		for i, in := range v.Inputs {
			q := fmt.Sprintf("%s.inputs[%d]", p, i)
			if (in.Item == "") == (in.Attribute == "") {
				add(s.File, q, "item と attribute のどちらか一方を指定")
			}
			if in.Item != "" {
				itemRef(s.File, q+".item", in.Item)
			}
			if in.Attribute != "" {
				attributeRef(s.File, q+".attribute", in.Attribute)
			}
			positive(s.File, q+".count", in.Count)
			r.Inputs = append(r.Inputs, defs.InputSlot{Item: in.Item, Attribute: in.Attribute, Count: in.Count})
		}
		for i, b := range v.Bonuses {
			q := fmt.Sprintf("%s.bonuses[%d]", p, i)
			attributeRef(s.File, q+".attribute", b.Attribute)
			positive(s.File, q+".min_count", b.MinCount)
			r.Bonuses = append(r.Bonuses, defs.Bonus{Attribute: b.Attribute, MinCount: b.MinCount, Add: b.Add})
		}
		c.Recipes[v.ID] = r
	}
	for _, s := range gathers {
		v := s.Value
		p := "gather[id=" + v.ID + "]"
		unique("gather", s.File, v.ID)
		positive(s.File, p+".rolls", v.Rolls)
		positive(s.File, p+".cooldown_seconds", v.CooldownSeconds)
		nonEmpty(s.File, p+".drops", len(v.Drops))
		g := defs.GatherDef{ID: v.ID, Name: v.Name, Rolls: v.Rolls, CooldownSeconds: v.CooldownSeconds}
		for i, d := range v.Drops {
			q := fmt.Sprintf("%s.drops[%d]", p, i)
			itemRef(s.File, q+".item", d.Item)
			positive(s.File, q+".weight", d.Weight)
			quality(s.File, q+".quality_min", d.QualityMin)
			quality(s.File, q+".quality_max", d.QualityMax)
			if d.QualityMin > d.QualityMax {
				add(s.File, q, "quality_min は quality_max 以下")
			}
			g.Drops = append(g.Drops, defs.Drop{Item: d.Item, Weight: d.Weight, QualityMin: d.QualityMin, QualityMax: d.QualityMax})
		}
		c.Gathers[v.ID] = g
	}
	convertMaterials := func(file, p string, values []itemAmount) []defs.ItemAmount {
		var out []defs.ItemAmount
		for i, m := range values {
			q := fmt.Sprintf("%s[%d]", p, i)
			itemRef(file, q+".item", m.Item)
			positive(file, q+".count", m.Count)
			out = append(out, defs.ItemAmount{Item: m.Item, Count: m.Count})
		}
		return out
	}
	for _, s := range golems {
		v := s.Value
		p := "golem[id=" + v.ID + "]"
		unique("golem", s.File, v.ID)
		// 所要 tick の計算で speed_percent で割るため、0 以下を弾く（DES-GOL-02、NFR-08）。
		positive(s.File, p+".speed_percent", v.SpeedPercent)
		nonEmpty(s.File, p+".materials", len(v.Materials))
		c.Golems[v.ID] = defs.GolemDef{ID: v.ID, Name: v.Name, Materials: convertMaterials(s.File, p+".materials", v.Materials), SpeedPercent: v.SpeedPercent}
	}
	for _, s := range facilities {
		v := s.Value
		p := "facility[id=" + v.ID + "]"
		unique("facility", s.File, v.ID)
		nonEmpty(s.File, p+".levels", len(v.Levels))
		f := defs.FacilityDef{ID: v.ID, Name: v.Name}
		for i, level := range v.Levels {
			q := fmt.Sprintf("%s.levels[%d]", p, i)
			l := defs.FacilityLevel{Money: level.Money, Materials: convertMaterials(s.File, q+".materials", level.Materials)}
			for j, e := range level.Effects {
				r := fmt.Sprintf("%s.effects[%d]", q, j)
				if e.Kind != "golem_slots" && e.Kind != "sell_rate" {
					add(s.File, r+".kind", "未知の効果: "+e.Kind)
				}
				l.Effects = append(l.Effects, defs.Effect{Kind: e.Kind, Value: e.Value})
			}
			f.Levels = append(f.Levels, l)
		}
		c.Facilities[v.ID] = f
	}
	if len(errs) > 0 {
		return defs.Catalog{}, errs
	}
	return c, nil
}
