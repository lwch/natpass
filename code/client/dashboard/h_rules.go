package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/lwch/natpass/code/client/rule"
)

// Rules get rule list
func (db *Dashboard) Rules(w http.ResponseWriter, r *http.Request) {
	type link struct {
		ID         string `json:"id"`
		SendBytes  uint64 `json:"send_bytes"`
		SendPacket uint64 `json:"send_packet"`
		RecvBytes  uint64 `json:"recv_bytes"`
		RecvPacket uint64 `json:"recv_packet"`
	}
	type item struct {
		Name   string `json:"name"`
		Remote string `json:"remote"`
		Port   uint16 `json:"port"`
		Type   string `json:"type"`
		Links  []link `json:"links"`
	}
	var ret []item
	db.mgr.Range(func(t rule.Rule) {
		var it item
		it.Name = t.GetName()
		it.Remote = t.GetRemote()
		it.Port = t.GetPort()
		it.Type = t.GetTypeName()
		for _, l := range t.GetLinks() {
			var lk link
			lk.ID = l.GetID()
			lk.RecvBytes, lk.SendBytes = l.GetBytes()
			lk.RecvPacket, lk.SendPacket = l.GetPackets()
			it.Links = append(it.Links, lk)
		}
		ret = append(ret, it)
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
