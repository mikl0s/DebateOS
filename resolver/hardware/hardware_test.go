package hardware_test

import (
	"os"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/hardware"
)

// loadProfile reads a HardwareProfile from a YAML testdata file.
func loadProfile(t *testing.T, path string) hardware.HardwareProfile {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("loadProfile: open %s: %v", path, err)
	}
	defer f.Close()
	var p hardware.HardwareProfile
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&p); err != nil {
		t.Fatalf("loadProfile: decode %s: %v", path, err)
	}
	return p
}

// TestHardwareEval covers EC-037 (NVIDIA skip) and EC-038 (Apple T2 apply).
func TestHardwareEval(t *testing.T) {
	t.Run("EC-037", func(t *testing.T) {
		// NVIDIA GPU leaf predicate evaluated against amd-radeon profile.
		// No NVIDIA in predicates or facts → condition is false → opinion is skipped.
		profile := loadProfile(t, "testdata/profile-amd.yaml")
		expr := resolver.HardwareExpr{
			Type:      resolver.HardwareExprLeaf,
			Predicate: "hw-nvidia-gpu",
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("EC-037: unexpected error: %v", err)
		}
		if got {
			t.Errorf("EC-037: expected false (no NVIDIA GPU in amd-radeon profile), got true")
		}
	})

	t.Run("EC-038", func(t *testing.T) {
		// PCI set-membership predicate: 106b:1801 OR 106b:1802.
		// profile-t2.yaml has pci_ids: [106b:1801] → condition is true → opinion is applied.
		profile := loadProfile(t, "testdata/profile-t2.yaml")
		expr := resolver.HardwareExpr{
			Type:      resolver.HardwareExprOr,
			Operands: []resolver.HardwareExpr{
				{
					Type:      resolver.HardwareExprLeaf,
					Predicate: "pci-id",
					Values:    []string{"106b:1801"},
				},
				{
					Type:      resolver.HardwareExprLeaf,
					Predicate: "pci-id",
					Values:    []string{"106b:1802"},
				},
			},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("EC-038: unexpected error: %v", err)
		}
		if !got {
			t.Errorf("EC-038: expected true (PCI 106b:1801 present in T2 profile), got false")
		}
	})
}

// TestHardwareEvalCompound tests a three-predicate AND with NOT (SR-005 OM-077 shape):
// intel-cpu AND battery-present AND NOT dmi-match "XPS"
func TestHardwareEvalCompound(t *testing.T) {
	// Build: intel-cpu AND battery-present AND (NOT dmi-match "XPS")
	intelAndBattery := resolver.HardwareExpr{
		Type: resolver.HardwareExprAnd,
		Operands: []resolver.HardwareExpr{
			{Type: resolver.HardwareExprLeaf, Predicate: "hw-intel-cpu"},
			{Type: resolver.HardwareExprLeaf, Predicate: "battery-present"},
			{
				Type: resolver.HardwareExprNot,
				Operands: []resolver.HardwareExpr{
					{
						Type:      resolver.HardwareExprLeaf,
						Predicate: "dmi-product-match",
						Match:     "XPS",
					},
				},
			},
		},
	}

	t.Run("matches_intel_battery_non_xps", func(t *testing.T) {
		// Profile: intel cpu + battery present + no XPS in product name → should be true
		profile := hardware.HardwareProfile{
			Predicates: []string{"hw-intel-cpu", "battery-present"},
			Facts:      map[string]string{"dmi_product_name": "ThinkPad X1 Carbon"},
		}
		got, err := hardware.EvalCondition(intelAndBattery, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("expected true for Intel+battery+non-XPS profile, got false")
		}
	})

	t.Run("excludes_xps", func(t *testing.T) {
		// Profile: intel cpu + battery present + XPS in product name → NOT XPS is false → AND is false
		profile := hardware.HardwareProfile{
			Predicates: []string{"hw-intel-cpu", "battery-present"},
			Facts:      map[string]string{"dmi_product_name": "XPS 15 9530"},
		}
		got, err := hardware.EvalCondition(intelAndBattery, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("expected false for XPS profile (NOT XPS fails), got true")
		}
	})

	t.Run("excludes_no_battery", func(t *testing.T) {
		// Profile: intel cpu + no battery → battery-present predicate missing → AND is false
		profile := hardware.HardwareProfile{
			Predicates: []string{"hw-intel-cpu"},
			Facts:      map[string]string{"dmi_product_name": "ThinkPad X1 Carbon"},
		}
		got, err := hardware.EvalCondition(intelAndBattery, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("expected false for no-battery profile, got true")
		}
	})
}

// TestHardwareEvalSetMembership tests a cpu-model set-membership leaf (SR-005 OM-071 shape).
// Returns true only when the declared cpu_model fact is in the set.
func TestHardwareEvalSetMembership(t *testing.T) {
	// cpu-model in [151, 154, 170, 172] (Intel 12th/13th gen Raptor/Alder Lake IDs)
	expr := resolver.HardwareExpr{
		Type:      resolver.HardwareExprLeaf,
		Predicate: "cpu-model-in-set",
		Values:    []string{"151", "154", "170", "172"},
	}

	t.Run("model_in_set", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"cpu_model": "154"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("expected true: cpu_model 154 is in set [151,154,170,172]")
		}
	})

	t.Run("model_not_in_set", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"cpu_model": "999"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("expected false: cpu_model 999 is not in set [151,154,170,172]")
		}
	})
}

// TestHardwareEvalErrors validates that EvalCondition returns an error on
// malformed expressions (unknown Type), satisfying T-01-08.
func TestHardwareEvalErrors(t *testing.T) {
	expr := resolver.HardwareExpr{
		Type: resolver.HardwareExprType("unknown-type"),
	}
	profile := hardware.HardwareProfile{}
	_, err := hardware.EvalCondition(expr, profile)
	if err == nil {
		t.Error("expected error for unknown HardwareExprType, got nil")
	}
}

// ─── Gap-closure tests (01-05 supplemental) ───────────────────────────────

// TestHardwareEvalOR exercises the OR branch in EvalCondition.
func TestHardwareEvalOR(t *testing.T) {
	// OR(hw-intel-cpu, hw-amd-cpu): true when either predicate is present.
	expr := resolver.HardwareExpr{
		Type: resolver.HardwareExprOr,
		Operands: []resolver.HardwareExpr{
			{Type: resolver.HardwareExprLeaf, Predicate: "hw-intel-cpu"},
			{Type: resolver.HardwareExprLeaf, Predicate: "hw-amd-cpu"},
		},
	}

	t.Run("first_operand_matches", func(t *testing.T) {
		profile := hardware.HardwareProfile{Predicates: []string{"hw-intel-cpu"}}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true: hw-intel-cpu matches first OR operand")
		}
	})

	t.Run("second_operand_matches", func(t *testing.T) {
		profile := hardware.HardwareProfile{Predicates: []string{"hw-amd-cpu"}}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true: hw-amd-cpu matches second OR operand")
		}
	})

	t.Run("no_operand_matches", func(t *testing.T) {
		profile := hardware.HardwareProfile{Predicates: []string{"battery-present"}}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Error("expected false: neither hw-intel-cpu nor hw-amd-cpu is in profile")
		}
	})
}

// TestHardwareEvalNOTErrorPath exercises the NOT-with-wrong-operand-count error path.
func TestHardwareEvalNOTErrorPath(t *testing.T) {
	// NOT with 0 operands → error
	expr := resolver.HardwareExpr{
		Type:     resolver.HardwareExprNot,
		Operands: []resolver.HardwareExpr{}, // 0 operands — invalid
	}
	profile := hardware.HardwareProfile{}
	_, err := hardware.EvalCondition(expr, profile)
	if err == nil {
		t.Error("expected error for NOT with 0 operands, got nil")
	}
}

// TestHardwareEvalGenericFactValues exercises the generic fact-set-membership
// evalLeaf path (Values non-empty, predicate is not pci-id or cpu-model-in-set).
func TestHardwareEvalGenericFactValues(t *testing.T) {
	expr := resolver.HardwareExpr{
		Type:      resolver.HardwareExprLeaf,
		Predicate: "gpu",
		Values:    []string{"amd-radeon-rx-7600", "amd-radeon-rx-7700"},
	}

	t.Run("fact_in_values", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"gpu": "amd-radeon-rx-7600"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true: gpu fact is in values set")
		}
	})

	t.Run("fact_not_in_values", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"gpu": "nvidia-rtx-3090"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Error("expected false: nvidia-rtx-3090 is not in amd values set")
		}
	})

	t.Run("fact_key_missing", func(t *testing.T) {
		// If the fact key is absent, should return false (not error).
		profile := hardware.HardwareProfile{
			Facts: map[string]string{},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Error("expected false when fact key is missing from profile")
		}
	})
}

// TestHardwareEvalGenericMatch exercises the generic fact-string-match
// evalLeaf path (Match non-empty, predicate is not dmi-product-match).
func TestHardwareEvalGenericMatch(t *testing.T) {
	expr := resolver.HardwareExpr{
		Type:      resolver.HardwareExprLeaf,
		Predicate: "cpu_model_name",
		Match:     "Intel",
	}

	t.Run("fact_contains_match", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"cpu_model_name": "Intel Core i7-12700H"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("expected true: cpu_model_name contains 'Intel'")
		}
	})

	t.Run("fact_does_not_contain_match", func(t *testing.T) {
		profile := hardware.HardwareProfile{
			Facts: map[string]string{"cpu_model_name": "AMD Ryzen 7 7840U"},
		}
		got, err := hardware.EvalCondition(expr, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Error("expected false: AMD cpu_model_name does not contain 'Intel'")
		}
	})
}

// TestHardwareEvalCPUModelMissing exercises the cpu-model-in-set path when the
// fact key is absent from the profile.
func TestHardwareEvalCPUModelMissing(t *testing.T) {
	expr := resolver.HardwareExpr{
		Type:      resolver.HardwareExprLeaf,
		Predicate: "cpu-model-in-set",
		Values:    []string{"151", "154"},
	}
	profile := hardware.HardwareProfile{
		Facts: map[string]string{}, // no cpu_model key
	}
	got, err := hardware.EvalCondition(expr, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false when cpu_model fact is absent")
	}
}

// TestHardwareEvalDepthLimit verifies WR-03: an expression tree nested beyond
// maxHardwareExprDepth must return an error rather than overflowing the goroutine
// stack. A depth of 100 (well above the 32-frame limit) must produce an error.
func TestHardwareEvalDepthLimit(t *testing.T) {
	// Build a deeply nested NOT(NOT(NOT(...NOT(leaf)...))) chain 100 levels deep.
	// This is far beyond the allowed maxHardwareExprDepth (32).
	leaf := resolver.HardwareExpr{
		Type:      resolver.HardwareExprLeaf,
		Predicate: "hw-intel-cpu",
	}
	expr := leaf
	for i := 0; i < 100; i++ {
		expr = resolver.HardwareExpr{
			Type:     resolver.HardwareExprNot,
			Operands: []resolver.HardwareExpr{expr},
		}
	}

	profile := hardware.HardwareProfile{
		Predicates: []string{"hw-intel-cpu"},
	}

	_, err := hardware.EvalCondition(expr, profile)
	if err == nil {
		t.Fatal("DepthLimit: expected error for expression depth > 32, got nil")
	}
	if !strings.Contains(err.Error(), "depth") {
		t.Errorf("DepthLimit: error %q should mention depth limit", err.Error())
	}
}
