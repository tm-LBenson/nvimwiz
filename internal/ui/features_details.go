package ui

import "strings"

func (w *Wizard) renderDetails(it itemRef) {
	title := w.itemTitle(it)

	lines := []string{title, ""}

	if it.Kind == itemFeature {
		if f, ok := w.cat.Features[it.ID]; ok {
			note := beginnerNote(it.ID, f.Title)
			if note != "" {
				lines = append(lines, "Info:", note, "")
			}

			desc := strings.TrimSpace(f.Long)
			if desc == "" {
				desc = strings.TrimSpace(f.Short)
			}
			if desc != "" {
				lines = append(lines, desc, "")
			}

			if len(f.Requires) > 0 {
				lines = append(lines, "Requires: "+strings.Join(f.Requires, ", "), "")
			}
		}
		lines = append(lines, "Current: "+w.itemActionLabel(it))
		w.detailView.SetText(strings.Join(trimTrailingEmpty(lines), "\n"))
		return
	}

	if it.Kind == itemChoice {
		if c, ok := w.cat.Choices[it.ID]; ok {
			note := beginnerNote(it.ID, c.Title)
			if note != "" {
				lines = append(lines, "Info:", note, "")
			}

			desc := strings.TrimSpace(c.Long)
			if desc == "" {
				desc = strings.TrimSpace(c.Short)
			}
			if desc != "" {
				lines = append(lines, desc, "")
			}

			lines = append(lines, "Options:")
			for _, opt := range c.Options {
				lines = append(lines, " - "+opt.Title)
			}
			lines = append(lines, "")
		}
		lines = append(lines, "Current: "+w.itemActionLabel(it))
		w.detailView.SetText(strings.Join(trimTrailingEmpty(lines), "\n"))
		return
	}
}

func beginnerNote(id, title string) string {
	low := strings.ToLower(title + " " + id)

	switch {
	case strings.Contains(low, "telescope"):
		return "Telescope is a fast finder. Use it to open files, search text (grep), and jump around your project without a mouse."
	case strings.Contains(low, "treesitter"):
		return "Treesitter improves syntax highlighting and code navigation using real parsing (not regex). It usually makes editing feel more \"IDE-like\"."
	case strings.Contains(low, "lsp"):
		return "LSP adds IDE features: autocomplete, go-to-definition, rename, diagnostics, hover docs. If you code, you almost always want this."
	case strings.Contains(low, "mason"):
		return "Mason installs language servers and dev tools (like TypeScript/Go/Python servers) so Neovim can provide IDE features."
	case strings.Contains(low, "nvim-tree") || strings.Contains(low, "file tree") || strings.Contains(low, "tree"):
		return "A file explorer sidebar (like VS Code). Useful if you prefer browsing folders visually."
	case strings.Contains(low, "harpoon"):
		return "Harpoon is a fast \"bookmark\" system for files. Mark a few files and jump between them instantly.\n\nKeys: <leader>ha add, <leader>hm menu, <leader>h1..h4 jump."
	case strings.Contains(low, "toggleterm") || strings.Contains(low, "terminal"):
		return "Adds an integrated terminal you can toggle from Neovim (bottom split), similar to VS Code's terminal.\n\nKey: <leader>tt toggles. In terminal mode, press <esc><esc> to return to normal mode."
	case strings.Contains(low, "ripgrep") || strings.Contains(low, " rg"):
		return "ripgrep (rg) is a very fast search tool. Many Neovim search features use it under the hood."
	case strings.Contains(low, "fd"):
		return "fd is a faster, friendlier 'find' command. Plugins use it to list files quickly."
	case strings.Contains(low, "neovim") || strings.Contains(low, "nvim"):
		return "This is the editor itself. Installing a stable version avoids random breakage from old distro packages."
	case strings.Contains(low, "statusline"):
		return "A statusline shows mode, file info, git branch, diagnostics, and more at the bottom of the screen."
	case strings.Contains(low, "theme") || strings.Contains(low, "colorscheme"):
		return "A theme changes the colors of the UI and syntax highlighting. Pick one that is comfortable for your eyes."
	}

	return ""
}
