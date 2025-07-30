# 🎯 T-9: Ultimate Tic-Tac-Toe

A modern, real-time multiplayer implementation of Ultimate Tic-Tac-Toe with Go backend and SvelteKit frontend.

## 🎮 What is Ultimate Tic-Tac-Toe?

Ultimate Tic-Tac-Toe (also known as "Super Tic-Tac-Toe" or "Meta Tic-Tac-Toe") is a strategic variant that consists of nine tic-tac-toe boards arranged in a 3×3 grid.

### Rules:
- **Local boards**: Each small 3×3 grid is a regular tic-tac-toe game
- **Global board**: The 3×3 arrangement of local boards forms the ultimate game
- **Movement constraint**: Your move determines which board your opponent must play in next
- **Winning**: Win three local boards in a row (horizontally, vertically, or diagonally) to win the game
- **Anarchy mode**: If sent to a completed board, you can play anywhere legal!

## ✨ Features

- 🌐 **Real-time multiplayer** with WebSocket communication
- 🍎 **Friendly room codes** (apple, broccoli, mango, etc.)
- 🎯 **Complete Ultimate TTT rules** implementation
- 🎨 **Beautiful, responsive UI** built with SvelteKit
- 🔄 **Live game state synchronization**
- 📱 **Mobile-friendly design**
- 🏠 **Single-player mode** for local play
- 🚀 **Fast, modern tech stack**

## 🛠️ Tech Stack

**Backend:**
- Go with Gin framework
- WebSocket support via Gorilla WebSocket
- Real-time game room management

**Frontend:**
- SvelteKit with TypeScript
- Svelte 5 with runes for reactivity
- Real-time WebSocket client

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/DukeofWaterloo/T-9.git
   cd T-9
   ```

2. **Start the backend**
   ```bash
   cd backend
   go mod tidy
   go run cmd/main.go
   ```
   Server starts on `http://localhost:8080`

3. **Start the frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   App runs on `http://localhost:5173`

## 🎯 How to Play

### Single Player
1. Visit `http://localhost:5173`
2. Click **"Single Player"**
3. Play locally on the same device

### Multiplayer
1. **Player 1**: Click **"Multiplayer"** → **"Create New Game"**
2. **Share the room code** (e.g., "apple", "broccoli")
3. **Player 2**: Click **"Multiplayer"** → Enter room code → **"Join Game"**
4. **Play in real-time!** 🎮

## 🏗️ Project Structure

```
T-9/
├── backend/           # Go server
│   ├── cmd/          # Main application
│   ├── internal/
│   │   ├── game/     # Core game logic
│   │   ├── api/      # REST API handlers
│   │   └── ws/       # WebSocket management
│   └── go.mod
├── frontend/         # SvelteKit app
│   ├── src/
│   │   ├── lib/      # Game components
│   │   └── routes/   # Pages
│   └── package.json
└── README.md
```

## 🎨 Game Logic

The game implements all Ultimate Tic-Tac-Toe rules:

- **Move validation**: Ensures moves are legal according to current constraints
- **Board state tracking**: Manages both local and global board states
- **Win detection**: Checks for wins in both local boards and the ultimate game
- **Turn management**: Handles player turns and move constraints
- **Anarchy mode**: Allows free play when sent to completed boards

## 🔄 Real-time Features

- **WebSocket connections** for instant move synchronization
- **Game room management** with unique produce-themed codes
- **Player assignment** (X/O) on join
- **Connection status** indicators
- **Automatic reconnection** on disconnect

## 🎯 Future Plans

- 🤖 **AI opponent** with machine learning
- 📊 **Game statistics** and history
- 🏆 **Tournament mode**
- 🎨 **Custom themes**
- 📱 **Mobile app**

## 🤝 Contributing

Contributions welcome! This project was built as a learning exercise for Go and Svelte.

## 📝 License

MIT License - feel free to use and modify!

---

**Built with ❤️ using Go + SvelteKit**
TicTicTicTacTacTacToeToeToe
