// Package hardware provides compound hardware predicate evaluation against a
// declared hardware profile (SR-004/SR-005). The resolver calls EvalCondition
// to decide whether a hardware-conditional opinion is applied or skipped.
//
// Security note (T-01-07): expression depth is bounded by the validated
// document depth (JSON Schema validates the input before it reaches here);
// the recursive evaluator processes only schema-validated HardwareExpr trees.
// T-01-08: EvalCondition returns an error on unknown HardwareExprType rather
// than silently defaulting to true or false.
package hardware

import (
	"fmt"
	"strings"

	"github.com/mikkelraglan/debateos/resolver"
)

// HardwareProfile is the evaluated representation of a machine's declared
// hardware facts used to gate hardware-conditional opinions at composition
// time (SR-004/SR-005). It extends the speech-level resolver.HardwareProfile
// with structured fields for efficient predicate matching.
//
// Fields cover the hardware categories referenced by the EC-037/EC-038 corpus
// and the OM-NNN hardware helpers: GPU name, CPU vendor/model, named boolean
// predicates (hw-nvidia-gpu, battery-present, hw-intel-cpu, …), PCI device
// IDs, and DMI product name for string-match conditions.
type HardwareProfile struct {
	// Predicates is a set of boolean named predicates active on this machine
	// (e.g. "hw-intel-cpu", "battery-present", "apple-hardware").
	// A leaf predicate with no Values/Match is satisfied iff its name appears
	// in this list.
	Predicates []string `json:"predicates,omitempty" yaml:"predicates,omitempty"`

	// Facts holds arbitrary string key→value hardware facts (e.g.
	// "gpu" → "amd-radeon-rx-7600", "cpu_model" → "154",
	// "dmi_product_name" → "ThinkPad X1 Carbon").
	Facts map[string]string `json:"facts,omitempty" yaml:"facts,omitempty"`

	// PCIIDs lists raw PCI device IDs present on the machine in
	// "vendor:device" hex format (e.g. "106b:1801" for Apple T2 chip).
	// Used for pci-id set-membership predicates (EC-038).
	PCIIDs []string `json:"pci_ids,omitempty" yaml:"pci_ids,omitempty"`
}

// EvalCondition recursively evaluates a HardwareExpr discriminated-union
// predicate against the declared HardwareProfile. It returns (true, nil) when
// the condition is satisfied, (false, nil) when unsatisfied, and (false, err)
// when the expression is malformed (unknown Type — T-01-08).
//
// Supported node types:
//   - HardwareExprLeaf: evaluates a single named predicate.
//     If Values is non-empty → set-membership check (cpu-model-in-set, pci-id).
//     If Match is non-empty → substring match against a fact value
//     (dmi-product-match checks dmi_product_name).
//     Otherwise → boolean predicate check (predicate name in profile.Predicates).
//   - HardwareExprAnd: all operands must evaluate to true.
//   - HardwareExprOr: at least one operand must evaluate to true.
//   - HardwareExprNot: exactly one operand; result is negated.
func EvalCondition(expr resolver.HardwareExpr, hw HardwareProfile) (bool, error) {
	switch expr.Type {
	case resolver.HardwareExprLeaf:
		return evalLeaf(expr, hw)

	case resolver.HardwareExprAnd:
		for _, operand := range expr.Operands {
			ok, err := EvalCondition(operand, hw)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil // short-circuit
			}
		}
		return true, nil

	case resolver.HardwareExprOr:
		for _, operand := range expr.Operands {
			ok, err := EvalCondition(operand, hw)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil // short-circuit
			}
		}
		return false, nil

	case resolver.HardwareExprNot:
		if len(expr.Operands) != 1 {
			return false, fmt.Errorf("hardware: NOT expression requires exactly 1 operand, got %d", len(expr.Operands))
		}
		ok, err := EvalCondition(expr.Operands[0], hw)
		if err != nil {
			return false, err
		}
		return !ok, nil

	default:
		return false, fmt.Errorf("hardware: unknown HardwareExprType %q (T-01-08)", expr.Type)
	}
}

// evalLeaf evaluates a single leaf predicate against the hardware profile.
// The evaluation strategy depends on which optional fields are set on the leaf:
//
//  1. Values non-empty + Predicate == "pci-id": PCI device ID membership check
//     against profile.PCIIDs.
//  2. Values non-empty + Predicate == "cpu-model-in-set": cpu_model fact
//     membership check against the Values set.
//  3. Values non-empty (other predicates): generic fact membership — the fact
//     whose key matches the predicate must have a value in Values.
//  4. Match non-empty + Predicate == "dmi-product-match": substring match of
//     profile.Facts["dmi_product_name"] against the Match string.
//  5. Match non-empty (other predicates): substring match of the fact value
//     for the key matching the predicate.
//  6. Neither Values nor Match: boolean predicate — true iff the predicate
//     name appears in profile.Predicates.
func evalLeaf(expr resolver.HardwareExpr, hw HardwareProfile) (bool, error) {
	pred := expr.Predicate

	// --- set-membership evaluation ---
	if len(expr.Values) > 0 {
		switch pred {
		case "pci-id":
			// Any PCI ID in expr.Values must appear in profile.PCIIDs.
			pciSet := makeStringSet(hw.PCIIDs)
			for _, v := range expr.Values {
				if pciSet[v] {
					return true, nil
				}
			}
			return false, nil

		case "cpu-model-in-set":
			// profile.Facts["cpu_model"] must be in expr.Values.
			model, ok := hw.Facts["cpu_model"]
			if !ok {
				return false, nil
			}
			valSet := makeStringSet(expr.Values)
			return valSet[model], nil

		default:
			// Generic: the fact value for key=pred must be in Values.
			factVal, ok := hw.Facts[pred]
			if !ok {
				return false, nil
			}
			valSet := makeStringSet(expr.Values)
			return valSet[factVal], nil
		}
	}

	// --- string-match evaluation ---
	if expr.Match != "" {
		switch pred {
		case "dmi-product-match":
			productName := hw.Facts["dmi_product_name"]
			return strings.Contains(productName, expr.Match), nil
		default:
			factVal := hw.Facts[pred]
			return strings.Contains(factVal, expr.Match), nil
		}
	}

	// --- boolean predicate evaluation ---
	// True iff the predicate name appears in profile.Predicates.
	predSet := makeStringSet(hw.Predicates)
	return predSet[pred], nil
}

// makeStringSet builds a set (map[string]bool) from a slice of strings for
// O(1) membership testing. Deterministic: map is read-only after construction.
func makeStringSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}
