package game

// Player represents a player in the game
type Player int

const (
	Empty Player = iota
	X
	O
)

func (p Player) String() string {
	switch p {
	case X:
		return "X"
	case O:
		return "O"
	default:
		return " "
	}
}

// SmallBoard represents a 3x3 tic-tac-toe board
type SmallBoard [9]Player

// BigBoard represents the 3x3 grid of small boards
type BigBoard [9]SmallBoard

// BigBoardWins tracks who won each small board
type BigBoardWins [9]Player

// GameState represents the complete state of an Ultimate Tic-Tac-Toe game
type GameState struct {
	BigBoard      BigBoard     `json:"bigBoard"`
	BigBoardWins  BigBoardWins `json:"bigBoardWins"`
	ActiveBoard   int          `json:"activeBoard"`   // Which small board must be played (-1 = any)
	CurrentPlayer Player       `json:"currentPlayer"`
	GameWon       Player       `json:"gameWon"`
	GameOver      bool         `json:"gameOver"`
}

// Move represents a player's move
type Move struct {
	BigBoardIndex   int    `json:"bigBoardIndex"`
	SmallBoardIndex int    `json:"smallBoardIndex"`
	Player          Player `json:"player"`
}