package catalog

import "sort"

type Feature struct {
	ID       string
	Category string
	Title    string
	Short    string
	Long     string
	Default  bool
	Requires []string
	Modules  []string
}

type ChoiceOption struct {
	ID      string
	Title   string
	Short   string
	Long    string
	Modules []string
}

type Choice struct {
	Key      string
	Category string
	Title    string
	Short    string
	Long     string
	Options  []ChoiceOption
	Default  string
}

type Preset struct {
	ID       string
	Title    string
	Short    string
	Features map[string]bool
	Choices  map[string]string
}

type Catalog struct {
	Features   map[string]Feature
	Choices    map[string]Choice
	Presets    map[string]Preset
	Categories []string
}

func Get() Catalog {
	cats := []string{"Install", "Core", "UI", "LSP", "Extras"}

	features := []Feature{
		{
			ID:       "install.neovim",
			Category: "Install",
			Title:    "Neovim",
			Short:    "Install or update Neovim.",
			Long: `What it does
- Downloads the Neovim stable release for your OS/arch.
- Installs it under ~/.local/nvim/<version>/bin/nvim.
- Keeps your system package manager install untouched.

Why you want it
- You get a consistent Neovim version for this setup.
- New users avoid distro packages that can lag behind.

How to verify
- Run: nvim --version
- The wizard also detects the version and will skip if you're already up to date.`,
			Default: true,
			Modules: []string{"install/neovim"},
		},
		{
			ID:       "install.ripgrep",
			Category: "Install",
			Title:    "Ripgrep (rg)",
			Short:    "Fast text search used by Telescope.",
			Long: `What it does
- Installs ripgrep (rg) under ~/.local/bin.

Why you want it
- Telescope uses rg for fast project-wide searching (live_grep).
- Many plugins assume rg exists for text search.

Repo
- https://github.com/BurntSushi/ripgrep

How to verify
- Run: rg --version`,
			Default: true,
			Modules: []string{"install/rg"},
		},
		{
			ID:       "install.fd",
			Category: "Install",
			Title:    "fd",
			Short:    "Fast file finder used by Telescope.",
			Long: `What it does
- Installs fd under ~/.local/bin.

Why you want it
- Telescope uses fd for fast file picking (find_files).
- Much faster and nicer defaults than find.

Repo
- https://github.com/sharkdp/fd

How to verify
- Run: fd --version`,
			Default: true,
			Modules: []string{"install/fd"},
		},
		{
			ID:       "config.write",
			Category: "Core",
			Title:    "Write config",
			Short:    "Write the Neovim config for this profile.",
			Long: `What it does
- Copies the bundled Neovim config template.
- Writes generated config based on your preset, features, and choices.

Where it writes
- Target: system config
  - Writes to ~/.config/nvim
- Target: safe build
  - Writes to ~/.config/<build name>
  - Launch with NVIM_APPNAME=<build name> nvim

Why you want it
- This is the step that actually applies your selections.`,
			Default: true,
			Modules: []string{"config/write"},
		},
		{
			ID:       "config.lazysync",
			Category: "Core",
			Title:    "Sync plugins",
			Short:    "Install/update plugins after writing the config.",
			Long: `What it does
- Runs Neovim headless.
- Triggers lazy.nvim to install/update plugins.

Why you want it
- First-time setup pulls down all plugins so Neovim is ready immediately.
- Updates can be done without manually opening Neovim.

If something fails
- If a download times out (504) or your network drops, use Retry.
- You can also run this later inside Neovim:
  - :Lazy sync

Repo
- https://github.com/folke/lazy.nvim`,
			Default:  true,
			Requires: []string{"config.write"},
			Modules:  []string{"config/lazysync"},
		},
		{
			ID:       "core.dashboard",
			Category: "Core",
			Title:    "Dashboard",
			Short:    "Start screen with quick actions.",
			Long: `What it does
- Shows a start screen when Neovim launches.
- Typically includes recent files, projects, and common actions.

Why you want it
- Great for beginners: you start with obvious next steps.
- Makes it feel more like an IDE landing page.

Tip
- If you prefer a blank start, disable it and you'll land in an empty buffer.`,
			Default: true,
			Modules: []string{"core/dashboard"},
		},
		{
			ID:       "core.telescope",
			Category: "Core",
			Title:    "Telescope",
			Short:    "Fuzzy finder for files, text, help, etc.",
			Long: `What it does
- Adds Telescope pickers for files, grep, buffers, help, and more.

Why you want it
- This is the "VS Code Ctrl+P" experience in Neovim.
- It becomes the main way you jump around a project.

Dependencies
- ripgrep (rg) for live_grep
- fd for fast file searching

Repo
- https://github.com/nvim-telescope/telescope.nvim`,
			Default:  true,
			Requires: []string{"install.ripgrep", "install.fd"},
			Modules:  []string{"core/telescope"},
		},
		{
			ID:       "core.treesitter",
			Category: "Core",
			Title:    "Treesitter",
			Short:    "Better syntax highlighting and text objects.",
			Long: `What it does
- Uses Tree-sitter parsers for better highlighting and code awareness.
- Enables smarter indentation and additional text objects in many setups.

Why you want it
- Better highlighting than regex-based syntax.
- Makes editing feel more modern and IDE-like.

Repo
- https://github.com/nvim-treesitter/nvim-treesitter`,
			Default: true,
			Modules: []string{"core/treesitter"},
		},
		{
			ID:       "core.completion",
			Category: "Core",
			Title:    "Completion",
			Short:    "Autocomplete menu while you type.",
			Long: `What it does
- Adds completion UI (nvim-cmp).
- Adds snippet expansion (LuaSnip).
- Includes a large snippet collection (friendly-snippets).

Why you want it
- This is the biggest quality-of-life upgrade for VS Code users.
- Completion integrates with LSP so it feels like a modern IDE.

Repos
- https://github.com/hrsh7th/nvim-cmp
- https://github.com/L3MON4D3/LuaSnip
- https://github.com/rafamadriz/friendly-snippets`,
			Default: true,
			Modules: []string{"core/cmp"},
		},
		{
			ID:       "lsp.core",
			Category: "LSP",
			Title:    "LSP core",
			Short:    "Language Server Protocol support.",
			Long: `What it does
- Enables Neovim's built-in LSP client.
- Adds common LSP keymaps and sensible defaults.

Why you want it
- Hover docs, go-to-definition, rename, references, diagnostics.
- This is what makes Neovim behave like an IDE.

Important note
- This config sets up the client. You still need language servers installed
  on your system for each language.

Repo
- https://github.com/neovim/nvim-lspconfig`,
			Default:  true,
			Requires: []string{"core.completion"},
			Modules:  []string{"lsp/core"},
		},
		{
			ID:       "lsp.go",
			Category: "LSP",
			Title:    "Go",
			Short:    "gopls config and Go tooling.",
			Long: `What it does
- Configures the Go language server (gopls) and common defaults.

What you still need
- Install gopls:
  - go install golang.org/x/tools/gopls@latest

Repo
- https://github.com/golang/tools/tree/master/gopls`,
			Default:  false,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/go"},
		},
		{
			ID:       "lsp.python",
			Category: "LSP",
			Title:    "Python",
			Short:    "pyright config and Python tooling.",
			Long: `What it does
- Configures the Python language server (pyright).

What you still need
- Install pyright:
  - npm i -g pyright

Repo
- https://github.com/microsoft/pyright`,
			Default:  false,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/python"},
		},
		{
			ID:       "lsp.typescript",
			Category: "LSP",
			Title:    "TypeScript/JavaScript",
			Short:    "tsserver config and web tooling.",
			Long: `What it does
- Configures TypeScript/JavaScript LSP.

What you still need
- Requires Node.js and a language server installed via npm.

Common servers
- typescript-language-server

Repo
- https://github.com/typescript-language-server/typescript-language-server`,
			Default:  false,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/typescript"},
		},
		{
			ID:       "lsp.web",
			Category: "LSP",
			Title:    "Web (HTML/CSS/JSON)",
			Short:    "HTML/CSS/JSON language servers.",
			Long: `What it does
- Configures common web language servers:
  - html
  - cssls
  - jsonls

What you still need
- Requires Node.js and the servers installed via npm.

Tip
- Enable this if you want a VS Code-like HTML/CSS editing experience.

Repos
- https://github.com/hrsh7th/vscode-langservers-extracted
- https://github.com/microsoft/vscode-json-languageservice`,
			Default:  false,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/web"},
		},
		{
			ID:       "lsp.bash",
			Category: "LSP",
			Title:    "Bash",
			Short:    "bash-language-server config.",
			Long: `What it does
- Configures bash-language-server.

What you still need
- Install the server:
  - npm i -g bash-language-server

Repo
- https://github.com/bash-lsp/bash-language-server`,
			Default:  false,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/bash"},
		},
		{
			ID:       "lsp.lua",
			Category: "LSP",
			Title:    "Lua",
			Short:    "lua-language-server config.",
			Long: `What it does
- Configures lua-language-server for editing Neovim configs and Lua projects.

Why you want it
- If you customize your config (or write Lua), this makes it much nicer.

What you still need
- Install lua-language-server using your package manager.

Repo
- https://github.com/LuaLS/lua-language-server`,
			Default:  true,
			Requires: []string{"lsp.core"},
			Modules:  []string{"lsp/lua"},
		},
		{
			ID:       "extra.gitsigns",
			Category: "Extras",
			Title:    "Git signs",
			Short:    "Git gutter signs and hunk actions.",
			Long: `What it does
- Shows git diff indicators in the sign column.
- Provides simple hunk actions (stage/reset/preview).

Why you want it
- Similar feedback to VS Code's source control gutter.

Repo
- https://github.com/lewis6991/gitsigns.nvim`,
			Default: true,
			Modules: []string{"extras/gitsigns"},
		},
		{
			ID:       "extra.autopairs",
			Category: "Extras",
			Title:    "Autopairs",
			Short:    "Auto-close brackets and quotes.",
			Long: `What it does
- Automatically inserts closing pairs like (), {}, "".

Why you want it
- Keeps typing flow fast, especially for beginners.

Repo
- https://github.com/windwp/nvim-autopairs`,
			Default: true,
			Modules: []string{"extras/autopairs"},
		},
		{
			ID:       "extra.whichkey",
			Category: "Extras",
			Title:    "Which-key",
			Short:    "Popup of available keybinds.",
			Long: `What it does
- Shows available keybinds as you start a key sequence.

Why you want it
- Huge beginner help: you don't need to memorize everything.
- It makes Neovim feel discoverable.

Repo
- https://github.com/folke/which-key.nvim`,
			Default: true,
			Modules: []string{"extras/whichkey"},
		},
		{
			ID:       "extra.comment",
			Category: "Extras",
			Title:    "Comment",
			Short:    "Toggle comments easily.",
			Long: `What it does
- Adds gc mappings to comment/uncomment code quickly.

Why you want it
- VS Code-style comment toggling.

Repo
- https://github.com/numToStr/Comment.nvim`,
			Default: true,
			Modules: []string{"extras/comment"},
		},
		{
			ID:       "extra.emmet",
			Category: "Extras",
			Title:    "Emmet",
			Short:    "HTML/CSS abbreviation expansion.",
			Long: `What it does
- Adds Emmet-style abbreviation expansion for HTML/CSS (and often JSX/TSX).
- Example: div>ul>li*3 then expand.

Why you want it
- This is a common "missing piece" for VS Code users.

Notes
- This is best paired with Web LSP enabled.

Repo
- https://github.com/olrtg/nvim-emmet`,
			Default:  true,
			Requires: []string{"core.completion", "lsp.web"},
			Modules:  []string{"extras/emmet"},
		},
		{
			ID:       "extra.harpoon",
			Category: "Extras",
			Title:    "Harpoon",
			Short:    "Quick file marking and navigation.",
			Long: `What it does
- Lets you "mark" files and jump between them instantly.

Why you want it
- Great for working across a handful of files repeatedly.
- Common workflow: mark your main files, then bounce between them.

Notes
- Keymaps are configured by the nvimwiz modules.
- If you want to change them, edit the Harpoon module config in the generated config.

Repo
- https://github.com/ThePrimeagen/harpoon`,
			Default: true,
			Modules: []string{"extras/harpoon"},
		},
		{
			ID:       "extra.terminal",
			Category: "Extras",
			Title:    "Terminal",
			Short:    "Toggle a terminal inside Neovim.",
			Long: `What it does
- Adds a toggleable terminal panel inside Neovim.

Why you want it
- Run tests, build, git, or a REPL without leaving Neovim.
- Makes the editor feel more like a full IDE.

Notes
- The terminal can be opened/closed with a keybind configured in the module.

Repo
- https://github.com/akinsho/toggleterm.nvim`,
			Default: true,
			Modules: []string{"extras/toggleterm"},
		},
	}

	choices := []Choice{
		{
			Key:      "ui.theme",
			Category: "UI",
			Title:    "Theme",
			Short:    "Pick a colorscheme.",
			Long: `What it controls
- The colorscheme used by Neovim.

How to think about it
- Pick the theme that feels easiest on your eyes.
- If you're coming from GitHub/VS Code, GitHub Dark is familiar.

Tip
- You can always switch later; this is a safe, reversible choice.`,
			Default: "tokyonight",
			Options: []ChoiceOption{
				{
					ID:    "tokyonight",
					Title: "TokyoNight",
					Short: "Dark theme with good contrast.",
					Long: `Why pick it
- Strong contrast and very popular defaults.
- Multiple built-in "styles" in many configs.

Repo
- https://github.com/folke/tokyonight.nvim`,
					Modules: []string{"ui/themes/tokyonight"},
				},
				{
					ID:    "catppuccin",
					Title: "Catppuccin",
					Short: "Pastel theme with many flavors.",
					Long: `Why pick it
- Soft/pastel look with multiple variants (latte/frappe/macchiato/mocha).
- Great if you want less harsh contrast.

Repo
- https://github.com/catppuccin/nvim`,
					Modules: []string{"ui/themes/catppuccin"},
				},
				{
					ID:    "gruvbox",
					Title: "Gruvbox",
					Short: "Classic warm retro theme.",
					Long: `Why pick it
- Warm, retro palette.
- Popular across Vim and terminal communities.

Repo
- https://github.com/ellisonleao/gruvbox.nvim`,
					Modules: []string{"ui/themes/gruvbox"},
				},
				{
					ID:    "github_dark",
					Title: "GitHub Dark",
					Short: "GitHub’s dark colorscheme.",
					Long: `Why pick it
- Familiar if you spend a lot of time on GitHub.
- Feels close to GitHub/VS Code styling.

Repo
- https://github.com/projekt0n/github-nvim-theme`,
					Modules: []string{"ui/themes/github"},
				},
				{
					ID:    "rose_pine",
					Title: "Rose Pine",
					Short: "Soft and elegant dark theme.",
					Long: `Why pick it
- Calm, low-contrast feel.
- Great for long sessions.

Repo
- https://github.com/rose-pine/neovim`,
					Modules: []string{"ui/themes/rose_pine"},
				},
				{
					ID:    "kanagawa",
					Title: "Kanagawa",
					Short: "Japanese-inspired theme.",
					Long: `Why pick it
- Nice balance of contrast and color.
- Very popular dark theme.

Repo
- https://github.com/rebelot/kanagawa.nvim`,
					Modules: []string{"ui/themes/kanagawa"},
				},
				{
					ID:    "none",
					Title: "None",
					Short: "Use Neovim defaults.",
					Long: `Why pick it
- You want the default Neovim look, with no theme plugin.

Tradeoffs
- You miss out on polished highlight groups many themes provide.`,
					Modules: []string{},
				},
			},
		},
		{
			Key:      "ui.explorer",
			Category: "UI",
			Title:    "File explorer",
			Short:    "Pick a file tree / explorer UI.",
			Long: `What it controls
- How you browse and open files from inside Neovim.

How to think about it
- If you want a VS Code-like sidebar, choose nvim-tree.
- If you want zero plugin overhead, choose netrw.

Tip
- Even with an explorer, many people primarily use Telescope for navigation.`,
			Default: "nvimtree",
			Options: []ChoiceOption{
				{
					ID:    "nvimtree",
					Title: "nvim-tree",
					Short: "Sidebar tree (VS Code-like).",
					Long: `Layout
- Usually a left-side panel you can toggle on/off.

Why pick it
- The closest mental model to VS Code's explorer.
- Easy to discover and use for beginners.

Tradeoffs
- Adds a plugin and some startup/runtime overhead compared to netrw.

Repo
- https://github.com/nvim-tree/nvim-tree.lua`,
					Modules: []string{"ui/explorer/nvimtree"},
				},
				{
					ID:    "netrw",
					Title: "netrw",
					Short: "Built-in explorer (no plugin).",
					Long: `Layout
- Built into Neovim. Opens in a normal buffer/window.

Why pick it
- Always available, simple, fast.
- Zero plugin dependencies.

Tradeoffs
- Less featureful and less "IDE-like" than a sidebar tree.

Docs
- :help netrw`,
					Modules: []string{"ui/explorer/netrw"},
				},
				{
					ID:    "none",
					Title: "None",
					Short: "No explorer plugin.",
					Long: `Why pick it
- You want a minimal setup.
- You plan to navigate mostly via Telescope (:Telescope find_files) or the terminal.

Tip
- You can still use :e, :Ex (netrw), and other built-ins.`,
					Modules: []string{},
				},
			},
		},
		{
			Key:      "ui.statusline",
			Category: "UI",
			Title:    "Status line",
			Short:    "Pick a status line plugin.",
			Long: `What it controls
- The information bar at the bottom of the Neovim window.

How to think about it
- If you want a polished, informative bar, choose lualine.
- If you want minimal/no plugin overhead, choose None.`,
			Default: "lualine",
			Options: []ChoiceOption{
				{
					ID:    "lualine",
					Title: "lualine",
					Short: "Popular Lua statusline.",
					Long: `Why pick it
- Widely used, fast, and very configurable.
- Commonly shows mode, git branch, diagnostics, file info, etc.

Repo
- https://github.com/nvim-lualine/lualine.nvim`,
					Modules: []string{"ui/statusline/lualine"},
				},
				{
					ID:    "none",
					Title: "None",
					Short: "Neovim defaults.",
					Long: `Why pick it
- You want to keep things simple.
- You prefer the built-in statusline and plan to customize later.`,
					Modules: []string{},
				},
			},
		},
	}

	presets := []Preset{
		{
			ID:    "kickstart",
			Title: "Kickstart-like",
			Short: `Who it's for
- Beginners who want a minimal, understandable starting point.

What it feels like
- Small set of core tools (LSP + Telescope) with a clean baseline.

Upstream inspiration
- https://github.com/nvim-lua/kickstart.nvim`,
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"core.completion": true,
				"lsp.core":        true,
				"lsp.lua":         true,
				"extra.whichkey":  true,
			},
			Choices: map[string]string{
				"ui.theme":      "tokyonight",
				"ui.explorer":   "nvimtree",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "lazyvim",
			Title: "LazyVim-like",
			Short: `Who it's for
- Users who want a more "IDE-like" default without doing a lot of setup.

What it feels like
- A solid baseline with common quality-of-life extras.
- Great "first Neovim" preset.

Upstream inspiration
- https://github.com/LazyVim/LazyVim`,
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"core.completion": true,
				"lsp.core":        true,
				"lsp.lua":         true,
				"extra.whichkey":  true,
				"extra.gitsigns":  true,
				"extra.autopairs": true,
			},
			Choices: map[string]string{
				"ui.theme":      "tokyonight",
				"ui.explorer":   "nvimtree",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "astronvim",
			Title: "AstroNvim-like",
			Short: `Who it's for
- Users who want a modular feel and strong UI defaults.

What it feels like
- A bit more opinionated and featureful.

Upstream inspiration
- https://github.com/AstroNvim/AstroNvim`,
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"core.completion": true,
				"lsp.core":        true,
				"lsp.lua":         true,
				"extra.whichkey":  true,
				"extra.gitsigns":  true,
				"extra.autopairs": true,
				"extra.comment":   true,
			},
			Choices: map[string]string{
				"ui.theme":      "tokyonight",
				"ui.explorer":   "nvimtree",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "nvchad",
			Title: "NvChad-like",
			Short: `Who it's for
- Users who want a fast, UI-forward Neovim experience.

What it feels like
- Snappy and polished.

Upstream inspiration
- https://github.com/NvChad/NvChad`,
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"core.completion": true,
				"lsp.core":        true,
				"lsp.lua":         true,
				"extra.whichkey":  true,
				"extra.gitsigns":  true,
			},
			Choices: map[string]string{
				"ui.theme":      "github_dark",
				"ui.explorer":   "nvimtree",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "lunarvim",
			Title: "LunarVim-like",
			Short: `Who it's for
- Users who want lots of features enabled out of the box.

What it feels like
- More plugins and more opinionated behavior.
- Closer to a "batteries included" IDE.

Upstream inspiration
- https://github.com/LunarVim/LunarVim`,
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"core.completion": true,
				"lsp.core":        true,
				"lsp.lua":         true,
				"extra.whichkey":  true,
				"extra.gitsigns":  true,
				"extra.autopairs": true,
				"extra.comment":   true,
				"extra.harpoon":   true,
				"extra.terminal":  true,
				"extra.emmet":     true,
			},
			Choices: map[string]string{
				"ui.theme":      "tokyonight",
				"ui.explorer":   "nvimtree",
				"ui.statusline": "lualine",
			},
		},
	}

	cat := Catalog{
		Features:   map[string]Feature{},
		Choices:    map[string]Choice{},
		Presets:    map[string]Preset{},
		Categories: cats,
	}

	for _, f := range features {
		cat.Features[f.ID] = f
	}

	for _, c := range choices {
		sort.Slice(c.Options, func(i, j int) bool {
			return c.Options[i].Title < c.Options[j].Title
		})
		cat.Choices[c.Key] = c
	}

	for _, p := range presets {
		cat.Presets[p.ID] = p
	}

	return cat
}
