# no-peek

A simple Bubble Tea TUI with two focused tools:

- **Puzzle mode**: for LeetCode, coding interviews, math puzzles, or any problem where you want a structured attempt before reading the editorial.
- **Deep work mode**: for sustained work sessions with deliberate focus blocks and quiet breaks.

You must choose a tool with `--mode puzzle` or `--mode deep`.

## Tools / Modes

### Puzzle mode

For interview-style problems:

1. Start with a 30 minute focus round.
2. When time is up, no-peek asks whether you're still thinking or stuck.
3. If you're still thinking, it starts another 30 minute focus round.
4. If you're stuck, it starts a 15 minute rescue round.
5. After rescue time, if you're still stuck, it tells you to read the editorial.

### Deep work mode

A quiet deep-work cycle:

1. 45 minutes focus — no distractions.
2. 5 minutes break — rest, but don't distract yourself.
3. 45 minutes focus — no distractions.
4. 20 minutes break — rest, but don't distract yourself.
5. Repeat.

Deep work mode does not ring notifications between phases; the TUI simply advances to the next focus or rest period.

## Install

```bash
go install github.com/0venburn/no-peek@latest
```

Make sure your Go bin directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Run

`--mode` is required.

Run puzzle mode:

```bash
no-peek --mode puzzle "Two Sum"
```

Run deep work mode:

```bash
no-peek --mode deep "Write design doc"
```

Quick puzzle test with short timers:

```bash
no-peek --mode puzzle --focus 1 --rescue 1 "Test Problem"
```

Quick deep work test with short timers:

```bash
no-peek --mode deep --deep-focus 1 --short-break 1 --long-break 1 "Test Task"
```

Customize puzzle durations:

```bash
no-peek --mode puzzle --focus 25 --rescue 10 "Binary Search"
```

Customize deep work durations:

```bash
no-peek --mode deep --deep-focus 45 --short-break 5 --long-break 20 "Write design doc"
```

## Build from source

```bash
git clone git@github.com:0venburn/no-peek.git
cd no-peek
go build -o no-peek
./no-peek --mode puzzle "Two Sum"
```

## Controls

Common controls:

- `p` or `space` pause/resume timer
- `q` quit

Puzzle mode controls:

- `t` still thinking / new ideas
- `s` stuck
- `r` restart from editorial screen

Deep work mode runs automatically through focus and break periods.

## Notifications

Puzzle mode rings the terminal bell when a timer ends. It also tries a best-effort desktop notification with built-in OS commands:

- Linux: `notify-send`
- macOS: `osascript`
- Windows: PowerShell

No extra notification dependency is required.

On Linux, install `notify-send` if you want desktop notifications:

```bash
sudo apt install libnotify-bin
```
