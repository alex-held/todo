package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "adds a todo",
		Long:  "adds a todo",
		Run: func(cmd *cobra.Command, args []string) {
			_ = strings.Join(args, " ")

		},
	}
}
