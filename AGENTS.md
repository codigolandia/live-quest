# LiveQuest AI Developer Guide

This document is designed for AI agents and developer tools working on this codebase. It provides a structured view of the repository layout, package architecture, component dependencies, and extension guidelines.

---

## Repository Structure & Package Maps

The codebase is structured as a single Go module (`github.com/codigolandia/live-quest`).

```
.
├── cmd/
│   └── live-quest/         # Core application entry point and game logic
│       ├── main.go         # Initialization, Ebitengine game loop, HTTP API
│       ├── viewer.go       # Viewer (player) state, physics, rendering
│       ├── gopher.go       # Parsing color modification command
│       ├── http.go         # Serves static assets & JSON chat history APIs
│       └── cmdcheck.go     # Go Playground challenge worker queue & execution
├── assets/                 # Embedded graphic & web resources
│   ├── animations/         # Gopher animations frames (gray-scale & skins)
│   ├── img/                # Fixed UI sprites (HP, XP, Twitch/YouTube icons)
│   ├── web/                # Front-end chat/leaderboard assets (HTML/JS/CSS)
│   ├── assets.go           # Embedding & utility functions for loading assets
│   └── animation.go        # Dynamic sprite sheet and skin overlays loader
├── oauth/                  # OAuth2 local server and storage mechanics
├── twitch/                 # Twitch IRC socket-based client
├── youtube/                # YouTube Live Chat API & gRPC client
│   └── proto/              # Protobuf & gRPC definitions for YouTube Live Chat service
├── challenges.json         # Programming challenges definitions database
└── go.mod                  # Package dependency configuration
```

---

## Package Breakdown

### 1. Game State & Logic (`package main`)
Located in [cmd/live-quest/](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest).
* **[main.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/main.go)**: Holds the `Game` struct which manages all viewers (`Viewers`), chat history, fighting queue, active challenge queue, and persistence. Orchestrates Ebitengine's `Update()` and `Draw()` ticks.
* **[viewer.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/viewer.go)**: Defines the `Viewer` type. Manages individual viewer physics (position, velocity, gravity), stats (XP, HP), and renders the color-tinted sprites, health/XP bars, and badges.
* **[cmdcheck.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/cmdcheck.go)**: Launches a consumer goroutine reading from `g.queue`. Handles remote code fetching, execution, formatting, and vetting via the Go Playground API to validate chat submissions for the `!check` command.
* **[http.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/http.go)**: Hosts a local server mapping a chat history JSON endpoint at `/chat` and servicing local HTML files from `assets/web/`.
* **[gopher.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/gopher.go)**: Helper function to parse hex values from the chat to colorize Gophers.

### 2. Embedded Resources (`package assets`)
Located in [assets/](file:///home/ronoaldo/workspace/codigolandia/live-quest/assets).
* Utilizes `go:embed` to package game UI, styles, icons, and sprite animations inside the final compiled binary.
* **Gray-scale & tinting pattern**: Animations frames (prefixed with `frame`) are loaded in grayscale to enable real-time multiplication rendering (`colorm.ColorM.ScaleWithColor()`). Skins (prefixed with `skin`) containing static colored segments (like eyes) are layered on top without any modification.

### 3. Authentication (`package oauth`)
Located in [oauth/](file:///home/ronoaldo/workspace/codigolandia/live-quest/oauth).
* Implements a local callback server `localhost:8089` to capture OAuth2 authorization codes when configuring Twitch or YouTube permissions.
* Keeps cached credentials persistent by encoding the tokens to `~/.live-quest.[Platform].json`.

### 4. Twitch Client (`package twitch`)
Located in [twitch/](file:///home/ronoaldo/workspace/codigolandia/live-quest/twitch).
* Integrates directly via raw TCP socket connection using IRC protocols (`irc.chat.twitch.tv:6667`).
* Starts an async read loop decoding `PRIVMSG` logs and responding to server `PING` calls with `PONG`.

### 5. YouTube Client (`package youtube`)
Located in [youtube/](file:///home/ronoaldo/workspace/codigolandia/live-quest/youtube).
* Leverages gRPC clients structured under [youtube/proto/](file:///home/ronoaldo/workspace/codigolandia/live-quest/youtube/proto) mapping direct requests to `youtube.googleapis.com:443`.
* Subscribes to chat stream feeds via `StreamList` gRPC endpoints.

---

## Extension Guidelines for AI Agents

When modifying or introducing new features, respect the following design invariants:

### Adding New Commands
To register new chat commands, update the `ParseCommands` method in [main.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/main.go#L209):
```go
switch {
// ...
case strings.Contains(m.Text, "!newcmd"):
    // Process new action here
}
```

### Adding Coding Challenges
Challenges are loaded dynamically from [challenges.json](file:///home/ronoaldo/workspace/codigolandia/live-quest/challenges.json). To add a new one:
1. Define a unique string identifier (`code`).
2. Specify constraints: `codeContains` (regex search on raw submission string) and `output` (regex matching stdout/stderr during compiling execution).
3. Set the reward value in XP.

### Rendering Changes
* Always use gray-scale frames in `assets/animations` if the element's color should dynamically adapt to user configurations.
* UI and fonts scaling configurations are declared globally at [main.go](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/main.go#L403). Use the `DrawTextAt()` wrapper to render high-contrast text with shadows.

### Testing
* Run `go test ./...` in the root workspace to validate changes.
* Package [oauth](file:///home/ronoaldo/workspace/codigolandia/live-quest/oauth) and [cmdcheck](file:///home/ronoaldo/workspace/codigolandia/live-quest/cmd/live-quest/cmdcheck_test.go) tests execute external network requests (testing real endpoints). Ensure internet connectivity is active during tests.

### Commit Message Guidelines
* Commits in this repository must **always** be written in Portuguese.
* The prefix of the commit type must be in English (e.g., `feat`, `fix`, `docs`, `chore`, `infra`, etc.), followed by a colon and a space, with the description in Portuguese.
* Example: `docs: adiciona diretrizes de commit no AGENTS.md`
