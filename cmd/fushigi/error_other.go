//go:build !windows

package main

import (
	"fmt"
	"os"
)

// reportDataError は本体データの誤りを標準エラー出力に表示する（DES-PKG-01、UC-09）。
func reportDataError(message string) {
	fmt.Fprintln(os.Stderr, message)
}
