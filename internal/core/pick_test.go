package core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

// 計算例 F の地図と固定品質の採取表を組み立てる（DES-FLD-05・07、DES-PCK-01・02）。
func pickCatalog() defs.Catalog {
	c := battleCatalog()
	c.Game.DaySeconds = 1
	c.Game.RankReputation = []int{3}
	c.Maps = map[string]defs.MapDef{
		"town":   {ID: "town", StartX: 5, StartY: 5, Rows: []string{"TTTTTTTTT", "T,aaa,,QT", "T,aaa,,,T", "T,aaa,,DT", "T,,,,,,MT", "T,,,,,,,T", "TTTT2TTTT"}, Lots: []defs.LotDef{{Mark: 'a', Facility: "market", X: 2, Y: 1}}, Exits: []defs.ExitDef{{Mark: '2', To: SceneGather, Dir: defs.DirUp, X: 4, Y: 6}}},
		"meadow": {ID: "meadow", Rows: []string{"TTTTTTT", "T*,,,ET", "T,,o,,T", "T*,,,,T", "TTT1TTT"}, Exits: []defs.ExitDef{{Mark: '1', To: "town", Dir: defs.DirUp, X: 3, Y: 4}}},
	}
	s := c.Gathers["field"]
	s.ID, s.Name, s.Rolls, s.RespawnSeconds = "meadow", "草原", 3, 2
	s.Drops = []defs.Drop{{Item: "herb", Weight: 1, QualityMin: 40, QualityMax: 40}}
	c.Gathers["meadow"] = s
	return c
}

func TestPickExamples(t *testing.T) {
	for _, tc := range []struct {
		id  string
		run func(*testing.T, *World)
	}{
		{"TC-FR-FLD-03-80 F-1", func(t *testing.T, w *World) {
			if w.Player != (Player{"town", 528, 528, defs.DirUp}) {
				t.Fatal(w.Player)
			}
		}},
		{"TC-IF-08-70 F-2", func(t *testing.T, w *World) {
			w.Player = Player{"town", 512, 240, defs.DirLeft}
			assertTarget(t, w, Target{TargetLot, "market", 4, 2})
		}},
		{"TC-IF-08-71 F-3", func(t *testing.T, w *World) {
			w.Player = Player{"town", 640, 336, defs.DirRight}
			assertTarget(t, w, Target{TargetFixture, defs.FixtureWorkshopDoor, 7, 3})
		}},
		{"TC-FR-FLD-03-81 F-4", func(t *testing.T, w *World) {
			w.Player.Scene = SceneGather
			if err := w.EnterGather("meadow"); err != nil {
				t.Fatal(err)
			}
			if w.Player != (Player{"meadow", 336, 336, defs.DirUp}) {
				t.Fatal(w.Player)
			}
		}},
		{"TC-FR-FLD-03-82 F-5", func(t *testing.T, w *World) {
			before := w.Player
			if err := w.EnterGather("meadow"); err != ErrNotInGather || w.Player != before {
				t.Fatal(err, w.Player)
			}
		}},
		{"TC-FR-FLD-03-83 F-6", func(t *testing.T, w *World) {
			s := w.catalog.Gathers["meadow"]
			s.Rank = 2
			w.catalog.Gathers["meadow"] = s
			w.Player.Scene = SceneGather
			before := w.Player
			if err := w.EnterGather("meadow"); !errors.Is(err, ErrLocked) || w.Player != before {
				t.Fatal(err, w.Player)
			}
		}},
		{"TC-FR-FLD-02-80 F-7", func(t *testing.T, w *World) {
			w.Player = Player{"meadow", 336, 336, defs.DirUp}
			for i := 0; i < 3; i++ {
				if err := w.MovePlayer(0, -1); err != nil {
					t.Fatal(err)
				}
			}
			if w.Player != (Player{"meadow", 336, 320, defs.DirUp}) {
				t.Fatal(w.Player)
			}
		}},
		{"TC-IF-08-72 F-8", func(t *testing.T, w *World) {
			w.Player = Player{"meadow", 336, 320, defs.DirUp}
			assertTarget(t, w, Target{TargetPick, "", 3, 2})
		}},
		{"TC-FR-MAT-01-50 F-9", func(t *testing.T, w *World) {
			pickExample(t, w)
			if w.pickReady[PickKey{"meadow", 3, 2}] != 120 {
				t.Fatal(w.pickReady)
			}
		}},
		{"TC-FR-MAT-01-51 F-10", func(t *testing.T, w *World) {
			pickExample(t, w)
			w.Now = 50
			before := w.Inventory.Stacks()
			_, err := w.PickFacing()
			if !errors.Is(err, ErrPickNotReady) || err.Error() != "あと 2 秒で拾えます: まだ拾えません" || !reflect.DeepEqual(before, w.Inventory.Stacks()) || w.pickReady[PickKey{"meadow", 3, 2}] != 120 {
				t.Fatal(err, w.pickReady)
			}
		}},
		{"TC-FR-MAT-01-52 F-11", func(t *testing.T, w *World) {
			pickExample(t, w)
			w.Now = 120
			got, err := w.PickFacing()
			if err != nil || !reflect.DeepEqual(got, []Stack{{ItemKey{"herb", 40}, 3}}) || w.Inventory.Count(ItemKey{"herb", 40}) != 6 || w.pickReady[PickKey{"meadow", 3, 2}] != 240 {
				t.Fatal(got, err, w.pickReady)
			}
		}},
		{"TC-FR-MAT-01-53 F-12", func(t *testing.T, w *World) {
			w.Player = Player{"meadow", 336, 336, defs.DirUp}
			if _, err := w.PickFacing(); err != ErrNothingToPick {
				t.Fatal(err)
			}
		}},
		{"TC-IF-08-73 F-13", func(t *testing.T, w *World) {
			w.Player = Player{"meadow", 528, 224, defs.DirUp}
			assertTarget(t, w, Target{TargetEnemy, "", 5, 1})
		}},
		{"TC-FR-BTL-10-50 F-14", func(t *testing.T, w *World) { startFacingExample(t, w) }},
		{"TC-FR-BTL-10-51 F-15", func(t *testing.T, w *World) {
			startFacingExample(t, w)
			for i := range w.Battle.Units {
				if !w.Battle.Units[i].Ally {
					w.Battle.Units[i].HP = 0
				}
			}
			before := w.Player
			if !w.finishBattle() || w.Player != before || w.Battle != nil || w.LastBattle == nil || !w.LastBattle.Won {
				t.Fatal(w.Player, w.LastBattle)
			}
			startFacingExample(t, w)
		}},
		{"TC-FR-MAT-01-54 F-16", func(t *testing.T, w *World) {
			startFacingExample(t, w)
			if _, err := w.PickFacing(); err != ErrInBattle {
				t.Fatal(err)
			}
		}},
		{"TC-FR-FLD-03-84 F-17", func(t *testing.T, w *World) {
			w.Player = Player{"meadow", 336, 336, defs.DirDown}
			for i := 0; i < 7; i++ {
				if err := w.MovePlayer(0, 1); err != nil {
					t.Fatal(err)
				}
				want := Player{"meadow", 336, 343 + 7*i, defs.DirDown}
				if i == 6 {
					want = Player{"town", 432, 528, defs.DirUp}
				}
				if w.Player != want {
					t.Fatal(i, w.Player, want)
				}
			}
		}},
		{"TC-FR-FLD-05-01 F-18", func(t *testing.T, w *World) {
			pickExample(t, w)
			w.Player = Player{"meadow", 336, 336, defs.DirDown}
			for i := 0; i < 7; i++ {
				w.MovePlayer(0, 1)
			}
			for i := 0; i < 7; i++ {
				if err := w.MovePlayer(0, 1); err != nil {
					t.Fatal(err)
				}
			}
			if w.Player.Scene != SceneGather {
				t.Fatal(w.Player)
			}
			if err := w.EnterGather("meadow"); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				w.MovePlayer(0, -1)
			}
			w.Now = 50
			if _, err := w.PickFacing(); !errors.Is(err, ErrPickNotReady) {
				t.Fatal(err)
			}
			if w.Player != (Player{"meadow", 336, 320, defs.DirUp}) || w.PickRemaining("meadow", 3, 2) != 70 {
				t.Fatal(w.Player, w.pickReady)
			}
		}},
	} {
		t.Run(tc.id, func(t *testing.T) { tc.run(t, NewWorld(pickCatalog(), 42)) })
	}
}

func assertTarget(t *testing.T, w *World, want Target) {
	t.Helper()
	before := w.Player
	got, ok := w.FacingTarget()
	if !ok || got != want || w.Player != before {
		t.Fatalf("target=%+v,%v want=%+v", got, ok, want)
	}
}
func pickExample(t *testing.T, w *World) {
	t.Helper()
	w.Player = Player{"meadow", 336, 320, defs.DirUp}
	got, err := w.PickFacing()
	if err != nil || !reflect.DeepEqual(got, []Stack{{ItemKey{"herb", 40}, 3}}) || !reflect.DeepEqual(w.Inventory.Stacks(), got) {
		t.Fatal(got, err, w.Inventory.Stacks())
	}
}
func startFacingExample(t *testing.T, w *World) {
	t.Helper()
	w.Player = Player{"meadow", 528, 224, defs.DirUp}
	before := w.Player
	if err := w.StartBattleFacing(); err != nil {
		t.Fatal(err)
	}
	if w.Battle == nil || w.Battle.GatherID != "meadow" || w.Player != before {
		t.Fatal(w.Battle, w.Player)
	}
}

// 既存の抽選試験を拾う場所へ移す（DES-PCK-01）。
func preparePick(w *World, id string) {
	if w.catalog.Maps == nil {
		w.catalog.Maps = map[string]defs.MapDef{}
	}
	w.catalog.Maps[id] = defs.MapDef{ID: id, Rows: []string{"TTT", "T*T", "T,T", "TTT"}}
	s, exists := w.catalog.Gathers[id]
	if !exists {
		w.Player = Player{id, 144, 224, defs.DirUp}
		return
	}
	if s.RespawnSeconds == 0 {
		s.RespawnSeconds = 1
	}
	w.catalog.Gathers[id] = s
	w.Player = Player{id, 144, 224, defs.DirUp}
}

// 既存試験用に拾うマスを用意する。再出現状態は保持する（DES-PCK-01・02）。
func pickAtTestSite(w *World, id string) ([]Stack, error) { preparePick(w, id); return w.PickFacing() }

// 失敗時に在庫・乱数・場面・拾う場所・戦闘など全状態を比較する（DES-WLD-01）。
func assertPickFailure(t *testing.T, w *World, call func() error, want error, wantText string) {
	t.Helper()
	before := *w
	if w.Inventory.stacks != nil {
		before.Inventory = cloneInventory(w.Inventory)
	}
	if w.pickReady != nil {
		before.pickReady = make(map[PickKey]uint64)
		for k, v := range w.pickReady {
			before.pickReady[k] = v
		}
	}
	if w.Battle != nil {
		b := copyBattle(w.Battle)
		before.Battle = &b
	}
	rngBefore, _ := w.rngSource.MarshalBinary()
	err := call()
	rngAfter, _ := w.rngSource.MarshalBinary()
	if !errors.Is(err, want) || err.Error() != wantText || !reflect.DeepEqual(before, *w) || !reflect.DeepEqual(rngBefore, rngAfter) {
		t.Fatalf("error=%v want=%s 状態または乱数の変化=%v", err, wantText, !reflect.DeepEqual(before, *w))
	}
}

func TestPickFailureRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*World)
		want  error
		text  string
	}{
		{"TC-FR-MAT-01-55 戦闘優先", func(w *World) { w.Player.Scene = SceneGather; w.Battle = &Battle{} }, ErrInBattle, "戦闘中です"},
		{"TC-FR-MAT-01-55 拾う場所の検査優先", func(w *World) { w.Player.Scene = SceneGather; delete(w.catalog.Gathers, "meadow") }, ErrNothingToPick, "拾える物がありません"},
		{"TC-FR-MAT-01-55 空の検査優先", func(w *World) {
			w.pickReady = map[PickKey]uint64{{"meadow", 3, 2}: 120}
			delete(w.catalog.Gathers, "meadow")
		}, ErrPickNotReady, "あと 2 秒で拾えます: まだ拾えません"},
		{"TC-FR-MAT-01-56 未知の採取地", func(w *World) { delete(w.catalog.Gathers, "meadow") }, ErrUnknownID, "採取地 meadow: ID が見つかりません"},
		{"TC-FR-MAT-01-57 再出現ゼロ", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.RespawnSeconds = 0
			w.catalog.Gathers["meadow"] = s
		}, ErrInvalidCount, "個数が正しくありません"},
		{"TC-FR-MAT-01-57 再出現負数", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.RespawnSeconds = -1
			w.catalog.Gathers["meadow"] = s
		}, ErrInvalidCount, "個数が正しくありません"},
		{"TC-FR-MAT-01-57 秒からtickあふれ", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.RespawnSeconds = int(^uint(0) >> 1)
			w.catalog.Gathers["meadow"] = s
		}, ErrInvalidCount, "個数が正しくありません"},
		{"TC-FR-MAT-01-57 Now加算あふれ", func(w *World) { w.Now = ^uint64(0) - 119 }, ErrInvalidCount, "個数が正しくありません"},
		{"TC-FR-MAT-01-58 在庫あふれ", func(w *World) { w.Inventory.Add(ItemKey{"herb", 40}, int(^uint(0)>>1)-2) }, ErrInvalidCount, "追加後の個数が範囲外: 個数が正しくありません"},
		{"TC-FR-MAT-01-59 不正な採取表", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.Drops[0].QualityMax = 201
			w.catalog.Gathers["meadow"] = s
		}, ErrInvalidQuality, "品質は 1〜200 で指定してください"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(pickCatalog(), 42)
			w.Player = Player{"meadow", 336, 320, defs.DirUp}
			tc.setup(w)
			assertPickFailure(t, w, func() error { _, err := w.PickFacing(); return err }, tc.want, tc.text)
		})
	}
}

func TestEnterGatherFailureRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*World)
		id    string
		want  error
		text  string
	}{
		{"TC-FR-FLD-03-85 場面優先", func(w *World) { w.Player.Scene = "town"; w.Battle = &Battle{} }, "missing", ErrNotInGather, "採取地の一覧にいません"},
		{"TC-FR-FLD-03-85 戦闘優先", func(w *World) { w.Battle = &Battle{} }, "missing", ErrInBattle, "戦闘中です"},
		{"TC-FR-FLD-03-86 未知の採取地", func(w *World) {}, "missing", ErrUnknownID, "採取地 missing: ID が見つかりません"},
		{"TC-FR-FLD-03-86 解放検査優先", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.Rank = 2
			w.catalog.Gathers["meadow"] = s
			delete(w.catalog.Maps, "meadow")
		}, "meadow", ErrLocked, "草原はランク 2 で解放されます: まだ解放されていません"},
		{"TC-FR-FLD-03-87 地図なし", func(w *World) { delete(w.catalog.Maps, "meadow") }, "meadow", ErrNoGatherMap, "この採取地には地図がありません"},
		{"TC-FR-FLD-03-87 街への出口なし", func(w *World) { m := w.catalog.Maps["meadow"]; m.Exits = nil; w.catalog.Maps["meadow"] = m }, "meadow", ErrNoGatherMap, "この採取地には地図がありません"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(pickCatalog(), 42)
			w.Player.Scene = SceneGather
			tc.setup(w)
			assertPickFailure(t, w, func() error { return w.EnterGather(tc.id) }, tc.want, tc.text)
		})
	}
}

func TestBattleFacingRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*World)
		want  error
		text  string
	}{
		{"TC-FR-BTL-10-52 戦闘優先", func(w *World) { w.Battle = &Battle{}; w.Player.Scene = SceneGather }, ErrInBattle, "戦闘中です"},
		{"TC-FR-BTL-10-52 敵なし", func(w *World) { w.Player = Player{"meadow", 336, 320, defs.DirUp} }, ErrNoEnemy, "戦う相手がいません"},
		{"TC-FR-BTL-10-53 解放検査", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.Rank = 2
			s.Battle = nil
			w.catalog.Gathers["meadow"] = s
			w.InjuredUntil = 100
		}, ErrLocked, "草原はランク 2 で解放されます: まだ解放されていません"},
		{"TC-FR-BTL-10-53 戦闘定義なし", func(w *World) {
			s := w.catalog.Gathers["meadow"]
			s.Battle = nil
			w.catalog.Gathers["meadow"] = s
			w.InjuredUntil = 100
		}, ErrNoBattle, ErrNoBattle.Error()},
		{"TC-FR-BTL-10-54 負傷中", func(w *World) { w.InjuredUntil = 100 }, ErrInjured, ErrInjured.Error()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(pickCatalog(), 42)
			w.Player = Player{"meadow", 528, 224, defs.DirUp}
			tc.setup(w)
			assertPickFailure(t, w, w.StartBattleFacing, tc.want, tc.text)
		})
	}
	t.Run("TC-FR-BTL-10-55 接近だけでは開始しない", func(t *testing.T) {
		w := NewWorld(pickCatalog(), 42)
		w.Player = Player{"meadow", 528, 240, defs.DirUp}
		for i := 0; i < 10; i++ {
			w.MovePlayer(0, -1)
			w.Tick()
		}
		if w.Battle != nil || w.Player != (Player{"meadow", 528, 224, defs.DirUp}) {
			t.Fatal(w.Player, w.Battle)
		}
	})
	t.Run("TC-FR-BTL-10-56 戦闘中の移動と終了後の位置", func(t *testing.T) {
		w := NewWorld(pickCatalog(), 42)
		startFacingExample(t, w)
		assertPickFailure(t, w, func() error { return w.MovePlayer(1, 0) }, ErrInBattle, "戦闘中です")
		before := w.Player
		for i := range w.Battle.Units {
			if w.Battle.Units[i].Ally {
				w.Battle.Units[i].HP = 0
			}
		}
		if !w.finishBattle() || w.Player != before || w.Battle != nil || w.LastBattle.Won || w.InjuredUntil != 3600 {
			t.Fatal(w.Player, w.LastBattle)
		}
		if err := w.MovePlayer(0, 1); err != nil || w.Player != (Player{"meadow", 528, 231, defs.DirDown}) {
			t.Fatal(err, w.Player)
		}
	})
}

func TestPickPersistence(t *testing.T) {
	t.Run("TC-FR-FLD-05-02 場所と地図ごとに独立", func(t *testing.T) {
		w := NewWorld(pickCatalog(), 42)
		pickExample(t, w)
		w.Player = Player{"meadow", 144, 224, defs.DirUp}
		if _, err := w.PickFacing(); err != nil {
			t.Fatal(err)
		}
		m := w.catalog.Maps["meadow"]
		m.ID = "other"
		w.catalog.Maps["other"] = m
		w.catalog.Gathers["other"] = w.catalog.Gathers["meadow"]
		w.Player = Player{"other", 336, 320, defs.DirUp}
		if _, err := w.PickFacing(); err != nil {
			t.Fatal(err)
		}
		want := map[PickKey]uint64{{"meadow", 3, 2}: 120, {"meadow", 1, 1}: 120, {"other", 3, 2}: 120}
		if !reflect.DeepEqual(w.pickReady, want) || w.Inventory.Count(ItemKey{"herb", 40}) != 9 {
			t.Fatal(w.pickReady, w.Inventory.Stacks())
		}
	})
	t.Run("TC-FR-FLD-05-03 日替わりと残りtick", func(t *testing.T) {
		w := NewWorld(pickCatalog(), 42)
		pickExample(t, w)
		w.Now = 59
		w.Tick()
		if w.Day() != 2 || w.PickRemaining("meadow", 3, 2) != 60 || w.pickReady[PickKey{"meadow", 3, 2}] != 120 {
			t.Fatal(w.Now, w.pickReady)
		}
		for _, p := range []PickKey{{"meadow", 0, 0}, {"missing", 3, 2}, {"meadow", 5, 1}} {
			if got := w.PickRemaining(p.MapID, p.X, p.Y); got != 0 {
				t.Fatal(p, got)
			}
		}
		w.Now = 120
		if w.PickRemaining("meadow", 3, 2) != 0 {
			t.Fatal(w.pickReady)
		}
		w.Now = 121
		if w.PickRemaining("meadow", 3, 2) != 0 || w.pickReady[PickKey{"meadow", 3, 2}] != 120 {
			t.Fatal(w.pickReady)
		}
	})
	t.Run("TC-FR-FLD-03-88 負傷中に入って拾える", func(t *testing.T) {
		w := NewWorld(pickCatalog(), 42)
		w.InjuredUntil = 100
		w.Player.Scene = SceneGather
		if err := w.EnterGather("meadow"); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 3; i++ {
			w.MovePlayer(0, -1)
		}
		pickExample(t, w)
		if w.InjuredUntil != 100 {
			t.Fatal(w.InjuredUntil)
		}
	})
	t.Run("TC-FR-MAT-01-60 同じシードと入力", func(t *testing.T) {
		a, b := NewWorld(economyCatalog(), 42), NewWorld(economyCatalog(), 42)
		for _, w := range []*World{a, b} {
			preparePick(w, "field")
		}
		x, e1 := a.PickFacing()
		y, e2 := b.PickFacing()
		want := []Stack{{ItemKey{"herb", 2}, 1}, {ItemKey{"potion", 50}, 2}}
		if e1 != nil || e2 != nil || !reflect.DeepEqual(x, want) || !reflect.DeepEqual(y, want) || !reflect.DeepEqual(a.Inventory.Stacks(), want) || !reflect.DeepEqual(b.Inventory.Stacks(), want) {
			t.Fatal(x, y, e1, e2)
		}
		sa, _ := a.rngSource.MarshalBinary()
		sb, _ := b.rngSource.MarshalBinary()
		if !reflect.DeepEqual(sa, sb) {
			t.Fatal("乱数不一致")
		}
	})
}

func TestFacingTargetRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		p     Player
		want  Target
		found bool
	}{
		{"TC-IF-08-74 同面積なら左の拾う場所", Player{"mixed", 288, 224, defs.DirUp}, Target{TargetPick, "", 2, 1}, true},
		{"TC-IF-08-74 面積が大きい敵", Player{"mixed", 290, 224, defs.DirUp}, Target{TargetEnemy, "", 3, 1}, true},
		{"TC-IF-08-75 同面積なら上の敵", Player{"mixed", 256, 192, defs.DirRight}, Target{TargetEnemy, "", 3, 1}, true},
		{"TC-IF-08-75 面積が大きい設備", Player{"mixed", 256, 194, defs.DirRight}, Target{TargetFixture, defs.FixtureQuestBoard, 3, 2}, true},
		{"TC-IF-08-76 区画を返す", Player{"mixed", 240, 320, defs.DirUp}, Target{TargetLot, "market", 2, 2}, true},
		{"TC-IF-08-77 地図なし", Player{"missing", 240, 320, defs.DirUp}, Target{}, false},
		{"TC-IF-08-77 一覧では返さない", Player{SceneGather, 240, 320, defs.DirUp}, Target{}, false},
		{"TC-IF-08-77 不正な向き", Player{"mixed", 240, 320, defs.Dir("bad")}, Target{}, false},
		{"TC-IF-08-77 壁だけ", Player{"mixed", 128, 144, defs.DirLeft}, Target{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := pickCatalog()
			c.Maps["mixed"] = defs.MapDef{ID: "mixed", Rows: []string{"TTTTT", "T,*ET", "T,aQT", "T,,,T", "TTTTT"}, Lots: []defs.LotDef{{Mark: 'a', Facility: "market", X: 2, Y: 2}}}
			w := NewWorld(c, 42)
			w.Player = tc.p
			got, ok := w.FacingTarget()
			if got != tc.want || ok != tc.found || w.Player != tc.p {
				t.Fatal(got, ok, w.Player)
			}
		})
	}
}

func TestGatherMapObstacles(t *testing.T) {
	for _, tc := range []struct {
		name   string
		p      Player
		dx, dy int
		want   Player
	}{
		{"TC-FR-FLD-02-81 草の拾う場所", Player{"meadow", 144, 224, defs.DirUp}, 0, -1, Player{"meadow", 144, 224, defs.DirUp}},
		{"TC-FR-FLD-02-81 敵", Player{"meadow", 528, 224, defs.DirUp}, 0, -1, Player{"meadow", 528, 224, defs.DirUp}},
		{"TC-FR-FLD-02-82 区画は未建設でも通れない", Player{"town", 512, 240, defs.DirLeft}, -1, 0, Player{"town", 512, 240, defs.DirLeft}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWorld(pickCatalog(), 42)
			w.Player = tc.p
			if err := w.MovePlayer(tc.dx, tc.dy); err != nil || w.Player != tc.want {
				t.Fatal(err, w.Player)
			}
			w.Facilities = map[string]int{"market": 1}
			if err := w.MovePlayer(tc.dx, tc.dy); err != nil || w.Player != tc.want {
				t.Fatal(err, w.Player)
			}
		})
	}
}
