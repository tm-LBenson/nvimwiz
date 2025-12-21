package tasks

import "context"

type Options struct {
	ProjectsDir string
	Leader      string
	LocalLeader string
	ConfigMode  string // "managed" or "integrate"

	InstallNeovim   bool
	InstallRipgrep  bool
	InstallFd       bool
	WriteNvimConfig bool

	EnableTree      bool
	EnableTelescope bool
	EnableLSP       bool
	EnableJava      bool

	RunLazySync      bool
	RequireChecksums bool
}

type Task struct {
	Name string
	Run  func(ctx context.Context, logf func(string)) error
}

func Plan(opts Options) []Task {
	var out []Task

	out = append(out, Task{
		Name: "Ensure ~/.local/bin exists",
		Run: func(ctx context.Context, logf func(string)) error {
			return ensureLocalBin(logf)
		},
	})

	if opts.InstallNeovim {
		out = append(out, Task{Name: "Install Neovim (stable) to ~/.local", Run: func(ctx context.Context, logf func(string)) error {
			return installNeovim(ctx, logf, opts.RequireChecksums)
		}})
	}
	if opts.InstallRipgrep {
		out = append(out, Task{Name: "Install ripgrep (rg) to ~/.local/bin", Run: func(ctx context.Context, logf func(string)) error {
			return installRipgrep(ctx, logf, opts.RequireChecksums)
		}})
	}
	if opts.InstallFd {
		out = append(out, Task{Name: "Install fd to ~/.local/bin", Run: func(ctx context.Context, logf func(string)) error {
			return installFd(ctx, logf, opts.RequireChecksums)
		}})
	}
	if opts.WriteNvimConfig {
		out = append(out, Task{Name: "Write Neovim config to ~/.config/nvim", Run: func(ctx context.Context, logf func(string)) error {
			return writeConfig(logf, opts)
		}})
	}
	if opts.RunLazySync {
		out = append(out, Task{Name: "Run :Lazy sync headless (best-effort)", Run: func(ctx context.Context, logf func(string)) error {
			return runLazySync(ctx, logf)
		}})
	}

	return out
}
