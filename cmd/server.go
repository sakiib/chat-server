/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sakiib/chat-server/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts the chat server",
	Long:  `starts the chat server`,
	Run: func(cmd *cobra.Command, args []string) {
		server.StartChatServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
