# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.1.5] - 2026-06-23

### Changed

- The `update` command will preview the release notes before performing an update

## [0.1.4] - 2026-06-13

### Fixed

- File picker arrow key navigation now works — arrow keys were being intercepted by the history handler before reaching the picker

## [0.1.3] - 2026-06-13

### Added

- File sharing: `/upload [path]` sends a file to all connected peers (max 10 MB); omitting the path opens an interactive file picker
- `/save <filename>` downloads a received file to `~/Downloads/lol/`
- Drag-and-drop: dropping a file onto the terminal pre-fills `/upload <path>` in the input
- Command argument completion: Tab now completes arguments in addition to command names — file-system paths for `/upload`, buffered filenames for `/save`, `name#shortid` for `/dm`, theme names for `/theme`
- Rate limiting: per-IP connection limit (1 new connection/sec, burst 3) and per-client message limit (10 msg/sec, burst 20) to prevent flooding
- All message senders now display `name#shortid` for unambiguous identification when multiple users share the same hostname

### Changed

- Tab key completes the current argument or command name before falling back to toggling scroll mode
- Help bar reflects the updated Tab behaviour: `tab: complete/scroll`

### Fixed

- `/dm` completions include `#shortid` to avoid ambiguity with duplicate names

## [0.1.2] - 2026-06-12

### Fixed

- `lol update <version>` now works with or without a `v` prefix (e.g. `0.1.0` and `v0.1.0` are both accepted)
- `lol update <version>` no longer skips installation when the requested version is older than the current version (downgrade support)

## [0.1.1] - 2026-06-12

### Added

- `/theme <name>` command to switch color themes at runtime (`dark`, `dracula`, `nord`)
- the `--no-changelog` flag of the `update` command use `-n` as a shorthand
- Help bar at the bottom of the chat UI showing active keybindings for the current mode
- Copy button (`⎘`) in every message bubble header — click to copy the message body to clipboard

### Changed

- `update` now open the changelogs in a new terminal screen
- `help` command now displays flags and available subcommands as tables
- Scroll mode is now toggled with `tab` instead of `esc`
- In scroll mode, `esc`, `q`, `ctrl+c`, and `ctrl+q` quit the app
- In input mode, `ctrl+c`, `ctrl+a`, and `ctrl+v` now work as standard textarea shortcuts
- Theme background color is now fully controlled by the theme — no hardcoded black background

### Fixed

- Single newlines in chat messages no longer collapse into one line when rendered as markdown


## [0.1.0] - 2026-06-11

### Added

- Initial release
