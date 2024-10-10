package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	C "github.com/ArchiveNetwork/wgcf-cli/constant"
	E "github.com/ArchiveNetwork/wgcf-cli/enum"
)

func constructAddress(resStruct C.Response, endpoint_type E.EndpointType) string {
	var peer *C.ResponsePeer = &resStruct.Config.Peers[0]
	switch endpoint_type {
	case E.Domain:
		return peer.Endpoint.Host
	case E.IPv4:
		return peer.Endpoint.V4 + ":" + strconv.Itoa(int(peer.Endpoint.Ports[0]))
	case E.IPv6:
		return peer.Endpoint.V6 + ":" + strconv.Itoa(int(peer.Endpoint.Ports[0]))
	}
	return ""
}

func GenXray(resStruct C.Response, tag string, config_module string, indent_size uint8, endpoint_type E.EndpointType) (body []byte, err error) {
	config_body_json := C.Xray{
		Protocol: "wireguard",
		Settings: C.XraySettings{
			SecretKey: resStruct.Config.PrivateKey,
			Address:   []string{resStruct.Config.Interface.Addresses.V4 + "/32", resStruct.Config.Interface.Addresses.V6 + "/128"},
			Peers: []struct {
				PublicKey  string   `json:"publicKey"`
				AllowedIPs []string `json:"allowedIPs"`
				Endpoint   string   `json:"endpoint"`
			}{
				{
					PublicKey:  resStruct.Config.Peers[0].PublicKey,
					AllowedIPs: []string{"0.0.0.0/0", "::/0"},
					Endpoint:   constructAddress(resStruct, endpoint_type),
				},
			},
			Reserved: resStruct.Config.ReservedDec,
			MTU:      1280,
		},
		Tag: tag,
	}

	indent := strings.Repeat(" ", int(indent_size))
	if config_module == "" {
		body, err = json.MarshalIndent(config_body_json, "", indent)
	} else {
		var config_json = map[string][]C.Xray{config_module: {config_body_json}}
		body, err = json.MarshalIndent(config_json, "", indent)
	}
	return
}

func GenSing(resStruct C.Response) (body []byte, err error) {
	in_struct := C.Sing{
		Type:          "wireguard",
		Tag:           "wireguard-out",
		Server:        resStruct.Config.Peers[0].Endpoint.Host,
		ServerPort:    2408,
		LocalAddress:  []string{resStruct.Config.Interface.Addresses.V4 + "/32", resStruct.Config.Interface.Addresses.V6 + "/128"},
		PrivateKey:    resStruct.Config.PrivateKey,
		PeerPublicKey: "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
		Reserved:      resStruct.Config.ClientID,
		MTU:           1280,
	}

	body, err = json.MarshalIndent(in_struct, "", "    ")
	return
}

func GenWgQuick(resStruct C.Response) (body []byte, err error) {
	in_str := fmt.Sprint(`
[Interface]
PrivateKey = ` + resStruct.Config.PrivateKey + `
Address = ` + resStruct.Config.Interface.Addresses.V4 + `/32
Address = ` + resStruct.Config.Interface.Addresses.V6 + `/128
MTU = 1280

Table = 300

PreUp = ip rule add oif %i lookup 300
PostDown = ip rule del oif %i lookup 300
PreUp = ip -6 rule add oif %i lookup 300
PostDown = ip -6 rule del oif %i lookup 300

PreUp = ip rule add fwmark 32975 lookup 300
PostDown = ip rule del fwmark 32975 lookup 300
PreUp = ip -6 rule add fwmark 32975 lookup 300
PostDown = ip -6 rule del fwmark 32975 lookup 300

#PreUp = ip rule add from ` + resStruct.Config.Interface.Addresses.V4 + `/32 lookup 300
#PostDown = ip rule del from ` + resStruct.Config.Interface.Addresses.V4 + `/32 lookup 300
#PreUp = ip -6 rule add from ` + resStruct.Config.Interface.Addresses.V6 + `/128 lookup 300
#PostDown = ip -6 rule del from ` + resStruct.Config.Interface.Addresses.V6 + `/128 lookup 300
# Alternative

PostUp = iptables -t mangle -A OUTPUT -s ` + resStruct.Config.Interface.Addresses.V4 + ` -j MARK --set-mark 32975
PreDown = iptables -t mangle -D OUTPUT -s ` + resStruct.Config.Interface.Addresses.V4 + ` -j MARK --set-mark 32975
PostUp = ip6tables -t mangle -A OUTPUT -s ` + resStruct.Config.Interface.Addresses.V6 + ` -j MARK --set-mark 32975
PreDown = ip6tables -t mangle -D OUTPUT -s ` + resStruct.Config.Interface.Addresses.V6 + ` -j MARK --set-mark 32975

[Peer]
PublicKey = ` + resStruct.Config.Peers[0].PublicKey + `
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = ` + resStruct.Config.Peers[0].Endpoint.V4 + `
`)
	body = []byte(in_str)
	return
}
