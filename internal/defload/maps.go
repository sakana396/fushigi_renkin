package defload

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

type mapStart struct {
	X *int `toml:"x"`
	Y *int `toml:"y"`
}
type mapExit struct {
	Mark string `toml:"mark"`
	To   string `toml:"to"`
	Dir  string `toml:"dir"`
	Look string `toml:"look"`
}
type mapLot struct {
	Mark     string `toml:"mark"`
	Facility string `toml:"facility"`
}
type mapDef struct {
	Lots  []mapLot  `toml:"lots"`
	ID    string    `toml:"id"`
	Name  string    `toml:"name"`
	Start *mapStart `toml:"start"`
	Rows  []string  `toml:"rows"`
	Exits []mapExit `toml:"exits"`
}
type mapSource struct {
	file  string
	value defs.MapDef
	valid bool
}

// validateMaps は地図の形、参照と到達性を順に検査する（DES-VAL-17〜28）。
func validateMaps(c *defs.Catalog, maps []source[mapDef], add func(string, string, string)) {
	if len(maps) == 0 {
		return
	}
	maps = append([]source[mapDef](nil), maps...)
	sort.SliceStable(maps, func(i, j int) bool {
		if maps[i].Value.ID != maps[j].Value.ID {
			return maps[i].Value.ID < maps[j].Value.ID
		}
		return maps[i].File < maps[j].File
	})
	ids := []string{defs.MapTown}
	for id := range c.Gathers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	sources := make(map[string]*mapSource)
	for _, s := range maps {
		v := s.Value
		p := "map[id=" + v.ID + "]"
		m := defs.MapDef{ID: v.ID, Name: v.Name, Rows: v.Rows}
		bad := false
		report := func(key, reason string) { add(s.File, key, reason); bad = true }
		if _, ok := c.Gathers[v.ID]; v.ID != defs.MapTown && (!ok || v.ID == defs.ExitGather) {
			report(p+".id", "地図の ID は town または採取地の ID")
		}
		_, duplicate := sources[v.ID]
		if duplicate {
			report(p+".id", "ID が重複")
		} else {
			sources[v.ID] = &mapSource{file: s.File, value: m}
		}
		if len(v.Rows) < 1 || len(v.Rows) > 100 {
			report(p+".rows", "行数は 1〜100")
		}
		width := 0
		if len(v.Rows) > 0 {
			width = utf8.RuneCountInString(v.Rows[0])
		}
		for y, row := range v.Rows {
			n := utf8.RuneCountInString(row)
			if n < 1 || n > 100 {
				report(fmt.Sprintf("%s.rows[%d]", p, y), "行の長さは 1〜100 文字")
			}
			if y > 0 && n != width {
				report(fmt.Sprintf("%s.rows[%d]", p, y), fmt.Sprintf("行の長さが rows[0] と違う（%d 文字。rows[0] は %d 文字）", n, width))
			}
			for x, ch := range []rune(row) {
				if ch > 127 || defs.TileInfo(byte(ch)).Kind == defs.CellUnknown {
					report(fmt.Sprintf("%s.rows[%d][%d]", p, y, x), fmt.Sprintf("知らない文字 %q", string(ch)))
					break
				}
			}
		}
		if bad {
			continue
		}
		if v.ID == defs.MapTown {
			if v.Start == nil || v.Start.X == nil || v.Start.Y == nil {
				report(p+".start", "開始位置が書かれていない")
			} else {
				m.StartX, m.StartY = *v.Start.X, *v.Start.Y
				if m.StartX < 0 || m.StartX >= m.Width() || m.StartY < 0 || m.StartY >= m.Height() {
					report(p+".start", fmt.Sprintf("開始位置 (%d, %d) が地図の外", m.StartX, m.StartY))
				} else if ch := m.Cell(m.StartX, m.StartY); defs.TileInfo(ch).Kind != defs.CellFloor {
					report(p+".start", fmt.Sprintf("開始位置 (%d, %d) が床ではない（%q）", m.StartX, m.StartY, string(ch)))
				}
			}
			for _, ch := range "DMQ" {
				if !strings.ContainsRune(strings.Join(v.Rows, ""), ch) {
					report(p, fmt.Sprintf("%s（%c）がない", defs.TileInfo(byte(ch)).Name, ch))
				}
			}
		} else if v.Start != nil {
			report(p+".start", "採取地の地図には開始位置を書かない")
		}
		picks, enemies := 0, 0
		for y, row := range v.Rows {
			for x := 0; x < len(row); x++ {
				kind := defs.TileInfo(row[x]).Kind
				if kind == defs.CellPick {
					picks++
				}
				if kind == defs.CellEnemy {
					enemies++
				}
				if v.ID == defs.MapTown && (kind == defs.CellPick || kind == defs.CellEnemy) {
					report(fmt.Sprintf("%s.rows[%d][%d]", p, y, x), fmt.Sprintf("%s（%c）は採取地にしか置けない", defs.TileInfo(row[x]).Name, row[x]))
				}
				if v.ID != defs.MapTown && kind == defs.CellLot {
					report(fmt.Sprintf("%s.rows[%d][%d]", p, y, x), fmt.Sprintf("区画（%c）は街にしか置けない", row[x]))
				}
			}
		}
		if v.ID != defs.MapTown {
			if picks == 0 {
				add(s.File, p+".rows", "拾う場所は 1 マス以上必要")
			}
			want := 0
			if c.Gathers[v.ID].Battle != nil {
				want = 1
			}
			if enemies != want {
				add(s.File, p+".rows", fmt.Sprintf("敵（E）が %d マスある（必要なマス数は %d）", enemies, want))
			}
		}
		validateLots(&m, v.Lots, *c, report)
		counts := map[byte]int{}
		coords := map[byte][2]int{}
		for y, row := range v.Rows {
			for x := 0; x < len(row); x++ {
				ch := row[x]
				if defs.TileInfo(ch).Kind == defs.CellExit {
					counts[ch]++
					coords[ch] = [2]int{x, y}
				}
			}
		}
		marks := map[byte]bool{}
		tos := map[string]bool{}
		for i, e := range v.Exits {
			q := fmt.Sprintf("%s.exits[%d]", p, i)
			if len(e.Mark) != 1 || e.Mark[0] < '1' || e.Mark[0] > '9' {
				report(q+".mark", "mark は 1〜9 の 1 文字")
				continue
			}
			mark := e.Mark[0]
			if marks[mark] {
				report(q+".mark", "mark が重複")
			}
			marks[mark] = true
			if counts[mark] != 1 {
				report(q+".mark", fmt.Sprintf("出入り口 %q は rows にちょうど 1 マス必要", e.Mark))
			}
			wantTo := defs.MapTown
			if v.ID == defs.MapTown {
				wantTo = defs.ExitGather
			}
			if e.To == v.ID {
				report(q+".to", "自分の地図には行けない")
			} else if e.To != wantTo {
				report(q+".to", "行き先がない: "+e.To+"（必要な行き先は "+wantTo+"）")
			}
			if tos[e.To] {
				report(q+".to", "同じ行き先が重複")
			}
			tos[e.To] = true
			dir := defs.Dir(e.Dir)
			dx, dy := 0, 0
			switch dir {
			case defs.DirUp:
				dy = -1
			case defs.DirDown:
				dy = 1
			case defs.DirLeft:
				dx = -1
			case defs.DirRight:
				dx = 1
			default:
				report(q+".dir", "向きは up・down・left・right")
			}
			if e.Look != "stone_door" && e.Look != "wood_door" && e.Look != "road" {
				report(q+".look", "絵は stone_door・wood_door・road")
			}
			xy := coords[mark]
			if counts[mark] == 1 && dir != "" && (dx != 0 || dy != 0) && defs.TileInfo(m.Cell(xy[0]+dx, xy[1]+dy)).Kind != defs.CellFloor {
				report(q+".dir", fmt.Sprintf("到着のマス (%d, %d) が床ではない", xy[0]+dx, xy[1]+dy))
			}
			m.Exits = append(m.Exits, defs.ExitDef{Mark: mark, To: e.To, Dir: dir, Look: e.Look, X: xy[0], Y: xy[1]})
		}
		for mark := byte('1'); mark <= '9'; mark++ {
			count := counts[mark]
			if count == 0 {
				continue
			}
			if count != 1 {
				xy := coords[mark]
				report(fmt.Sprintf("%s.rows[%d][%d]", p, xy[1], xy[0]), fmt.Sprintf("出入り口 %q が %d マスある", string(mark), count))
			}
			if !marks[mark] {
				xy := coords[mark]
				report(fmt.Sprintf("%s.rows[%d][%d]", p, xy[1], xy[0]), fmt.Sprintf("出入り口 %q の定義がない", string(mark)))
			}
		}
		wantTo := defs.MapTown
		if v.ID == defs.MapTown {
			wantTo = defs.ExitGather
		}
		if len(v.Exits) != 1 || !tos[wantTo] {
			report(p+".exits", wantTo+" 行きがちょうど 1 つ必要")
		}
		if v.ID != defs.MapTown && !bad {
			e := m.Exits[0]
			dx, dy := mapDirection(e.Dir)
			m.StartX, m.StartY = e.X+dx, e.Y+dy
		}
		if src := sources[v.ID]; src != nil && !duplicate {
			src.value = m
			src.valid = !bad
		}
	}
	for _, id := range ids {
		if _, ok := sources[id]; !ok {
			add(maps[0].File, "map[id="+id+"]", "地図がない")
		}
	}
	for _, id := range ids {
		src := sources[id]
		if src == nil {
			continue
		}
		if src.valid {
			validateMapReachability(src, add)
		}
		if src.valid {
			c.Maps[src.value.ID] = src.value
		}
	}
}

// validateMapReachability は床だけを塗りつぶして到達性を調べる（DES-VAL-23）。
func validateMapReachability(src *mapSource, add func(string, string, string)) {
	m := src.value
	p := "map[id=" + m.ID + "]"
	w, h := m.Width(), m.Height()
	seen := make([]bool, w*h)
	queue := [][2]int{{m.StartX, m.StartY}}
	seen[m.StartY*w+m.StartX] = true
	dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for len(queue) > 0 {
		xy := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			x, y := xy[0]+d[0], xy[1]+d[1]
			if x >= 0 && x < w && y >= 0 && y < h && defs.TileInfo(m.Cell(x, y)).Kind == defs.CellFloor && !seen[y*w+x] {
				seen[y*w+x] = true
				queue = append(queue, [2]int{x, y})
			}
		}
	}
	reachable := func(x, y int) bool { return x >= 0 && x < w && y >= 0 && y < h && seen[y*w+x] }
	for y, row := range m.Rows {
		for x := 0; x < len(row); x++ {
			info := defs.TileInfo(row[x])
			if info.Kind != defs.CellFixture && info.Kind != defs.CellPick && info.Kind != defs.CellEnemy {
				continue
			}
			ok := false
			for _, d := range dirs {
				if reachable(x+d[0], y+d[1]) {
					ok = true
				}
			}
			if !ok {
				add(src.file, fmt.Sprintf("%s.rows[%d][%d]", p, y, x), "開始位置から行けない: "+info.Name)
			}
		}
	}
	for i, lot := range m.Lots {
		ok := false
		for y := lot.Y; y < lot.Y+3; y++ {
			for x := lot.X; x < lot.X+3; x++ {
				for _, d := range dirs {
					if reachable(x+d[0], y+d[1]) {
						ok = true
					}
				}
			}
		}
		if !ok {
			add(src.file, fmt.Sprintf("%s.lots[%d]", p, i), "開始位置から行けない: 区画")
		}
	}
	for i, e := range m.Exits {
		dx, dy := 0, 0
		switch e.Dir {
		case defs.DirUp:
			dy = -1
		case defs.DirDown:
			dy = 1
		case defs.DirLeft:
			dx = -1
		case defs.DirRight:
			dx = 1
		}
		if !reachable(e.X+dx, e.Y+dy) {
			add(src.file, fmt.Sprintf("%s.exits[%d]", p, i), fmt.Sprintf("開始位置から到着のマス (%d, %d) へ行けない", e.X+dx, e.Y+dy))
		}
	}
}

// mapDirection は出口から到着マスへの差分を返す（DES-VAL-22）。
func mapDirection(dir defs.Dir) (int, int) {
	switch dir {
	case defs.DirUp:
		return 0, -1
	case defs.DirDown:
		return 0, 1
	case defs.DirLeft:
		return -1, 0
	case defs.DirRight:
		return 1, 0
	}
	return 0, 0
}

// validateLots は文字と施設の対応、3×3 の形を検証し左上を保存する（DES-DEF-11、DES-VAL-25）。
func validateLots(m *defs.MapDef, lots []mapLot, c defs.Catalog, report func(string, string)) {
	p := "map[id=" + m.ID + "]"
	if m.ID != defs.MapTown {
		if lots != nil {
			report(p+".lots", "lots は街にだけ書く")
		}
		return
	}
	marks := map[byte]bool{}
	facilities := map[string]bool{}
	for i, l := range lots {
		q := fmt.Sprintf("%s.lots[%d]", p, i)
		if len(l.Mark) != 1 || l.Mark[0] < 'a' || l.Mark[0] > 'i' {
			report(q+".mark", "mark は a〜i の 1 文字")
			continue
		}
		mark := l.Mark[0]
		if marks[mark] {
			report(q+".mark", "mark が重複")
		}
		marks[mark] = true
		if _, ok := c.Facilities[l.Facility]; !ok {
			report(q+".facility", "存在しない施設: "+l.Facility)
		}
		if facilities[l.Facility] {
			report(q+".facility", "同じ施設の区画が重複")
		}
		facilities[l.Facility] = true
		minX, minY, count := m.Width(), m.Height(), 0
		for y, row := range m.Rows {
			for x := range row {
				if row[x] == mark {
					count++
					if x < minX {
						minX = x
					}
					if y < minY {
						minY = y
					}
				}
			}
		}
		square := count == 9
		for y := minY; y < minY+3; y++ {
			for x := minX; x < minX+3; x++ {
				if m.Cell(x, y) != mark {
					square = false
				}
			}
		}
		if !square {
			report(q, fmt.Sprintf("区画 %q が 3 × 3 の四角になっていない", l.Mark))
		}
		m.Lots = append(m.Lots, defs.LotDef{Mark: mark, Facility: l.Facility, X: minX, Y: minY})
	}
	for y, row := range m.Rows {
		for x := range row {
			if defs.TileInfo(row[x]).Kind == defs.CellLot && !marks[row[x]] {
				report(fmt.Sprintf("%s.rows[%d][%d]", p, y, x), "区画の定義がない")
			}
		}
	}
	ids := make([]string, 0, len(c.Facilities))
	for id := range c.Facilities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if !facilities[id] {
			report(p+".lots", "施設 "+id+" の区画がない")
		}
	}
}
