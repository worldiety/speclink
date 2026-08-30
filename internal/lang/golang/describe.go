package golang

import (
	"path"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// DescribeArchitecture states the shape this project is held to.
//
// Every sentence here corresponds to a rule that actually runs against this
// configuration. That is the whole discipline of the file: a description is a
// promise to the reader that something is checked, and one that outruns the
// checks is a lie that a clean run will never catch. So the layering rules
// appear only where the layering rules run, and every path comes from the
// project's own configuration rather than the profile's defaults, because a
// project that moved its contexts elsewhere is held to the place it moved
// them to.
func (m *Model) DescribeArchitecture() ir.Architecture {
	a := ir.Architecture{Style: m.Style.Name}
	cfg := m.Layout

	ctxRoot := or(cfg.ContextRoot, "app")
	cmdRoot := or(cfg.CmdRoot, "cmd")

	a.Layers = append(a.Layers, ir.Layer{
		Name:    "Bounded contexts",
		Where:   ctxRoot + "/<context>",
		Purpose: "One context per part of the business. What a context knows is its own; nothing reaches across.",
	})
	if m.Layered {
		a.Layers = append(a.Layers,
			ir.Layer{
				Name:    "Domain",
				Where:   ctxRoot + "/<context>",
				Purpose: "The rules of the business, in terms of the business. It names no protocol and no store.",
			},
			ir.Layer{
				Name:    "Presentation",
				Where:   ctxRoot + "/<context>/{rest,cli}",
				Purpose: "How the outside reaches a use case. One package per way in.",
			},
			ir.Layer{
				Name:    "Adapter",
				Where:   ctxRoot + "/<context>/adapter/<technology>",
				Purpose: "The other side of a port: the code that speaks to a database, a file system or another service.",
			})
	}
	a.Layers = append(a.Layers, ir.Layer{
		Name:    "Entry points",
		Where:   cmdRoot + "/<binary>",
		Purpose: "Where the program is assembled. The only place allowed to choose which adapter is used.",
	})
	// One row, however many directories are reserved. Repeating the same name
	// and the same sentence once per path reads as several different layers.
	if infra := infraRoots(cfg.InfraRoots, cfg.FoundationRoot); len(infra) > 0 {
		a.Layers = append(a.Layers, ir.Layer{
			Name:    "Infrastructure",
			Where:   strings.Join(infra, ", "),
			Purpose: "Technical helpers that know nothing about the business, and may not learn. Shared by every context, which is why they must stay ignorant of all of them.",
		})
	}

	// Rules that run for every Go profile.
	a.Rules = append(a.Rules,
		ir.Rule{
			ID:        "K8-MAIN-EXISTS",
			Statement: "The module has a main package.",
			Why:       "A module with no entry point is a library, and every statement in this document about what the system does would be about something that never runs.",
		},
		ir.Rule{
			ID:        "K8-MAIN-LOCATION",
			Statement: "Every main package lives under " + cmdRoot + "/.",
			Why:       "The place a program is assembled is the place its dependencies are chosen. Scattering that makes the wiring impossible to find and impossible to review.",
		},
		ir.Rule{
			ID:        "K7-INFRA-DOMAIN-FREE",
			Statement: infraSentence(cfg.InfraRoots, cfg.FoundationRoot),
			Why:       "Infrastructure is shared by every context. The moment it knows about one, every other context inherits that knowledge and the contexts stop being separate.",
		},
		ir.Rule{
			ID:        "K6-CTX-NO-UI-IMPORT",
			Statement: "A context does not import its own user interface.",
			Why:       "The direction of that dependency is the whole point of the separation: the interface is built on the rules, and rules that reach back into a screen cannot be reused behind a second one.",
		},
		ir.Rule{
			ID:        "K5-UC-FILE",
			Statement: "A use case is declared in a file of its own, named after it.",
			Why:       "It is the unit this document is organised around and the unit a reviewer is asked to read. A file holding three of them cannot be reviewed as any of them.",
		},
		ir.Rule{
			ID:        "K5-UC-AUTHZ",
			Statement: "A use case checks a permission before it does anything.",
			Why:       "Authorisation placed anywhere else is authorisation that a second caller can skip.",
		},
	)

	if m.Layered {
		a.Rules = append(a.Rules,
			ir.Rule{
				ID:        "K6-CTX-PRESENTATION-NO-IMPORT",
				Statement: "Domain code does not import rest/ or cli/.",
				Why:       "A rule that knows how it is being called is a rule that can only be called that way.",
			},
			ir.Rule{
				ID:        "K6-ADAPTER-WIRED-IN-CMD",
				Statement: "Nothing but " + cmdRoot + "/ imports an adapter.",
				Why:       "A port exists so the choice of technology is made once, where the program is assembled. An import anywhere else makes that choice permanent and invisible.",
			},
			ir.Rule{
				ID:        "K6-PRESENTATION-NO-BUNDLE",
				Statement: "A handler takes the use cases it calls, not the whole bundle.",
				Why:       "A handler holding everything has no readable dependencies, so nothing can be said about what it may reach.",
			},
		)
	}

	if m.Framework.Rest != "" {
		a.Rules = append(a.Rules, ir.Rule{
			ID:        "K4-NO-GENERIC-CRUD",
			Statement: "The framework's generic create-read-update-delete factories are not used.",
			Why:       "A screen generated from a type is a screen with no use case behind it, and nothing this document could trace a requirement to.",
		})
	}
	return a
}

// infraRoots is every directory this project reserves for infrastructure,
// each named once.
func infraRoots(roots []string, foundation string) []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range append(append([]string(nil), roots...), foundation) {
		r = strings.TrimSpace(r)
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, path.Clean(r)+"/")
	}
	return out
}

// infraSentence names the directories this project actually reserves.
func infraSentence(roots []string, foundation string) string {
	all := infraRoots(roots, foundation)
	if len(all) == 0 {
		return "Infrastructure holds no business knowledge."
	}
	return strings.Join(all, " and ") + " hold no business knowledge and declare no use case."
}

func or(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
