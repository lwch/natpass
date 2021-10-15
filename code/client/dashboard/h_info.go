package dashboard

import (
	"encoding/json"
	"natpass/code/client/tunnel"
	"net/http"
)

// Info information data
func (db *Dashboard) Info(w http.ResponseWriter, r *http.Request) {
	var ret struct {
		Tunnels       int `json:"tunnels"`
		PhysicalLinks int `json:"physical_links"`
		VirtualLinks  int `json:"virtual_links"`
		Session       int `json:"sessions"`
	}
	ret.Tunnels = len(db.cfg.Tunnels)
	ret.PhysicalLinks = db.pl.Size()
	db.mgr.Range(func(t tunnel.Tunnel) {
		n := len(t.GetLinks())
		ret.VirtualLinks += n
		if t.GetTypeName() == "shell" {
			ret.Session += n
		}
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
