# Gomoku — A LAN-Enabled Gomoku Game in Go with Ebitengine

Gomoku is a classic board game (Five-in-a-Row) where players take turns placing black or white stones on a grid.  
This project implements a graphical version of Gomoku using the [Ebitengine (Ebiten)](https://ebitengine.org/en/) game engine in Go.

It is developed as part of a team-based software engineering course project using Agile methodology.  
Key features include AI difficulty levels, LAN multiplayer support, undo functionality, and clean UI.

---

## Game Overview

Gomoku is a traditional two-player strategy game in which players take turns placing stones on a board.  
The objective is to form an unbroken line of 3 to 5 stones horizontally, vertically, or diagonally to win.

---

## Technologies Used

- **Programming Language**: Go (Golang)
- **Game Engine**: [Ebitengine (Ebiten)](https://ebitengine.org/en/)
- **Other Tools**:
  - Go modules
  - Ebiten audio/image packages
  - `net` package for LAN communication
  - `encoding/json` for multiplayer message synchronization

---

## Getting Started

### 1. Install Go  
Follow the official instructions: https://golang.org/doc/install  
**Recommended version**: Go 1.21+

### 2. Clone the project
```bash
git clone https://github.com/yourusername/gomoku-ebiten
cd gomoku-ebiten
```

### 3. Run the game
```bash
go run main.go
```

### 4. Build a standalone executable
```bash
go build -o gomoku
```

---

## Releases

### Release 1 (v1) – Console-Based Playable Prototype
**Release Date**: 2025.7.9  
**Goal**: Build a simple prototype playable via the command line, with core Gomoku logic.

#### Features
- Basic input/output handling via console
- Turn-based placement logic
- Win detection (5-in-a-row)
- Display game result

---

### Release 2 (v2) – GUI Interface & Local PvP Mode
**Release Date**: 2025.7.12  
**Goal**: Add a graphical interface and enable two local players to play using mouse input.

#### Features
- All features of **v1**
- Graphical title screen and board UI
- Mouse-based stone placement
- Local two-player mode (PvP)
- Restart and Exit buttons
- Turn control and player indicator

---

### Release 3 (v3) – AI Opponent & Regret Functionality
**Release Date**: 2025.7.16  
**Goal**: Introduce computer opponent with multiple AI difficulty levels and allow move undoing.

#### Features
- All features of **v1** and **v2**
- Player vs AI (PvE) mode
- Three difficulty levels:
  - **Easy**: Monte Carlo AI with 1s time limit
  - **Medium**: Monte Carlo AI with 3s time limit
  - **Hard**: AlphaZero-inspired AI with model-based prediction
- Regret (Undo) function for canceling previous move
- AI time control and logic improvements

---

### Release 4 (v4) – LAN Multiplayer with Sync Logic
**Release Date**: 2025.7.23  
**Goal**: Enable two players to connect and play over a local network using synchronized game state.

#### Features
- All features from **v1** to **v3**
- TCP-based LAN multiplayer mode (host/client)
- Turn data exchange via JSON protocol
- Synchronized regret (undo) system
- Regret confirmation from opponent
- Room-based connection system for joining games

---

### Release 5 (v5) – Final Polish & Presentation Prep
**Expected Release Date**: Week 7  
**Goal**: Refine game presentation and prepare for final demonstration.

#### Features
- All features from **v1** to **v4**
- Background music and audio effects
- UI/UX polish and layout improvement
- Code optimization and cleanup
- Planned: Export standalone executable
