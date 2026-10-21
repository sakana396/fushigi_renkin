package defload

import (
	"strings"
	"testing"
)

// TestMapReview は開始位置の境界と診断形式を確かめる（DES-VAL-20）。
func TestMapReview(t *testing.T) {
	for _, tc := range []struct{ id, value string }{
		{"TC-FR-FLD-01-06", "y = 2"}, {"TC-FR-FLD-01-07", "x = 1"}, {"TC-FR-FLD-01-08", "x = 7, y = 2"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			_, err := loadMapText(strings.Replace(validMaps, "x = 1, y = 2", tc.value, 1))
			mapDiagnostics(t, err, ".start")
		})
	}
	t.Run("TC-FR-FLD-01-09", func(t *testing.T) {
		_, err := loadMapText(strings.Replace(validMaps, "x = 1, y = 2", "x = -1, y = 2", 1))
		mapDiagnostics(t, err, ".start")
	})
}
