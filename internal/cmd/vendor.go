package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/gosharmot/clog"
	"github.com/gosharmot/pbtree/internal/buf"
	"github.com/gosharmot/pbtree/internal/config"
	"github.com/gosharmot/pbtree/internal/fetcher"
	"github.com/gosharmot/pbtree/internal/tree"
	"github.com/spf13/cobra"
)

var Vendor = &cobra.Command{
	Use:   "vendor",
	Short: "Vendor and generate proto files",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	SilenceUsage: true,
	RunE:         vendorF,
}

func init() {
	Vendor.Flags().StringVar(&bufBin, "buf", "./bin/buf", "buf binary directory")
	Vendor.Flags().StringVar(&template, "template", "buf.gen.yaml", "buf template")
	Vendor.Flags().StringVar(&configFile, "config", "pbtree.yaml", "pbtree config file")
	Vendor.Flags().StringVar(&vendorDir, "vendor-dir", ".vendorpb", "folder for vendoring files")
	Vendor.Flags().StringVar(&outputDir, "output", "internal/pb", "folder for generated files")
	Vendor.Flags().StringVar(&projectRepo, "project", "", "project name")
	Vendor.Flags().StringVar(&tokenKey, "token-key", "GITHUB_TOKEN", "env key for github token (without token-key && token available only local provider)")
	Vendor.Flags().StringVar(&token, "token", "", "github token (without token-key && token available only local provider)")
	_ = Vendor.MarkFlagRequired("project")
	Vendor.SetErrPrefix(color.RedString("Fatal error:"))
}

var (
	wdFunc = os.Getwd
)

func vendorF(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	wd, err := wdFunc()
	if err != nil {
		return fmt.Errorf("get wd: %s", err)
	}

	if fromEnv := os.Getenv(tokenKey); fromEnv != "" {
		token = fromEnv
	}

	// parse buf.gen.yaml
	localBufPath := filepath.Join(wd, template)
	bufGen, err := buf.Parse(localBufPath)
	if err != nil {
		return fmt.Errorf("parse buf template: %s", err)
	}

	// parse -M flags
	mFlags, err := bufGen.MFlags()
	if err != nil {
		return fmt.Errorf("parse mFlags: %s", err)
	}

	// get buf.gen.yaml data with `go, grpc` plugins only
	externalTemplateContent, err := bufGen.ExternalPluginsOnly()
	if err != nil {
		return fmt.Errorf("generate buf template: %s", err)
	}

	externalBufPath := filepath.Join(wd, vendorDir, tree.GenerateDir, template)
	if dir := filepath.Dir(externalBufPath); dir != "" {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir: %s", err)
		}
	}

	// create buf.gen.yaml for external proto
	templateFile, err := os.Create(externalBufPath)
	if err != nil {
		return fmt.Errorf("generate buf template file: %s", err)
	}
	defer func() { _ = templateFile.Close() }()

	_, err = templateFile.Write(externalTemplateContent)
	if err != nil {
		return fmt.Errorf("write data to template: %s", err)
	}

	// parse pbtree.yaml
	cfg, err := config.Parse(filepath.Join(wd, configFile))
	if err != nil {
		return fmt.Errorf("parse config: %s", err)
	}
	if len(cfg.LocalProto) == 0 && len(cfg.ExternalProto) == 0 {
		clog.Warning("no proto in config")
		return nil
	}

	fetchers := []fetcher.Fetcher{fetcher.NewLocal(wd)}
	if token != "" {
		fetchers = append(fetchers, fetcher.NewGithub(token))
	}

	treeManager := tree.New(fetcher.NewCompoundFetcher(fetchers...), wd, vendorDir, outputDir, projectRepo)

	clog.Info("Vendoring...")
	templateTargets, err := treeManager.Vendor(ctx, tree.VendorCommand{
		Config:           cfg,
		LocalTemplate:    localBufPath,
		ExternalTemplate: externalBufPath,
		MFlags:           mFlags,
	})
	if err != nil {
		return fmt.Errorf("vendor proto: %s", err)
	}

	toMigrate := make([]tree.Target, 0, len(cfg.LocalProto)+len(cfg.ExternalProto))
	for bufTemplatePath, targets := range templateTargets {
		err = treeManager.Generate(cmd, tree.GenerateCommand{
			BufBin:   bufBin,
			Template: bufTemplatePath,
			Targets:  targets,
		})
		if err != nil {
			return fmt.Errorf("generate: %s", err)
		}
		toMigrate = append(toMigrate, targets...)
	}

	// migrate to dst and remove .generate dir
	if err = treeManager.Migrate(toMigrate); err != nil {
		return fmt.Errorf("migrate: %s", err)
	}

	return nil
}
