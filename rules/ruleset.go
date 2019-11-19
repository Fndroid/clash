package rules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
)

type Ruleset struct {
	url         string
	rules       []C.Rule
	adapter     string
	isNoResolve bool
	lastUpdate  time.Time
}

func (r *Ruleset) RuleType() C.RuleType {
	return C.Ruleset
}

func (r *Ruleset) Match(metadata *C.Metadata) bool {
	for _, rule := range r.rules {
		if rule.Match(metadata) {
			return true
		}
	}
	return false
}

func (r *Ruleset) Adapter() string {
	return r.adapter
}

func (r *Ruleset) Payload() string {
	return r.url
}

func (r *Ruleset) NoResolveIP() bool {
	return r.isNoResolve
}

func trimArr(arr []string) (r []string) {
	for _, e := range arr {
		r = append(r, strings.Trim(e, " "))
	}
	return
}

func (r *Ruleset) Update() {
	isNoResolve := true

	rules := []C.Rule{}

	resp, err := http.Get(r.url)
	if err != nil {
		log.Errorln(err.Error())
		return
	}
	rawRules, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Errorln(err.Error())
		return
	}

	rulesConfig := strings.Split(string(rawRules), "\n")
	// parse rules
	for _, line := range rulesConfig {
		rule := trimArr(strings.Split(line, ","))
		var (
			payload string
			target  string
			params  = []string{}
		)

		target = r.adapter

		switch l := len(rule); {
		case l == 1:
		case l == 2:
			payload = rule[1]
		case l >= 3:
			payload = rule[1]
			params = rule[2:]
		default:
			continue
		}

		rule = trimArr(rule)
		params = trimArr(params)
		var (
			parseErr error
			parsed   C.Rule
		)

		switch rule[0] {
		case "DOMAIN":
			parsed = NewDomain(payload, target)
		case "DOMAIN-SUFFIX":
			parsed = NewDomainSuffix(payload, target)
		case "DOMAIN-KEYWORD":
			parsed = NewDomainKeyword(payload, target)
		case "GEOIP":
			noResolve := HasNoResolve(params)
			if !noResolve {
				isNoResolve = false
			}
			parsed = NewGEOIP(payload, target, noResolve)
		case "IP-CIDR", "IP-CIDR6":
			noResolve := HasNoResolve(params)
			if !noResolve {
				isNoResolve = false
			}
			parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRNoResolve(noResolve))
		// deprecated when bump to 1.0
		case "SOURCE-IP-CIDR":
			fallthrough
		case "SRC-IP-CIDR":
			parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))
		case "SRC-PORT":
			parsed, parseErr = NewPort(payload, target, true)
		case "DST-PORT":
			parsed, parseErr = NewPort(payload, target, false)
		case "MATCH":
			fallthrough
		// deprecated when bump to 1.0
		case "FINAL":
			parsed = NewMatch(target)
		default:
			parseErr = fmt.Errorf("unsupported rule type %s", rule[0])
		}

		if parseErr != nil {
			continue
		}

		rules = append(rules, parsed)
	}

	r.isNoResolve = isNoResolve
	r.rules = rules
	r.lastUpdate = time.Now()
}

func (r *Ruleset) LastUpdate() string {
	return r.lastUpdate.Format(time.UnixDate)
}

func NewRuleset(url string, adapter string) *Ruleset {
	rs := Ruleset{
		url:     url,
		adapter: adapter,
	}
	go rs.Update()
	return &rs
}
