// Package defs は定義データの型と ID で引ける目録を提供する（DES-PKG-02）。
package defs

// ItemDef はアイテムの種類を定義する（DES-IT01 2.1、DES-DEF-02）。
type ItemDef struct {
	ID         string
	Name       string
	Attributes []string
	BasePrice  int
}

// RecipeDef は錬金の入出力と手作りボーナスを定義する（DES-IT01 2.1）。
type RecipeDef struct {
	ID          string
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

// GatherDef は採取地を定義する（DES-IT01 2.1）。
type GatherDef struct {
	ID              string
	Name            string
	Rolls           int
	CooldownSeconds int
	Drops           []Drop
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
	Money     int
	Materials []ItemAmount
	Effects   []Effect
}

// Effect は施設の効果の種類と値を表す（DES-IT01 2.1、DES-DEF-03）。
type Effect struct {
	Kind  string
	Value int
}

// GameDef はゲーム開始時の設定と店の品ぞろえを定義する（DES-IT01 2.1）。
type GameDef struct {
	StartMoney     int
	BaseGolemSlots int
	Shop           []ShopOffer
}

// Catalog は ID をキーにした定義の目録を保持する（DES-PKG-02、DES-IT01 2.1）。
type Catalog struct {
	Game       GameDef
	Items      map[string]ItemDef
	Recipes    map[string]RecipeDef
	Gathers    map[string]GatherDef
	Golems     map[string]GolemDef
	Facilities map[string]FacilityDef
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
