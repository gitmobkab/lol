# lol

A minimal terminal chat app for local networks.

## How it works

One person starts the server, shares their IP, everyone else joins.

```
# Start the server (auto-detects your LAN IP)
lol serve

# Join from another machine
lol join 192.168.1.42
lol join 192.168.1.42 --name Alice
lol join 192.168.1.42:9090        # if serve was started with --port 9090
```

## Install

**macOS — Apple Silicon**
```bash
curl -Lo lol https://github.com/gitmobkab/lol/releases/latest/download/lol-darwin-arm64
chmod +x lol && sudo mv lol /usr/local/bin/
```

**macOS — Intel**
```bash
curl -Lo lol https://github.com/gitmobkab/lol/releases/latest/download/lol-darwin-amd64
chmod +x lol && sudo mv lol /usr/local/bin/
```

**Linux — amd64**
```bash
curl -Lo lol https://github.com/gitmobkab/lol/releases/latest/download/lol-linux-amd64
chmod +x lol && sudo mv lol /usr/local/bin/
```

**Linux — arm64**
```bash
curl -Lo lol https://github.com/gitmobkab/lol/releases/latest/download/lol-linux-arm64
chmod +x lol && sudo mv lol /usr/local/bin/
```

**Windows — PowerShell**
```powershell
Invoke-WebRequest -Uri https://github.com/gitmobkab/lol/releases/latest/download/lol-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\Microsoft\WindowsApps\lol.exe"
```

`$env:LOCALAPPDATA\Microsoft\WindowsApps` is on `PATH` by default — no admin or extra config needed.

### Update

```
lol update              # latest release
lol update 0.1.0        # specific version
lol update --no-changelog
```

## Key bindings

| Key | Action |
|-----|--------|
| `enter` | Send message |
| `shift+enter` | New line |
| `↑` / `↓` | Browse sent message history |
| `esc` | Enter scroll mode (scroll through chat) |
| `esc` / `tab` | Exit scroll mode |
| `ctrl+q` | Quit |

## Commands

Type `/` in the input to see the autocomplete overlay.

| Command | Description |
|---------|-------------|
| `/ping` | Ping the server — replies with Pong |
| `/dm <user> <message>` | Send a direct message |
| `/die` | Quit the app |

For DMs with name collisions, append the short ID: `/dm alice#a3f2c1 hey`.

## Build from source

```
git clone https://github.com/gitmobkab/lol
cd lol
go build -o lol .
```
