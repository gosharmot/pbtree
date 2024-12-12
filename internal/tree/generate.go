package tree

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gosharmot/clog"
	"github.com/spf13/cobra"
)

const GenerateDir = ".generate"

type GenerateCommand struct {
	BufBin   string
	Template string
	Targets  []Target
}

func (t Tree) Generate(c *cobra.Command, command GenerateCommand) error {
	if len(command.Targets) == 0 {
		return nil
	}

	cmdArgs := []string{
		"generate",
		"--template", command.Template,
		"--output", filepath.Join(t.wd, t.vendorDir, GenerateDir),
	}

	var targetsAdded int
	for _, target := range command.Targets {
		if !target.NeedGenerate {
			continue
		}
		cmdArgs = append(cmdArgs, "--path", filepath.Join(t.wd, t.vendorDir, target.Module))
		targetsAdded++
	}

	if targetsAdded == 0 {
		return nil
	}

	cmdArgs = append(cmdArgs, filepath.Join(t.wd, t.vendorDir))

	cmd := exec.Command(command.BufBin, cmdArgs...)
	cmd.Stderr = c.OutOrStderr()
	cmd.Stdout = c.OutOrStdout()

	clog.Info(strings.ReplaceAll(cmd.String(), " -", "\n\t -"))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
