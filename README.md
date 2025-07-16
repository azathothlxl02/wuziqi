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

---

## Releases

### Release 1 (v1) - Basic PvP Gameplay
**Release Date**: 2025.7.9  
**Goal**: Build a minimal playable version of the Gomoku game with local two-player support.

#### Features
- Title screen before the game starts
- Local two-player (PVP) mode: alternate turns between black and white players
- Turn management logic
- Win condition: judge and announce winner when a player gets 5 in a row
- Display current turn
- Game end screen with result
