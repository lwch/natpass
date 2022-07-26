package conn

import (
	"net/http"
	"time"

	"github.com/lwch/logging"
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
			Body:      dup(body),
			Header:    makeCodeHeader(header),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		logging.Info("send: droped %s", network.Msg_code_request.String())
		return 0
	}
}

// SendCodeConnect send connect
func (conn *Conn) SendCodeConnect(to, linkID string, requestID uint64,
	uri string, header http.Header) uint64 {
	var msg network.Msg
	msg.To = to
	msg.XType = network.Msg_code_connect
	msg.LinkId = linkID
	msg.Payload = &network.Msg_Csconn{
		Csconn: &network.CodeConnect{
			RequestId: requestID,
			Uri:       uri,
			Header:    makeCodeHeader(header),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		logging.Info("send: droped %s", network.Msg_code_connect.String())
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
		logging.Info("send: droped %s", network.Msg_code_response_hdr.String())
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
			Body:      dup(data),
		},
	}
	select {
	case conn.write <- &msg:
		data, _ := proto.Marshal(&msg)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		logging.Info("send: droped %s", network.Msg_code_response_body.String())
		return 0
	}
}

// SendCodeResponseConnect send response connect
func (conn *Conn) SendCodeResponseConnect(to, linkID string, requestID uint64,
	ok bool, msg string, header http.Header) uint64 {
	var m network.Msg
	m.To = to
	m.XType = network.Msg_code_connect_response
	m.LinkId = linkID
	m.Payload = &network.Msg_CsconnRep{
		CsconnRep: &network.CodeConnectResponse{
			RequestId: requestID,
			Ok:        ok,
			Msg:       msg,
			Header:    makeCodeHeader(header),
		},
	}
	select {
	case conn.write <- &m:
		data, _ := proto.Marshal(&m)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		logging.Info("send: droped %s", network.Msg_code_connect_response.String())
		return 0
	}
}

// SendCodeData send data
func (conn *Conn) SendCodeData(to, linkID string, requestID uint64,
	ok bool, t int, body []byte) uint64 {
	var m network.Msg
	m.To = to
	m.XType = network.Msg_code_data
	m.LinkId = linkID
	m.Payload = &network.Msg_Csdata{
		Csdata: &network.CodeData{
			RequestId: requestID,
			Ok:        ok,
			Type:      uint32(t),
			Data:      dup(body),
		},
	}
	select {
	case conn.write <- &m:
		data, _ := proto.Marshal(&m)
		return uint64(len(data))
	case <-time.After(conn.cfg.WriteTimeout):
		logging.Info("send: droped %s", network.Msg_code_data.String())
		return 0
	}
}
