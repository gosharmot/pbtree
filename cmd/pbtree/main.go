package main

import (
	"github.com/gosharmot/pbtree/internal/cmd"
	"github.com/spf13/cobra"
)

var pbtree = &cobra.Command{
	Use:   "pbtree",
	Short: "Proto generator",
	Long:  "Tool for vendoring and generation proto files",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	pbtree.SetHelpCommand(&cobra.Command{Hidden: true})
	pbtree.AddCommand(cmd.Vendor)
	pbtree.AddCommand(cmd.Init)
	pbtree.AddCommand(cmd.Add)
	_ = pbtree.Execute()
}
