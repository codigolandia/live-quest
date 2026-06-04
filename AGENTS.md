# LiveQuest AI Developer Guide

This document provides high-signal instructions for AI agents working on this codebase.

---

## 🚀 Quick Start & Commands

* **Test:** `go test ./...` in the root workspace.
* **Note:** Package `oauth` and `cmdcheck` tests execute external network requests. Ensure internet connectivity is active during tests.

## 🏗️ Architecture & Logic

### Game Core (`package main` in `cmd/live-quest/`)
* **State Management:** `main.go` holds the `Game` struct (Viewers, chat history, queues, persistence) and manages the Ebitengine loop.
* **Physics/Rendering:** `viewer.go` handles player physics (position, velocity, gravity), stats (XP, HP), and rendering logic.
* **Command Parsing:** Custom commands are registered in `ParseCommands` within `main.go`.

### Embedded Assets (`package assets`)
* Uses `go:embed` for all game resources.
* **Tinting Pattern:** Animation frames (prefixed `frame`) are loaded in grayscale to enable real-time color multiplication via `colorm.ColorM.ScaleWithColor()`. Skins (prefixed `skin`) are layered on top.

### Integrations
* **Twitch:** Raw TCP socket connection using IRC protocols (`irc.chat.twitch.tv:6667`).
* **YouTube:** gRPC clients mapping to YouTube Live Chat API.
* **Challenges:** Loaded from `challenges.json`. Validation is handled via the Go Playground API in `cmdcheck.go`.

## 🛠️ Development Guidelines

### Adding New Commands
Update the `ParseCommands` method in `cmd/live-quest/main.go`:
```go
switch {
// ...
case strings.Contains(m.Text, "!newcmd"):
    // Process new action here
}
```

### Adding Coding Challenges
1. Define a unique string identifier (`code`) in `challenges.json`.
2. Specify constraints: `codeContains` (regex search) and `output` (regex matching stdout/stderr).
3. Set the reward value in XP.

### Rendering Rules
* Use gray-scale frames from `assets/animations` for dynamic coloring.
* Use the `DrawTextAt()` wrapper to render high-contrast text with shadows. Global scaling configs are in `cmd/live-quest/main.go`.

## 📝 Commit Convention
* **Format:** `<type>: <description in Portuguese>` (e.g., `feat: adiciona novo comando`).
* **Types:** `feat`, `fix`, `docs`, `chore`, `infra`, etc.

