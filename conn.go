package main

import (
    "github.com/gorilla/websocket"
    "log"
    "net/http"
    "time"
)

const (

    // Time allowed to read the next pong message from the peer.
    pongWait = 60 * time.Second

    // Send pings to peer with this period. Must be less than pongWait.
    pingPeriod = (pongWait * 9) / 10
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
    // The websocket connection.
    ws  *websocket.Conn

    // Buffered channel of outbound messages.
    send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
    defer func() {
        h.unregister <- c
        c.ws.Close()
    }()
    c.ws.SetReadLimit(Conf.Websocket.MaxMessageSize)
    c.ws.SetReadDeadline(time.Now().Add(pongWait))
    c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
    for {
        _, message, err := c.ws.ReadMessage()
        if err != nil {
            break
        }
        h.broadcast <- message
    }
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
    c.ws.SetWriteDeadline(time.Now().Add(time.Duration(Conf.Websocket.WriteWait) * time.Second))
    return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.ws.Close()
    }()
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.write(websocket.CloseMessage, []byte{})
                return
            }
            if err := c.write(websocket.TextMessage, message); err != nil {
                return
            }
        case <-ticker.C:
            if err := c.write(websocket.PingMessage, []byte{}); err != nil {
                return
            }
        }
    }
}

// serverWs handles webocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", 405)
        return
    }
    if r.Header.Get("Origin") != "http://"+r.Host {
        http.Error(w, "Origin not allowed", 403)
        return
    }
    ws, err := websocket.Upgrade(w, r, nil, Conf.Websocket.ReadBufSize, Conf.Websocket.WriteBufSize)
    if _, ok := err.(websocket.HandshakeError); ok {
        http.Error(w, "Not a websocket handshake", 400)
        return
    } else if err != nil {
        log.Println(err)
        return
    }
    c := &connection{send: make(chan []byte, 256), ws: ws}
    h.register <- c
    go c.writePump()
    c.readPump()
}
