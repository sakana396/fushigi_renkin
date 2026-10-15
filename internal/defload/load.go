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
	Units      []unitDef     `toml:"unit"`
	Quests     []questDef    `toml:"quest"`
}
type gameDef struct {
	BattleConfigured bool        `toml:"-"`
	RankReputation   []int       `toml:"rank_reputation"`
	StartMoney       int         `toml:"start_money"`
	BaseGolemSlots   int         `toml:"base_golem_slots"`
	Shop             []shopOffer `toml:"shop"`
	Party            []placement `toml:"party"`
	BattleWidth      int         `toml:"battle_width"`
	BattleHeight     int         `toml:"battle_height"`
	InjurySeconds    int         `toml:"injury_seconds"`
}
type itemDef struct {
	ID         string   `toml:"id"`
	Name       string   `toml:"name"`
	Attributes []string `toml:"attributes"`
	BasePrice  int      `toml:"base_price"`
	Use        *useDef  `toml:"use"`
}
type useDef struct {
	Target string     `toml:"target"`
	Range  int        `toml:"range"`
	Stages []useStage `toml:"stages"`
}
type useStage struct {
	MinQuality int         `toml:"min_quality"`
	Radius     int         `toml:"radius"`
	Effects    []useEffect `toml:"effects"`
}
type useEffect struct {
	Kind  string `toml:"kind"`
	Value int    `toml:"value"`
}
type recipeDef struct {
	ID          string      `toml:"id"`
	Rank        *int        `toml:"rank"`
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
	ID              string     `toml:"id"`
	Rank            *int       `toml:"rank"`
	Name            string     `toml:"name"`
	Rolls           int        `toml:"rolls"`
	CooldownSeconds int        `toml:"cooldown_seconds"`
	Drops           []drop     `toml:"drops"`
	Battle          *battleDef `toml:"battle"`
}
type unitDef struct {
	ID     string `toml:"id"`
	Name   string `toml:"name"`
	Role   string `toml:"role"`
	HP     int    `toml:"hp"`
	Attack int    `toml:"attack"`
	Speed  int    `toml:"speed"`
	Move   int    `toml:"move"`
	Range  int    `toml:"range"`
}
type battleDef struct {
	Enemies []placement `toml:"enemies"`
	Rolls   int         `toml:"rolls"`
	Drops   []drop      `toml:"drops"`
}
type placement struct {
	Unit string `toml:"unit"`
	X    int    `toml:"x"`
	Y    int    `toml:"y"`
}
type questDef struct {
	ID         string `toml:"id"`
	Rank       *int   `toml:"rank"`
	Reputation int    `toml:"reputation"`
	Item       string `toml:"item"`
	MinQuality int    `toml:"min_quality"`
	Count      int    `toml:"count"`
	Reward     int    `toml:"reward"`
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

// Load は fs.FS 内の TOML をすべて読み、検証済みの目録を返す（DES-PKG-03、DES-VAL-01〜16）。
func Load(fsys fs.FS) (defs.Catalog, error) {
	c := defs.Catalog{Items: map[string]defs.ItemDef{}, Recipes: map[string]defs.RecipeDef{}, Gathers: map[string]defs.GatherDef{}, Golems: map[string]defs.GolemDef{}, Facilities: map[string]defs.FacilityDef{}, Units: map[string]defs.UnitDef{}, Quests: map[string]defs.QuestDef{}}
	var errs Errors
	var games []source[gameDef]
	var items []source[itemDef]
	var recipes []source[recipeDef]
	var gathers []source[gatherDef]
	var golems []source[golemDef]
	var facilities []source[facilityDef]
	var units []source[unitDef]
	var quests []source[questDef]
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
			d.Game.BattleConfigured = meta.IsDefined("game", "party") || meta.IsDefined("game", "battle_width") || meta.IsDefined("game", "battle_height") || meta.IsDefined("game", "injury_seconds")
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
		for _, v := range d.Units {
			units = append(units, source[unitDef]{name, v})
		}
		for _, v := range d.Quests {
			quests = append(quests, source[questDef]{name, v})
		}
	}
	add := func(file, key, reason string) { errs = append(errs, Diagnostic{file, key, reason}) }
	positive := func(file, key string, n int) {
		if n < 1 {
			add(file, key, "1 以上であること")
		}
	}
	// 省略された rank は 1 とし、明示された値だけを範囲検査する（DES-DEF-08、DES-VAL-14）。
	rank := func(file, key string, value *int, max int) int {
		if value == nil {
			return 1
		}
		if *value < 1 || *value > max {
			add(file, key, fmt.Sprintf("1〜%d であること", max))
		}
		return *value
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
		item := defs.ItemDef{ID: v.ID, Name: v.Name, Attributes: v.Attributes, BasePrice: v.BasePrice}
		if v.Use != nil {
			u := v.Use
			q := p + ".use"
			if u.Target != "ally" && u.Target != "enemy" {
				add(s.File, q+".target", "ally または enemy であること")
			}
			if u.Range < 0 || (u.Target == "enemy" && u.Range < 1) {
				add(s.File, q+".range", "ally は 0 以上、enemy は 1 以上であること")
			}
			use := &defs.UseDef{Target: u.Target, Range: u.Range}
			stageSeen := map[int]bool{}
			baseFound := false
			for i, st := range u.Stages {
				r := fmt.Sprintf("%s.stages[%d]", q, i)
				if st.MinQuality != 1 && st.MinQuality != 50 && st.MinQuality != 100 && st.MinQuality != 150 && st.MinQuality != 200 {
					add(s.File, r+".min_quality", "1・50・100・150・200 のいずれかであること")
				}
				if stageSeen[st.MinQuality] {
					add(s.File, r+".min_quality", "品質段階が重複")
				}
				stageSeen[st.MinQuality] = true
				if st.MinQuality == 1 {
					baseFound = true
					if len(st.Effects) == 0 {
						add(s.File, r+".effects", "基本の段階に効果が 1 個以上必要")
					}
				}
				if st.Radius < 0 {
					add(s.File, r+".radius", "0 以上であること")
				}
				stage := defs.UseStage{MinQuality: st.MinQuality, Radius: st.Radius}
				for j, e := range st.Effects {
					k := fmt.Sprintf("%s.effects[%d]", r, j)
					if e.Kind != "heal" && e.Kind != "damage" && e.Kind != "attack_up" && e.Kind != "speed_up" {
						add(s.File, k+".kind", "未知の効果: "+e.Kind)
					}
					positive(s.File, k+".value", e.Value)
					stage.Effects = append(stage.Effects, defs.UseEffect{Kind: e.Kind, Value: e.Value})
				}
				use.Stages = append(use.Stages, stage)
			}
			if !baseFound {
				add(s.File, q+".stages", "品質 1 の基本の段階が必要")
			}
			item.Use = use
		}
		c.Items[v.ID] = item
	}
	attributeRef := func(file, key, attr string) {
		if !attrs[attr] {
			add(file, key, "存在しない属性: "+attr)
		}
	}
	for _, s := range units {
		v := s.Value
		p := "unit[id=" + v.ID + "]"
		unique("unit", s.File, v.ID)
		if v.Role != "front" && v.Role != "back" && v.Role != "enemy" {
			add(s.File, p+".role", "front・back・enemy のいずれかであること")
		}
		for _, field := range []struct {
			name  string
			value int
		}{{"hp", v.HP}, {"attack", v.Attack}, {"speed", v.Speed}, {"move", v.Move}, {"range", v.Range}} {
			positive(s.File, p+"."+field.name, field.value)
		}
		c.Units[v.ID] = defs.UnitDef{ID: v.ID, Name: v.Name, Role: v.Role, HP: v.HP, Attack: v.Attack, Speed: v.Speed, Move: v.Move, Range: v.Range}
	}
	// 配置は仲間と敵の双方で同じ盤面を使う（DES-VAL-12）。
	placements := func(file, p string, values []placement, wantRole string, width, height int, occupied map[[2]int]bool) []defs.Placement {
		out := make([]defs.Placement, 0, len(values))
		for i, v := range values {
			q := fmt.Sprintf("%s[%d]", p, i)
			u, ok := c.Units[v.Unit]
			if !ok {
				add(file, q+".unit", "存在しないユニット ID: "+v.Unit)
			} else if wantRole == "enemy" && u.Role != "enemy" || wantRole == "party" && u.Role != "front" && u.Role != "back" {
				add(file, q+".unit", "配置に使えない役割: "+u.Role)
			}
			if v.X < 0 || v.X >= width {
				add(file, q+".x", "盤面の範囲外")
			}
			if v.Y < 0 || v.Y >= height {
				add(file, q+".y", "盤面の範囲外")
			}
			xy := [2]int{v.X, v.Y}
			if occupied[xy] {
				add(file, q, "初期位置が重複")
			}
			occupied[xy] = true
			out = append(out, defs.Placement{Unit: v.Unit, X: v.X, Y: v.Y})
		}
		return out
	}
	// 反復 1・2 の定義には戦闘設定がないため、戦闘の項目が現れたとき検証する。
	battleConfigured := false
	for _, s := range games {
		if s.Value.BattleConfigured {
			battleConfigured = true
		}
	}
	for _, s := range gathers {
		if s.Value.Battle != nil {
			battleConfigured = true
		}
	}
	for _, s := range games {
		v := s.Value
		if len(games) > 1 {
			unique("game", s.File, "game")
		}
		c.Game = defs.GameDef{StartMoney: v.StartMoney, RankReputation: v.RankReputation, BaseGolemSlots: v.BaseGolemSlots, BattleWidth: v.BattleWidth, BattleHeight: v.BattleHeight, InjurySeconds: v.InjurySeconds}
		for i, n := range v.RankReputation {
			if n < 1 || i > 0 && n <= v.RankReputation[i-1] {
				add(s.File, fmt.Sprintf("game.rank_reputation[%d]", i), "1 以上で狭義に増えること")
			}
		}
		if battleConfigured {
			positive(s.File, "game.battle_width", v.BattleWidth)
			positive(s.File, "game.battle_height", v.BattleHeight)
			positive(s.File, "game.injury_seconds", v.InjurySeconds)
			nonEmpty(s.File, "game.party", len(v.Party))
			c.Game.Party = placements(s.File, "game.party", v.Party, "party", v.BattleWidth, v.BattleHeight, map[[2]int]bool{})
		}
		for i, o := range v.Shop {
			p := fmt.Sprintf("game.shop[%d]", i)
			itemRef(s.File, p+".item", o.Item)
			quality(s.File, p+".quality", o.Quality)
			c.Game.Shop = append(c.Game.Shop, defs.ShopOffer{Item: o.Item, Quality: o.Quality, Price: o.Price})
		}
	}
	if battleConfigured && len(games) == 0 {
		for _, s := range gathers {
			if s.Value.Battle != nil {
				add(s.File, "gather[id="+s.Value.ID+"].battle", "戦闘には game の設定が必要")
				break
			}
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
		r := defs.RecipeDef{ID: v.ID, Rank: rank(s.File, p+".rank", v.Rank, len(c.Game.RankReputation)+1), Output: v.Output, OutputCount: v.OutputCount, Seconds: v.Seconds}
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
	for _, s := range quests {
		v := s.Value
		p := "quest[id=" + v.ID + "]"
		unique("quest", s.File, v.ID)
		product := false
		for _, r := range c.Recipes {
			if r.Output == v.Item {
				product = true
				break
			}
		}
		if !product {
			add(s.File, p+".item", "レシピの出力ではないアイテム ID: "+v.Item)
		}
		quality(s.File, p+".min_quality", v.MinQuality)
		positive(s.File, p+".count", v.Count)
		positive(s.File, p+".reward", v.Reward)
		if v.Reputation < 0 {
			add(s.File, p+".reputation", "0 以上であること")
		}
		c.Quests[v.ID] = defs.QuestDef{ID: v.ID, Rank: rank(s.File, p+".rank", v.Rank, len(c.Game.RankReputation)+1), Reputation: v.Reputation, Item: v.Item, MinQuality: v.MinQuality, Count: v.Count, Reward: v.Reward}
	}
	validateDrops := func(file, p string, values []drop) []defs.Drop {
		out := make([]defs.Drop, 0, len(values))
		for i, d := range values {
			q := fmt.Sprintf("%s[%d]", p, i)
			itemRef(file, q+".item", d.Item)
			positive(file, q+".weight", d.Weight)
			quality(file, q+".quality_min", d.QualityMin)
			quality(file, q+".quality_max", d.QualityMax)
			if d.QualityMin > d.QualityMax {
				add(file, q, "quality_min は quality_max 以下")
			}
			out = append(out, defs.Drop{Item: d.Item, Weight: d.Weight, QualityMin: d.QualityMin, QualityMax: d.QualityMax})
		}
		return out
	}
	for _, s := range gathers {
		v := s.Value
		p := "gather[id=" + v.ID + "]"
		unique("gather", s.File, v.ID)
		positive(s.File, p+".rolls", v.Rolls)
		positive(s.File, p+".cooldown_seconds", v.CooldownSeconds)
		nonEmpty(s.File, p+".drops", len(v.Drops))
		g := defs.GatherDef{ID: v.ID, Rank: rank(s.File, p+".rank", v.Rank, len(c.Game.RankReputation)+1), Name: v.Name, Rolls: v.Rolls, CooldownSeconds: v.CooldownSeconds}
		g.Drops = validateDrops(s.File, p+".drops", v.Drops)
		if v.Battle != nil {
			b := v.Battle
			q := p + ".battle"
			nonEmpty(s.File, q+".enemies", len(b.Enemies))
			positive(s.File, q+".rolls", b.Rolls)
			nonEmpty(s.File, q+".drops", len(b.Drops))
			occupied := map[[2]int]bool{}
			for _, member := range c.Game.Party {
				occupied[[2]int{member.X, member.Y}] = true
			}
			g.Battle = &defs.BattleDef{Enemies: placements(s.File, q+".enemies", b.Enemies, "enemy", c.Game.BattleWidth, c.Game.BattleHeight, occupied), Rolls: b.Rolls, Drops: validateDrops(s.File, q+".drops", b.Drops)}
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
	validateProgression(c, recipes, quests, add)
	if len(errs) > 0 {
		return defs.Catalog{}, errs
	}
	return c, nil
}

// validateProgression はランクごとに入手可能な物を閉包として求め、進行不能な定義を検出する（DES-VAL-15、DES-VAL-16）。
func validateProgression(c defs.Catalog, recipes []source[recipeDef], quests []source[questDef], add func(string, string, string)) {
	if len(c.Game.RankReputation) == 0 {
		return
	}
	rankOne := 0
	positiveQuest := make([]bool, len(c.Game.RankReputation)+1)
	for _, q := range c.Quests {
		if q.Rank == 1 {
			rankOne++
		}
		if q.Rank >= 1 && q.Rank < len(positiveQuest) && q.Reputation >= 1 {
			positiveQuest[q.Rank] = true
		}
	}
	if rankOne < 3 {
		add(questFile(quests), "quest", "ランク 1 の依頼は 3 件以上必要")
	}
	for r := 1; r <= len(c.Game.RankReputation); r++ {
		if !positiveQuest[r] {
			add(questFile(quests), "quest", fmt.Sprintf("ランク %d に評判を得られる依頼がありません", r))
		}
	}
	for r := 1; r <= len(c.Game.RankReputation)+1; r++ {
		available := map[string]bool{}
		for _, offer := range c.Game.Shop {
			available[offer.Item] = true
		}
		for _, g := range c.Gathers {
			if g.Rank > r {
				continue
			}
			for _, d := range g.Drops {
				available[d.Item] = true
			}
			if g.Battle != nil {
				for _, d := range g.Battle.Drops {
					available[d.Item] = true
				}
			}
		}
		for changed := true; changed; {
			changed = false
			for _, recipe := range c.Recipes {
				if recipe.Rank > r || available[recipe.Output] || !inputsAvailable(recipe.Inputs, available, c.Items) {
					continue
				}
				available[recipe.Output] = true
				changed = true
			}
		}
		for _, s := range recipes {
			v := c.Recipes[s.Value.ID]
			if v.Rank == r && !available[v.Output] {
				add(s.File, "recipe[id="+v.ID+"].inputs", fmt.Sprintf("ランク %d で材料が手に入りません", r))
			}
		}
		for _, s := range quests {
			v := c.Quests[s.Value.ID]
			if v.Rank == r && !available[v.Item] {
				add(s.File, "quest[id="+v.ID+"].item", fmt.Sprintf("ランク %d で手に入りません", r))
			}
		}
	}
}

// inputsAvailable は各入力枠に合う物が集合にあるかを調べる（DES-VAL-16）。
func inputsAvailable(inputs []defs.InputSlot, available map[string]bool, items map[string]defs.ItemDef) bool {
	for _, input := range inputs {
		if input.Item != "" {
			if !available[input.Item] {
				return false
			}
			continue
		}
		found := false
		for id := range available {
			for _, attr := range items[id].Attributes {
				if attr == input.Attribute {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// questFile は依頼の集合に関する誤りのファイル名を返す（DES-PKG-03、DES-VAL-15）。
func questFile(quests []source[questDef]) string {
	if len(quests) == 0 {
		return "<definitions>"
	}
	return quests[0].File
}
