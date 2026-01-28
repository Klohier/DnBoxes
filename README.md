# DnBoxes (Dots and Boxes)

A full-stack, real-time implementation of the Dots and Boxes game. This project demonstrates backend game logic, real-time communication, and a modern frontend.

## Highlights

* Real-time multiplayer gameplay using WebSockets
* Server-authoritative game logic and move validation
* Interactive SVG-based UI in React
* Persistent game state and chat stored in a database
* Solo play against a bot

## Tech Stack

**Backend**
* Go (Echo framework)
* WebSockets for real-time gameplay and chat

**Frontend**
* React
* SVG-based board rendering
* Real-time updates via WebSockets

**Database**
* Postgres: persistent storage for users, game state, and chat history


**Infrastructure**
* Redis: in-memory Pub/Sub for events
* Docker: containerized environment
* Caddy: reverse proxy and HTTPS handling

## Setup & Run

> Prerequisites: Go (1.24+), Node.js (18+), npm

## Docker Setup (Recommended)

> One-command setup for running the full stack

```bash
# clone repository
# from project root
docker compose --env-file .env up --build  
```

This will:

* Build and run the Go API server
* Build the React frontend
* Start Caddy, Redis, and Postgres

Access the frontend at https://localhost (Caddy handles routing to backend API and WebSockets)

### Docker Architecture

* **backend**: Go + Echo backend (serves API and WebSockets)
* **frontend**: React frontend container
* **db**: Postgres container
* **caddy**: Caddy reverse proxy container
* **redis** Redis container


## Manual Setup (Needs updating)

### Backend

```bash
# clone repository

# install dependencies
go mod tidy

# run API server
go run ./cmd
```

The backend runs on `http://localhost:8484` and serves both the API and WebSocket connections.
### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Testing

### Backend
Must have docker setup running, tests  need a running redis setup.

- `User` tests (needs updating will fail)
- `Game` tests
- `Infra` tests
- `Lobby` tests


```bash
go test ./internal/user
go test ./internal/game
go test ./internal/infra
go test ./internal/lobby
```




## Gameplay Overview

1. Users authenticate and join a global chat
2. Players create or join a game session or play against a bot
3. Each move selects a side of a square
4. The server validates moves and updates state
5. Completing a box awards a point and an extra turn

## What This Project Demonstrates

* Designing a turn-based multiplayer game
* Real-time synchronization across clients
* Managing shared state safely on the server
* Deploying full-stack applications with Docker and Caddy


## Status

Actively developed. Core gameplay and multiplayer features are complete.
Chat currently down due to iterating on it.
