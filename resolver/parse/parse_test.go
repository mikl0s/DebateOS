package parse

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	resolver "github.com/mikkelraglan/debateos/resolver"
)

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// TestParseOpinionSR proves the full SR-001..SR-020 floor is expressible:
// every field of the kitchen-sink fixture round-trips into the typed struct.
func TestParseOpinionSR(t *testing.T) {
	op, err := ParseOpinion(openFixture(t, "opinion-sr-full.yaml"))
	if err != nil {
		t.Fatalf("ParseOpinion: %v", err)
	}
	if op.Schema != 1 {
		t.Errorf("schema = %d, want 1", op.Schema)
	}
	if op.ID != resolver.OpinionID("example/desktop/compositor-settings") {
		t.Errorf("id = %q", op.ID)
	}
	if op.Status != resolver.StatusRequired {
		t.Errorf("status = %q, want required (SR-001)", op.Status)
	}
	if len(op.DependsOn) != 1 || len(op.Conflicts) != 1 {
		t.Errorf("depends_on/conflicts not populated (SR-002/SR-003)")
	}
	if len(op.KnownPatches) != 1 || op.KnownPatches[0].Resolves == "" {
		t.Errorf("known_patches not populated (SR-003)")
	}
	hc := op.HardwareCondition
	if hc == nil || hc.Type != resolver.HardwareExprAnd || len(hc.Operands) != 3 {
		t.Fatalf("compound hardware_condition not populated (SR-004/SR-005): %+v", hc)
	}
	if hc.Operands[1].Type != resolver.HardwareExprNot {
		t.Errorf("NOT combinator lost (SR-005)")
	}
	if got := hc.Operands[2].Values; len(got) != 3 {
		t.Errorf("set-membership values lost (SR-005): %v", got)
	}
	if op.InstallPhase != "config" || op.Ordering == nil || len(op.Ordering.After) != 1 || len(op.Ordering.Before) != 1 {
		t.Errorf("install_phase/ordering not populated (SR-006)")
	}
	if len(op.TranslatorCapabilities) != 2 {
		t.Errorf("translator_capabilities not populated (SR-007)")
	}
	if len(op.FileAssets) != 1 || op.FileAssets[0].Dst == "" {
		t.Errorf("file_assets not populated (SR-008)")
	}
	if len(op.CustomRepos) != 1 || op.CustomRepos[0].SigLevel != resolver.SigLevelNever || op.CustomRepos[0].Priority != 10 {
		t.Errorf("custom_repos not populated (SR-009): %+v", op.CustomRepos)
	}
	if len(op.RuntimeToolInstalls) != 1 || op.RuntimeToolInstalls[0].Manager != "npm" {
		t.Errorf("runtime_tool_installs not populated (SR-010)")
	}
	if op.ExecutionPhase != "install-time" {
		t.Errorf("execution_phase not populated (SR-011)")
	}
	if op.ScriptPayload == nil || len(op.ScriptPayload.Capabilities) != 2 {
		t.Errorf("script_payload not populated (SR-012)")
	}
	if op.DisplayManager == nil || !op.DisplayManager.AutoLogin {
		t.Errorf("display_manager not populated (SR-013)")
	}
	if op.Bootloader == nil || op.Bootloader.Timeout != 3 || !op.Bootloader.Snapshot {
		t.Errorf("bootloader not populated (SR-014)")
	}
	if len(op.Services) != 2 || !op.Services[1].Deferred {
		t.Errorf("services not populated (SR-015)")
	}
	if len(op.SysctlParams) != 1 || op.SysctlParams[0].Key != "fs.inotify.max_user_watches" {
		t.Errorf("sysctl_params not populated (SR-016)")
	}
	if len(op.KernelParams) != 2 {
		t.Errorf("kernel_params not populated (SR-017)")
	}
	if len(op.GroupMemberships) != 1 {
		t.Errorf("group_memberships not populated (SR-018)")
	}
	if len(op.MimeAssociations) != 1 {
		t.Errorf("mime_associations not populated (SR-019)")
	}
	if op.Theme == nil || !op.Theme.IsDefault || op.Theme.Symlinks["current"] == "" {
		t.Errorf("theme not populated (SR-020)")
	}
}

// TestParsePointValid covers SR-021.
func TestParsePointValid(t *testing.T) {
	p, err := ParsePoint(openFixture(t, "point-valid.yaml"))
	if err != nil {
		t.Fatalf("ParsePoint: %v", err)
	}
	if p.Schema != 1 || p.Name == "" || p.Intent == "" {
		t.Errorf("point header not populated (SR-021): %+v", p)
	}
	if len(p.Members) != 3 || p.Members[2].Status != resolver.StatusNiceToHave {
		t.Errorf("point members not populated (SR-021): %+v", p.Members)
	}
}

// TestParseSpeechValid covers SR-022 (foundation target + schema version).
func TestParseSpeechValid(t *testing.T) {
	s, err := ParseSpeech(openFixture(t, "speech-valid.yaml"))
	if err != nil {
		t.Fatalf("ParseSpeech: %v", err)
	}
	if s.Schema != 1 || s.Foundation != "example-foundation" {
		t.Errorf("speech foundation/schema not populated (SR-022): %+v", s)
	}
	if len(s.Points) != 1 || len(s.Opinions) != 1 {
		t.Errorf("speech points/opinions not populated: %+v", s)
	}
	if s.Hardware == nil || s.Hardware.Facts["cpu-model"] != "151" {
		t.Errorf("speech hardware profile not populated: %+v", s.Hardware)
	}
}

// TestParseUnknownFieldRejected: strict decoding (KnownFields) rejects typos.
func TestParseUnknownFieldRejected(t *testing.T) {
	_, err := ParseOpinion(openFixture(t, "opinion-unknown-field.yaml"))
	if err == nil {
		t.Fatal("expected error for unknown field 'statues', got nil")
	}
	if !strings.Contains(err.Error(), "statues") && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "field") {
		t.Errorf("error should mention the unknown field: %v", err)
	}
}

// TestParseSchemaViolationRejected: a document missing required fields fails
// JSON Schema validation even though it YAML-decodes cleanly.
func TestParseSchemaViolationRejected(t *testing.T) {
	doc := "schema: 1\nid: example/no-status\nname: Missing status\ncategory: config-deployment\n"
	_, err := ParseOpinion(strings.NewReader(doc))
	if err == nil {
		t.Fatal("expected schema validation error for missing 'status', got nil")
	}
}

// TestSchemaOSAgnostic enforces invariant 1 / SCHM-02 on the schema files.
func TestSchemaOSAgnostic(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)pacman|mkarchiso|\baur\b|dpkg|\bapt\b|debian|\barch \b`)
	for _, name := range []string{"opinion", "point", "speech"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "schemas", name+".schema.json"))
		if err != nil {
			t.Fatalf("read schema %s: %v", name, err)
		}
		if loc := forbidden.FindString(string(b)); loc != "" {
			t.Errorf("schemas/%s.schema.json contains OS-specific token %q (invariant 1)", name, loc)
		}
	}
}
