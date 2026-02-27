# VGA-GO

A demoscene demo engine in Go, inspired by classic DOS demos like Future Crew's Second Reality. Renders to a 320x200 256-color software framebuffer with MOD tracker music playback and beat-synced visual effects.

## Build Dependencies

### Go
Go 1.21 or later. Install from https://go.dev/dl/

### Linux (Ubuntu/Debian)
```bash
sudo apt install gcc pkg-config libxmp-dev libasound2-dev \
    libc6-dev libgl1-mesa-dev libxcursor-dev libxrandr-dev \
    libxinerama-dev libxi-dev libxxf86vm-dev
```

### Linux (Fedora/RHEL)
```bash
sudo dnf install gcc pkg-config libxmp-devel alsa-lib-devel \
    mesa-libGL-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel libXi-devel libXxf86vm-devel
```

### macOS
```bash
brew install libxmp pkg-config
```
No ALSA needed on macOS (oto uses CoreAudio).

### Windows
Install MSYS2/MinGW-w64, then:
```bash
pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config mingw-w64-x86_64-libxmp
```

### Verify dependencies
```bash
make check-deps
```

## Building

```bash
# Native build
make

# Build and run (no music)
make run

# Build and run with a tracker module
make run-mod MOD=assets/song.mod

# Or run directly
./build/vga-demo -mod assets/song.mod
```

### Cross-compilation
```bash
make build-linux        # Linux amd64
make build-mac-intel    # macOS amd64 (requires osxcross)
make build-mac-arm      # macOS arm64 (requires osxcross)
make build-windows      # Windows amd64 (requires mingw-w64)
make build-all          # All platforms
```

Cross-compiling with CGo requires the target platform's toolchain and libraries. Native builds on each platform are the most reliable approach.

## Usage

```
./build/vga-demo [flags]

Flags:
  -mod string        Path to MOD/S3M/XM/IT tracker module file
  -cue string        Path to JSON cue file (demo timeline)
  -fullscreen        Start in fullscreen mode
  -debug             Enable debug logging for music playback
```

### Controls
| Key   | Action                      |
|-------|-----------------------------|
| ESC   | Quit                        |
| F11   | Toggle fullscreen           |
| F1    | Toggle debug overlay (shows music position, FPS) |

Ctrl+C also triggers a graceful shutdown.

## How It Works

- **VGA Mode 13h emulation**: 320x200 pixels, 256-color palette, all effects write to an indexed byte buffer
- **Ebitengine**: Creates the window, scales the 320x200 buffer to display resolution with nearest-neighbor filtering
- **libxmp**: Plays MOD/S3M/XM/IT tracker modules and exposes per-frame sync data (order, pattern, row, BPM, channel volumes)
- **Sequencer**: Chains effects based on tracker position, with cut/fade/crossfade transitions

## Effects

| Effect       | Description                                              |
|--------------|----------------------------------------------------------|
| Plasma       | Classic sine-based color cycling                         |
| Fire         | Bottom-up heat propagation with palette gradient         |
| Tunnel       | Texture-mapped tunnel with XOR pattern                   |
| Starfield    | 3D parallax starfield flying through space               |
| SineScroller | Horizontal text scroller with per-character sine wave    |
| BigScroller  | Large scaled-up text scroller                            |

All effects react to music sync state (BPM, beats, channel volumes).

## Arranging Effects with Music (Cue Files)

### Tracker Music Concepts

If you're new to tracker music, here's what the sync terms mean:

- **Order**: The top-level playlist of a tracker module. Order 0 plays first, then order 1, etc. Think of it as "which section of the song."
- **Pattern**: A grid of note data, typically 64 rows tall. Each order position plays one pattern.
- **Row**: A single step within a pattern. At the default 125 BPM / speed 6, each row lasts ~120ms. There are usually 64 rows per pattern.
- **BPM / Speed**: BPM controls the tick rate, Speed controls ticks per row. Together they determine how fast the music plays.
- **Channel Volumes**: Each tracker channel (instrument) has a volume level (0-255) that can drive visual reactivity.

### Cue File Format

Effects are arranged with a JSON cue file. Each cue triggers an effect at a specific `order:row` position in the tracker module:

```json
{
  "effects": ["plasma", "fire", "tunnel", "starfield", "scroller", "bigscroller"],
  "cues": [
    {"order": 0, "row": 0,  "effect": "plasma",     "transition": "cut"},
    {"order": 1, "row": 0,  "effect": "starfield",  "transition": "cut"},
    {"order": 2, "row": 0,  "effect": "tunnel",     "transition": "fade", "fade_dur": 1.5},
    {"order": 2, "row": 32, "effect": "scroller",   "transition": "cut"},
    {"order": 3, "row": 0,  "effect": "fire",       "transition": "cut"},
    {"order": 4, "row": 0,  "effect": "plasma",     "transition": "fade", "fade_dur": 2.0}
  ]
}
```

**Fields:**
- `order` / `row`: The tracker position where this cue triggers
- `effect`: Name of the effect (must match a registered effect name)
- `transition`: How to switch — `"cut"` (instant), `"fade"` (crossfade to new effect)
- `fade_dur`: Duration of fade in seconds (only used with `"fade"` transition)

Cues are evaluated in order. When the tracker reaches or passes a cue's `order:row`, that effect becomes active.

### How to Sync Your Demo

1. **Open your MOD/S3M/XM/IT file** in a tracker (MilkyTracker, OpenMPT, etc.) or play it with `-debug` to see positions
2. **Note the order positions** where you want effects to change (e.g., intro at order 0, chorus at order 4)
3. **Create a cue file** mapping those positions to effects
4. **Run:** `./build/vga-demo -mod song.mod -cue demo.json`
5. **Iterate:** Adjust order/row values and transitions until it feels right

Use `-debug` to see the current order/pattern/row in the terminal while the demo runs.

### Without a Cue File

If no `-cue` flag is given, a default timeline is used that cycles through all effects every order position. This is useful for testing.

## Project Structure

```
cmd/demo/main.go          Entry point, game loop, audio setup
internal/vga/              Framebuffer (320x200), palettes, font, sprites
internal/music/            libxmp CGo bindings and audio pipeline
internal/sync/             Music-to-visual sync system, sequencer, cue file loader
internal/effects/          Demo effects (plasma, fire, tunnel, starfield, scrollers)
assets/                    Tracker modules, cue files, and other assets
```
