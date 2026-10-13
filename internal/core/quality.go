package core

import "fmt"

// AverageQuality は投入個数で重み付けした品質の平均を端数切り捨てで返す（DES-QLT-01、DES-QLT-04）。
func AverageQuality(stacks []Stack) (int, error) {
	if len(stacks) == 0 {
		return 0, ErrInvalidCount
	}
	var sum, total int
	for _, stack := range stacks {
		if stack.Count <= 0 {
			return 0, fmt.Errorf("素材の個数 %d: %w", stack.Count, ErrInvalidCount)
		}
		if stack.Key.Quality < 1 || stack.Key.Quality > 200 {
			return 0, fmt.Errorf("素材の品質 %d: %w", stack.Key.Quality, ErrInvalidQuality)
		}
		maxInt := int(^uint(0) >> 1)
		if stack.Count > maxInt-total || stack.Count > (maxInt-sum)/stack.Key.Quality {
			return 0, fmt.Errorf("素材の個数が範囲外: %w", ErrInvalidCount)
		}
		sum += stack.Key.Quality * stack.Count
		total += stack.Count
	}
	return sum / total, nil
}

// ClampQuality は品質を 1〜200 に収める（DES-QLT-02）。
func ClampQuality(quality int) int {
	if quality < 1 {
		return 1
	}
	if quality > 200 {
		return 200
	}
	return quality
}
