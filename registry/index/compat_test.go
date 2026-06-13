package index_test

import (
	"testing"

	resolver "github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/registry/index"
)

func TestComputeCompat(t *testing.T) {
	// caps simulates translators/arch/capabilities.json + translators/debian/capabilities.json
	caps := map[string][]string{
		"arch": {
			"install-packages",
			"deploy-config-file-tree",
			"enable-systemd-service",
			"write-sysctl-drop-in",
			"deploy-sddm-theme", // arch-only
		},
		"debian": {
			"install-packages",
			"deploy-config-file-tree",
			"enable-systemd-service",
			"write-sysctl-drop-in",
			// deploy-sddm-theme is NOT in debian
		},
	}

	t.Run("both_foundations_compatible", func(t *testing.T) {
		// A point whose members only require tokens in BOTH arch and debian.
		point := resolver.Point{
			Schema:  1,
			ID:      "test/both-compat",
			Name:    "Both Compat",
			Curator: "test@example.com",
			Members: []resolver.PointMember{
				{ID: "OP-001", Status: "required"},
				{ID: "OP-002", Status: "required"},
			},
		}
		opinions := map[resolver.OpinionID]resolver.Opinion{
			"OP-001": {
				Schema: 1, ID: "OP-001", Name: "Op1", Category: "pkg", Status: "required",
				TranslatorCapabilities: []string{"install-packages", "deploy-config-file-tree"},
			},
			"OP-002": {
				Schema: 1, ID: "OP-002", Name: "Op2", Category: "svc", Status: "required",
				TranslatorCapabilities: []string{"enable-systemd-service", "write-sysctl-drop-in"},
			},
		}

		result := index.ComputeCompat(point, opinions, caps)

		if len(result) != 2 {
			t.Fatalf("expected 2 FoundationCompat entries, got %d", len(result))
		}
		// Result must be sorted by Foundation: arch < debian
		if result[0].Foundation != "arch" {
			t.Errorf("expected result[0].Foundation == arch, got %s", result[0].Foundation)
		}
		if result[1].Foundation != "debian" {
			t.Errorf("expected result[1].Foundation == debian, got %s", result[1].Foundation)
		}
		if !result[0].Compatible {
			t.Errorf("arch should be compatible, got Compatible=false")
		}
		if !result[1].Compatible {
			t.Errorf("debian should be compatible, got Compatible=false")
		}
		if len(result[0].Missing) != 0 {
			t.Errorf("arch should have no missing caps, got %v", result[0].Missing)
		}
		if len(result[1].Missing) != 0 {
			t.Errorf("debian should have no missing caps, got %v", result[1].Missing)
		}
	})

	t.Run("arch_only_token_debian_incompatible", func(t *testing.T) {
		// A point that requires deploy-sddm-theme — only in arch, not debian.
		point := resolver.Point{
			Schema:  1,
			ID:      "test/arch-only",
			Name:    "Arch Only",
			Curator: "test@example.com",
			Members: []resolver.PointMember{
				{ID: "OP-SDDM", Status: "required"},
			},
		}
		opinions := map[resolver.OpinionID]resolver.Opinion{
			"OP-SDDM": {
				Schema: 1, ID: "OP-SDDM", Name: "SDDM", Category: "theming", Status: "required",
				TranslatorCapabilities: []string{"deploy-sddm-theme"},
			},
		}

		result := index.ComputeCompat(point, opinions, caps)

		if len(result) != 2 {
			t.Fatalf("expected 2 FoundationCompat entries, got %d", len(result))
		}

		var archFC, debianFC index.FoundationCompat
		for _, fc := range result {
			if fc.Foundation == "arch" {
				archFC = fc
			} else if fc.Foundation == "debian" {
				debianFC = fc
			}
		}

		if !archFC.Compatible {
			t.Errorf("arch should be compatible for sddm-theme point")
		}
		if len(archFC.Missing) != 0 {
			t.Errorf("arch should have no missing caps, got %v", archFC.Missing)
		}
		if debianFC.Compatible {
			t.Errorf("debian should NOT be compatible for sddm-theme point")
		}
		found := false
		for _, m := range debianFC.Missing {
			if m == "deploy-sddm-theme" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("debian Missing should contain 'deploy-sddm-theme', got %v", debianFC.Missing)
		}
	})

	t.Run("determinism_sorted_by_foundation", func(t *testing.T) {
		// Result must always be sorted by Foundation lexically.
		point := resolver.Point{
			Schema:  1,
			ID:      "test/det",
			Name:    "Det",
			Curator: "test@example.com",
			Members: []resolver.PointMember{
				{ID: "OP-A", Status: "required"},
			},
		}
		opinions := map[resolver.OpinionID]resolver.Opinion{
			"OP-A": {
				Schema: 1, ID: "OP-A", Name: "OpA", Category: "pkg", Status: "required",
				TranslatorCapabilities: []string{"install-packages"},
			},
		}

		r1 := index.ComputeCompat(point, opinions, caps)
		r2 := index.ComputeCompat(point, opinions, caps)

		if len(r1) != len(r2) {
			t.Fatalf("determinism: len mismatch %d vs %d", len(r1), len(r2))
		}
		for i := range r1 {
			if r1[i].Foundation != r2[i].Foundation {
				t.Errorf("determinism: result[%d].Foundation changed: %s vs %s", i, r1[i].Foundation, r2[i].Foundation)
			}
			if r1[i].Compatible != r2[i].Compatible {
				t.Errorf("determinism: result[%d].Compatible changed", i)
			}
		}
		// Also verify ordering: arch before debian
		if len(r1) >= 2 && r1[0].Foundation > r1[1].Foundation {
			t.Errorf("result not sorted by Foundation: %s > %s", r1[0].Foundation, r1[1].Foundation)
		}
	})

	t.Run("missing_sorted_lexically", func(t *testing.T) {
		// When multiple tokens are missing, Missing slice is sorted lexically.
		caps2 := map[string][]string{
			"arch": {
				"install-packages",
			},
			"minimal": {
				// has nothing
			},
		}
		point := resolver.Point{
			Schema:  1,
			ID:      "test/multi-miss",
			Name:    "Multi Miss",
			Curator: "test@example.com",
			Members: []resolver.PointMember{
				{ID: "OP-M", Status: "required"},
			},
		}
		opinions := map[resolver.OpinionID]resolver.Opinion{
			"OP-M": {
				Schema: 1, ID: "OP-M", Name: "OpM", Category: "pkg", Status: "required",
				TranslatorCapabilities: []string{"zoo-token", "alpha-token", "middle-token"},
			},
		}

		result := index.ComputeCompat(point, opinions, caps2)
		var minimalFC index.FoundationCompat
		for _, fc := range result {
			if fc.Foundation == "minimal" {
				minimalFC = fc
				break
			}
		}
		if minimalFC.Compatible {
			t.Error("minimal should not be compatible")
		}
		// Missing should be sorted: alpha-token, middle-token, zoo-token
		expected := []string{"alpha-token", "middle-token", "zoo-token"}
		if len(minimalFC.Missing) != len(expected) {
			t.Fatalf("missing count: expected %d got %d", len(expected), len(minimalFC.Missing))
		}
		for i, e := range expected {
			if minimalFC.Missing[i] != e {
				t.Errorf("missing[%d]: expected %s got %s", i, e, minimalFC.Missing[i])
			}
		}
	})
}
