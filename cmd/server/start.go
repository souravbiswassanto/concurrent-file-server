package server

import (
	"github.com/souravbiswassanto/concurrent-file-server/internal/server"
	"github.com/spf13/cobra"
	"log"
)

func AddStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Starts the file-server",
		Long:  "run will starts the file server, you need to run this command before upload or download a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := server.SetupAndRunServer(); err != nil {
				log.Println(err)
				return err
			}
			return nil
		},
	}
}
