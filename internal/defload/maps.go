package defload

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/sakana396/fushigi_renkin/internal/defs"
)

type mapStart struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}
type mapExit struct {
	Mark string `toml:"mark"`
	To   string `toml:"to"`
	Dir  string `toml:"dir"`
	Look string `toml:"look"`
}
type mapDef struct {
	ID    string    `toml:"id"`
	Name  string    `toml:"name"`
	Start *mapStart `toml:"start"`
	Rows  []string  `toml:"rows"`
	Exits []mapExit `toml:"exits"`
}
type mapSource struct {
	file       string
	value      defs.MapDef
	valid      bool
	exitsValid bool
}

// validateMaps は地図の形、参照と到達性を順に検査する（DES-VAL-17〜23）。
func validateMaps(c *defs.Catalog, maps []source[mapDef], add func(string, string, string)) {
	if len(maps) == 0 {
		return
	}
	sources := make(map[string]*mapSource)
	for _, s := range maps {
		v := s.Value
		p := "map[id=" + v.ID + "]"
		m := defs.MapDef{ID: v.ID, Name: v.Name, Rows: v.Rows}
		bad := false
		report := func(key, reason string) { add(s.File, key, reason); bad = true }
		if v.ID != defs.MapWorkshop && v.ID != defs.MapTown {
			report(p+".id", "地図の ID は workshop または town")
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
		if v.Start == nil {
			report(p+".start", "開始位置が書かれていない")
		} else {
			m.StartX, m.StartY = v.Start.X, v.Start.Y
			ch := m.Cell(m.StartX, m.StartY)
			if defs.TileInfo(ch).Kind != defs.CellFloor {
				report(p+".start", fmt.Sprintf("開始位置 (%d, %d) が床ではない（%q）", m.StartX, m.StartY, string(ch)))
			}
		}
		required := "KSG"
		if v.ID == defs.MapTown {
			required = "MQF"
		}
		for _, ch := range required {
			if !strings.ContainsRune(strings.Join(v.Rows, ""), ch) {
				report(p, fmt.Sprintf("%s（%c）がない", defs.TileInfo(byte(ch)).Name, ch))
			}
		}
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
			if e.To == v.ID {
				report(q+".to", "自分の地図には行けない")
			} else if e.To != defs.MapWorkshop && e.To != defs.MapTown && e.To != defs.ExitGather {
				report(q+".to", "行き先がない: "+e.To)
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
		for mark, count := range counts {
			if count != 1 {
				xy := coords[mark]
				report(fmt.Sprintf("%s.rows[%d][%d]", p, xy[1], xy[0]), fmt.Sprintf("出入り口 %q が %d マスある", string(mark), count))
			}
			if !marks[mark] {
				xy := coords[mark]
				report(fmt.Sprintf("%s.rows[%d][%d]", p, xy[1], xy[0]), fmt.Sprintf("出入り口 %q の定義がない", string(mark)))
			}
		}
		if v.ID == defs.MapWorkshop && !tos[defs.MapTown] {
			report(p+".exits", "town 行きがない")
		}
		if v.ID == defs.MapTown {
			if !tos[defs.MapWorkshop] {
				report(p+".exits", "workshop 行きがない")
			}
			if !tos[defs.ExitGather] {
				report(p+".exits", "gather 行きがない")
			}
		}
		if v.ID != defs.MapTown && tos[defs.ExitGather] {
			report(p+".exits", "gather 行きは town にだけ置く")
		}
		if src := sources[v.ID]; src != nil && !duplicate {
			src.value = m
			src.valid = !bad
			src.exitsValid = !bad
		}
	}
	for _, id := range []string{defs.MapWorkshop, defs.MapTown} {
		if _, ok := sources[id]; !ok {
			add(maps[0].File, "map[id="+id+"]", "地図がない")
		}
	}
	for _, s := range maps {
		src := sources[s.Value.ID]
		if src == nil || src.file != s.File || !src.valid {
			continue
		}
		p := "map[id=" + s.Value.ID + "]"
		for i, e := range src.value.Exits {
			if e.To == defs.ExitGather {
				continue
			}
			other := sources[e.To]
			if other == nil {
				add(src.file, fmt.Sprintf("%s.exits[%d].to", p, i), "行き先がない: "+e.To)
				src.exitsValid = false
				continue
			}
			if !other.valid {
				src.exitsValid = false
				continue
			}
			n := 0
			for _, back := range other.value.Exits {
				if back.To == s.Value.ID {
					n++
				}
			}
			if n != 1 {
				add(src.file, fmt.Sprintf("%s.exits[%d].to", p, i), "戻り道がない: "+e.To)
				src.exitsValid = false
			}
		}
	}
	for _, src := range sources {
		if src.valid && src.exitsValid {
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
			if info.Kind != defs.CellFixture {
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
