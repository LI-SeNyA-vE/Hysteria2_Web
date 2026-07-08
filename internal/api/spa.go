package api

import (
	"io/fs"
	"net/http"
	"strings"

	"hysteria2-web/frontend"
)

// spaHandler раздаёт встроенный frontend/dist с fallback на index.html
// для любых путей, не начинающихся на /api или /sub.
func spaHandler() http.Handler {
	dist, err := fs.Sub(frontend.Dist, "dist")
	if err != nil {
		panic("spa: не удалось открыть frontend/dist: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// API и sub-ы не должны попадать сюда (chi разберёт раньше),
		// но добавим защиту на случай прямого использования хендлера.
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/sub") {
			http.NotFound(w, r)
			return
		}

		// Пробуем отдать файл как есть.
		f, err := dist.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Файл не найден — отдаём index.html (SPA routing).
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}
