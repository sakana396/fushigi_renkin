// Package defs は定義データの型と ID で引ける目録を提供する（DES-PKG-02）。
package defs

// ItemDef はアイテムの種類を定義する（DES-IT01 2.1、DES-DEF-02）。
type ItemDef struct {
	ID         string
	Name       string
	Attributes []string
	BasePrice  int
	Use        *UseDef
}

// UseDef は戦闘での使用条件と品質段階を定義する（DES-DEF-05）。
type UseDef struct {
	Target string
	Range  int
	Stages []UseStage
}

// UseStage は品質に応じて加わる効果を定義する（DES-DEF-05）。
type UseStage struct {
	MinQuality int
	Radius     int
	Effects    []UseEffect
}

// UseEffect は戦闘での効果の種類と値を表す（DES-DEF-05）。
type UseEffect struct {
	Kind  string
	Value int
}

// RecipeDef は錬金の入出力、手作りボーナス、解放ランクを定義する（DES-IT01 2.1、DES-DEF-08）。
type RecipeDef struct {
	ID          string
	Rank        int
	Output      string
	OutputCount int
	Inputs      []InputSlot
	Seconds     int
	Bonuses     []Bonus
}

// InputSlot はアイテムまたは属性で指定する入力枠を表す（DES-IT01 2.1、DES-DEF-01）。
type InputSlot struct {
	Item      string
	Attribute string
	Count     int
}

// Bonus は属性の個数による品質加算を定義する（DES-IT01 2.1）。
type Bonus struct {
	Attribute string
	MinCount  int
	Add       int
}

// GatherDef は採取地と解放ランクを定義する（DES-IT01 2.1、DES-DEF-08）。
type GatherDef struct {
	// 拾う場所が戻る秒数（DES-DEF-10、DES-VAL-28）。
	RespawnSeconds  int
	ID              string
	Rank            int
	Name            string
	Rolls           int
	CooldownSeconds int
	Drops           []Drop
	Battle          *BattleDef
}

// UnitDef は仲間と敵に共通する能力を定義する（DES-DEF-06）。
type UnitDef struct {
	ID     string
	Name   string
	Role   string
	HP     int
	Attack int
	Speed  int
	Move   int
	Range  int
}

// BattleDef は採取地で始められる戦闘と報酬を定義する（DES-DEF-06）。
type BattleDef struct {
	Enemies []Placement
	Rolls   int
	Drops   []Drop
}

// Placement はユニットの初期位置を定義する（DES-DEF-06）。
type Placement struct {
	Unit string
	X    int
	Y    int
}

// QuestDef は製品の納品条件、報酬、解放ランクを定義する（DES-DEF-07、DES-DEF-08）。
type QuestDef struct {
	ID         string
	Rank       int
	Reputation int
	Item       string
	MinQuality int
	Count      int
	Reward     int
}

// Drop は採取で得られるアイテムの候補を定義する（DES-IT01 2.1）。
type Drop struct {
	Item       string
	Weight     int
	QualityMin int
	QualityMax int
}

// ShopOffer は店で購入できるアイテムを定義する（DES-IT01 2.1）。
type ShopOffer struct {
	Item    string
	Quality int
	Price   int
}

// ItemAmount はアイテムの種類と個数を表す。品質は問わない（DES-IT01 2.1、DES-GOL-06）。
type ItemAmount struct {
	Item  string
	Count int
}

// GolemDef はゴーレムの建設素材と速度を定義する（DES-IT01 2.1）。
type GolemDef struct {
	ID           string
	Name         string
	Materials    []ItemAmount
	SpeedPercent int
}

// FacilityDef は施設とその強化段階を定義する（DES-IT01 2.1）。
type FacilityDef struct {
	ID     string
	Name   string
	Levels []FacilityLevel
}

// FacilityLevel は施設の強化に必要なものと効果を定義する（DES-IT01 2.1）。
type FacilityLevel struct {
	// 段階ごとの区画の見た目（DES-DEF-12）。
	Look      string
	Money     int
	Materials []ItemAmount
	Effects   []Effect
}

// Effect は施設の効果の種類と値を表す（DES-IT01 2.1、DES-DEF-03）。
type Effect struct {
	Kind  string
	Value int
}

// GameDef はゲーム開始時の設定、店の品ぞろえ、ランクの閾値を定義する（DES-IT01 2.1、DES-DEF-08）。
type GameDef struct {
	// 1 日の秒数。0 は日替わりなし（DES-DAY-01、DES-VAL-29）。
	DaySeconds     int
	StartMoney     int
	RankReputation []int
	BaseGolemSlots int
	Shop           []ShopOffer
	Party          []Placement
	BattleWidth    int
	BattleHeight   int
	InjurySeconds  int
}

// Catalog は ID をキーにした定義の目録を保持する（DES-PKG-02、DES-IT01 2.1）。
type Catalog struct {
	Game       GameDef
	Items      map[string]ItemDef
	Recipes    map[string]RecipeDef
	Gathers    map[string]GatherDef
	Golems     map[string]GolemDef
	Facilities map[string]FacilityDef
	Units      map[string]UnitDef
	Quests     map[string]QuestDef
	Maps       map[string]MapDef
}

// Unit は ID に対応するユニット定義と、見つかったかどうかを返す（DES-DEF-04、DES-DEF-06）。
func (c Catalog) Unit(id string) (UnitDef, bool) {
	d, ok := c.Units[id]
	return d, ok
}

// Quest は ID に対応する依頼定義と、見つかったかどうかを返す（DES-DEF-04、DES-DEF-07）。
func (c Catalog) Quest(id string) (QuestDef, bool) {
	d, ok := c.Quests[id]
	return d, ok
}

// Item は ID に対応するアイテム定義と、見つかったかどうかを返す（DES-PKG-02、DES-DEF-04）。
func (c Catalog) Item(id string) (ItemDef, bool) {
	d, ok := c.Items[id]
	return d, ok
}

// Recipe は ID に対応するレシピ定義と、見つかったかどうかを返す（DES-PKG-02、DES-DEF-04）。
func (c Catalog) Recipe(id string) (RecipeDef, bool) {
	d, ok := c.Recipes[id]
	return d, ok
}

// Gather は ID に対応する採取地定義と、見つかったかどうかを返す（DES-PKG-02、DES-DEF-04）。
func (c Catalog) Gather(id string) (GatherDef, bool) {
	d, ok := c.Gathers[id]
	return d, ok
}

// Golem は ID に対応するゴーレム定義と、見つかったかどうかを返す（DES-PKG-02、DES-DEF-04）。
func (c Catalog) Golem(id string) (GolemDef, bool) {
	d, ok := c.Golems[id]
	return d, ok
}

// Facility は ID に対応する施設定義と、見つかったかどうかを返す（DES-PKG-02、DES-DEF-04）。
func (c Catalog) Facility(id string) (FacilityDef, bool) {
	d, ok := c.Facilities[id]
	return d, ok
}
