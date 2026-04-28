# no-peek

A simple Bubble Tea TUI for solving interview-style problems without peeking too early.

Use it for LeetCode, coding interviews, math puzzles, or any problem where you want a structured attempt before reading the editorial.

## Flow

1. Start with a 30 minute focus round.
2. When time is up, no-peek asks whether you're still thinking or stuck.
3. If you're still thinking, it starts another 30 minute focus round.
4. If you're stuck, it starts a 15 minute rescue round.
5. After rescue time, if you're still stuck, it tells you to read the editorial.

## Run

```bash
go run . "Two Sum"
```

Quick test with short timers:

```bash
go run . --focus 1 --rescue 1 "Test Problem"
```

Customize durations:

```bash
go run . --focus 25 --rescue 10 "Binary Search"
```

## Build

```bash
go build -o no-peek
./no-peek "Two Sum"
```

## Controls

- `t` still thinking / new ideas
- `s` stuck
- `p` or `space` pause timer
- `r` restart from editorial screen
- `q` quit

## Notifications

When time is up, no-peek always rings the terminal bell. It also tries a best-effort desktop notification with built-in OS commands:

- Linux: `notify-send`
- macOS: `osascript`
- Windows: PowerShell

No extra notification dependency is required.

On Linux, install `notify-send` if you want desktop notifications:

```bash
sudo apt install libnotify-bin
```
