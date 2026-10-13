// Package data は本体の定義データを埋め込む（DES-PKG-06）。
package data

import "embed"

// Base は本体の定義データを含む（DES-PKG-06）。
//
//go:embed base/*.toml
var Base embed.FS
