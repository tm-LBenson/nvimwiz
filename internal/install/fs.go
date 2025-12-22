package install

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func findFile(root, name string) (string, error) {
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == name {
			found = path
			return errors.New("stop")
		}
		return nil
	})
	if found == "" {
		return "", os.ErrNotExist
	}
	return found, nil
}

func replaceSymlink(linkPath, target string) error {
	_ = os.Remove(linkPath)
	return os.Symlink(target, linkPath)
}
