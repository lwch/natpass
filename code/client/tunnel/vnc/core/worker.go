package core

import "github.com/gorilla/websocket"

type DesktopInfo struct {
	Bits   int
	Width  int
	Height int
}

type Context struct {
	ctxOsBased
	Info DesktopInfo
}

func NewContext() *Context {
	ctx := &Context{}
	err := ctx.init()
	if err != nil {
		return nil
	}
	return ctx
}

func (ctx *Context) Do(conn *websocket.Conn) {
	defer conn.Close()
}
