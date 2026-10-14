package core

import "github.com/sakana396/fushigi_renkin/internal/defs"

// UseEffectResult は品質を反映した効果と、狙ったマスからの範囲を表す（DES-QLT-05）。
type UseEffectResult struct {
	Radius  int
	Effects []defs.UseEffect
}

// UseEffectOf は品質以下の段階の効果を合算し、品質に応じた値を返す（DES-QLT-05）。
// Use がないアイテムでは第 2 戻り値が false になる。
func UseEffectOf(item defs.ItemDef, quality int) (UseEffectResult, bool) {
	if item.Use == nil {
		return UseEffectResult{}, false
	}

	result := UseEffectResult{}
	indices := make(map[string]int)
	for _, stage := range item.Use.Stages {
		if stage.MinQuality > quality {
			continue
		}
		if stage.Radius > result.Radius {
			result.Radius = stage.Radius
		}
		for _, effect := range stage.Effects {
			if i, ok := indices[effect.Kind]; ok {
				result.Effects[i].Value += effect.Value
			} else {
				indices[effect.Kind] = len(result.Effects)
				result.Effects = append(result.Effects, effect)
			}
		}
	}
	for i := range result.Effects {
		value := result.Effects[i].Value * quality / 100
		if value < 1 {
			value = 1
		}
		result.Effects[i].Value = value
	}
	return result, true
}
