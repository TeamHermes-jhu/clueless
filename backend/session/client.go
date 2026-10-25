package session

import (
	"context"

	"github.com/gorilla/websocket"
)

const clientSendBuffer = 8

// Client represents a single connected client.
type Client struct {
	id      string
	conn    *websocket.Conn
	session *Session
	send    chan []byte
}

// NewClient creates a client with the id, websocket connection and session provided.
func NewClient(id string, conn *websocket.Conn, sess *Session) *Client {
	return &Client{
		id:      id,
		conn:    conn,
		session: sess,
		send:    make(chan []byte, clientSendBuffer),
	}
}

// Read reads messages coming from a client and
// sends them to the session.
func (c *Client) Read() {
	defer func() {
		_ = c.conn.Close()
		c.session.unregister <- c.id
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		c.session.inbound <- inboundMessage{from: c.id, data: data}
	}
}

// Write sends messages to the client.
func (c *Client) Write() {
	defer func() { _ = c.conn.Close() }()

	for data := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

func (c *Client) Close(_ context.Context) {
	c.conn.Close()
}
