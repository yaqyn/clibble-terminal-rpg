---
version: alpha
name: Clibble
description: An original wilderness adventure that begins at the terminal prompt.
colors:
  primary: "#365b43"
  text: "#142a25"
  background: "#eef3e7"
  surface: "#fbfcf5"
  muted: "#526b5d"
  terminal-accent: "#a7c78b"
typography:
  sans:
    fontFamily: "Manrope, system-ui, sans-serif"
  mono:
    fontFamily: "ui-monospace, monospace"
rounded:
  DEFAULT: "0.75rem"
  control: "0.5rem"
spacing:
  section-gap: "2rem"
  page-max: "76rem"
components:
  navigation:
    backgroundColor: "{colors.background}"
    textColor: "{colors.text}"
  creature-record:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
  caption:
    textColor: "{colors.muted}"
  terminal:
    textColor: "{colors.terminal-accent}"
---

# Clibble design system

## Creative direction

The reference is a friendly terminal game's title screen: a cursor leading into
character-art grass, a tiny creature glimpsed along the trail, and room to type.
The memorable feature is the prompt becoming a path. Clibble is pronounced
“clib-ul”; CLI is tucked into a name that sounds like a small companion.
“A little wild in your terminal.” is the invitation.

The terminal is the main product. Use a full-window RPG screen for interactive play, with a highlighted choice,
a centered clearing framed by green pixel pine trees, and a second-person narration
box. The opaque forest uses the user's terminal font. A soft fern fill and dark pine
text mark the selected action; the chevron also works without color.
The read-only web companion is a quiet collection album with pine and fern
accents, pale leaf surfaces, and the same prompt-and-cursor mark.

The primary audience is terminal-comfortable players who enjoy quiet nature games. English is the
supported interface language; the data source does not make this a Japan-market
product. The brand is playful; errors and recovery instructions remain precise.

## Ownership and mapping

[internal/brand/identity.json](internal/brand/identity.json) is the canonical
identity and palette. This file describes its roles. The Go brand package embeds
those tokens. `scripts/brand_assets.py` generates web CSS variables and vector
marks from the same source. The command renderer maps Fern to headings on the terminal's existing background.
The RPG renderer uses the identity file's `forest` palette: pine #10291e,
canopy #17392a, fern #20503a, leaf #3e6a43, selection #b1d18e, ivory #e3efd4.
These six colors have one canonical owner and are adapted to ANSI RGB in forest.go.
The README cover is generated raster artwork, with the exact name visible.

## Typography, layout, and states

The web view uses locally bundled Manrope with a system sans fallback. Headlines
are sentence case; numbers align with tabular figures. Body lines stay below
80 characters. The page uses an asymmetric field-guide composition: a collection
index and an adjacent creature record on wide screens, stacked on narrow screens.
Records stay readable at 320px; details have no fixed viewport height.

The default interactive CLI uses Bubble Tea's alternate screen. Menus use arrows
and Enter, a text chevron marks selection without relying on color, and the
narration addresses the player as you. Procedural half-block pixel scenery frames
centered menus, with a low-detail clearing behind text. Sparse drizzle, canopy
shimmer, and motes animate at 12 fps without changing gameplay. Focus loss stops
ticks; a saved Motion setting or session `-no-motion` flag makes the scene still.
NO_COLOR uses a static character backdrop. At 40 × 18 the same actions and
narration stay available; below that, show a resize hint and allow safe exit.
Long lists and records scroll; returning restores the prior selected item.
Long narration and errors have a visible Tab continuation hint and message pages.
The `-plain` mode keeps the linear command prompt, Tab completion, and bounded
in-memory history. Bubble Tea or the line editor restores terminal mode on exit.
Color is optional and never carries meaning alone. Piped output remains plain.

Web controls have explicit hover, focus, selected, disabled, and loading states.
Search has a clear action; filtering distinguishes an empty collection from no
matches. Failed refreshes preserve the last readable collection. Status is textual
and announced. The web app is read-only; collecting stays in the CLI.

## Shape and restraint

The mark is a chevron prompt followed by a block cursor on a pine tile. Its
square-ended strokes recall a character grid. Use it once in the header. Keep
records readable and controls familiar; the companion uses spacing and borders
rather than simulated window chrome. No glossy dioramas, neon, cinematic effects,
or third-party creature artwork in the product interface.

The cover uses pine #142a25, ivory #eef3d9, fern #a7c78b, sage #617e68,
and a small honey #e7be73 catch accent. These illustration colors are documented
separately from the runtime palette. Runtime primary #365b43 provides legible
links/buttons on the light companion; terminal fern #a7c78b is an optional accent.

## Verification

Use the Go test suite for save compatibility, command modes, and local web API
boundaries. Run the design-context linter and premium static audit. Inspect the
terminal at wide/narrow sizes, color off, and piped output. Browser verification
covers filtering, detail selection, keyboard focus, empty/error/loading states,
narrow layout, and reduced motion when a browser connection is available.
