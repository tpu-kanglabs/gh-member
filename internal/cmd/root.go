package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "member",
	Short: "Manage GitHub Organization members",
}

// Execute はルートコマンドを実行する。main.go から呼ぶ。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newListCmd())
}
