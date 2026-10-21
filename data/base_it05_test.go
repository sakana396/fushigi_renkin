package data_test

import (
	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/defs"
	"strings"
	"testing"
)

// TestBaseMapsIT05 は反復 6 の最小移行データを確認する（DES-VAL-17〜29）。
// 工房地図と旧距離制約は DES-IT06 で廃止。完成データの試験は #106。
func TestBaseMapsIT05(t *testing.T) {
	c, err := defload.Load(data.Base)
	t.Run("TC-FR-FLD-04-40", func(t *testing.T) {
		if err != nil {
			t.Fatal(err)
		}
		if len(c.Maps) != len(c.Gathers)+1 {
			t.Fatal(c.Maps)
		}
	})
	if err != nil {
		return
	}
	t.Run("TC-FR-FLD-01-30 / TC-FR-FLD-04-41", func(t *testing.T) {
		if _, ok := c.Maps["workshop"]; ok {
			t.Fatal("工房地図が残っている")
		}
		for id := range c.Gathers {
			m, ok := c.Maps[id]
			if !ok || len(m.Exits) != 1 || m.Exits[0].To != defs.MapTown {
				t.Fatalf("採取地 %s: %+v", id, m)
			}
		}
	})
	t.Run("TC-FR-FLD-01-31 / TC-FR-FLD-04-42", func(t *testing.T) {
		m := c.Maps[defs.MapTown]
		if len(m.Lots) != len(c.Facilities) || len(m.Exits) != 1 || m.Exits[0].To != defs.ExitGather {
			t.Fatal(m)
		}
		for _, ch := range "DMQ" {
			if !strings.ContainsRune(strings.Join(m.Rows, ""), ch) {
				t.Fatalf("設備 %c がない", ch)
			}
		}
	})
}
