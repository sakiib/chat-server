/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sakiib/chat-server/client"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "connects the client to the chat server",
	Long:  `connects the client to the chat server`,
	Run: func(cmd *cobra.Command, args []string) {
		client.ConnectClient()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
