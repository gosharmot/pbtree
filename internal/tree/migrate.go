package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (t Tree) Migrate(targets []Target) error {
	for _, target := range targets {
		if !target.NeedGenerate {
			continue
		}

		from := filepath.Dir(filepath.Join(t.wd, t.vendorDir, GenerateDir, target.Module))
		to := filepath.Dir(filepath.Join(t.wd, t.outputDir, target.Module))
		if target.Destination != "" {
			var ok bool
			_, target.Destination, ok = strings.Cut(target.Destination, t.projectRepo)
			if !ok {
				return errors.New("invalid destination")
			}
			to = filepath.Join(t.wd, target.Destination)
		}
		if err := os.MkdirAll(to, 0750); err != nil {
			return fmt.Errorf("failed to create dir %s: %v", to, err)
		}

		_, protoFilepath := filepath.Split(target.Module)
		protoFileName := strings.TrimSuffix(protoFilepath, filepath.Ext(protoFilepath))

		err := filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if from != path {
					return filepath.SkipDir
				}
				return nil
			}

			_, name := filepath.Split(path)
			dest := filepath.Join(to, filepath.Base(path))
			if strings.HasPrefix(name, protoFileName) {
				return os.Rename(path, dest)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	if err := os.RemoveAll(filepath.Join(t.wd, t.vendorDir, GenerateDir)); err != nil {
		return errors.New("failed to remove temp generated dir")
	}

	return nil
}
