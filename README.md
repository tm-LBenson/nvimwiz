# nvimwiz (refactor)

A TUI wizard that installs Neovim + a few CLI tools (user-local) and writes a modular Neovim config driven by a profile JSON.

## Build

```bash
./build.sh
```

Or:

```bash
go mod download
go build -o nvimwiz ./cmd/nvimwiz
```

## Run

```bash
./nvimwiz
```

The wizard stores its profile at:

- `$XDG_CONFIG_HOME/nvimwiz/profile.json` (if set)
- otherwise `~/.config/nvimwiz/profile.json`

## What it changes

- Installs binaries to `~/.local/bin` (symlink for Neovim points into `~/.local/nvim/<tag>/bin/nvim`)
- Writes Neovim config to `~/.config/nvim`
- Generated settings live at `~/.config/nvim/lua/nvimwiz/generated/config.lua`
- Safe user override file: `~/.config/nvim/lua/nvimwiz/user.lua` (never overwritten if it already exists)

## Config modes

- **managed**: writes `~/.config/nvim/init.lua` that loads `nvimwiz.loader`
- **integrate**: does not touch your `init.lua`. Add this yourself:

```lua
require("nvimwiz.loader")
```

## Adding a new feature/module

1. Create a module under `assets/nvim/lua/nvimwiz/modules/...` that exports either:
   - `spec()` returning a lazy.nvim spec list, and/or
   - `setup()` for non-plugin runtime setup
2. Register it in `internal/catalog/catalog.go` by adding:
   - a `Feature` (toggle) or
   - a `Choice` option

That is all. The wizard UI and config generator use the catalog as the single source of truth.

## Presets

Presets are “starting points” (Kickstart-like, LazyVim-like, AstroNvim-like, NvChad-like, LunarVim-like). They map onto this wizard’s feature/choice set and are not a copy of those projects.

