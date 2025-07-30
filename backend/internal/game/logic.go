package game

import "errors"

var (
	ErrInvalidMove     = errors.New("invalid move")
	ErrGameOver        = errors.New("game is over")
	ErrWrongBoard      = errors.New("must play in specified board")
	ErrPositionTaken   = errors.New("position already taken")
)

// NewGame creates a new Ultimate Tic-Tac-Toe game
func NewGame() *GameState {
	return &GameState{
		BigBoard:      BigBoard{},
		BigBoardWins:  BigBoardWins{},
		ActiveBoard:   -1, // Can play anywhere initially
		CurrentPlayer: X,
		GameWon:       Empty,
		GameOver:      false,
	}
}

// IsValidMove checks if a move is valid
func (g *GameState) IsValidMove(move Move) error {
	if g.GameOver {
		return ErrGameOver
	}

	if move.Player != g.CurrentPlayer {
		return ErrInvalidMove
	}

	// Check if must play in specific board
	if g.ActiveBoard != -1 && move.BigBoardIndex != g.ActiveBoard {
		return ErrWrongBoard
	}

	// Check if big board position is already won
	if g.BigBoardWins[move.BigBoardIndex] != Empty {
		// If forced to play in won board, can play anywhere legal
		if g.ActiveBoard == move.BigBoardIndex {
			return g.findLegalMove(move)
		}
		return ErrInvalidMove
	}

	// Check if small board position is taken
	if g.BigBoard[move.BigBoardIndex][move.SmallBoardIndex] != Empty {
		return ErrPositionTaken
	}

	return nil
}

// findLegalMove finds if there's any legal move when sent to won board
func (g *GameState) findLegalMove(move Move) error {
	// Check if the intended move is in a legal small board
	for i := 0; i < 9; i++ {
		if g.BigBoardWins[i] == Empty {
			for j := 0; j < 9; j++ {
				if g.BigBoard[i][j] == Empty {
					// There's at least one legal move, allow any legal move
					if move.BigBoardIndex == i && g.BigBoard[i][move.SmallBoardIndex] == Empty {
						return nil
					}
				}
			}
		}
	}
	return ErrInvalidMove
}

// MakeMove applies a move to the game state
func (g *GameState) MakeMove(move Move) error {
	if err := g.IsValidMove(move); err != nil {
		return err
	}

	// Apply the move
	g.BigBoard[move.BigBoardIndex][move.SmallBoardIndex] = move.Player

	// Check if small board is won
	if winner := checkSmallBoardWin(&g.BigBoard[move.BigBoardIndex]); winner != Empty {
		g.BigBoardWins[move.BigBoardIndex] = winner
	}

	// Check if game is won
	if winner := checkBigBoardWin(&g.BigBoardWins); winner != Empty {
		g.GameWon = winner
		g.GameOver = true
		return nil
	}

	// Check for tie
	if g.isBoardFull() {
		g.GameOver = true
		return nil
	}

	// Set next active board
	if g.BigBoardWins[move.SmallBoardIndex] != Empty || g.isSmallBoardFull(move.SmallBoardIndex) {
		g.ActiveBoard = -1 // Can play anywhere
	} else {
		g.ActiveBoard = move.SmallBoardIndex
	}

	// Switch players
	if g.CurrentPlayer == X {
		g.CurrentPlayer = O
	} else {
		g.CurrentPlayer = X
	}

	return nil
}

// checkSmallBoardWin checks if a 3x3 board has a winner
func checkSmallBoardWin(board *SmallBoard) Player {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i*3] != Empty && board[i*3] == board[i*3+1] && board[i*3+1] == board[i*3+2] {
			return board[i*3]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[i] != Empty && board[i] == board[i+3] && board[i+3] == board[i+6] {
			return board[i]
		}
	}

	// Check diagonals
	if board[0] != Empty && board[0] == board[4] && board[4] == board[8] {
		return board[0]
	}
	if board[2] != Empty && board[2] == board[4] && board[4] == board[6] {
		return board[2]
	}

	return Empty
}

// checkBigBoardWin checks if the big board has a winner
func checkBigBoardWin(wins *BigBoardWins) Player {
	// Check rows
	for i := 0; i < 3; i++ {
		if wins[i*3] != Empty && wins[i*3] == wins[i*3+1] && wins[i*3+1] == wins[i*3+2] {
			return wins[i*3]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if wins[i] != Empty && wins[i] == wins[i+3] && wins[i+3] == wins[i+6] {
			return wins[i]
		}
	}

	// Check diagonals
	if wins[0] != Empty && wins[0] == wins[4] && wins[4] == wins[8] {
		return wins[0]
	}
	if wins[2] != Empty && wins[2] == wins[4] && wins[4] == wins[6] {
		return wins[2]
	}

	return Empty
}

// isBoardFull checks if all legal moves are exhausted
func (g *GameState) isBoardFull() bool {
	for i := 0; i < 9; i++ {
		if g.BigBoardWins[i] == Empty && !g.isSmallBoardFull(i) {
			return false
		}
	}
	return true
}

// isSmallBoardFull checks if a small board is full
func (g *GameState) isSmallBoardFull(boardIndex int) bool {
	for i := 0; i < 9; i++ {
		if g.BigBoard[boardIndex][i] == Empty {
			return false
		}
	}
	return true
}