package ws

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type WSConnHub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func New() *WSConnHub {
	return &WSConnHub{conns: make(map[string]*websocket.Conn)}
}

func (ws *WSConnHub) Register(fileID string, con *websocket.Conn) {
	ws.mu.Lock()
	ws.conns[fileID] = con
	ws.mu.Unlock()
}

func (ws *WSConnHub) Delete(fileID string, filename string) error {
	ws.mu.Lock()
	conn := ws.conns[fileID]
	delete(ws.conns, fileID)
	ws.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("There is no such connection")
	}
	msg := fmt.Sprintf("Your cards from %v are ready", filename)
	err := conn.WriteJSON(map[string]string{"message": msg})
	if err != nil {
		conn.Close()
		return err
	}
	return conn.Close()
}
