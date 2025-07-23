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

---

## Build Instructions

To generate a standalone executable:

```bash
go build -o gomoku
