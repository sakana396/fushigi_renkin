package defs

// 地図と採取地の ID（DES-DEF-09）。
const (
	MapTown    = "town"
	ExitGather = "gather"
)

// 設備の種類（DES-DEF-09）。
const (
	FixtureWorkshopDoor  = "workshop_door"
	FixtureCauldron      = "cauldron"
	FixtureShelf         = "shelf"
	FixtureWorkbench     = "workbench"
	FixtureShop          = "shop"
	FixtureQuestBoard    = "quest_board"
	FixtureFacilityBoard = "facility_board"
)

// Dir は地図内を向く方向を表す（DES-DEF-09）。
type Dir string

const (
	DirDown  Dir = "down"
	DirUp    Dir = "up"
	DirLeft  Dir = "left"
	DirRight Dir = "right"
)

// CellKind は地図の文字の種類を表す（DES-DEF-09・10）。
type CellKind uint8

const (
	CellUnknown CellKind = iota
	CellFloor
	CellWall
	CellFixture
	CellExit
	CellLot
	CellPick
	CellEnemy
)

// CellInfo は文字に対応する通行と設備の情報を持つ（DES-DEF-09）。
type CellInfo struct {
	Kind    CellKind
	Fixture string
	Name    string
}

// TileInfo は地図の文字を固定の表から引く（DES-DEF-09・10）。
func TileInfo(ch byte) CellInfo {
	switch ch {
	case '.', ',', ':':
		return CellInfo{Kind: CellFloor}
	case '#', 'W', '^', 'T', '~':
		return CellInfo{Kind: CellWall}
	case 'D':
		return CellInfo{CellFixture, FixtureWorkshopDoor, "工房の扉"}
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i':
		return CellInfo{Kind: CellLot, Name: "区画"}
	case '*', 'o':
		return CellInfo{Kind: CellPick, Name: "拾う場所"}
	case 'E':
		return CellInfo{Kind: CellEnemy, Name: "敵"}
	case 'M':
		return CellInfo{CellFixture, FixtureShop, "店"}
	case 'Q':
		return CellInfo{CellFixture, FixtureQuestBoard, "依頼板"}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return CellInfo{Kind: CellExit}
	default:
		return CellInfo{}
	}
}

// MapDef は検証済みの文字の地図と施設の区画を表す（DES-DEF-09・11）。
type MapDef struct {
	ID, Name       string
	Rows           []string
	StartX, StartY int
	Lots           []LotDef
	Exits          []ExitDef
}

// ExitDef は地図の出入り口を表す（DES-DEF-09）。
type ExitDef struct {
	Mark byte
	To   string
	Dir  Dir
	Look string
	X, Y int
}

// Width は地図の幅を返す（DES-DEF-09）。
func (m MapDef) Width() int {
	if len(m.Rows) == 0 {
		return 0
	}
	return len(m.Rows[0])
}

// Height は地図の高さを返す（DES-DEF-09）。
func (m MapDef) Height() int { return len(m.Rows) }

// Cell は地図外では 0 を返す（DES-DEF-09）。
func (m MapDef) Cell(x, y int) byte {
	if y < 0 || y >= len(m.Rows) || x < 0 || x >= len(m.Rows[y]) {
		return 0
	}
	return m.Rows[y][x]
}

// Map は ID に対応する地図を返す（DES-DEF-09）。
func (c Catalog) Map(id string) (MapDef, bool) { m, ok := c.Maps[id]; return m, ok }

// LotDef は施設の区画と左上のマスを表す（DES-DEF-11）。
type LotDef struct {
	Mark     byte
	Facility string
	X, Y     int
}
