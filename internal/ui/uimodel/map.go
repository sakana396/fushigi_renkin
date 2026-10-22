package uimodel

import "github.com/sakana396/fushigi_renkin/internal/defs"

// TilePixels は16pxのタイルを3倍した地図上の大きさ（DES-UI-16）。
const TilePixels = 48

// CameraOffset は小さい地図を中央に置き、大きい地図の端で止める（DES-UI-16）。
func CameraOffset(size, viewport, player int) int {
	if size <= viewport {
		return (viewport - size) / 2
	}
	offset := viewport/2 - player
	if offset > 0 {
		return 0
	}
	if offset < viewport-size {
		return viewport - size
	}
	return offset
}

// VisibleRange は表示枠に交わるマスの範囲を求める（DES-UI-16）。
func VisibleRange(offset, count, viewport int) (int, int) {
	first := 0
	if offset < 0 {
		first = (-offset) / TilePixels
	}
	last := (viewport - offset + TilePixels - 1) / TilePixels
	if last > count {
		last = count
	}
	return first, last
}

// DominantFloor は下地の床を個数と優先順で選ぶ（DES-UI-15）。
func DominantFloor(m defs.MapDef) byte {
	count := map[byte]int{'.': 0, ',': 0, ':': 0}
	for _, row := range m.Rows {
		for i := 0; i < len(row); i++ {
			if _, ok := count[row[i]]; ok {
				count[row[i]]++
			}
		}
	}
	best := byte('.')
	for _, ch := range []byte{',', ':'} {
		if count[ch] > count[best] {
			best = ch
		}
	}
	return best
}

func GroundTile(ch byte) (bool, int) {
	switch ch {
	case ',':
		return true, 0
	case ':':
		return true, 25
	default:
		return false, 48
	}
}

// MapTile は文字と隣の文字を絵番号に対応付ける（DES-UI-15、4.3 節）。
func MapTile(m defs.MapDef, x, y int, ch byte) (bool, int) {
	switch ch {
	case '.':
		return false, 48
	case ',':
		return true, 0
	case ':':
		return true, 25
	case '#':
		return false, 40
	case 'W':
		left, right := m.Cell(x-1, y), m.Cell(x+1, y)
		if left != 'W' && !(left >= '1' && left <= '9') && (right == 'W' || right >= '1' && right <= '9') {
			return true, 72
		}
		if right != 'W' && !(right >= '1' && right <= '9') && (left == 'W' || left >= '1' && left <= '9') {
			return true, 75
		}
		return true, 73
	case '^':
		n := 53
		if m.Cell(x-1, y) != '^' && m.Cell(x+1, y) == '^' {
			n = 52
		}
		if m.Cell(x+1, y) != '^' && m.Cell(x-1, y) == '^' {
			n = 54
		}
		if m.Cell(x, y-1) == '^' {
			n += 12
		}
		return true, n
	case '*':
		return true, 5
	case 'o':
		return false, 6
	case 'T':
		return true, 4
	case 'D':
		return true, 86
	case 'M':
		return true, 57
	case 'Q':
		return true, 83
	}
	return true, 25
}

func ExitTile(look string) (bool, int) {
	switch look {
	case "stone_door":
		return false, 45
	case "wood_door":
		return true, 86
	default:
		return true, 25
	}
}

// MapNamePosition は名前をマスの上に置き、地図表示枠の左右と上に収める（DES-UI-15・16）。
func MapNamePosition(x, y, boxW, boxH, viewport int) (int, int) {
	if x+boxW > viewport {
		x = viewport - boxW
	}
	if x < 0 {
		x = 0
	}
	y -= boxH
	if y < 0 {
		y = 0
	}
	return x, y
}

// ExitArrow は出口のマスに重ねる下向き矢印の軸と左の羽を返す（DES-UI-28 U-3）。
func ExitArrow(x, y int) [6]float32 {
	return [6]float32{float32(x + 24), float32(y + 9), float32(x + 24), float32(y + 37), float32(x + 12), float32(y + 25)}
}
