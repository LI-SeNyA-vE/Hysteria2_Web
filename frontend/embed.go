// Package frontend встраивает собранный React-фронтенд (dist) в бинарь панели.
// Файл лежит в frontend/, потому что go:embed не умеет ссылаться на "..".
package frontend

import "embed"

//go:embed all:dist
var Dist embed.FS
