package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// reportDataError は本体データの誤りを Windows の画面に表示する（DES-PKG-01、UC-09）。
func reportDataError(message string) {
	caption, err := windows.UTF16PtrFromString("ふしぎ錬金")
	if err == nil {
		var body *uint16
		body, err = windows.UTF16PtrFromString(message)
		if err == nil {
			_, err = windows.MessageBox(0, body, caption, windows.MB_OK|windows.MB_ICONERROR)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, message)
		fmt.Fprintln(os.Stderr, err)
	}
}
