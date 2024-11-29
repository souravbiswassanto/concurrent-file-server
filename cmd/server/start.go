package server

import (
	"github.com/souravbiswassanto/concurrent-file-server/internal/server"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"github.com/spf13/cobra"
)

func AddStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Starts the file-server",
		Long:  "run will starts the file server, you need to run this command before upload or download a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.SetupAndRunServer(util.HandleFunc{})
		},
	}
}
