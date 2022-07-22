package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/lwch/natpass/code/client/rule"
)

// Info information data
func (db *Dashboard) Info(w http.ResponseWriter, r *http.Request) {
	var ret struct {
		Rules        int `json:"rules"`
		VirtualLinks int `json:"virtual_links"`
		Session      int `json:"sessions"`
	}
	ret.Rules = len(db.cfg.Rules)
	db.mgr.Range(func(t rule.Rule) {
		lr, ok := t.(rule.LinkedRule)
		if ok {
			n := len(lr.GetLinks())
			ret.VirtualLinks += n
			if t.GetTypeName() == "shell" ||
				t.GetTypeName() == "vnc" {
				ret.Session += n
			}
		}
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
