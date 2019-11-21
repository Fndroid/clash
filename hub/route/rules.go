package route

import (
	"context"
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

	success := make(chan C.RemoteRule)
	ctx, cancel := context.WithTimeout(context.Background(), C.DownloadTimeout)
	count := 0
	defer cancel()
	for _, rule := range rawRules {
		if rule.RuleType() == C.Ruleset {
			if r, ok := rule.(C.RemoteRule); ok {
				count++
				go r.Update(ctx, success)
			}
		}
	}
	rules := []R{}
	if count == 0 {
		render.JSON(w, r, render.M{
			"rules": rules,
		})
		return
	}
	for {
		select {
		case rule := <-success:
			rules = append(rules, RemoteRule{
				Rule: Rule{
					Type:    rule.RuleType().String(),
					Payload: rule.Payload(),
					Proxy:   rule.Adapter(),
				},
				LastUpdate: rule.LastUpdate(),
			})
			if count == len(rules) {
				render.JSON(w, r, render.M{
					"rules": rules,
				})
				return
			}
		case <-ctx.Done():
			render.JSON(w, r, render.M{
				"rules": rules,
			})
			return
		}
	}
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
