package core

import "github.com/gorilla/websocket"

type desktopInfo struct {
	bits   int
	width  int
	height int
}

type Worker struct {
	workerOsBased
	info desktopInfo
}

func NewWorker() *Worker {
	worker := &Worker{}
	err := worker.init()
	if err != nil {
		return nil
	}
	return worker
}

func (worker *Worker) Do(conn *websocket.Conn) {
	defer conn.Close()
	// TODO: handle msg
}
