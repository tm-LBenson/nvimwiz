package tasks

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"

	"nvimwiz/internal/catalog"
	"nvimwiz/internal/install"
	"nvimwiz/internal/nvimcfg"
	"nvimwiz/internal/profile"
)

type State struct {
	NvimPath string
	RgPath   string
	FdPath   string
}

type Task struct {
	Name string
	Run  func(ctx context.Context, st *State, log func(string)) error
}

func Plan(p profile.Profile, cat catalog.Catalog) []Task {
	plan := []Task{}

	if p.Features["install.neovim"] {
		plan = append(plan, Task{
			Name: "Install Neovim",
			Run: func(ctx context.Context, st *State, log func(string)) error {
				path, err := install.InstallNeovim(ctx, p.Verify, log)
				if err != nil {
					return err
				}
				st.NvimPath = path
				return nil
			},
		})
	}

	if p.Features["install.ripgrep"] {
		plan = append(plan, Task{
			Name: "Install ripgrep",
			Run: func(ctx context.Context, st *State, log func(string)) error {
				path, err := install.InstallRipgrep(ctx, p.Verify, log)
				if err != nil {
					return err
				}
				st.RgPath = path
				return nil
			},
		})
	}

	if p.Features["install.fd"] {
		plan = append(plan, Task{
			Name: "Install fd",
			Run: func(ctx context.Context, st *State, log func(string)) error {
				path, err := install.InstallFd(ctx, p.Verify, log)
				if err != nil {
					return err
				}
				st.FdPath = path
				return nil
			},
		})
	}

	if p.Features["config.write"] {
		plan = append(plan, Task{
			Name: "Write Neovim config",
			Run: func(ctx context.Context, st *State, log func(string)) error {
				_ = ctx
				return nvimcfg.Write(p, cat, log)
			},
		})
	}

	if p.Features["config.lazysync"] && p.Features["config.write"] {
		plan = append(plan, Task{
			Name: "Sync plugins",
			Run: func(ctx context.Context, st *State, log func(string)) error {
				bin := st.NvimPath
				if bin == "" {
					p, err := exec.LookPath("nvim")
					if err == nil {
						bin = p
					}
				}
				if bin == "" {
					return errors.New("nvim not found for plugin sync")
				}
				cfgDir, err := nvimcfg.ConfigDir()
				if err != nil {
					return err
				}
				headless := filepath.Join(cfgDir, "nvimwiz_headless_init.vim")
				cmd := exec.CommandContext(ctx, bin, "--headless", "-u", headless, "+Lazy! sync", "+qa")
				b, err := cmd.CombinedOutput()
				if log != nil && len(b) > 0 {
					log(string(b))
				}
				return err
			},
		})
	}

	return plan
}

func RunAll(ctx context.Context, plan []Task, log func(string), progress func(done, total int)) error {
	st := &State{}
	total := len(plan)
	for i, t := range plan {
		if log != nil {
			log("== " + t.Name + " ==")
		}
		if err := t.Run(ctx, st, log); err != nil {
			if log != nil {
				log("Error: " + err.Error())
			}
			return err
		}
		if progress != nil {
			progress(i+1, total)
		}
	}
	return nil
}
