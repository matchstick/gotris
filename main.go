package main

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
	rand.Seed(time.Now().UnixNano())

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

	log.Println("Started ticker with speed ", g.speed)
	g.ticker = time.NewTicker(g.speed)

	// Send initial state
	log.Println("Sending Initial State")
	g.SendState()

	// Start game loop
	go func() {
		for {
			select {
			case <-g.ticker.C:
				if !g.state.GameOver {
					g.MovePiece(Down)
					g.SendState()
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

		log.Println("Message Read")
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
			g.SendState()

		case NewGame:
			log.Println("New Game Recvd")
			g.Reset()
			if g.ticker != nil {
				log.Println("TICK STOPPED")
				g.ticker.Stop()
			}
			log.Println("Sending New game")
			g.ticker = time.NewTicker(g.speed)
			g.SendState()
		}
	}

	log.Println("hey")
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

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	// Set up static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Handle WebSocket connection
	http.HandleFunc("/ws", handleWebSocket)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
