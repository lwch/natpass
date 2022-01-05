package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/jkstack/natpass/code/client/rule"
)

// Info information data
func (db *Dashboard) Info(w http.ResponseWriter, r *http.Request) {
	var ret struct {
		Rules         int `json:"rules"`
		PhysicalLinks int `json:"physical_links"`
		VirtualLinks  int `json:"virtual_links"`
		Session       int `json:"sessions"`
	}
	ret.Rules = len(db.cfg.Rules)
	ret.PhysicalLinks = db.pl.Size()
	db.mgr.Range(func(t rule.Rule) {
		n := len(t.GetLinks())
		ret.VirtualLinks += n
		if t.GetTypeName() == "shell" {
			ret.Session += n
		}
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
