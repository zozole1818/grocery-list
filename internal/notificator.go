package internal

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Notificator struct {
	clients map[*websocket.Conn]int
	mu      sync.Mutex
	in      chan Item
	Out     chan Item
}

func NewNotificator() *Notificator {
	return &Notificator{
		clients: make(map[*websocket.Conn]int),
		in:      make(chan Item, 10),
		Out:     make(chan Item, 10),
	}
}

func (n *Notificator) Notify(item Item) {
	n.Out <- item
}

func (n *Notificator) AddClient(conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.clients[conn] = 1
}

func (n *Notificator) RemoveClient(conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.clients, conn)
}

func (n *Notificator) GetClients() map[*websocket.Conn]int {
	n.mu.Lock()
	defer n.mu.Unlock()
	clientCopy := make(map[*websocket.Conn]int, len(n.clients))
	for k, v := range n.clients {
		clientCopy[k] = v
	}
	return clientCopy
}
