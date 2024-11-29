package client

import (
	"github.com/spf13/cobra"
)

func UploadCMD() *cobra.Command {
	var file string
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file",
		Long:  "Upload a file to a server",
		RunE: func(cmd *cobra.Command, args []string) error {

			return c
		},
	}
	uploadCmd.Flags().StringVarP(&file, "file", "-f", "", "upload a file")
	return uploadCmd
}
