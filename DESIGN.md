---
version: alpha
name: Clibble
description: A small Pokemon catching adventure that begins at the terminal prompt.
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

The terminal is the main product. Keep its actual output linear, quick to scan,
and respectful of the user's font and background. The cover can express the
imagined world; it must not claim an ASCII map renderer, battles, or movement.
The read-only web companion is a quiet collection album with pine and fern
accents, pale leaf surfaces, and the same prompt-and-cursor mark.

The primary audience is terminal-comfortable Pokemon fans. English is the
supported interface language; the data source does not make this a Japan-market
product. The brand is playful; errors and recovery instructions remain precise.

## Ownership and mapping

[internal/brand/identity.json](internal/brand/identity.json) is the canonical
identity and palette. This file describes its roles. The Go brand package embeds
those tokens. `scripts/brand_assets.py` generates web CSS variables and vector
marks from the same source. The terminal renderer maps Fern to headings on a
terminal's existing background; normal text inherits the terminal foreground.
The README cover is generated raster artwork, with the exact name visible.

## Typography, layout, and states

The web view uses locally bundled Manrope with a system sans fallback. Headlines
are sentence case; numbers align with tabular figures. Body lines stay below
80 characters. The page uses an asymmetric field-guide composition: a collection
index and an adjacent creature record on wide screens, stacked on narrow screens.
Records stay readable at 320px; details have no fixed viewport height.

The CLI uses the user's terminal font. It does not clear the screen automatically, animate, or hide the cursor.
The line editor temporarily uses raw input and restores the previous terminal
state on exit. It provides Tab completion and a bounded in-memory history. Narrow terminals get linear records.
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
