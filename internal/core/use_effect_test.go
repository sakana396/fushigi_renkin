package core

import (
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

func sameUseEffects(got, want []defs.UseEffect) bool {
	if len(got) != len(want) {
		return false
	}
	values := make(map[string]int, len(want))
	for _, effect := range want {
		values[effect.Kind] = effect.Value
	}
	for _, effect := range got {
		if value, ok := values[effect.Kind]; !ok || value != effect.Value {
			return false
		}
		delete(values, effect.Kind)
	}
	return len(values) == 0
}

func testBombUse() defs.ItemDef {
	return defs.ItemDef{ID: "bomb", Use: &defs.UseDef{Target: "enemy", Range: 3, Stages: []defs.UseStage{
		{MinQuality: 1, Radius: 0, Effects: []defs.UseEffect{{Kind: "damage", Value: 40}}},
		{MinQuality: 50, Radius: 1},
		{MinQuality: 100, Effects: []defs.UseEffect{{Kind: "damage", Value: 20}}},
	}}}
}

func testBurgerUse() defs.ItemDef {
	return defs.ItemDef{ID: "burger", Use: &defs.UseDef{Target: "ally", Range: 1, Stages: []defs.UseStage{
		{MinQuality: 1, Effects: []defs.UseEffect{{Kind: "heal", Value: 30}}},
		{MinQuality: 100, Effects: []defs.UseEffect{{Kind: "attack_up", Value: 5}}},
		{MinQuality: 150, Effects: []defs.UseEffect{{Kind: "speed_up", Value: 3}}},
	}}}
}

func TestUseEffectOfPower(t *testing.T) {
	bomb, burger := testBombUse(), testBurgerUse()
	twoEffects := defs.ItemDef{Use: &defs.UseDef{Stages: []defs.UseStage{{MinQuality: 1, Effects: []defs.UseEffect{{Kind: "damage", Value: 1}, {Kind: "damage", Value: 1}}}}}}
	tests := []struct {
		name    string
		item    defs.ItemDef
		quality int
		want    []defs.UseEffect
	}{
		{"TC-FR-ITM-08-01 爆弾の品質20はダメージ8", bomb, 20, []defs.UseEffect{{Kind: "damage", Value: 8}}},
		{"TC-FR-ITM-08-02 爆弾の品質60はダメージ24", bomb, 60, []defs.UseEffect{{Kind: "damage", Value: 24}}},
		{"TC-FR-ITM-08-03 爆弾の品質120は合計してダメージ72", bomb, 120, []defs.UseEffect{{Kind: "damage", Value: 72}}},
		{"TC-FR-ITM-08-04 ハンバーガーの品質30は回復9", burger, 30, []defs.UseEffect{{Kind: "heal", Value: 9}}},
		{"TC-FR-ITM-08-05 ハンバーガーの品質100は回復30と攻撃5", burger, 100, []defs.UseEffect{{Kind: "heal", Value: 30}, {Kind: "attack_up", Value: 5}}},
		{"TC-FR-ITM-08-06 ハンバーガーの品質160は回復48と攻撃8と素早さ4", burger, 160, []defs.UseEffect{{Kind: "heal", Value: 48}, {Kind: "attack_up", Value: 8}, {Kind: "speed_up", Value: 4}}},
		{"TC-FR-ITM-08-07 品質1でも効果の最小値は1", burger, 1, []defs.UseEffect{{Kind: "heal", Value: 1}}},
		{"TC-FR-ITM-08-08 品質49は端数を切り捨てる", bomb, 49, []defs.UseEffect{{Kind: "damage", Value: 19}}},
		{"TC-FR-ITM-08-09 品質200は合計値を2倍する", bomb, 200, []defs.UseEffect{{Kind: "damage", Value: 120}}},
		{"TC-FR-ITM-08-10 同種の効果は合計後に最小値を適用する", twoEffects, 1, []defs.UseEffect{{Kind: "damage", Value: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := UseEffectOf(tt.item, tt.quality)
			if !ok || !sameUseEffects(got.Effects, tt.want) {
				t.Fatalf("UseEffectOf = (%+v, %t), want effects %v", got, ok, tt.want)
			}
		})
	}
}

func TestUseEffectOfStages(t *testing.T) {
	bomb, burger := testBombUse(), testBurgerUse()
	unordered := defs.ItemDef{Use: &defs.UseDef{Stages: []defs.UseStage{
		{MinQuality: 200, Radius: 0, Effects: []defs.UseEffect{{Kind: "heal", Value: 1}}},
		{MinQuality: 50, Radius: 2},
		{MinQuality: 1, Radius: 0, Effects: []defs.UseEffect{{Kind: "damage", Value: 40}}},
		{MinQuality: 150, Radius: 1},
	}}}
	tests := []struct {
		name    string
		item    defs.ItemDef
		quality int
		want    UseEffectResult
		ok      bool
	}{
		{"TC-FR-ITM-09-01 品質20は基本段階の一マス", bomb, 20, UseEffectResult{Effects: []defs.UseEffect{{Kind: "damage", Value: 8}}}, true},
		{"TC-FR-ITM-09-02 品質49では50の段階を選ばない", bomb, 49, UseEffectResult{Effects: []defs.UseEffect{{Kind: "damage", Value: 19}}}, true},
		{"TC-FR-ITM-09-03 品質50で周囲四マスを加える", bomb, 50, UseEffectResult{Radius: 1, Effects: []defs.UseEffect{{Kind: "damage", Value: 20}}}, true},
		{"TC-FR-ITM-09-04 品質120で同種効果を合計する", bomb, 120, UseEffectResult{Radius: 1, Effects: []defs.UseEffect{{Kind: "damage", Value: 72}}}, true},
		{"TC-FR-ITM-09-05 品質160で三種類の効果を選ぶ", burger, 160, UseEffectResult{Effects: []defs.UseEffect{{Kind: "heal", Value: 48}, {Kind: "attack_up", Value: 8}, {Kind: "speed_up", Value: 4}}}, true},
		{"TC-FR-ITM-09-06 品質200は順不同の段階を選び範囲の最大値を使う", unordered, 200, UseEffectResult{Radius: 2, Effects: []defs.UseEffect{{Kind: "heal", Value: 2}, {Kind: "damage", Value: 80}}}, true},
		{"TC-FR-ITM-09-07 Useがなければ使えない", defs.ItemDef{ID: "ore"}, 100, UseEffectResult{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := UseEffectOf(tt.item, tt.quality)
			if ok != tt.ok || got.Radius != tt.want.Radius || !sameUseEffects(got.Effects, tt.want.Effects) {
				t.Fatalf("UseEffectOf = (%+v, %t), want (%+v, %t)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
