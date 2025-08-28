package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Notificator struct {
	clients map[*websocket.Conn]string
	mu      sync.Mutex
	in      chan Item
	Out     chan Item
}

func NewNotificator() *Notificator {
	return &Notificator{
		clients: make(map[*websocket.Conn]string),
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
	n.clients[conn] = conn.RemoteAddr().String()
	fmt.Printf("Client added %s\n", conn.RemoteAddr().String())
}

func (n *Notificator) RemoveClient(conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.clients, conn)
	fmt.Printf("Client removed %s\n", conn.RemoteAddr().String())
}

func (n *Notificator) GetClients() map[*websocket.Conn]string {
	n.mu.Lock()
	defer n.mu.Unlock()
	clientCopy := make(map[*websocket.Conn]string, len(n.clients))
	for k, v := range n.clients {
		clientCopy[k] = v
	}
	fmt.Printf("Clients to broadcast: %v\n", clientCopy)
	return clientCopy
}
