# wgcf-cli
A command-line tool for Cloudflare-WARP API, built using Cobra.<br>
Thanks: [@fscarmen](https://github.com/fscarmen/), [fscarmen/warp](https://gitlab.com/fscarmen/warp/), [@badafans](https://github.com/badafans), [badafans/warp-reg](https://github.com/badafans/warp-reg)<br>
The name was taken from [ViRb3/wgcf](https://github.com/ViRb3/wgcf), but aside from the similar names, they are not actually related.
```
❯ wgcf-cli 
A command-line tool for Cloudflare-WARP API, built using Cobra.

Usage:
  wgcf-cli [command]

Available Commands:
  bind        Check current bind devices
  cancel      Cancel a account
  completion  Generate the autocompletion script for the specified shell
  generate    Generate a xray/sing-box config
  help        Help about any command
  license     Change to a new license
  name        Change the device name
  plus        Recharge your account indefinitely
  register    Register a new WARP account
  simplify    Simplify a config from original one
  unbind      Unbind from original license
  update      Update a config
  version     Print the version of wgcf-cli

Flags:
  -c, --config string   set configuration file path (default "wgcf.json")
  -h, --help            help for wgcf-cli

Use "wgcf-cli [command] --help" for more information about a command.
```
## Build 
```bash
make
```
- Available environment variable:
1. `GOFLAGS`, Default: 
```bash
-trimpath $(BUILD_MODE) -tags=$(TAGS) -mod=readonly -modcacherw -v -ldflags "$(LDFLAGS)"
```
2. `CGO_ENABLED`, Default: 
```bash
$(shell go env CGO_ENABLED)
```
3. `TAGS`, Default: ` `
