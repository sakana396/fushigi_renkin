package core

import (
	"math/rand/v2"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// World はゲームの状態とシード付き乱数、依頼板を保持する（DES-IT01 3.1、DES-TIME-01、DES-QST-04、DES-FLD-01）。
type World struct {
	catalog      defs.Catalog
	Player       Player
	rng          *rand.Rand
	rngSource    *rand.PCG
	questRNG     *rand.Rand
	Inventory    Inventory
	Quests       [3]string
	Accepted     []string
	Battle       *Battle
	LastBattle   *BattleReport
	InjuredUntil uint64
	Golems       []Golem
	Money        int
	Reputation   int
	Now          uint64
	rankUps      []int
	newDays      []int
	Facilities   map[string]int
	gatherReady  map[string]uint64
}

// NewWorld は目録とシードから初期状態と依頼板を作る（DES-IT01 3.1、DES-TIME-01、DES-QST-04、DES-FLD-01）。
func NewWorld(catalog defs.Catalog, seed uint64) *World {
	source := rand.NewPCG(seed, 0)
	w := &World{catalog: catalog, rng: rand.New(source), rngSource: source, questRNG: rand.New(rand.NewPCG(seed, 1)), Money: catalog.Game.StartMoney}
	if m, ok := catalog.Map(defs.MapWorkshop); ok {
		w.Player = Player{Scene: defs.MapWorkshop, X: m.StartX*playerTile + playerTile/2, Y: m.StartY*playerTile + playerTile/2, Facing: defs.DirDown}
	}
	for i := range w.Quests {
		w.Quests[i] = w.drawQuest()
	}
	return w
}

// Tick は時刻とゴーレムを進め、最後に日替わりを行う（DES-TIME-01、DES-DAY-01・02）。
func (w *World) Tick() {
	previousDay := w.Day()
	w.Now++
	for i := range w.Golems {
		w.stepGolem(i)
	}
	if day := w.Day(); day != previousDay {
		w.Quests = [3]string{}
		for i := range w.Quests {
			w.Quests[i] = w.drawQuest()
		}
		w.newDays = append(w.newDays, day)
	}
}
