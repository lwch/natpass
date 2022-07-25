package conn

import (
	"net/http"
	"time"

	"github.com/lwch/natpass/code/network"
	"google.golang.org/protobuf/proto"
)

func makeCodeHeader(header http.Header) map[string]*network.CodeHeaderValues {
	ret := make(map[string]*network.CodeHeaderValues)
	for key, values := range header {
		data := make([]string, len(values))
		copy(data, values)
		ret[key] = &network.CodeHeaderValues{
			Values: data,
		}
	}
	return ret
}

// SendCodeRequest send request
func (conn *Conn) SendCodeRequest(to, linkID string, requestID uint64,
	method, uri string, body []byte, header http.Header) uint64 {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_code_request
	msg.LinkId = linkID
	msg.Payload = &network.Msg_Csreq{
		Csreq: &network.CodeRequest{
			RequestId: requestID,
			Method:    method,
			Uri:       uri,
			Body:      body,
			Header:    makeCodeHeader(header),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		return 0
	}
}

// SendCodeResponseHeader send response header
func (conn *Conn) SendCodeResponseHeader(to, linkID string, requestID uint64,
	code uint32, header http.Header) uint64 {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_code_response_hdr
	msg.LinkId = linkID
	msg.Payload = &network.Msg_CsrepHdr{
		CsrepHdr: &network.CodeResponseHeader{
			RequestId: requestID,
			Code:      code,
			Header:    makeCodeHeader(header),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		return 0
	}
}

// SendCodeResponseBody send response body
func (conn *Conn) SendCodeResponseBody(to, linkID string, requestID uint64,
	idx uint32, ok, done bool, data []byte) uint64 {
	var mask uint32
	if ok {
		mask |= 1
	}
	if done {
		mask |= 2
	}
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_code_response_body
	msg.LinkId = linkID
	msg.Payload = &network.Msg_CsrepBody{
		CsrepBody: &network.CodeResponseBody{
			RequestId: requestID,
			Index:     idx,
			Mask:      mask,
			Body:      data,
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		return 0
	}
}
