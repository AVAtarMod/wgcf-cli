package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"path"
	"strings"

	C "github.com/ArchiveNetwork/wgcf-cli/constant"
	"github.com/ArchiveNetwork/wgcf-cli/utils"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:       "generate",
	Short:     "Generate a xray/sing-box/wg-quick config",
	Run:       generate,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"--xray", "--xray-module", "--xray-tag", "--xray-indent-width", "--sing-box", "--wg", "--wg-quick", "--output-file"},
}

type OutputFileType int8

const (
	Stdout OutputFileType = iota
	Default
	Custom
)

type GeneratorType int8

const (
	Xray GeneratorType = iota
	SingBox
	WgQuick
	None
)

func (t GeneratorType) String() string {
	switch t {
	case Xray:
		return "xray"
	case SingBox:
		return "sing-box"
	case WgQuick:
		return "wg-quick"
	}
	return "unknown"
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().Bool(asString(Xray), false, "generate a xray config")
	generateCmd.Flags().Bool(asString(SingBox), false, "generate a sing-box config")
	generateCmd.Flags().Bool(asString(WgQuick), false, "generate a wg-quick config")
	generateCmd.Flags().Bool("wg", false, "see --"+asString(WgQuick))

	generateCmd.Flags().String("output-file", "default", "output file name. Supported values: 'default'/'stdout'/any file path")
	generateCmd.Flags().String(asString(Xray)+"-module", "", "xray top-level config module ('inbounds' as example). By default generate no top-level module")
	generateCmd.Flags().String(asString(Xray)+"-tag", "wireguard", "'Tag' field of xray config")
	generateCmd.Flags().Uint8(asString(Xray)+"-indent-width", 4, "indentation size for xray config")
}

func asString[V fmt.Stringer](object V) string {
	return V.String(object)
}

func detectOutputFileType(cmd *cobra.Command) (OutputFileType, error) {
	var err error
	path, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return Stdout, err
	}
	switch path {
	case "stdout":
		return Stdout, nil
	case "default":
		return Default, nil
	}
	return Custom, nil
}

func ternary[V any](condition bool, onTrue V, onFalse V) V {
	if condition {
		return onTrue
	}
	return onFalse
}

func detectGeneratorType(cmd *cobra.Command) (GeneratorType, error) {
	xray, _ := cmd.Flags().GetBool(asString(Xray))
	sing, _ := cmd.Flags().GetBool(asString(SingBox))
	wg, _ := cmd.Flags().GetBool(asString(WgQuick))
	if !wg {
		wg, _ = cmd.Flags().GetBool("wg")
	}

	var options uint8 = 0
	options |= ternary(xray, uint8(0b001), 0)
	options |= ternary(sing, uint8(0b010), 0)
	options |= ternary(wg, uint8(0b100), 0)
	if c := bits.OnesCount8(options); c != 1 {
		if c == 0 {
			return None, errors.New("generator not specified")
		} else {
			return None, errors.New("multiple generators not supported")
		}
	}

	if xray {
		return Xray, nil
	} else if sing {
		return SingBox, nil
	} else if wg {
		return WgQuick, nil
	}
	return None, nil
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

func getDefaultFilePath(generator GeneratorType) string {
	var baseName = strings.TrimSuffix(configPath, path.Ext(configPath))
	switch generator {
	case Xray:
		return baseName + ".xray.json"
	case SingBox:
		return baseName + ".sing-box.json"
	case WgQuick:
		return baseName + ".ini"
	}
	return ""
}

func Exit(err error, exitCode int) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(exitCode)
}
func ExitDefault(err error) {
	Exit(err, 1)
}

func generate(cmd *cobra.Command, args []string) {
	var err error
	var generator GeneratorType
	var outputType OutputFileType

	outputType, err = detectOutputFileType(cmd)
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
	case Xray:
		confModule, _ := cmd.Flags().GetString(asString(Xray) + "-module")
		tag, _ := cmd.Flags().GetString(asString(Xray) + "-tag")
		indentWidth, _ := cmd.Flags().GetUint8(asString(Xray) + "-indent-width")

		body, err = utils.GenXray(resStruct, tag, confModule, indentWidth)
		if err != nil {
			ExitDefault(err)
		}
	case SingBox:
		body, err = utils.GenSing(resStruct)
	case WgQuick:
		body, err = utils.GenWgQuick(resStruct)
	}
	if err != nil {
		ExitDefault(err)
	}

	switch outputType {
	case Stdout:
		_, err = fmt.Print(string(body))
		if err != nil {
			ExitDefault(err)
		}
	case Default:
		var filepath = getDefaultFilePath(generator)
		askOutputOverwrite(filepath)
		err = os.WriteFile(filepath, body, 0600)
		if err != nil {
			ExitDefault(err)
		}
		fmt.Printf("Generate %s configuration file '%s' (ID: %s) successfully\n", asString(generator), filepath, resStruct.ID)
	case Custom:
		filepath, _ := cmd.Flags().GetString("output-file")
		askOutputOverwrite(filepath)
		err = os.WriteFile(filepath, body, 0600)
		if err != nil {
			ExitDefault(err)
		}
		fmt.Printf("Generate %s configuration file '%s' (ID: %s) successfully\n", asString(generator), filepath, resStruct.ID)
	}
}
