package constant

import (
	"context"
	"time"
)

// Rule Type
const (
	Domain RuleType = iota
	DomainSuffix
	DomainKeyword
	Ruleset
	GEOIP
	IPCIDR
	SrcIPCIDR
	SrcPort
	DstPort
	MATCH
)

const DownloadTimeout = 3 * time.Second
const UpdateInterval = 24 * time.Hour

type RuleType int

func (rt RuleType) String() string {
	switch rt {
	case Domain:
		return "Domain"
	case DomainSuffix:
		return "DomainSuffix"
	case DomainKeyword:
		return "DomainKeyword"
	case Ruleset:
		return "Ruleset"
	case GEOIP:
		return "GeoIP"
	case IPCIDR:
		return "IPCIDR"
	case SrcIPCIDR:
		return "SrcIPCIDR"
	case SrcPort:
		return "SrcPort"
	case DstPort:
		return "DstPort"
	case MATCH:
		return "Match"
	default:
		return "Unknown"
	}
}

type Rule interface {
	RuleType() RuleType
	Match(metadata *Metadata) bool
	Adapter() string
	Payload() string
	NoResolveIP() bool
}

type RemoteRule interface {
	Rule
	Update(context.Context, chan RemoteRule)
	LastUpdate() string
	Destroy()
}
