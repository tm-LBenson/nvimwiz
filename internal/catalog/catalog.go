package catalog

type Feature struct {
	ID       string
	Category string
	Title    string
	Short    string
	Long     string
	Default  bool
	Modules  []string
	Requires []string
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
	Default  string
	Options  []ChoiceOption
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
	categories := []string{"Install", "Core", "UI", "LSP", "Extras"}

	features := []Feature{
		{ID: "install.neovim", Category: "Install", Title: "Neovim", Short: "Install Neovim (user-local)", Long: "Installs Neovim to ~/.local/nvim and symlinks ~/.local/bin/nvim", Default: true},
		{ID: "install.ripgrep", Category: "Install", Title: "ripgrep", Short: "Install rg (user-local)", Long: "Installs ripgrep to ~/.local/bin/rg", Default: true},
		{ID: "install.fd", Category: "Install", Title: "fd", Short: "Install fd (user-local)", Long: "Installs fd to ~/.local/bin/fd", Default: true},

		{ID: "config.write", Category: "Core", Title: "Write config", Short: "Write/update Neovim config files", Long: "Writes a managed Neovim config under ~/.config/nvim and a safe user file under lua/nvimwiz/user.lua", Default: true},
		{ID: "config.lazysync", Category: "Core", Title: "Sync plugins", Short: "Run :Lazy sync headless", Long: "Runs Neovim headless to install/update plugins after writing config", Default: true},
		{ID: "core.dashboard", Category: "Core", Title: "Projects dashboard", Short: "Projects dashboard on startup", Long: "Shows a projects screen on startup and lets you pick or create projects", Default: true, Modules: []string{"nvimwiz.modules.core.dashboard_projects"}},
		{ID: "core.telescope", Category: "Core", Title: "Telescope", Short: "Fuzzy finder for files/grep", Long: "Adds Telescope with keymaps <leader>ff, <leader>fg, <leader>fb", Default: true, Modules: []string{"nvimwiz.modules.core.telescope"}},
		{ID: "core.treesitter", Category: "Core", Title: "Treesitter", Short: "Better syntax highlighting", Long: "Adds nvim-treesitter for improved highlighting and parsing", Default: true, Modules: []string{"nvimwiz.modules.core.treesitter"}},
		{ID: "core.completion", Category: "Core", Title: "Autocomplete", Short: "Autocomplete menu + snippets (VS Code-style)", Long: "Adds an autocomplete menu while you type (nvim-cmp) with snippet expansion (LuaSnip) and a large snippet collection (friendly-snippets). Recommended for new Neovim users.", Default: true, Modules: []string{"nvimwiz.modules.core.completion"}},

		{ID: "lsp.core", Category: "LSP", Title: "LSP core", Short: "Mason + lspconfig baseline", Long: "Enables Mason and nvim-lspconfig with sane defaults", Default: true, Modules: []string{"nvimwiz.modules.lsp.core"}},
		{ID: "lsp.typescript", Category: "LSP", Title: "TypeScript/JavaScript", Short: "TypeScript/JavaScript language server", Long: "Enables a TypeScript/JavaScript language server when available in your lspconfig version", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.python", Category: "LSP", Title: "Python", Short: "Python language server", Long: "Enables Pyright", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.web", Category: "LSP", Title: "HTML/CSS", Short: "HTML + CSS language servers", Long: "Enables html and cssls servers", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.emmet", Category: "LSP", Title: "Emmet", Short: "Emmet completions for HTML/CSS", Long: "Enables Emmet completions for HTML/CSS and related filetypes. Works best with Autocomplete enabled.", Default: false, Requires: []string{"lsp.core", "lsp.web", "core.completion"}},
		{ID: "lsp.go", Category: "LSP", Title: "Go", Short: "Go language server", Long: "Enables gopls", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.bash", Category: "LSP", Title: "Bash", Short: "Bash language server", Long: "Enables bashls", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.lua", Category: "LSP", Title: "Lua", Short: "Lua language server", Long: "Enables lua_ls and configures vim globals", Default: true, Requires: []string{"lsp.core"}},
		{ID: "lsp.java", Category: "LSP", Title: "Java", Short: "Java via jdtls", Long: "Enables Java support via nvim-jdtls and jdtls", Default: false, Requires: []string{"lsp.core"}, Modules: []string{"nvimwiz.modules.lsp.java"}},

		{ID: "qol.gitsigns", Category: "Extras", Title: "Git signs", Short: "Git hunk signs and actions", Long: "Adds gitsigns for inline git change indicators and actions", Default: true, Modules: []string{"nvimwiz.modules.extras.gitsigns"}},
		{ID: "qol.autopairs", Category: "Extras", Title: "Autopairs", Short: "Auto-close brackets/quotes", Long: "Adds autopairs for bracket/quote pairing", Default: true, Modules: []string{"nvimwiz.modules.extras.autopairs"}},
		{ID: "qol.whichkey", Category: "Extras", Title: "Which-key", Short: "Keymap helper popup", Long: "Adds which-key to help discover keymaps", Default: false, Modules: []string{"nvimwiz.modules.extras.whichkey"}},
		{ID: "qol.comment", Category: "Extras", Title: "Comment", Short: "Toggle comments easily", Long: "Adds comment.nvim for quick commenting", Default: false, Modules: []string{"nvimwiz.modules.extras.comment"}},
		{ID: "qol.harpoon", Category: "Extras", Title: "Harpoon", Short: "Quick file marks + jump list", Long: "Adds Harpoon for fast file bookmarking and jumping", Default: false, Modules: []string{"nvimwiz.modules.extras.harpoon"}},
		{ID: "qol.terminal", Category: "Extras", Title: "Terminal", Short: "Toggle a bottom terminal", Long: "Adds a toggleable terminal split (like VS Code integrated terminal)", Default: false, Modules: []string{"nvimwiz.modules.extras.terminal"}},
	}

	choices := []Choice{
		{
			Key:      "ui.explorer",
			Category: "UI",
			Title:    "Explorer",
			Short:    "File explorer UI",
			Long:     "Choose a file explorer behavior similar to VS Code or keep it minimal",
			Default:  "nvimtree",
			Options: []ChoiceOption{
				{ID: "nvimtree", Title: "nvim-tree", Short: "Tree on the left", Long: "A VS Code-like file tree on the left with toggle <leader>e", Modules: []string{"nvimwiz.modules.ui.explorer_nvimtree"}},
				{ID: "netrw", Title: "netrw", Short: "Built-in explorer", Long: "Uses Neovim's built-in netrw explorer", Modules: []string{"nvimwiz.modules.ui.explorer_netrw"}},
				{ID: "none", Title: "None", Short: "No explorer", Long: "No explorer configuration", Modules: nil},
			},
		},
		{
			Key:      "ui.theme",
			Category: "UI",
			Title:    "Theme",
			Short:    "Colorscheme",
			Long:     "Choose a colorscheme",
			Default:  "tokyonight",
			Options: []ChoiceOption{
				{ID: "tokyonight", Title: "Tokyo Night", Short: "tokyonight", Long: "tokyonight colorscheme", Modules: []string{"nvimwiz.modules.ui.theme_tokyonight"}},
				{ID: "catppuccin", Title: "Catppuccin", Short: "catppuccin", Long: "catppuccin colorscheme", Modules: []string{"nvimwiz.modules.ui.theme_catppuccin"}},
				{ID: "gruvbox", Title: "Gruvbox", Short: "gruvbox", Long: "gruvbox colorscheme", Modules: []string{"nvimwiz.modules.ui.theme_gruvbox"}},
				{ID: "github_dark", Title: "GitHub Dark", Short: "GitHub dark theme", Long: "GitHub Dark (similar to GitHub UI)", Modules: []string{"nvimwiz.modules.ui.theme_github"}},
				{ID: "rose_pine", Title: "Rose Pine", Short: "Soft pink/purple theme", Long: "Rose Pine (calm, muted colors)", Modules: []string{"nvimwiz.modules.ui.theme_rose_pine"}},
				{ID: "kanagawa", Title: "Kanagawa", Short: "Japanese ink-inspired theme", Long: "Kanagawa (high quality dark theme)", Modules: []string{"nvimwiz.modules.ui.theme_kanagawa"}},
				{ID: "none", Title: "None", Short: "default", Long: "Do not set a colorscheme", Modules: nil},
			},
		},
		{
			Key:      "ui.statusline",
			Category: "UI",
			Title:    "Statusline",
			Short:    "Statusline UI",
			Long:     "Choose a statusline",
			Default:  "lualine",
			Options: []ChoiceOption{
				{ID: "lualine", Title: "lualine", Short: "lualine", Long: "Adds lualine statusline", Modules: []string{"nvimwiz.modules.ui.statusline_lualine"}},
				{ID: "none", Title: "None", Short: "built-in", Long: "Use the built-in statusline", Modules: nil},
			},
		},
	}

	presets := []Preset{
		{
			ID:    "kickstart",
			Title: "Kickstart-like",
			Short: "Minimal starter with LSP + Telescope",
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"lsp.core":        true,
				"lsp.typescript":  true,
				"lsp.python":      true,
				"lsp.web":         true,
				"lsp.go":          true,
				"lsp.bash":        true,
				"lsp.lua":         true,
				"lsp.java":        false,
				"qol.gitsigns":    false,
				"qol.autopairs":   false,
				"qol.whichkey":    false,
				"qol.comment":     false,
			},
			Choices: map[string]string{
				"ui.explorer":   "netrw",
				"ui.theme":      "tokyonight",
				"ui.statusline": "none",
			},
		},
		{
			ID:    "lazyvim",
			Title: "LazyVim-like",
			Short: "IDE-like defaults with extras-style features",
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"lsp.core":        true,
				"lsp.typescript":  true,
				"lsp.python":      true,
				"lsp.web":         true,
				"lsp.go":          true,
				"lsp.bash":        true,
				"lsp.lua":         true,
				"lsp.java":        false,
				"qol.gitsigns":    true,
				"qol.autopairs":   true,
				"qol.whichkey":    true,
				"qol.comment":     true,
			},
			Choices: map[string]string{
				"ui.explorer":   "nvimtree",
				"ui.theme":      "tokyonight",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "astronvim",
			Title: "AstroNvim-like",
			Short: "Modular feel with strong UI defaults",
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"lsp.core":        true,
				"lsp.typescript":  true,
				"lsp.python":      true,
				"lsp.web":         true,
				"lsp.go":          true,
				"lsp.bash":        true,
				"lsp.lua":         true,
				"lsp.java":        false,
				"qol.gitsigns":    true,
				"qol.autopairs":   true,
				"qol.whichkey":    true,
				"qol.comment":     false,
			},
			Choices: map[string]string{
				"ui.explorer":   "nvimtree",
				"ui.theme":      "catppuccin",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "nvchad",
			Title: "NvChad-like",
			Short: "UI and theme forward defaults",
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"lsp.core":        true,
				"lsp.typescript":  true,
				"lsp.python":      true,
				"lsp.web":         true,
				"lsp.go":          true,
				"lsp.bash":        true,
				"lsp.lua":         true,
				"lsp.java":        false,
				"qol.gitsigns":    true,
				"qol.autopairs":   true,
				"qol.whichkey":    false,
				"qol.comment":     false,
			},
			Choices: map[string]string{
				"ui.explorer":   "nvimtree",
				"ui.theme":      "catppuccin",
				"ui.statusline": "lualine",
			},
		},
		{
			ID:    "lunarvim",
			Title: "LunarVim-like",
			Short: "More IDE-like defaults and helpers",
			Features: map[string]bool{
				"core.dashboard":  true,
				"core.telescope":  true,
				"core.treesitter": true,
				"lsp.core":        true,
				"lsp.typescript":  true,
				"lsp.python":      true,
				"lsp.web":         true,
				"lsp.go":          true,
				"lsp.bash":        true,
				"lsp.lua":         true,
				"lsp.java":        true,
				"qol.gitsigns":    true,
				"qol.autopairs":   true,
				"qol.whichkey":    true,
				"qol.comment":     true,
			},
			Choices: map[string]string{
				"ui.explorer":   "nvimtree",
				"ui.theme":      "gruvbox",
				"ui.statusline": "lualine",
			},
		},
	}

	fm := map[string]Feature{}
	for _, f := range features {
		fm[f.ID] = f
	}

	cm := map[string]Choice{}
	for _, c := range choices {
		cm[c.Key] = c
	}

	pm := map[string]Preset{}
	for _, p := range presets {
		pm[p.ID] = p
	}

	return Catalog{
		Features:   fm,
		Choices:    cm,
		Presets:    pm,
		Categories: categories,
	}
}
