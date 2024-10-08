package main

import (
	"fmt"
	"os"

	"github.com/ArchiveNetwork/wgcf-cli/utils"
	"github.com/spf13/cobra"
)

var unbindCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Unbind from original license",
	PreRun: func(cmd *cobra.Command, args []string) {
		client.New()
	},
	Run:     unbind,
	PostRun: update,
}

func init() {
	rootCmd.AddCommand(unbindCmd)
	unbindCmd.PersistentFlags().Bool("yes", false, "confirm that you want to unbind from original license")
	unbindCmd.MarkPersistentFlagRequired("yes")
}

func unbind(cmd *cobra.Command, args []string) {
	token, id := utils.GetTokenID(configPath)

	r := utils.Request{
		Action: "unbind",
		Payload: []byte(
			`{
				"active": false
			 }`,
		),
		ID:    id,
		Token: token,
	}
	requset, err := r.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if _, err = client.Do(requset); err != nil {
		client.HandleBody()
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Account unbinded (ID: %s) successfully\n", id)
}
