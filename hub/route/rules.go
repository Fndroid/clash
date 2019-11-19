package route

import (
	"net/http"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/tunnel"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func ruleRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getRules)
	r.Put("/", updateRulesets)
	return r
}

type R interface{}

type Rule struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
}

type RemoteRule struct {
	Rule
	LastUpdate string `json:"last-update"`
}

func updateRulesets(w http.ResponseWriter, r *http.Request) {
	rawRules := tunnel.Rules()

	for _, rule := range rawRules {
		if rule.RuleType() == C.Ruleset {
			if r, ok := rule.(C.RemoteRule); ok {
				go r.Update()
			}
		}
	}

	render.Status(r, 204)
}

func getRules(w http.ResponseWriter, r *http.Request) {
	rawRules := tunnel.Rules()

	rules := []R{}
	for _, rule := range rawRules {
		r := Rule{
			Type:    rule.RuleType().String(),
			Payload: rule.Payload(),
			Proxy:   rule.Adapter(),
		}
		if mr, ok := rule.(C.RemoteRule); ok {
			rules = append(rules, RemoteRule{
				Rule:       r,
				LastUpdate: mr.LastUpdate(),
			})
		} else {
			rules = append(rules, r)
		}
	}

	render.JSON(w, r, render.M{
		"rules": rules,
	})
}
