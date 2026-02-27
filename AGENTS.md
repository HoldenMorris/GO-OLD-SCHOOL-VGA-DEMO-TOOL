# VGA-GO: Demoscene Demo Engine in Go

## Work Flow
Always update the TODO.md doc **BEFORE** AND **AFTER** PLANING AND WORKING!

## Project Overview
A fullscreen/windowed VGA-style demo engine inspired by classic DOS demos (Future Crew's Second Reality, etc.). Renders to a 320x200 256-color software framebuffer, plays MOD tracker music, and syncs visual effects to the music.

## Architecture

### Tech Stack
- **Graphics:** Ebitengine (ebiten) — cross-platform 2D game library, pure Go
- **Audio Playback:** libxmp via CGo bindings — plays MOD/S3M/XM/IT tracker files
- **Audio Output:** ebitengine/oto v3 — cross-platform PCM audio output
- **Build:** Makefile with cross-compilation targets for Linux, macOS, Windows

### Directory Layout
```
cmd/demo/main.go          — entry point, Ebitengine game loop
internal/vga/             — framebuffer (320x200), palette (256 colors)
internal/music/           — libxmp CGo bindings + audio pipeline
internal/sync/            — music-to-visual sync system
internal/effects/         — demo effects (plasma, fire, tunnel, starfield)
assets/                   — MOD/S3M/XM files, textures
```

### Key Design Decisions
- **Software rendering only** — all effects write to a `[]byte` framebuffer (indexed color), converted to RGBA for display
- **VGA Mode 13h emulation** — 320x200, 256 colors, nearest-neighbor upscale to display resolution
- **libxmp for tracker playback** — chosen for sync API: `xmp_frame_info` exposes row, pattern, order, BPM, speed, per-channel volumes every frame
- **Frame-based sync** — effects receive a `SyncState` struct each frame with music position and beat info

### Build Prerequisites
- Go 1.21+
- C compiler (gcc) and pkg-config
- `libxmp-dev` — tracker module playback (system package, linked via pkg-config)
- `libasound2-dev` — ALSA audio output on Linux (not needed on macOS/Windows)
- Ebitengine system deps on Linux: `libc6-dev libgl1-mesa-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev`
- Run `make check-deps` to verify everything is installed
- See README.md for per-distro install commands

### Cross-Compilation Notes
- Linux → Linux: native build
- Linux → Windows: requires `x86_64-w64-mingw32-gcc` and Windows-compiled libxmp
- Linux → macOS: requires osxcross toolchain (or build on macOS natively)
- CGo cross-compilation is tricky; native builds on each platform are most reliable

### Conventions
- Effects implement an `Effect` interface: `Init(fb)`, `Update(dt, sync)`, `Draw(fb)`
- All pixel work happens on the indexed `[]byte` buffer, never directly on Ebitengine images
- Palette changes are cheap (just swap the lookup table)
- Music position is the source of truth for demo timing (not wall clock)
