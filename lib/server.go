// This code was generated with assistance from Claude AI by Anthropic.
// It is provided under the MIT License, which allows for free use, modification,
// and distribution with proper attribution.
//
// MIT License
//
// Copyright (c) [2025] [Michael Rubin]
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gotris

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type session struct {
	conn *websocket.Conn
	id   string
	game *game
}

type sessionManager struct {
	sessions      map[string]*session
	mutex         sync.RWMutex
	readySessions chan *session
}

var registry = sessionManager{
	sessions:      make(map[string]*session),
	readySessions: make(chan *session),
}

func registerSession(c *websocket.Conn, id string) {
	s := &session{
		conn: c,
		id:   id,
		game: MakeNewGame(c, id),
	}

	registry.mutex.Lock()
	_, exists := registry.sessions[id]
	if !exists {
		registry.sessions[id] = s
	} else {
		log.Printf("Session registered with same id %s\n", id)
		os.Exit(1)
	}
	registry.mutex.Unlock()

	registry.readySessions <- s

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query params or cookie
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = "anonymous-" + r.RemoteAddr // Fallback
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	registerSession(conn, sessionID)
}

func NewServer(port int, numPlayers int) error {
	// Set up static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Handle WebSocket connection
	http.HandleFunc("/ws", handleWebSocket)

	// Start server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return fmt.Errorf("Server error: %w", err)
	}

	return nil
}
