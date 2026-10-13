// 起動処理を置くパッケージ（DES-PKG-01）。
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sakana396/fushigi_renkin/data"
	"github.com/sakana396/fushigi_renkin/internal/core"
	"github.com/sakana396/fushigi_renkin/internal/defload"
	"github.com/sakana396/fushigi_renkin/internal/ui"
)

// main は本体データの検証、World の作成、画面の起動を順に行う（DES-PKG-01、DES-PKG-06）。
func main() {
	catalog, err := defload.Load(data.Base)
	if err != nil {
		reportDataError(fmt.Sprintf("本体の定義データに誤りがあります:\n%s", err))
		os.Exit(1)
	}
	if err := ui.Run(core.NewWorld(catalog, 1), catalog); err != nil {
		log.Fatal(err)
	}
}
