package code

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lwch/natpass/code/client/conn"
)

// Render render code-server
func (code *Code) Render(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	dir := strings.TrimPrefix(r.URL.Path, "/")
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
	data, _ = Asset("index.html")
	tpl, err := template.New("all").Parse(string(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, code)
}
