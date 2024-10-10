package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	C "github.com/ArchiveNetwork/wgcf-cli/constant"
	"github.com/ArchiveNetwork/wgcf-cli/utils"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:    "register",
	Short:  "Register a new WARP account",
	PreRun: pre_register,
	Run:    register,
}

var (
	teamToken string
)

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.PersistentFlags().StringVarP(&teamToken, "token", "t", "", "set register ZeroTrust Token")
}

func pre_register(cmd *cobra.Command, args []string) {
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		var input string
		fmt.Fprintf(os.Stderr, "Warn: File %s exist, are you sure to continue? [y/N]: ", configPath)
		fmt.Scanln(&input)
		input = strings.ToLower(input)
		if input != "y" {
			os.Exit(1)
		}
	}
	client.New()
}

func removePortFromIp(address string) (string, error) {
	const sep string = ":"
	slice := strings.Split(address, sep)
	if len(slice) < 2 {
		return "", errors.New("invalid address " + address)
	}
	return strings.Join(slice[0:len(slice)-1], sep), nil
}

func register(cmd *cobra.Command, args []string) {
	privateKey, publicKey := utils.GenerateKey()
	fmt.Printf("Generated public key: %s", publicKey)
	r := utils.Request{
		Payload: []byte(
			`{
				"key":"` + publicKey + `",
				"install_id":"",
				"fcm_token":"",
				"model":"",
				"serial_number":""
			}`,
		),
		Action:    "register",
		TeamToken: teamToken,
	}

	request, err := r.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	body, err := client.Do(request)
	if err != nil {
		client.HandleBody()
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	var resStruct C.Response
	if err = json.Unmarshal(body, &resStruct); err != nil {
		ExitDefault(err)
	}

	err_handler := func(err error) {
		if err != nil {
			ExitDefault(errors.New("cannot process API response. Reason: " + err.Error()))
		}
	}
	processed_peer_v4, err := removePortFromIp(resStruct.Config.Peers[0].Endpoint.V4)
	err_handler(err)
	processed_peer_v6, err := removePortFromIp(resStruct.Config.Peers[0].Endpoint.V6)
	err_handler(err)

	resStruct.Config.ReservedDec, resStruct.Config.ReservedHex = utils.ClientIDtoReserved(resStruct.Config.ClientID)
	resStruct.Config.PrivateKey = privateKey
	resStruct.Config.Peers[0].Endpoint.V4 = processed_peer_v4
	resStruct.Config.Peers[0].Endpoint.V6 = processed_peer_v6

	utils.SimplifyOutput(resStruct)

	store, err := json.MarshalIndent(resStruct, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err = os.WriteFile(configPath, store, 0600); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

}
