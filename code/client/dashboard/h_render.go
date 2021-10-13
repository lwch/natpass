package dashboard

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Render render asset file
func (db *Dashboard) Render(w http.ResponseWriter, r *http.Request) {
	dir := strings.TrimPrefix(r.URL.Path, "/")
	if filepath.Ext(dir) == ".html" {
		db.renderHtml(w, r, dir)
		return
	}
	data, err := Asset(dir)
	if err == nil {
		ctype := mime.TypeByExtension(filepath.Ext(dir))
		if ctype == "" {
			ctype = http.DetectContentType(data)
		}
		w.Header().Set("Content-Type", ctype)
		io.Copy(w, bytes.NewReader(data))
		return
	}
	db.renderHtml(w, r, "index.html")
}

func (db *Dashboard) renderHtml(w http.ResponseWriter, r *http.Request, name string) {
	data, err := Asset(name)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	header, _ := Asset("templates/header.html")
	aside, _ := Asset("templates/aside.html")
	footer, _ := Asset("templates/footer.html")

	tpl := template.New("all")
	tpl, err = tpl.Parse(string(header))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl, err = tpl.Parse(string(aside))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl, err = tpl.Parse(string(footer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl, err = tpl.Parse(string(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.Execute(w, db)
}
