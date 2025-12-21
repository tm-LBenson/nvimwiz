# nvimwiz

A small Go + tview wizard that can:

- detect basic system info & tool availability
- optionally install user-local binaries (Neovim, ripgrep, fd) from GitHub releases
- manage a Neovim IDE config via generated modules + safe user overrides
- persist your choices so you can rerun it later

## Build

From the project root (the folder with `go.mod`):

```bash
go mod tidy
go build -o nvimwiz ./cmd/nvimwiz
./nvimwiz
```

## What gets written

### Profile

- `~/.config/nvimwiz/profile.json` (or `$XDG_CONFIG_HOME/nvimwiz/profile.json`)

### Neovim config (managed)

nvimwiz writes a small loader plus generated modules:

- `~/.config/nvim/init.lua` (managed stub, optional)
- `~/.config/nvim/lua/nvimwiz/loader.lua`
- `~/.config/nvim/lua/nvimwiz/generated/*.lua`
- `~/.config/nvim/lua/nvimwiz/user.lua` (never overwritten)

## Config modes

- **Managed**: nvimwiz writes `init.lua` (backs up your existing `init.lua` if it's not already managed).
- **Integrate**: nvimwiz does *not* touch `init.lua`. You add one line to your existing config:

```lua
require("nvimwiz.loader")
```

## Notes

- Installs binaries into `~/.local/bin` and `~/.local/nvim/...`
- nvimwiz updates `PATH` for the running wizard process so follow-up steps work without restarting.
