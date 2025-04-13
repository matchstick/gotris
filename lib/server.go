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
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// Game constants
const (
	BoardWidth  = 10
	BoardHeight = 20
)

// TetrominoType represents different shapes
type TetrominoType int

const (
	I TetrominoType = iota
	J
	L
	O
	S
	T
	Z
)

// Direction for movement
type Direction int

const (
	Left Direction = iota
	Right
	Down
	Rotate
)

// Tetromino represents a tetris piece
type Tetromino struct {
	Type     TetrominoType `json:"type"`
	X        int           `json:"x"`
	Y        int           `json:"y"`
	Rotation int           `json:"rotation"`
}

// GameState represents the current state of the game
type GameState struct {
	Board        [BoardHeight][BoardWidth]int `json:"board"`
	CurrentPiece Tetromino                    `json:"current_piece"`
	NextPiece    TetrominoType                `json:"next_piece"`
	Score        int                          `json:"score"`
	Level        int                          `json:"level"`
	LinesCleared int                          `json:"lines_cleared"`
	GameOver     bool                         `json:"game_over"`
}

// Message types for websocket communication
type MessageType string

const (
	StateUpdate MessageType = "state_update"
	Move        MessageType = "move"
	NewGame     MessageType = "new_game"
	GameOverMsg MessageType = "game_over"
)

// Message is the websocket message format
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Message is the websocket message format
type RecvMessage struct {
	Type    MessageType `json:"type"`
	Payload Direction   `json:"payload"`
}

// Websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// Tetromino shapes - each shape has 4 rotations
var tetrominoShapes = map[TetrominoType][4][][]int{
	I: {
		{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		{{1, -1}, {1, 0}, {1, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
		{{2, -1}, {2, 0}, {2, 1}, {2, 2}},
	},
	J: {
		{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
		{{1, 0}, {2, 0}, {1, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {2, 2}},
		{{1, 0}, {1, 1}, {0, 2}, {1, 2}},
	},
	L: {
		{{2, 0}, {0, 1}, {1, 1}, {2, 1}},
		{{1, 0}, {1, 1}, {1, 2}, {2, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {0, 2}},
		{{0, 0}, {1, 0}, {1, 1}, {1, 2}},
	},
	O: {
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}},
	},
	S: {
		{{1, 0}, {2, 0}, {0, 1}, {1, 1}},
		{{1, 0}, {1, 1}, {2, 1}, {2, 2}},
		{{1, 1}, {2, 1}, {0, 2}, {1, 2}},
		{{0, 0}, {0, 1}, {1, 1}, {1, 2}},
	},
	T: {
		{{1, 0}, {0, 1}, {1, 1}, {2, 1}},
		{{1, 0}, {1, 1}, {2, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {1, 2}},
		{{1, 0}, {0, 1}, {1, 1}, {1, 2}},
	},
	Z: {
		{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
		{{2, 0}, {1, 1}, {2, 1}, {1, 2}},
		{{0, 1}, {1, 1}, {1, 2}, {2, 2}},
		{{1, 0}, {0, 1}, {1, 1}, {0, 2}},
	},
}

// Game represents a single player's game session
type Game struct {
	state  GameState
	conn   *websocket.Conn
	ticker *time.Ticker
	done   chan bool
	speed  time.Duration
}

// NewGame creates a new game instance
func MakeNewGame(conn *websocket.Conn) *Game {
	fmt.Println("New Game clicked")
	game := &Game{
		conn:  conn,
		done:  make(chan bool),
		speed: 800 * time.Millisecond, // Starting speed
	}

	game.Reset()
	return game
}

// Reset the game to starting state
func (g *Game) Reset() {
	g.state = GameState{
		Level:        1,
		Score:        0,
		LinesCleared: 0,
		GameOver:     false,
	}

	// Clear the board
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			g.state.Board[y][x] = 0
		}
	}

	// Generate first pieces
	g.state.NextPiece = TetrominoType(rand.Intn(7))
	g.SpawnNewPiece()
}

// SpawnNewPiece creates a new tetromino at the top of the board
func (g *Game) SpawnNewPiece() {
	g.state.CurrentPiece = Tetromino{
		Type:     g.state.NextPiece,
		X:        BoardWidth/2 - 1,
		Y:        0,
		Rotation: 0,
	}

	// Generate next piece
	g.state.NextPiece = TetrominoType(rand.Intn(7))

	// Check if the new piece can be placed - if not, game over
	if !g.isValidPosition(g.state.CurrentPiece) {
		g.state.GameOver = true
	}
}

// isValidPosition checks if a tetromino's position is valid
func (g *Game) isValidPosition(t Tetromino) bool {
	shape := tetrominoShapes[t.Type][t.Rotation]

	for _, block := range shape {
		x := t.X + block[0]
		y := t.Y + block[1]

		// Check boundaries
		if x < 0 || x >= BoardWidth || y < 0 || y >= BoardHeight {
			return false
		}

		// Check collision with existing blocks
		if y >= 0 && g.state.Board[y][x] != 0 {
			return false
		}
	}

	return true
}

// MovePiece tries to move the current piece
func (g *Game) MovePiece(dir Direction) bool {
	// Create a copy of current piece
	newPiece := g.state.CurrentPiece

	switch dir {
	case Left:
		newPiece.X--
	case Right:
		newPiece.X++
	case Down:
		newPiece.Y++
	case Rotate:
		newPiece.Rotation = (newPiece.Rotation + 1) % 4
	}

	// Check if new position is valid
	if g.isValidPosition(newPiece) {
		g.state.CurrentPiece = newPiece
		return true
	}

	// If down movement is blocked, lock the piece in place
	if dir == Down {
		g.LockPiece()
		return false
	}

	return false
}

// LockPiece fixes the current piece to the board
func (g *Game) LockPiece() {
	// Add the piece to the board
	shape := tetrominoShapes[g.state.CurrentPiece.Type][g.state.CurrentPiece.Rotation]

	for _, block := range shape {
		x := g.state.CurrentPiece.X + block[0]
		y := g.state.CurrentPiece.Y + block[1]

		if x >= 0 && x < BoardWidth && y >= 0 && y < BoardHeight {
			g.state.Board[y][x] = int(g.state.CurrentPiece.Type) + 1
		}
	}

	// Check for completed lines
	linesCleared := g.ClearLines()

	// Update score
	if linesCleared > 0 {
		g.UpdateScore(linesCleared)
	}

	// Spawn new piece
	g.SpawnNewPiece()
}

// ClearLines checks and clears completed lines
func (g *Game) ClearLines() int {
	linesCleared := 0

	for y := BoardHeight - 1; y >= 0; y-- {
		// Check if line is full
		full := true
		for x := 0; x < BoardWidth; x++ {
			if g.state.Board[y][x] == 0 {
				full = false
				break
			}
		}

		if full {
			// Clear this line
			linesCleared++

			// Move all lines above down
			for y2 := y; y2 > 0; y2-- {
				for x := 0; x < BoardWidth; x++ {
					g.state.Board[y2][x] = g.state.Board[y2-1][x]
				}
			}

			// Clear top line
			for x := 0; x < BoardWidth; x++ {
				g.state.Board[0][x] = 0
			}

			// Check the same line again after shifting
			y++
		}
	}

	return linesCleared
}

// UpdateScore calculates new score based on lines cleared
func (g *Game) UpdateScore(linesCleared int) {
	// Classic Tetris scoring
	basePoints := 0
	switch linesCleared {
	case 1:
		basePoints = 40
	case 2:
		basePoints = 100
	case 3:
		basePoints = 300
	case 4:
		basePoints = 1200
	}

	g.state.Score += basePoints * g.state.Level
	g.state.LinesCleared += linesCleared

	// Level up every 10 lines
	newLevel := (g.state.LinesCleared / 10) + 1
	if newLevel > g.state.Level {
		g.state.Level = newLevel
		g.speed = time.Duration(800-((newLevel-1)*50)) * time.Millisecond
		if g.speed < 100*time.Millisecond {
			g.speed = 100 * time.Millisecond
		}

		// Reset ticker with new speed
		if g.ticker != nil {
			g.ticker.Stop()
			g.ticker = time.NewTicker(g.speed)
		}
	}
}

// SendState sends the current game state to the client
func (g *Game) SendState() error {
	stateJSON, err := json.Marshal(g.state)
	if err != nil {
		return err
	}

	msg := Message{
		Type:    StateUpdate,
		Payload: stateJSON,
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return g.conn.WriteMessage(websocket.TextMessage, msgJSON)
}

// Start begins the game loop
func (g *Game) Start() {
	g.ticker = time.NewTicker(g.speed)

	// Send initial state
	err := g.SendState()
	if err != nil {
		log.Printf("Error in Start when sending State %v\n", err)
		os.Exit(1)
	}

	// Start game loop
	go func() {
		for {
			select {
			case <-g.ticker.C:
				if !g.state.GameOver {
					g.MovePiece(Down)
					err := g.SendState()
					if err != nil {
						log.Printf("Error in Start when sending State %v\n", err)
						os.Exit(1)
					}

				}
			case <-g.done:
				g.ticker.Stop()
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Listen for client messages
	for {
		_, rawMessage, err := g.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var message RecvMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Println("JSON error:", err)
			continue
		}

		switch message.Type {
		case Move:
			if g.state.GameOver {
				continue
			}

			g.MovePiece(message.Payload)
			err := g.SendState()
			if err != nil {
				log.Printf("Error in Start when sending State %v\n", err)
				os.Exit(1)
			}

		case NewGame:
			g.Reset()
			if g.ticker != nil {
				g.ticker.Stop()
			}
			g.ticker = time.NewTicker(g.speed)
			err := g.SendState()
			if err != nil {
				log.Printf("Error in Start when sending State %v\n", err)
				os.Exit(1)
			}
		}
	}

	g.done <- true
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	game := MakeNewGame(conn)
	game.Start()
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
