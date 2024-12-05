/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	cmc "github.com/souravbiswassanto/concurrent-file-server/cmd/client"
	cms "github.com/souravbiswassanto/concurrent-file-server/cmd/server"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "file-server",
	Short: "Short",
	Long:  `Long`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ip := os.Getenv("SERVER_IP")
		port := os.Getenv("SERVER_PORT")
		if port == "" || ip == "" {
			return fmt.Errorf("SERVER_IP and SERVER_PORT env variable can't be empty")
		}
		return nil
	},
	//Run: func(cmd *cobra.Command, args []string) {
	//	cmd.Help()
	//},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cms.AddStartCmd())
	rootCmd.AddCommand(cmc.UploadCMD())
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
