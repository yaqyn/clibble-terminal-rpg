![Clibble — A little wild in your terminal.](readme-assets/cover.webp)

# Clibble — A little wild in your terminal

Explore Pokemon locations, catch what you find, and build a collection worth coming back to.
Clibble is a terminal-first Pokemon catching adventure: discover locations, build a collection,
and read creature records without leaving your shell. An optional local browser view
complements the CLI when you want to browse saved records visually.
Your collection stays on your device. No account, API key, or subscription.

[Download the CLI](https://github.com/yaqyn/clibble/releases/latest) · [Start from source](#-installation) · [User guide](docs/GUIDE.md)

## <img src="readme-assets/icons/features.svg" width="24" height="24" alt=""> Features

- **Play in the terminal.** Choose with arrows and Enter, follow the narration, and inspect your catches.
- **Follow a new trail.** Browse location areas and discover the Pokemon that live there.
- **Make the catch.** Try your luck, save successful catches, and return to their records offline.
- **See your collection.** Search by name or type in the read-only web companion; compare stats and measurements.
- **Keep control.** Local saves, single-writer locks, bounded requests, and plain script output protect the useful parts of a small tool.

## <img src="readme-assets/icons/installation.svg" width="24" height="24" alt=""> Installation

Download the archive for your operating system and architecture from the
[releases page](https://github.com/yaqyn/clibble/releases/latest), verify it against
`SHA256SUMS`, and extract it. Run `clibble` on Linux/macOS or `clibble.exe` on Windows.
The compiled app includes the web interface and font; it needs no Go, Node, or Python runtime.
New exploration and catch data require internet access to PokéAPI.

To build from source, install the Go version declared in [go.mod](go.mod) or newer:

```bash
git clone https://github.com/yaqyn/clibble.git
cd clibble
go build -trimpath -o clibble .
```

On Windows, use `-o clibble.exe`. Go downloads pinned build dependencies on the first build.

## <img src="readme-assets/icons/launch.svg" width="24" height="24" alt=""> Quick start

```bash
./clibble
```

You arrive at the trailhead. Choose **Travel**, select a destination, then approach
a Pokemon and choose **Throw a Pokeball**. After a successful catch, choose
**Inspect your catch** to open its saved record. Your catches save automatically.

| Key | Action |
| --- | --- |
| Up / Down or `j` / `k` | Move through choices; scroll records and help |
| Enter / Space | Choose the highlighted action |
| Esc / Backspace | Go back; cancel travel or exploration |
| Left / Right | Browse destination pages |
| `t` / `c` / `?` | Open travel / collection / field guide |
| `r` | Retry a failed action |
| `q` / Ctrl+C | Leave safely |

The RPG screen needs a terminal at least 40 columns × 18 rows. It adapts to resizing
and restores your terminal when you leave. Location and menu position last for the
session; your caught Pokemon remain saved. Catches can escape; you can try again.

For the command prompt, use `./clibble -plain`. It keeps grouped help, Tab command
completion, and the last 100 commands in memory. Scripts and one-shot commands
continue to work directly.

## <img src="readme-assets/icons/usage.svg" width="24" height="24" alt=""> Usage

| Command | What it does |
| --- | --- |
| `help` | Show the field guide |
| `map` / `mapb` | Browse the next / previous page of location areas |
| `explore <area>` | List Pokemon in a location area |
| `catch <pokemon>` | Attempt a catch; higher base experience lowers the chance |
| `inspect <pokemon>` | Read a caught Pokemon's record without a network request |
| `collection` | List your caught Pokemon |
| `count` | Print the number of distinct caught Pokemon offline |
| `serve` | Start the local web companion; run this from your shell |
| `exit` | Leave the interactive session |

Commands and identifiers are case-insensitive. The original `pokedex` command
remains an alias for `collection`. Use single commands or pipe a batch:

```bash
./clibble inspect pidgey
./clibble collection
printf 'collection\nexit\n' | ./clibble
```

Pipes omit banners, color, and prompts. Results go to stdout; errors go to stderr.
A failed command produces a nonzero exit status. Batches continue after errors
and report failure at the end; an escaped catch is a valid outcome, not a command error.

### Optional browser companion

To open the collection in a browser, run this in another terminal:

```bash
./clibble serve
```

Open **http://127.0.0.1:8787**. Search your saved Pokemon by name or type, then choose
one to read its record. Catch in the CLI and use **Refresh collection** to see new
records. The companion is local and read-only; it does not expose writes or share your collection online.

See the [user guide](docs/GUIDE.md) for flags, storage locations, recovery, and examples.

## <img src="readme-assets/icons/configuration.svg" width="24" height="24" alt=""> Configuration

| Flag | Default | Purpose |
| --- | --- | --- |
| `-data <path>` | Platform user directory | Choose a collection file |
| `-plain` | Off | Use the command prompt instead of the RPG screen |
| `-ephemeral` | Off | Use a temporary CLI collection without saves |
| `-timeout <duration>` | `10s` | Bound each API request |
| `-color auto\|always\|never` | `auto` | Control interactive terminal color |
| `-port <number>` | `8787` | Select the local companion's port |
| `-version` / `-h` | — | Print version / flags |

Flags precede commands: `./clibble -data ./collection.json serve`.
Existing Pokedex saves remain accessible automatically; no collection is moved or reset.
See [save compatibility](docs/GUIDE.md#save-compatibility).

## <img src="readme-assets/icons/architecture.svg" width="24" height="24" alt=""> Architecture

The same Go binary runs the CLI and serves the companion. Typed PokéAPI responses
pass through a bounded cache; catches are stored as versioned local JSON. The web
view reads that collection and filters records locally, without an external service.

Read the [architecture guide](docs/ARCHITECTURE.md), [design system](DESIGN.md), and
[interaction contract](UX-CONTRACT.md). Brand tokens live in one
[identity file](internal/brand/identity.json) and generate the marks and palette adapters.

## <img src="readme-assets/icons/verification.svg" width="24" height="24" alt=""> Verification

```bash
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

For web interaction tests, install the development-only Node dependencies:

```bash
npm ci
npm test
```

Tests cover RPG navigation, cancellation, stale results, shutdown during saves,
small-screen layouts, catch/escape outcomes, pagination, cached requests, malformed responses,
timeouts, save recovery and locking, legacy-save discovery, plain CLI output,
local web API boundaries, search, selection, empty states, and refresh failures.
[CI](https://github.com/yaqyn/clibble/actions/workflows/ci.yml) runs native Go checks
on Linux, macOS, and Windows. See [release preparation](docs/RELEASE.md) for packaging.

## <img src="readme-assets/icons/security.svg" width="24" height="24" alt=""> Security and limitations

- This is an independent, single-user Pokemon tool. There are no accounts, sync service, battle simulation, or automatic backups.
- The companion binds to loopback only and rejects foreign hosts/origins. Its content policy allows bundled assets only; it has no write endpoint or third-party tracking.
- Saves are validated, bounded to 4 MiB, flushed before replacement, and protected by cooperating-writer locks. This does not replace backups or a database durability guarantee.
- API requests use HTTPS, explicit timeouts, response limits, and origin/path validation. Cache entries expire after five minutes and are bounded by count and response bytes.
- Pokemon can escape. Records can be inspected offline after a successful catch; new world data depends on PokéAPI availability.

## <img src="readme-assets/icons/license.svg" width="24" height="24" alt=""> License

[MIT](LICENSE). Dependency and font licenses are included in
[third-party notices](THIRD_PARTY_NOTICES.md). Pokemon names and concepts belong to
their respective owners; data is provided by [PokéAPI](https://pokeapi.co/).

> Based on Boot.dev's [Build a Pokedex in Go](https://www.boot.dev/courses/build-pokedex-cli-golang),
> extended with persistent saves, scriptable commands, a local collection companion,
> and the Clibble identity.
