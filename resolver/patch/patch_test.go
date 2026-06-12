package patch_test

import (
	"os"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/mikkelraglan/debateos/resolver"
	"github.com/mikkelraglan/debateos/resolver/patch"
)

// loadOpinion reads a resolver.Opinion from a YAML testdata file.
func loadOpinion(t *testing.T, path string) resolver.Opinion {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("loadOpinion: open %s: %v", path, err)
	}
	defer f.Close()
	var op resolver.Opinion
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&op); err != nil {
		t.Fatalf("loadOpinion: decode %s: %v", path, err)
	}
	return op
}

// buildEC032Opinions constructs the three opinions for the EC-032 conflict:
// mkinitcpio (conflicting, declares known_patches),
// dracut (conflicting),
// dracut-omarchy-bridge (patch opinion, loaded from testdata).
func buildEC032Opinions(t *testing.T) []resolver.Opinion {
	t.Helper()

	patchOp := loadOpinion(t, "testdata/dracut-bridge.yaml")

	mkinitcpio := resolver.Opinion{
		Schema:   1,
		ID:       "initramfs-method/mkinitcpio",
		Name:     "Omarchy mkinitcpio initramfs",
		Category: "initramfs-method",
		Status:   resolver.StatusRequired,
		Conflicts: []resolver.OpinionRef{
			{ID: "initramfs-method/dracut"},
		},
		KnownPatches: []resolver.PatchRef{
			{
				ID:       "patch/dracut-omarchy-bridge",
				Resolves: "initramfs-method/dracut",
			},
		},
	}

	dracut := resolver.Opinion{
		Schema:   1,
		ID:       "initramfs-method/dracut",
		Name:     "Garuda dracut initramfs",
		Category: "initramfs-method",
		Status:   resolver.StatusRequired,
		Conflicts: []resolver.OpinionRef{
			{ID: "initramfs-method/mkinitcpio"},
		},
	}

	return []resolver.Opinion{mkinitcpio, dracut, patchOp}
}

// TestPatchDiscovery covers EC-032: a patch opinion exists for the
// mkinitcpio↔dracut conflict pair and FindPatch returns a non-nil PatchOffer.
func TestPatchDiscovery(t *testing.T) {
	t.Run("EC-032", func(t *testing.T) {
		opinions := buildEC032Opinions(t)

		offer := patch.FindPatch(
			"initramfs-method/mkinitcpio",
			"initramfs-method/dracut",
			opinions,
		)
		if offer == nil {
			t.Fatal("EC-032: expected non-nil PatchOffer, got nil")
		}
		if offer.PatchID != "patch/dracut-omarchy-bridge" {
			t.Errorf("EC-032: PatchOffer.PatchID = %q, want %q",
				offer.PatchID, "patch/dracut-omarchy-bridge")
		}
		pairSet := map[resolver.OpinionID]bool{
			offer.Pair[0]: true,
			offer.Pair[1]: true,
		}
		if !pairSet["initramfs-method/mkinitcpio"] || !pairSet["initramfs-method/dracut"] {
			t.Errorf("EC-032: PatchOffer.Pair = %v, want both mkinitcpio and dracut IDs",
				offer.Pair)
		}
	})
}

// TestPatchDiscoveryNoPatch verifies that a conflict pair with no attached
// patch returns nil.
func TestPatchDiscoveryNoPatch(t *testing.T) {
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "package-install/foo",
			Name:     "Foo",
			Category: "package-install",
			Status:   resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{
				{ID: "package-install/bar"},
			},
			// No KnownPatches declared
		},
		{
			Schema:   1,
			ID:       "package-install/bar",
			Name:     "Bar",
			Category: "package-install",
			Status:   resolver.StatusRequired,
		},
	}

	offer := patch.FindPatch("package-install/foo", "package-install/bar", opinions)
	if offer != nil {
		t.Errorf("expected nil PatchOffer for pair with no patch, got %+v", offer)
	}
}

// TestPatchDiscoverySymmetric verifies that FindPatch(a,b,...) and
// FindPatch(b,a,...) return the same offer (pair order-independent).
func TestPatchDiscoverySymmetric(t *testing.T) {
	opinions := buildEC032Opinions(t)

	offerAB := patch.FindPatch(
		"initramfs-method/mkinitcpio",
		"initramfs-method/dracut",
		opinions,
	)
	offerBA := patch.FindPatch(
		"initramfs-method/dracut",
		"initramfs-method/mkinitcpio",
		opinions,
	)

	if offerAB == nil || offerBA == nil {
		t.Fatalf("expected non-nil offers in both directions, got ab=%v ba=%v", offerAB, offerBA)
	}
	if offerAB.PatchID != offerBA.PatchID {
		t.Errorf("symmetric: PatchID mismatch: ab=%q ba=%q", offerAB.PatchID, offerBA.PatchID)
	}
	if offerAB.Pair != offerBA.Pair {
		t.Errorf("symmetric: Pair mismatch: ab=%v ba=%v", offerAB.Pair, offerBA.Pair)
	}
}
