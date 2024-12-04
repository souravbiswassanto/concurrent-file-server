package client

import (
	"github.com/souravbiswassanto/concurrent-file-server/internal/handler"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"github.com/spf13/cobra"
)

func UploadCMD() *cobra.Command {
	var uc util.UploadConfig
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file",
		Long:  "Upload a file to a server",
		RunE: func(cmd *cobra.Command, args []string) error {

			return handler.HandleUpload(uc)
		},
	}
	uploadCmd.Flags().StringVarP(&uc.File, "file", "f", "", "upload a file")
	uploadCmd.Flags().StringVarP(&uc.Protocol, "proto", "t", "tcp", "connection protocol type")
	uploadCmd.Flags().Int32VarP(&uc.ChunkSize, "buf", "b", 10240, "chunk to send each tcp iteration")

	return uploadCmd
}
