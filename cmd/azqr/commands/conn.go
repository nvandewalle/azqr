// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(conCmd)
}

var conCmd = &cobra.Command{
	Use:   "con",
	Short: "Scan Connection",
	Long:  "Scan Connection",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"con"})
	},
}
