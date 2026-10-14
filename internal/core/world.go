package core

import (
	"math/rand/v2"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// World はゲームの状態とシード付き乱数を保持する（DES-IT01 3.1、DES-TIME-01）。
type World struct {
	catalog      defs.Catalog
	rng          *rand.Rand
	rngSource    *rand.PCG
	Inventory    Inventory
	Battle       *Battle
	LastBattle   *BattleReport
	InjuredUntil uint64
	Golems       []Golem
	Money        int
	Now          uint64
	Facilities   map[string]int
	gatherReady  map[string]uint64
}

// NewWorld は目録とシードから初期状態を作る（DES-IT01 3.1、DES-TIME-01）。
func NewWorld(catalog defs.Catalog, seed uint64) *World {
	source := rand.NewPCG(seed, 0)
	return &World{catalog: catalog, rng: rand.New(source), rngSource: source, Money: catalog.Game.StartMoney}
}

// Tick はゲーム内時刻を 1 tick 進める（DES-TIME-01）。
func (w *World) Tick() {
	w.Now++
	for i := range w.Golems {
		w.stepGolem(i)
	}
}
