package tree

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/gosharmot/clog"
	"github.com/gosharmot/pbtree/internal/config"
	"github.com/gosharmot/pbtree/internal/fetcher"
	"golang.org/x/sync/errgroup"
)

var (
	importRegexp    = regexp.MustCompile(`^import\s*"(.*/.*[.]proto)".*;$`)
	goPackageRegexp = regexp.MustCompile(`^option go_package = "(.*)";$`)
)

type TemplateTargets map[string][]Target

type Target struct {
	Module       string
	Destination  string
	Template     string
	NeedGenerate bool
}

func NewTarget(module, destination, template string, isLocal bool) Target {
	return Target{
		Module:       module,
		Destination:  destination,
		Template:     template,
		NeedGenerate: len(destination) != 0 || isLocal,
	}
}

type VendorCommand struct {
	Config           config.Config
	LocalTemplate    string
	ExternalTemplate string
	MFlags           map[string]string
}

type Tree struct {
	fetcher     fetcher.Fetcher
	vendored    map[Target]struct{}
	vendorDir   string
	wd          string
	outputDir   string
	projectRepo string
}

func New(fetcher fetcher.Fetcher, wd, vendorDir, outputDir, projectRepo string) Tree {
	return Tree{
		fetcher:     fetcher,
		vendored:    make(map[Target]struct{}),
		wd:          wd,
		vendorDir:   vendorDir,
		outputDir:   outputDir,
		projectRepo: projectRepo,
	}
}

func (t Tree) Vendor(ctx context.Context, command VendorCommand) (TemplateTargets, error) {
	eg := errgroup.Group{}
	mx := sync.RWMutex{}

	deps := make([]string, 0, 20)

	for _, proto := range command.Config.LocalProto {
		eg.Go(func() error {
			_deps, goPackage, err := t.vendor(ctx, proto)
			if err != nil {
				return fmt.Errorf("%s: %s", proto, err)
			}

			mx.Lock()
			defer mx.Unlock()
			t.vendored[NewTarget(proto, goPackage, command.LocalTemplate, true)] = struct{}{}

			deps = append(deps, _deps...)

			return nil
		})
	}

	for _, proto := range command.Config.ExternalProto {
		dst := command.MFlags[proto]
		target := NewTarget(proto, dst, command.ExternalTemplate, false)
		if _, ok := t.vendored[target]; ok {
			continue
		}
		t.vendored[target] = struct{}{}
		eg.Go(func() error {
			_deps, _, err := t.vendor(ctx, proto)
			if err != nil {
				return fmt.Errorf("%s: %s", proto, err)
			}

			if len(_deps) == 0 {
				return nil
			}

			mx.Lock()
			defer mx.Unlock()
			deps = append(deps, _deps...)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if len(deps) == 0 {
		targets := TemplateTargets{
			command.LocalTemplate:    make([]Target, 0, len(command.Config.LocalProto)),
			command.ExternalTemplate: make([]Target, 0, len(command.Config.ExternalProto)),
		}

		for target := range t.vendored {
			targets[target.Template] = append(targets[target.Template], target)
		}

		t.vendored = make(map[Target]struct{})
		return targets, nil
	}

	command.Config.LocalProto = nil
	command.Config.ExternalProto = deps
	return t.Vendor(ctx, command)
}

func (t Tree) vendor(ctx context.Context, proto string) ([]string, string, error) {
	imports := make([]string, 0)

	var goPackage string

	fetch, err := t.fetcher.Fetch(ctx, proto)
	if err != nil {
		if errors.Is(err, fetcher.ErrInappropriateFetcher) {
			clog.Errorf("fetch proto: %s: %s", proto, err)
			return nil, "", nil
		}
		return nil, "", err
	}
	defer func() { _ = fetch.Close() }()

	dst := filepath.Join(t.wd, t.vendorDir, proto)
	if dir := filepath.Dir(dst); dir != "" {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, "", err
		}
	}

	vendor, err := os.Create(dst)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = vendor.Close() }()

	scanner := bufio.NewScanner(fetch)
	for scanner.Scan() {
		line := scanner.Bytes()
		if _, err = vendor.Write(append(line, []byte("\n")...)); err != nil {
			return nil, "", err
		}

		if goPackageRegexp.Match(line) {
			goPackage = string(goPackageRegexp.ReplaceAll(line, []byte("$1")))
		}

		if importRegexp.Match(line) {
			imports = append(imports, string(importRegexp.ReplaceAll(line, []byte("$1"))))
		}
	}

	return imports, goPackage, nil
}
