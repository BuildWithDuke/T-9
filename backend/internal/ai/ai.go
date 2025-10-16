package ai

import (
	"math"
	"math/rand"
	"t-9/internal/game"
	"time"
)

// Difficulty levels for AI
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

// AIPlayer represents an AI opponent
type AIPlayer struct {
	difficulty Difficulty
	player     game.Player
	rng        *rand.Rand
}

// NewAIPlayer creates a new AI player
func NewAIPlayer(difficulty Difficulty, player game.Player) *AIPlayer {
	return &AIPlayer{
		difficulty: difficulty,
		player:     player,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetBestMove returns the best move for the current game state
func (ai *AIPlayer) GetBestMove(gameState *game.GameState) game.Move {
	// First check if there are any valid moves
	moves := ai.getValidMoves(gameState)
	if len(moves) == 0 {
		// Return empty move if no valid moves
		return game.Move{}
	}

	switch ai.difficulty {
	case Easy:
		return ai.getRandomMove(gameState)
	case Medium:
		return ai.getMediumMove(gameState)
	case Hard:
		return ai.getHardMove(gameState)
	default:
		return ai.getRandomMove(gameState)
	}
}

// getRandomMove returns a random valid move
func (ai *AIPlayer) getRandomMove(gameState *game.GameState) game.Move {
	moves := ai.getValidMoves(gameState)
	if len(moves) == 0 {
		return game.Move{}
	}
	return moves[ai.rng.Intn(len(moves))]
}

// getMediumMove uses strategic thinking with limited lookahead
func (ai *AIPlayer) getMediumMove(gameState *game.GameState) game.Move {
	moves := ai.getValidMoves(gameState)
	if len(moves) == 0 {
		return game.Move{}
	}

	// Use minimax but with limited depth for medium difficulty
	// 80% chance of strategic play, 20% chance of good-but-not-perfect move
	if ai.rng.Float64() < 0.8 {
		return ai.minimax(gameState, 4, math.Inf(-1), math.Inf(1), true).move
	} else {
		return ai.getStrategicMove(gameState, moves)
	}
}

// getHardMove uses advanced minimax algorithm with opening theory
func (ai *AIPlayer) getHardMove(gameState *game.GameState) game.Move {
	// Check if we're in opening - use opening theory
	totalMoves := ai.countTotalMoves(gameState)
	if totalMoves <= 6 {
		openingMove := ai.getOpeningMove(gameState)
		if openingMove.Player != 0 {
			return openingMove
		}
	}
	
	// Use deeper search for hard difficulty
	return ai.minimax(gameState, 10, math.Inf(-1), math.Inf(1), true).move
}

// MinimaxResult holds the result of minimax computation
type MinimaxResult struct {
	score float64
	move  game.Move
}

// minimax implements the minimax algorithm with alpha-beta pruning
func (ai *AIPlayer) minimax(gameState *game.GameState, depth int, alpha, beta float64, maximizing bool) MinimaxResult {
	if depth == 0 || gameState.GameOver {
		return MinimaxResult{
			score: ai.evaluatePosition(gameState),
			move:  game.Move{},
		}
	}

	moves := ai.getValidMoves(gameState)
	if len(moves) == 0 {
		return MinimaxResult{
			score: ai.evaluatePosition(gameState),
			move:  game.Move{},
		}
	}

	bestMove := moves[0]

	if maximizing {
		maxScore := math.Inf(-1)
		for _, move := range moves {
			// Create copy of game state
			newState := ai.copyGameState(gameState)
			newState.MakeMove(move)
			
			result := ai.minimax(newState, depth-1, alpha, beta, false)
			
			if result.score > maxScore {
				maxScore = result.score
				bestMove = move
			}
			
			alpha = math.Max(alpha, result.score)
			if beta <= alpha {
				break // Alpha-beta pruning
			}
		}
		return MinimaxResult{score: maxScore, move: bestMove}
	} else {
		minScore := math.Inf(1)
		for _, move := range moves {
			// Create copy of game state
			newState := ai.copyGameState(gameState)
			newState.MakeMove(move)
			
			result := ai.minimax(newState, depth-1, alpha, beta, true)
			
			if result.score < minScore {
				minScore = result.score
				bestMove = move
			}
			
			beta = math.Min(beta, result.score)
			if beta <= alpha {
				break // Alpha-beta pruning
			}
		}
		return MinimaxResult{score: minScore, move: bestMove}
	}
}

// getStrategicMove implements basic strategic thinking
func (ai *AIPlayer) getStrategicMove(gameState *game.GameState, moves []game.Move) game.Move {
	bestMove := moves[0]
	bestScore := float64(-1000)

	for _, move := range moves {
		score := ai.evaluateMove(gameState, move)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

// evaluateMove gives a score for a specific move using Ultimate TTT strategy
func (ai *AIPlayer) evaluateMove(gameState *game.GameState, move game.Move) float64 {
	score := 0.0
	opponent := game.X
	if ai.player == game.X {
		opponent = game.O
	}

	// Create a copy and make the move
	newState := ai.copyGameState(gameState)
	newState.MakeMove(move)

	// ULTIMATE WIN - highest priority
	if newState.GameWon == ai.player {
		return 10000
	}

	// BLOCK ULTIMATE WIN - second highest priority
	testState := ai.copyGameState(gameState)
	opponentMove := game.Move{
		BigBoardIndex:   move.BigBoardIndex,
		SmallBoardIndex: move.SmallBoardIndex,
		Player:          opponent,
	}
	if testState.IsValidMove(opponentMove) == nil {
		testState.MakeMove(opponentMove)
		if testState.GameWon == opponent {
			score += 5000 // Block opponent ultimate win
		}
	}

	// WIN SMALL BOARD - very important
	if newState.BigBoardWins[move.BigBoardIndex] == ai.player && 
	   gameState.BigBoardWins[move.BigBoardIndex] == game.Empty {
		score += 1000
	}

	// BLOCK OPPONENT SMALL BOARD WIN - very important
	if testState.IsValidMove(opponentMove) == nil {
		testState.MakeMove(opponentMove)
		if testState.BigBoardWins[move.BigBoardIndex] == opponent &&
		   gameState.BigBoardWins[move.BigBoardIndex] == game.Empty {
			score += 800
		}
	}

	// STRATEGIC BOARD CONTROL - this is the key to Ultimate TTT mastery!
	nextBoard := move.SmallBoardIndex
	if nextBoard >= 0 && nextBoard < 9 {
		// Evaluate where we're sending the opponent
		if newState.BigBoardWins[nextBoard] != game.Empty || ai.isSmallBoardFull(newState, nextBoard) {
			// Sending opponent to completed board = anarchy mode = VERY good for us!
			score += 800
		} else {
			// CRITICAL: Analyze the board we're sending opponent to
			sendingScore := ai.evaluateSendingTarget(gameState, newState, nextBoard, opponent)
			score += sendingScore
		}
	}

	// ANTI-CENTER BIAS - don't always default to center!
	if move.SmallBoardIndex == 4 && gameState.BigBoard[move.BigBoardIndex][4] == game.Empty {
		// Check if center is actually strategic here
		centerValue := ai.evaluateCenterMove(gameState, move, opponent)
		score += centerValue
	} else {
		// NON-CENTER MOVES - reward strategic positioning
		score += ai.evaluateNonCenterMove(gameState, move, opponent)
	}

	// BIG BOARD CENTER CONTROL - still important but context-dependent
	if move.BigBoardIndex == 4 { // Center big board
		// Only bonus if we're not helping opponent
		if !ai.isMoveDangerous(gameState, move, opponent) {
			score += 80
		}
	}

	// MULTIPLE THREAT CREATION
	threatsCreated := ai.countThreatsCreated(newState, move.BigBoardIndex)
	score += float64(threatsCreated) * 30

	// FORK OPPORTUNITIES (multiple ways to win)
	forkValue := ai.evaluateForkOpportunities(newState, move.BigBoardIndex)
	score += forkValue

	// ADVANCED CHAIN THINKING - plan 2-3 moves ahead
	chainValue := ai.evaluateChainStrategy(gameState, newState, move, opponent)
	score += chainValue

	// BOARD SACRIFICE STRATEGY - sometimes losing a board is good
	sacrificeValue := ai.evaluateSacrificeStrategy(gameState, newState, move, opponent)
	score += sacrificeValue

	// TEMPO CONTROL - prefer moves that give us options
	if newState.ActiveBoard == -1 { // Created anarchy mode
		score += 150
	}

	// ENDGAME STRATEGY - different logic when few moves left
	if ai.isEndgame(gameState) {
		endgameBonus := ai.evaluateEndgame(gameState, newState, move, opponent)
		score += endgameBonus
	}

	return score
}

// evaluateSmallBoardAdvantage returns positive if opponent has advantage, negative if we do
func (ai *AIPlayer) evaluateSmallBoardAdvantage(state *game.GameState, boardIndex int, opponent game.Player) float64 {
	board := &state.BigBoard[boardIndex]
	
	aiCount := 0
	opponentCount := 0
	
	for i := 0; i < 9; i++ {
		if board[i] == ai.player {
			aiCount++
		} else if board[i] == opponent {
			opponentCount++
		}
	}
	
	// More pieces = advantage
	return float64(opponentCount - aiCount)
}

// countThreatsCreated counts how many new winning threats this move creates
func (ai *AIPlayer) countThreatsCreated(state *game.GameState, boardIndex int) int {
	board := &state.BigBoard[boardIndex]
	threats := 0
	
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		aiCount := 0
		emptyCount := 0
		
		for _, pos := range line {
			if board[pos] == ai.player {
				aiCount++
			} else if board[pos] == game.Empty {
				emptyCount++
			}
		}
		
		// Two in a row with one empty = threat
		if aiCount == 2 && emptyCount == 1 {
			threats++
		}
	}
	
	return threats
}

// evaluateForkOpportunities looks for fork setups (multiple winning threats)
func (ai *AIPlayer) evaluateForkOpportunities(state *game.GameState, boardIndex int) float64 {
	// This is a simplified fork detection
	// In a full implementation, you'd check for positions that create multiple threats
	threats := ai.countThreatsCreated(state, boardIndex)
	if threats >= 2 {
		return 100 // Fork opportunity!
	}
	return 0
}

// isSmallBoardFull helper function
func (ai *AIPlayer) isSmallBoardFull(state *game.GameState, boardIndex int) bool {
	for i := 0; i < 9; i++ {
		if state.BigBoard[boardIndex][i] == game.Empty {
			return false
		}
	}
	return true
}

// evaluateSendingTarget - CRITICAL: Where are we sending the opponent?
func (ai *AIPlayer) evaluateSendingTarget(originalState, newState *game.GameState, targetBoard int, opponent game.Player) float64 {
	score := 0.0
	
	// Analyze the target board
	board := &originalState.BigBoard[targetBoard]
	
	// Count pieces and threats in target board
	aiPieces := 0
	opponentPieces := 0
	opponentThreats := 0
	aiThreats := 0
	
	// Count current pieces
	for i := 0; i < 9; i++ {
		if board[i] == ai.player {
			aiPieces++
		} else if board[i] == opponent {
			opponentPieces++
		}
	}
	
	// Count threats in target board
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		aiCount := 0
		oppCount := 0
		emptyCount := 0
		
		for _, pos := range line {
			if board[pos] == ai.player {
				aiCount++
			} else if board[pos] == opponent {
				oppCount++
			} else {
				emptyCount++
			}
		}
		
		// Opponent can create threat
		if oppCount == 2 && emptyCount == 1 && aiCount == 0 {
			opponentThreats++
		}
		// We have threat there
		if aiCount == 2 && emptyCount == 1 && oppCount == 0 {
			aiThreats++
		}
	}
	
	// SCORING:
	// BAD: Sending opponent where they have more pieces or threats
	if opponentPieces > aiPieces {
		score -= 300 // They dominate this board
	}
	
	if opponentThreats > 0 {
		score -= 400 * float64(opponentThreats) // They can threaten to win it
	}
	
	// GOOD: Sending them where we dominate
	if aiPieces > opponentPieces && aiThreats > 0 {
		score += 200 // We control this board
	}
	
	// STRATEGIC: Board position value
	if targetBoard == 4 { // Center board
		if opponentPieces == 0 && aiPieces > 0 {
			score += 150 // Good, we control center
		} else if opponentPieces > aiPieces {
			score -= 200 // Bad, giving them center advantage
		}
	}
	
	// TACTICAL: Check if they can immediately win the small board
	if ai.canOpponentWinBoard(originalState, targetBoard, opponent) {
		score -= 600 // NEVER send them where they can win immediately!
	}
	
	return score
}

// canOpponentWinBoard checks if opponent can win a small board on their next move
func (ai *AIPlayer) canOpponentWinBoard(state *game.GameState, boardIndex int, opponent game.Player) bool {
	board := &state.BigBoard[boardIndex]
	
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		oppCount := 0
		emptyCount := 0
		
		for _, pos := range line {
			if board[pos] == opponent {
				oppCount++
			} else if board[pos] == game.Empty {
				emptyCount++
			}
		}
		
		// Two in a row with one empty = can win
		if oppCount == 2 && emptyCount == 1 {
			return true
		}
	}
	
	return false
}

// evaluateCenterMove - is center actually good here?
func (ai *AIPlayer) evaluateCenterMove(state *game.GameState, move game.Move, opponent game.Player) float64 {
	score := 0.0
	board := &state.BigBoard[move.BigBoardIndex]
	
	// Center is good if:
	// 1. We don't already control this board
	// 2. It creates threats
	// 3. It doesn't send opponent somewhere dangerous
	
	// Check if center creates immediate threats
	testBoard := *board
	testBoard[4] = ai.player
	
	threats := 0
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		if ai.lineContainsCenter(line) {
			aiCount := 0
			emptyCount := 0
			for _, pos := range line {
				if testBoard[pos] == ai.player {
					aiCount++
				} else if testBoard[pos] == game.Empty {
					emptyCount++
				}
			}
			if aiCount == 2 && emptyCount == 1 {
				threats++
			}
		}
	}
	
	if threats >= 2 {
		score += 100 // Fork opportunity
	} else if threats == 1 {
		score += 40 // Single threat
	} else {
		score += 10 // Just positional
	}
	
	// Penalty if we're sending opponent to a good board for them
	nextBoard := move.SmallBoardIndex
	if nextBoard == 4 { // Sending them to center board
		if ai.evaluateSmallBoardAdvantage(state, 4, opponent) > 0 {
			score -= 60 // They have advantage in center board
		}
	}
	
	return score
}

// evaluateNonCenterMove - reward strategic non-center moves
func (ai *AIPlayer) evaluateNonCenterMove(state *game.GameState, move game.Move, opponent game.Player) float64 {
	score := 0.0
	
	// Reward corners if they create threats
	corners := []int{0, 2, 6, 8}
	for _, corner := range corners {
		if move.SmallBoardIndex == corner {
			// Corner is good if it blocks opponent or creates threats
			if ai.doesMoveBlockOpponent(state, move, opponent) {
				score += 40
			} else {
				score += 15 // Basic corner value
			}
		}
	}
	
	// Reward edges if they're strategic
	edges := []int{1, 3, 5, 7}
	for _, edge := range edges {
		if move.SmallBoardIndex == edge {
			if ai.doesMoveBlockOpponent(state, move, opponent) {
				score += 35
			} else {
				score += 8 // Basic edge value
			}
		}
	}
	
	return score
}

// doesMoveBlockOpponent checks if this move blocks an opponent threat
func (ai *AIPlayer) doesMoveBlockOpponent(state *game.GameState, move game.Move, opponent game.Player) bool {
	board := &state.BigBoard[move.BigBoardIndex]
	position := move.SmallBoardIndex
	
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		// Check if our position is in this line
		inLine := false
		for _, pos := range line {
			if pos == position {
				inLine = true
				break
			}
		}
		
		if inLine {
			// Count opponent pieces in this line
			oppCount := 0
			emptyCount := 0
			for _, pos := range line {
				if board[pos] == opponent {
					oppCount++
				} else if board[pos] == game.Empty {
					emptyCount++
				}
			}
			
			// Would block opponent threat
			if oppCount == 2 && emptyCount == 1 {
				return true
			}
		}
	}
	
	return false
}

// isMoveDangerous checks if this move helps the opponent too much
func (ai *AIPlayer) isMoveDangerous(state *game.GameState, move game.Move, opponent game.Player) bool {
	nextBoard := move.SmallBoardIndex
	
	// Dangerous if we're sending them to a board they can win
	if ai.canOpponentWinBoard(state, nextBoard, opponent) {
		return true
	}
	
	// Dangerous if they have big advantage in that board
	if ai.evaluateSmallBoardAdvantage(state, nextBoard, opponent) > 2 {
		return true
	}
	
	return false
}

// lineContainsCenter helper function
func (ai *AIPlayer) lineContainsCenter(line []int) bool {
	for _, pos := range line {
		if pos == 4 {
			return true
		}
	}
	return false
}

// countTotalMoves counts total moves made in game
func (ai *AIPlayer) countTotalMoves(state *game.GameState) int {
	count := 0
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if state.BigBoard[i][j] != game.Empty {
				count++
			}
		}
	}
	return count
}

// getOpeningMove implements opening theory for Ultimate TTT
func (ai *AIPlayer) getOpeningMove(state *game.GameState) game.Move {
	moves := ai.getValidMoves(state)
	if len(moves) == 0 {
		return game.Move{}
	}
	
	totalMoves := ai.countTotalMoves(state)
	
	// First move - always center of center board
	if totalMoves == 0 {
		return game.Move{
			BigBoardIndex:   4,
			SmallBoardIndex: 4,
			Player:          ai.player,
		}
	}
	
	// Second move - respond to opponent's first move
	if totalMoves == 1 {
		return ai.getSecondMoveResponse(state)
	}
	
	// Early game - control key boards
	if totalMoves <= 4 {
		return ai.getEarlyGameMove(state)
	}
	
	// Opening theory didn't apply
	return game.Move{}
}

// getSecondMoveResponse - critical opening response
func (ai *AIPlayer) getSecondMoveResponse(state *game.GameState) game.Move {
	moves := ai.getValidMoves(state)
	opponent := game.X
	if ai.player == game.X {
		opponent = game.O
	}
	
	// Find where opponent played
	var opponentMove game.Move
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if state.BigBoard[i][j] == opponent {
				opponentMove = game.Move{BigBoardIndex: i, SmallBoardIndex: j, Player: opponent}
				break
			}
		}
	}
	
	// If they took center-center, we need counter-strategy
	if opponentMove.BigBoardIndex == 4 && opponentMove.SmallBoardIndex == 4 {
		// Take a corner of center board to maintain some control
		corners := []int{0, 2, 6, 8}
		for _, corner := range corners {
			move := game.Move{BigBoardIndex: 4, SmallBoardIndex: corner, Player: ai.player}
			if ai.isValidMoveForAI(state, move) {
				return move
			}
		}
	}
	
	// If they took center of non-center board, mirror in center
	if opponentMove.SmallBoardIndex == 4 && opponentMove.BigBoardIndex != 4 {
		move := game.Move{BigBoardIndex: 4, SmallBoardIndex: 4, Player: ai.player}
		if ai.isValidMoveForAI(state, move) {
			return move
		}
	}
	
	// Default: take center of center if possible
	centerMove := game.Move{BigBoardIndex: 4, SmallBoardIndex: 4, Player: ai.player}
	if ai.isValidMoveForAI(state, centerMove) {
		return centerMove
	}
	
	// Fallback to best available
	return ai.getStrategicMove(state, moves)
}

// getEarlyGameMove - early game strategy
func (ai *AIPlayer) getEarlyGameMove(state *game.GameState) game.Move {
	moves := ai.getValidMoves(state)
	
	// Priority 1: Control center board if we don't have it
	if state.BigBoardWins[4] == game.Empty {
		for _, move := range moves {
			if move.BigBoardIndex == 4 {
				// Prefer center and corners of center board
				if move.SmallBoardIndex == 4 || 
				   move.SmallBoardIndex == 0 || move.SmallBoardIndex == 2 || 
				   move.SmallBoardIndex == 6 || move.SmallBoardIndex == 8 {
					return move
				}
			}
		}
	}
	
	// Priority 2: Control corner boards
	corners := []int{0, 2, 6, 8}
	for _, corner := range corners {
		if state.BigBoardWins[corner] == game.Empty {
			for _, move := range moves {
				if move.BigBoardIndex == corner && move.SmallBoardIndex == 4 {
					return move
				}
			}
		}
	}
	
	return game.Move{}
}

// evaluateChainStrategy - think 2-3 moves ahead
func (ai *AIPlayer) evaluateChainStrategy(originalState, newState *game.GameState, move game.Move, opponent game.Player) float64 {
	score := 0.0
	
	// Where will opponent be forced to play?
	nextBoard := move.SmallBoardIndex
	if nextBoard < 0 || nextBoard > 8 {
		return 0
	}
	
	// If opponent goes to completed board, they get anarchy
	if newState.BigBoardWins[nextBoard] != game.Empty || ai.isSmallBoardFull(newState, nextBoard) {
		return -50 // Giving them too much freedom is bad
	}
	
	// Simulate opponent's likely responses
	opponentMoves := ai.getValidMovesForPlayer(newState, nextBoard, opponent)
	if len(opponentMoves) == 0 {
		return 100 // Trapped them!
	}
	
	worstCase := 1000.0
	for _, oppMove := range opponentMoves {
		// Where would we be sent after their move?
		ourNextBoard := oppMove.SmallBoardIndex
		
		// Evaluate how good that board is for us
		chainState := ai.copyGameState(newState)
		chainState.MakeMove(oppMove)
		
		var chainScore float64
		if chainState.BigBoardWins[ourNextBoard] != game.Empty || ai.isSmallBoardFull(chainState, ourNextBoard) {
			// We'd get anarchy - good!
			chainScore = 80.0
		} else {
			// How good is that board for us?
			ourAdvantage := -ai.evaluateSmallBoardAdvantage(chainState, ourNextBoard, opponent)
			chainScore = ourAdvantage * 20
		}
		
		if chainScore < worstCase {
			worstCase = chainScore
		}
	}
	
	score += worstCase * 0.3 // Weight the chain thinking
	
	return score
}

// evaluateSacrificeStrategy - sometimes losing a board is strategic
func (ai *AIPlayer) evaluateSacrificeStrategy(originalState, newState *game.GameState, move game.Move, opponent game.Player) float64 {
	score := 0.0
	
	// Check if we're "sacrificing" this board to opponent
	if ai.canOpponentWinBoard(newState, move.BigBoardIndex, opponent) {
		// This looks bad, but might be strategic...
		
		// Is it a corner/edge board? (Less valuable)
		boardValue := 100.0
		if move.BigBoardIndex == 4 {
			boardValue = 200.0 // Center is more valuable
		}
		
		// What do we gain by sacrificing?
		// 1. Where do we send them?
		nextBoard := move.SmallBoardIndex
		if nextBoard >= 0 && nextBoard < 9 {
			// If we send them somewhere bad for them, sacrifice might be worth it
			sendingAdvantage := ai.evaluateSendingTarget(originalState, newState, nextBoard, opponent)
			if sendingAdvantage > boardValue/2 {
				score += 50 // Strategic sacrifice
			}
		}
		
		// 2. Do we create threats elsewhere?
		if ai.countThreatsCreated(newState, move.BigBoardIndex) > 0 {
			score += 30 // Creating threats while sacrificing
		}
	}
	
	return score
}

// isEndgame checks if we're in endgame
func (ai *AIPlayer) isEndgame(state *game.GameState) bool {
	wonBoards := 0
	for i := 0; i < 9; i++ {
		if state.BigBoardWins[i] != game.Empty {
			wonBoards++
		}
	}
	return wonBoards >= 5 // Endgame when many boards decided
}

// evaluateEndgame - endgame-specific evaluation
func (ai *AIPlayer) evaluateEndgame(originalState, newState *game.GameState, move game.Move, opponent game.Player) float64 {
	score := 0.0
	
	// In endgame, every move matters more
	// Focus on:
	// 1. Immediate threats on big board
	// 2. Blocking opponent big board threats
	// 3. Creating multiple big board threats
	
	bigBoardThreats := ai.evaluateBigBoardThreats(newState)
	score += bigBoardThreats * 200 // Double the importance
	
	// Check if this move creates or prevents immediate big board wins
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	
	for _, line := range lines {
		// Check our threats on this line
		aiCount := 0
		oppCount := 0
		emptyCount := 0
		
		for _, pos := range line {
			if newState.BigBoardWins[pos] == ai.player {
				aiCount++
			} else if newState.BigBoardWins[pos] == opponent {
				oppCount++
			} else {
				emptyCount++
			}
		}
		
		// Do we complete a line?
		if aiCount == 3 {
			score += 5000 // Win!
		} else if aiCount == 2 && emptyCount == 1 {
			score += 800 // Almost won!
		}
		
		// Check original state for this line to see if we blocked opponent
		origAiCount := 0
		origOppCount := 0
		origEmptyCount := 0
		
		for _, pos := range line {
			if originalState.BigBoardWins[pos] == ai.player {
				origAiCount++
			} else if originalState.BigBoardWins[pos] == opponent {
				origOppCount++
			} else {
				origEmptyCount++
			}
		}
		
		// Did we block opponent win?
		if origOppCount == 2 && origEmptyCount == 1 && oppCount == 2 && emptyCount == 0 {
			score += 1000 // Blocked opponent win!
		}
	}
	
	return score
}

// getValidMovesForPlayer gets valid moves for specific player in specific board
func (ai *AIPlayer) getValidMovesForPlayer(state *game.GameState, boardIndex int, player game.Player) []game.Move {
	var moves []game.Move
	
	if boardIndex < 0 || boardIndex > 8 {
		return moves
	}
	
	if state.BigBoardWins[boardIndex] != game.Empty {
		return moves
	}
	
	for i := 0; i < 9; i++ {
		if state.BigBoard[boardIndex][i] == game.Empty {
			move := game.Move{
				BigBoardIndex:   boardIndex,
				SmallBoardIndex: i,
				Player:          player,
			}
			moves = append(moves, move)
		}
	}
	
	return moves
}

// isValidMoveForAI checks if a specific move is valid
func (ai *AIPlayer) isValidMoveForAI(state *game.GameState, move game.Move) bool {
	testState := ai.copyGameState(state)
	return testState.IsValidMove(move) == nil
}

// evaluatePosition evaluates the current game position with Ultimate TTT strategy
func (ai *AIPlayer) evaluatePosition(gameState *game.GameState) float64 {
	if gameState.GameWon == ai.player {
		return 10000
	}
	if gameState.GameWon != game.Empty && gameState.GameWon != ai.player {
		return -10000
	}
	if gameState.GameOver {
		return 0 // Tie
	}

	score := 0.0
	opponent := game.X
	if ai.player == game.X {
		opponent = game.O
	}

	// BIG BOARD CONTROL - this is everything in Ultimate TTT!
	centerBoardBonus := 0.0
	cornerBoardBonus := 0.0
	
	for i := 0; i < 9; i++ {
		if gameState.BigBoardWins[i] == ai.player {
			baseScore := 500.0
			if i == 4 { // Center board
				baseScore = 800.0
				centerBoardBonus += 200.0
			}
			corners := []int{0, 2, 6, 8}
			for _, corner := range corners {
				if i == corner {
					baseScore = 600.0
					cornerBoardBonus += 100.0
				}
			}
			score += baseScore
		} else if gameState.BigBoardWins[i] == opponent {
			baseScore := -500.0
			if i == 4 { // Opponent has center
				baseScore = -800.0
			}
			corners := []int{0, 2, 6, 8}
			for _, corner := range corners {
				if i == corner {
					baseScore = -600.0
				}
			}
			score += baseScore
		}
	}

	// BIG BOARD THREATS - look for almost-wins
	bigBoardThreats := ai.evaluateBigBoardThreats(gameState)
	score += bigBoardThreats * 300 // Each big board threat is very valuable

	// SMALL BOARD EVALUATIONS - detailed analysis
	for i := 0; i < 9; i++ {
		if gameState.BigBoardWins[i] == game.Empty {
			smallBoardScore := ai.evaluateSmallBoardAdvanced(&gameState.BigBoard[i])
			
			// Weight small boards by strategic importance
			weight := 20.0
			if i == 4 { // Center board
				weight = 40.0
			}
			corners := []int{0, 2, 6, 8}
			for _, corner := range corners {
				if i == corner {
					weight = 30.0
				}
			}
			
			score += smallBoardScore * weight
		}
	}

	// TEMPO AND BOARD CONTROL
	if gameState.ActiveBoard == -1 {
		// Anarchy mode - we can play anywhere, slight advantage
		score += 50
	} else {
		// Evaluate the active board constraint
		activeBoard := gameState.ActiveBoard
		if gameState.BigBoardWins[activeBoard] != game.Empty {
			// Opponent is forced into anarchy - good for us!
			score += 100
		} else {
			// Evaluate if the active board favors us or opponent
			boardAdvantage := ai.evaluateSmallBoardAdvantage(gameState, activeBoard, opponent)
			if boardAdvantage < 0 {
				score += 30 // We have advantage in active board
			} else {
				score -= 30 // Opponent has advantage
			}
		}
	}

	// PATTERN BONUSES
	score += centerBoardBonus
	score += cornerBoardBonus

	return score
}

// evaluateBigBoardThreats counts threats on the big board level
func (ai *AIPlayer) evaluateBigBoardThreats(gameState *game.GameState) float64 {
	threats := 0.0
	opponent := game.X
	if ai.player == game.X {
		opponent = game.O
	}

	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}

	for _, line := range lines {
		aiCount := 0
		opponentCount := 0
		emptyCount := 0

		for _, pos := range line {
			if gameState.BigBoardWins[pos] == ai.player {
				aiCount++
			} else if gameState.BigBoardWins[pos] == opponent {
				opponentCount++
			} else {
				emptyCount++
			}
		}

		// Two in a row with one empty = big threat!
		if aiCount == 2 && emptyCount == 1 && opponentCount == 0 {
			threats += 2.0 // Our threat
		} else if opponentCount == 2 && emptyCount == 1 && aiCount == 0 {
			threats -= 2.5 // Opponent threat (more dangerous)
		} else if aiCount == 1 && emptyCount == 2 && opponentCount == 0 {
			threats += 0.3 // Our potential
		} else if opponentCount == 1 && emptyCount == 2 && aiCount == 0 {
			threats -= 0.3 // Opponent potential
		}
	}

	return threats
}

// evaluateSmallBoardAdvanced gives detailed small board evaluation  
func (ai *AIPlayer) evaluateSmallBoardAdvanced(board *game.SmallBoard) float64 {
	score := 0.0
	opponent := game.X
	if ai.player == game.X {
		opponent = game.O
	}

	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}

	for _, line := range lines {
		aiCount := 0
		opponentCount := 0
		emptyCount := 0

		for _, pos := range line {
			switch board[pos] {
			case ai.player:
				aiCount++
			case opponent:
				opponentCount++
			case game.Empty:
				emptyCount++
			}
		}

		// Score based on line control with Ultimate TTT weighting
		if aiCount == 3 {
			score += 100 // Won this small board
		} else if aiCount == 2 && emptyCount == 1 && opponentCount == 0 {
			score += 25 // Two in a row, can win
		} else if aiCount == 1 && emptyCount == 2 && opponentCount == 0 {
			score += 3 // One in a row with potential
		}

		if opponentCount == 3 {
			score -= 100 // Opponent won
		} else if opponentCount == 2 && emptyCount == 1 && aiCount == 0 {
			score -= 30 // Must block opponent
		} else if opponentCount == 1 && emptyCount == 2 && aiCount == 0 {
			score -= 3 // Opponent potential
		}
	}

	// CENTER AND CORNER BONUSES for small boards
	if board[4] == ai.player { // Center
		score += 8
	} else if board[4] == opponent {
		score -= 8
	}

	corners := []int{0, 2, 6, 8}
	for _, corner := range corners {
		if board[corner] == ai.player {
			score += 3
		} else if board[corner] == opponent {
			score -= 3
		}
	}

	return score
}


// getValidMoves returns all valid moves for the current game state
func (ai *AIPlayer) getValidMoves(gameState *game.GameState) []game.Move {
	var moves []game.Move

	if gameState.GameOver {
		return moves
	}

	// If must play in specific board
	if gameState.ActiveBoard != -1 {
		boardIndex := gameState.ActiveBoard
		if gameState.BigBoardWins[boardIndex] == game.Empty {
			for smallIndex := 0; smallIndex < 9; smallIndex++ {
				if gameState.BigBoard[boardIndex][smallIndex] == game.Empty {
					move := game.Move{
						BigBoardIndex:   boardIndex,
						SmallBoardIndex: smallIndex,
						Player:          ai.player,
					}
					if gameState.IsValidMove(move) == nil {
						moves = append(moves, move)
					}
				}
			}
		} else {
			// Board is won, can play anywhere legal
			return ai.getAllLegalMoves(gameState)
		}
	} else {
		// Can play in any board
		return ai.getAllLegalMoves(gameState)
	}

	return moves
}

// getAllLegalMoves returns all legal moves when can play anywhere
func (ai *AIPlayer) getAllLegalMoves(gameState *game.GameState) []game.Move {
	var moves []game.Move

	for bigIndex := 0; bigIndex < 9; bigIndex++ {
		if gameState.BigBoardWins[bigIndex] == game.Empty {
			for smallIndex := 0; smallIndex < 9; smallIndex++ {
				if gameState.BigBoard[bigIndex][smallIndex] == game.Empty {
					move := game.Move{
						BigBoardIndex:   bigIndex,
						SmallBoardIndex: smallIndex,
						Player:          ai.player,
					}
					if gameState.IsValidMove(move) == nil {
						moves = append(moves, move)
					}
				}
			}
		}
	}

	return moves
}

// copyGameState creates a deep copy of the game state
func (ai *AIPlayer) copyGameState(original *game.GameState) *game.GameState {
	newState := &game.GameState{
		BigBoard:      original.BigBoard,
		BigBoardWins:  original.BigBoardWins,
		ActiveBoard:   original.ActiveBoard,
		CurrentPlayer: original.CurrentPlayer,
		GameWon:       original.GameWon,
		GameOver:      original.GameOver,
	}
	return newState
}