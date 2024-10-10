package enum

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

type EndpointType uint8

const (
	Domain EndpointType = iota
	IPv4
	IPv6
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

func (ep EndpointType) String() string {
	switch ep {
	case Domain:
		return "domain"
	case IPv4:
		return "ip_v4"
	case IPv6:
		return "ip_v6"
	}
	return "unknown"
}
