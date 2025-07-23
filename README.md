=======
# Gomoku - Go + Ebitengine Project

A graphical Gomoku (Five-in-a-Row) game developed using Go and the [Ebitengine (Ebiten)](https://ebitengine.org/en/) game framework.

This is a team-based course project that follows Agile development with multiple feature releases.

---

## Game Overview

Gomoku is a traditional two-player game where players take turns placing stones on a board. The goal is to form a continuous line of a set number of stones (3–5) to win.

---

## Technologies Used

- **Programming Language**: Go (Golang)
- **Game Engine**: [Ebitengine](https://ebitengine.org/en/)
- **Others**: Go modules, Ebiten audio/image packages

---

## Getting Started

### 1. Install Go
https://golang.org/doc/install

### 2. Clone the project
```bash
git clone https://github.com/yourusername/gomoku-ebiten
cd gomoku-ebiten

```

## Releases

### Release 1 (v1) - Basic PvP Gameplay
**Release Date**: 2025.7.9  
**Goal**: Build a minimal playable version of the Gomoku game with local two-player support.

#### Features
- Title screen before the game starts
- Local two-player (PVP) mode: alternate turns between black and white players
- Turn management logic
- Win condition: judge and announce winner when a player gets 5 in a row
- Display the current turn
- Game ending screen

---

### Release 2 (v2) – Computer Opponent & Game Control Features
**Release Date**: 2025.7.12  
**Goal**: Implement player vs computer mode and enhance gameplay control.

#### Features
- Player vs CPU (PVE) mode: The player can play with a simple AI 
- Regret/Undo function: undo the previous move
- In-game Restart and Exit buttons
- The improved turn control and stone validation logic

---

### Release 3 (v3) – Game Experience Improvements & Rule Customization
**Expected Release**: 2025.7.16  
**Goal**: Improve game experience and allow more flexible win rules.

#### Features
- Highlight opponent's last move
- Optional background music (planned)
- Exportable game binary (planned)
- In-game help or README documentation (planned)

---

### Release 4 (v4) - Local Network Multiplayer
**Release Date**: 2025.7.23  
**Goal**: Enable real-time Gomoku gameplay between two players over a local network (LAN).

#### Features
- All features from **v1** to **v3**
- LAN-based multiplayer mode: two players can connect and play over the same local network
- Simple JSON-based message protocol for communication
- Host and client roles: one player creates the game as host, the other joins as client
- Basic connection logic and turn synchronization using JSON messages
- Prevent desync due to invalid actions or dropped messages

> Uses Go's `net` package and `encoding/json` to transmit move data such as stone position and player ID.

---

### Release 5 (v5) - Details Improvement

## Build Instructions

To generate a standalone executable:

```bash
go build -o gomoku
