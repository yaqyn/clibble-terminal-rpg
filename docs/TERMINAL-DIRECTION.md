> Implemented: the rainy forest direction now runs in the Go RPG. Clibble also
> has an original offline world: wildlings and catch seeds replace the earlier
> course terminology. Current rules and controls live in [GUIDE.md](GUIDE.md);
> this document preserves the design research and alternatives below.

# Clibble: centered forest terminal research

Status: design proposal, not implemented game behavior. The current released
interface is a left-aligned menu with a small ASCII scene. This proposal follows
the owner's revised brief: a green, animated, pixel-art forest; centered choices;
second-person narration; and catch attempts that cannot be reset by spamming.

## Recommended direction

A quiet forest clearing fills the screen. Trees frame the left and right edges;
the center remains a dark, low-detail clearing for the title and action menu.
Selection uses a soft fern fill, dark pine text, and a visible chevron. Keep every
main action in the same central position, including Start, Travel, Catch,
Collection, and encounter choices. Statistics are a small top-edge strip;
second-person messages occupy a stable bottom strip. Avoid a dashboard sidebar.

Use six scene colors: deep pine #10291e, distant canopy #17392a, mid fern #20503a,
leaf highlights #3e6a43, selection #b1d18e, and readable ivory #e3efd4. These are
proposed terminal scene tokens, not a replacement for the light web companion.

Three directions are explored in `concepts/forest-terminal.html`:

- Rainy forest: vertical centered menu, layered tree silhouettes, sparse soft rain.
- Sunlit clearing: compact two-column central choices, lighter foliage, drifting pollen.
- Moonlit grove: smaller title, low central menu, pond and sparse firefly particles.

Rainy forest most closely matches the brief. These are scene names, not new app
names. The interactive concepts demonstrate layout, motion, and a three-try
counter; their outcomes and budgets are illustrative, not game balance or a Go
runtime implementation. In the actual product, rain, wind, and particles should
be independent settings with a single low-motion preset.

## What the terminal can render

| Approach | Strength | Cost | Recommendation |
| --- | --- | --- | --- |
| Colored Unicode half blocks (`▀`, `▄`) | Two vertical color samples per character cell; deliberate chunky pixel art | Resolution still depends on terminal cells and font metrics | Default scene renderer |
| ASCII / block symbols with limited color | Readable fallback, works without image protocols | Less detailed scenery | Low-color, plain, and reduced-motion fallback |
| Kitty graphics protocol | Real pixel images and animation frames | Capability detection, emulator/multiplexer differences, image-layer lifecycle | Optional later enhancement |

Chafa documents half-block rendering, Unicode mosaics, color modes, and multiple
image protocols. That demonstrates the available rendering techniques; Clibble
does not need to add the Chafa executable or C library to its Go runtime.
The Kitty protocol supports animation, but should not become an install requirement.

## Go implementation direction

Keep Bubble Tea v2 for input, resize, state changes, and shutdown. Its official
cell-buffer example demonstrates an animation tick feeding a cellular canvas.
Use a small deterministic scene renderer with a fixed cell grid and precomputed
background layers. Lip Gloss v2 provides centering, sizing, color-profile
adaptation, and layered compositing if the project benefits from that dependency.
An application-owned compositor can also keep dependency additions smaller.

Scene composition, in order:

1. Opaque pine background and distant trees, cached until resize/theme changes.
2. Near trees and a few low-frequency leaf/wind frames.
3. Sparse rain or particles; never draw weather through menu or narration cells.
4. Stable centered menu, supplies/tries labels, and narration.

Start ambient updates around 10–12 frames per second as a performance target,
then profile. Do not make API requests, consume gameplay randomness, change catch
odds, or write saves on animation ticks. Separate cosmetic and gameplay RNG.
Keep controls immediately responsive; catch feedback animation is skippable and
must not reroll or determine the outcome. Do not use flashing, screen shake, or
rapidly blinking selections. Pause ambient work when unfocused; provide a
persistent motion-off setting and `-plain` as an independent access path.

Prefer a 80×24 scene target; keep a simplified centered menu at 40×18. Recompute
placement from the actual terminal dimensions. Use opaque low-detail menu
backgrounds for contrast. Crop scenery at the edges rather than shrinking the
font. Test truecolor, 256/16 colors, NO_COLOR, resizing, and normal restoration.

## Catch attempts and limits

Current code has unlimited attempts: `attemptCatch` in `catch.go` applies a catch
probability and saves successful catches, but does not save escaped attempts.
A new counter shown only in `game.go` would be bypassed by `-plain`, one-shot
commands, restarting, or leaving and reopening an encounter.

Recommended initial balance to prototype (numbers require playtesting):

- Three throws per active encounter, displayed as `Tries: 3 / 3`.
- A shared supply of 12 Pokeballs, one consumed per resolved throw, including escape.
- When the third throw fails, end that encounter and narrate the Pokemon leaving.
- A proposed persistent per-species/area encounter cooldown prevents immediate
  reopening of the same target. Changing menus or restarting does not reset it.
- Replenishment is a separate design choice. A possible first prototype is one
  ball every five minutes, capped at 12, using persisted timestamps; this needs
  playtesting to avoid turning a cozy game into an idle timer. A daily allowance
  is simpler but can end a short session abruptly. Awarding repeatable travel or
  free rests must not create an unlimited refill loop.

Limits belong in one domain service shared by the TUI and commands. Save inventory,
attempts, cooldowns, and caught records atomically under the same writer lock.
Migrate old collections explicitly, preserving their Pokemon and granting the
initial supplies once. The read-only companion must understand both supported
save schemas. Network failures before a resolved throw should not spend a try;
a failed persistence operation must not display a successful outcome or leave
only part of the state committed. Block repeated input while a throw is pending.

This is fair-play enforcement against normal interaction and accidental spam.
A fully local, editable single-player save cannot prevent its owner changing
files or the clock. Tamper-resistant anti-cheat would require authoritative
external state and is outside the requested local-game design.

## References and what each contributes

- [Bubble Tea cell-buffer animation](https://github.com/charmbracelet/bubbletea/blob/main/examples/cellbuffer/main.go): a Go tick/render loop on a cellular grid.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss): terminal centering, sizing, styling, color downsampling, and cell-based compositing.
- [Chafa](https://hpjansson.org/chafa/): comparison of half blocks, richer Unicode mosaics, image protocols, and color modes.
- [Kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/): real pixel transfer, animation frames, and capability queries.
- [Pokete](https://github.com/lxgr-linux/pokete): a game-specific terminal reference for creature collection and exploration; no code or artwork is copied.
- [TerminalTextEffects](https://github.com/ChrisBuilds/terminaltexteffects): reference for per-character motion and rain effects. It is Python; use as motion inspiration, not a dependency for Clibble.

## Next implementation slice

Implement one centered rainy-forest scene and menu composition first. Validate
its readability, resize behavior, motion settings, and CPU cost before creating
additional scenes. In a separate focused change, add the shared persisted catch
budget and schema migration, then connect both interfaces. Do not publish a
cosmetic try counter before its persisted rules are in place.
