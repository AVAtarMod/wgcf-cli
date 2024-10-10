package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	C "github.com/ArchiveNetwork/wgcf-cli/constant"
	E "github.com/ArchiveNetwork/wgcf-cli/enum"
	"github.com/ArchiveNetwork/wgcf-cli/utils"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:       "generate",
	Short:     "Generate a xray/sing-box/wg-quick config",
	Run:       generate,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"--xray", "--xray-module", "--xray-endpoint", "--xray-tag", "--xray-indent-width", "--sing-box", "--wg", "--wg-quick", "--output-file"},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().Bool(asString(E.Xray), false, "generate a xray config")
	generateCmd.Flags().Bool(asString(E.SingBox), false, "generate a sing-box config")
	generateCmd.Flags().Bool(asString(E.WgQuick), false, "generate a wg-quick config")
	generateCmd.Flags().Bool("wg", false, "see --"+asString(E.WgQuick))

	generateCmd.Flags().String("output-file", "default", "output file name. Supported values: 'default'/'stdout'/any file path")
	generateCmd.Flags().String(asString(E.Xray)+"-module", "", "xray top-level config module ('inbounds' as example). By default generate no top-level module")
	generateCmd.Flags().String(asString(E.Xray)+"-tag", "wireguard", "'Tag' field of xray config")
	generateCmd.Flags().Uint8(asString(E.Xray)+"-indent-width", 4, "indentation size for xray config")
	generateCmd.Flags().String(asString(E.Xray)+"-endpoint", "domain", "endpoint type to use. Supported values: 'domain'/'ip_v4'/'ip_v6'")
}

func asString[V fmt.Stringer](object V) string {
	return V.String(object)
}

func countTrue(args ...bool) uint {
	var true_count uint = 0
	for _, v := range args {
		if v {
			true_count += 1
		}
	}
	return true_count
}

func askOutputOverwrite(path string) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		var input string
		fmt.Fprintf(os.Stderr, "Warn: File %s exist, it will be overwritten. Continue? [y/N]: ", path)
		fmt.Scanln(&input)
		input = strings.ToLower(input)
		if input != "y" {
			os.Exit(1)
		}
	}
}

func getDefaultFilePath(generator E.GeneratorType) string {
	var base_name = strings.TrimSuffix(configPath, path.Ext(configPath))
	switch generator {
	case E.Xray:
		return base_name + ".xray.json"
	case E.SingBox:
		return base_name + ".sing-box.json"
	case E.WgQuick:
		return base_name + ".ini"
	}
	return ""
}

func Exit(err error, exit_code int) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(exit_code)
}
func ExitDefault(err error) {
	Exit(err, 1)
}

func generate(cmd *cobra.Command, args []string) {
	var err error
	var generator E.GeneratorType
	var output_type E.OutputFileType

	output_type, err = detectOutputFileType(cmd)
	if err != nil {
		ExitDefault(err)
	}
	generator, err = detectGeneratorType(cmd)
	if err != nil {
		ExitDefault(err)
	}

	var resStruct C.Response
	body := utils.ReadConfig(configPath)
	err = json.Unmarshal(body, &resStruct)
	if err != nil {
		ExitDefault(err)
	}

	switch generator {
	case E.Xray:
		conf_module, _ := cmd.Flags().GetString(asString(E.Xray) + "-module")
		tag, _ := cmd.Flags().GetString(asString(E.Xray) + "-tag")
		indent_width, _ := cmd.Flags().GetUint8(asString(E.Xray) + "-indent-width")
		endpoint_type, err := detectEndpointType(cmd)
		if err != nil {
			ExitDefault(err)
		}

		body, err = utils.GenXray(resStruct, tag, conf_module, indent_width, endpoint_type)
		if err != nil {
			ExitDefault(err)
		}
	case E.SingBox:
		body, err = utils.GenSing(resStruct)
	case E.WgQuick:
		body, err = utils.GenWgQuick(resStruct)
	}
	if err != nil {
		ExitDefault(err)
	}

	switch output_type {
	case E.Stdout:
		_, err = fmt.Print(string(body))
		if err != nil {
			ExitDefault(err)
		}
	case E.Default:
		var filepath = getDefaultFilePath(generator)
		askOutputOverwrite(filepath)
		err = os.WriteFile(filepath, body, 0600)
		if err != nil {
			ExitDefault(err)
		}
		fmt.Printf("Generate %s configuration file '%s' (ID: %s) successfully\n", asString(generator), filepath, resStruct.ID)
	case E.Custom:
		filepath, _ := cmd.Flags().GetString("output-file")
		askOutputOverwrite(filepath)
		err = os.WriteFile(filepath, body, 0600)
		if err != nil {
			ExitDefault(err)
		}
		fmt.Printf("Generate %s configuration file '%s' (ID: %s) successfully\n", asString(generator), filepath, resStruct.ID)
	}
}

func detectGeneratorType(cmd *cobra.Command) (E.GeneratorType, error) {
	xray, _ := cmd.Flags().GetBool(asString(E.Xray))
	sing, _ := cmd.Flags().GetBool(asString(E.SingBox))
	wg, _ := cmd.Flags().GetBool(asString(E.WgQuick))
	if !wg {
		wg, _ = cmd.Flags().GetBool("wg")
	}

	var flagsEnabled = countTrue(xray, sing, wg)
	if flagsEnabled != 1 {
		if flagsEnabled == 0 {
			return E.None, errors.New("generator not specified")
		} else {
			return E.None, errors.New("multiple generators not supported")
		}
	}

	if xray {
		return E.Xray, nil
	} else if sing {
		return E.SingBox, nil
	} else if wg {
		return E.WgQuick, nil
	}
	return E.None, nil
}

func detectOutputFileType(cmd *cobra.Command) (E.OutputFileType, error) {
	var err error
	path, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return E.Stdout, err
	}
	switch path {
	case "stdout":
		return E.Stdout, nil
	case "default":
		return E.Default, nil
	}
	return E.Custom, nil
}

func detectEndpointType(cmd *cobra.Command) (E.EndpointType, error) {
	endpoint_type, err := cmd.Flags().GetString(asString(E.Xray) + "-endpoint")
	if err != nil {
		return E.Domain, err
	}
	switch endpoint_type {
	case "domain":
		return E.Domain, nil
	case "ip_v4":
		return E.IPv4, nil
	case "ip_v6":
		return E.IPv6, nil
	}
	return E.Domain, errors.New("unsupported endpoint type")
}
