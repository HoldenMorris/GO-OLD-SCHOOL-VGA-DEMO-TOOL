# VGA-GO Task List

## Task 1: Initialize Go module and project structure [DONE]
Create go.mod, directory layout, Makefile with cross-compilation targets, add Ebitengine and oto dependencies.

## Task 2: Implement VGA framebuffer and palette system [DONE]
320x200 byte-array framebuffer, 256-color palette, Clear(), SetPixel(), WritePixels() for Ebitengine. Classic VGA palette presets.

## Task 3: Create Ebitengine game loop with fullscreen/windowed toggle [DONE]
ebiten.Game interface, F11 fullscreen toggle, nearest-neighbor scaling, 60fps with delta-time.

## Task 4: Write CGo bindings for libxmp [DONE]
Minimal wrapper using system libxmp-dev via pkg-config. Go-friendly FrameInfo struct with sync data.

## Task 5: Implement audio output pipeline with oto v3 [DONE]
Music player goroutine renders PCM from libxmp into ring buffer, consumed by oto.Player via io.Reader. Exposes SyncState via RWMutex.

## Task 6: Build sync system for music-driven effects [DONE]
Timeline maps tracker order:row positions to effect cues. BeatPulse/RowPulse/MaxChannelVolume helpers for reactive visuals.

## Task 7: Implement first demo effect: plasma [DONE]
Sine-based plasma with lookup table on VGA buffer. Speed driven by BPM, pulse on beat.

## Task 8: Implement additional classic effects (fire, tunnel, starfield) [DONE]
Fire (heat propagation), Tunnel (LUT-based with XOR texture), Starfield (3D parallax). All react to sync state.

## Task 9: Build demo timeline/sequencer to chain effects [DONE]
Sequencer chains effects at order:row cue points. Supports cut/fade/crossfade transitions. Manages effect init and alpha blending.

## Task 10: Integration test: full demo with music + synced effects [DONE]
Full wiring: cmd/demo/main.go loads MOD, starts oto audio, runs sequencer with all 4 effects. ESC quit, F11 fullscreen, F1 debug overlay. CLI flags: -mod, -fullscreen.

---

## Task 11: Graceful shutdown with OS signal handling [DONE]
- ESC or SIGINT/SIGTERM triggers graceful shutdown
- Music volume fades out over 5 seconds
- Screen fades to black over 5 seconds (palette darkens each frame)
- Debug logging during fade ("fading... 80%", etc.)
- Player cleanup with timeout after fade completes
- Implemented in: cmd/demo/main.go (startShutdown, updateFade)

## Task 12: VGA bitmap font system (8x8 CP437) [DONE]
- LoadFontFromPNG loads 256x64 sprite sheet (assets/font1.png)
- CP437Font array stores 8x8 bitmap for each char
- DrawChar/DrawString with transparent background (opts.Transparent)
- DrawCharBg/DrawStringBg with opaque background
- FontOptions struct for scale and color control
- Implemented in: internal/vga/font.go

## Task 13: Sprite system [DONE]
- Sprite struct (width, height, pixels slice)
- NewSprite, NewSpriteFromData constructors
- GetPixel/SetPixel with bounds checking
- DrawSprite draws sprite to framebuffer (color 0 = transparent)
- DrawSpriteScaled for integer-scaled rendering
- CharToSprite converts CP437 char to sprite
- Implemented in: internal/vga/sprite.go

## Task 14: Text scroller effects [DONE]
- SineScroller: horizontal scroll with per-character sine wave displacement
  - BPM-reactive speed
  - Configurable amplitude and frequency
- BigScroller: scaled-up font chars using DrawSpriteScaled
  - BPM-reactive speed
  - 3x scale by default
- Both implement Effect interface for sequencer
- Available as "sineScroller" and "bigScroller" in cue files
- Implemented in: internal/effects/scroller.go

## Task 15: Cue file system + documentation [DONE]
- JSON cue file loader without external deps (encoding/json)
- LoadCueFile parses assets/demo.json format
- Maps effect names to indices via "effects" array
- Supports order/row, effect name, transition type, fade_dur
- -cue CLI flag in main.go
- Example assets/demo.json with all effects
- Implemented in: internal/sync/cuefile.go

---

## All Tasks Completed
