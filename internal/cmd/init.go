package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gosharmot/clog"
	"github.com/gosharmot/pbtree/internal/config"
	"github.com/spf13/cobra"
)

const (
	bufTemplate = `version: v1
plugins:
  - name: go
    path: bin/protoc-gen-go
    out: .
    opt:
      - paths=source_relative
  - name: grpc
    path: bin/protoc-gen-go-grpc
    out: .
    opt:
      - paths=source_relative
  - name: gw
    path: bin/protoc-gen-grpc-gateway
    out: .
    opt:
      - logtostderr=true
      - paths=source_relative
      - generate_unbound_methods=true
  - name: swagger
    path: bin/protoc-gen-openapiv2
    out: .
    opt:
      - generate_unbound_methods=true`
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "Init pbtree config",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: initF,
}

func init() {
	Init.Flags().StringVar(&configFile, "config", "pbtree.yaml", "pbtree config file")
	Init.Flags().BoolVar(&force, "force", false, "init without config checking")
	Init.Flags().StringVar(&vendorDir, "vendor-dir", ".vendorpb", "folder for vendoring files")
}

func initF(*cobra.Command, []string) {
	wd, err := os.Getwd()
	if err != nil {
		clog.Fatalf("get wd: %s", err)
	}

	if err = addVendorDirToGitignore(wd); err != nil {
		clog.Fatal(err)
	}
	if err = createBufGenFile(wd); err != nil {
		clog.Fatal(err)
	}
	if !force {
		_, err = os.Stat(filepath.Join(wd, "pbtree.yaml"))
		if err == nil {
			clog.Warning("config already exists")
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			clog.Fatalf("get stat: %s", err)
		}
	}

	newConfig, err := createConfig(filepath.Join(wd, configFile))
	if err != nil {
		clog.Fatalf("create config: %s", err)
	}
	defer func() { _ = newConfig.Close() }()
}

func createBufGenFile(wd string) error {
	bufGenPath := filepath.Join(wd, "buf.gen.yaml")
	file, ok, err := getOrCreate(bufGenPath)
	if err != nil {
		return err
	}
	if ok {
		return file.Close()
	}
	if _, err = file.Write([]byte(bufTemplate)); err != nil {
		return fmt.Errorf("write bug.gen.yaml: %s", err)
	}

	return nil
}

func addVendorDirToGitignore(wd string) error {
	gitignorePath := filepath.Join(wd, ".gitignore")
	gitignore, _, err := getOrCreate(gitignorePath)
	if err != nil {
		return err
	}

	all, err := io.ReadAll(gitignore)
	if err != nil {
		return fmt.Errorf("read .gitignore: %s", err)
	}

	if bytes.Contains(all, []byte(vendorDir)) {
		return nil
	}
	sep := ""
	if len(all) != 0 {
		sep = "\n"
	}
	if _, err = gitignore.Write([]byte(sep + vendorDir + "\nbin")); err != nil {
		return fmt.Errorf("write .gitignore: %s", err)
	}
	return nil
}

func getOrCreate(f string) (io.ReadWriteCloser, bool, error) {
	var file *os.File
	_, err := os.Stat(f)
	exists := err == nil
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, exists, fmt.Errorf("get stat: %s", err)
		}
		file, err = os.Create(f)
		if err != nil {
			return nil, exists, fmt.Errorf("create '%s': %s", f, err)
		}
	} else {
		file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, exists, fmt.Errorf("open '%s': %s", f, err)
		}
	}
	return file, exists, nil
}

func createConfig(f string) (io.ReadWriteCloser, error) {
	newConfig, err := os.Create(f)
	if err != nil {
		clog.Fatalf("create config: %s", err)
	}
	if _, err = newConfig.Write(config.Template); err != nil {
		clog.Fatalf("write date to config: %s", err)
	}

	return newConfig, nil
}
