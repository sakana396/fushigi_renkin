package core

// Day は現在の tick から今日の日数を求める（DES-DAY-01）。
func (w *World) Day() int {
	if w.catalog.Game.DaySeconds == 0 {
		return 1
	}
	return int(w.Now/(uint64(w.catalog.Game.DaySeconds)*60)) + 1
}

// TakeNewDays は日替わりの知らせを取り出して空にする（DES-DAY-02）。
func (w *World) TakeNewDays() []int {
	days := w.newDays
	w.newDays = nil
	return days
}
