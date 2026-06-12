// Package resolver defines the shared domain types for DebateOS compositions:
// Opinions (atomic, OS-agnostic configuration decisions), Points (curated
// bundles of opinions), and Speeches (a user's complete composition).
//
// These types are the module-wide contract consumed by resolver/parse,
// resolver/graph, resolver/resolve, resolver/patch, and resolver/hardware.
// All numeric data uses ints or strings — never floating-point — so canonical
// JSON output is byte-identical between native and WASM builds.
package resolver

// OpinionID uniquely identifies an opinion within the registry namespace.
type OpinionID string

// Status declares how an opinion participates in conflict resolution (SR-001).
const (
	StatusRequired   = "required"
	StatusNiceToHave = "nice-to-have"
)

// SigLevel is the trust level an opinion declares for a custom package
// repository (SR-009). The four levels mirror the repo trust spectrum found
// in the Phase 0 evidence: full signature enforcement down to none.
type SigLevel string

const (
	SigLevelRequired                 SigLevel = "Required"
	SigLevelRequiredDatabaseOptional SigLevel = "RequiredDatabaseOptional"
	SigLevelOptionalTrustAll         SigLevel = "OptionalTrustAll"
	SigLevelNever                    SigLevel = "Never"
)

// HardwareExprType discriminates HardwareExpr tree nodes (SR-004/SR-005).
type HardwareExprType string

const (
	HardwareExprLeaf HardwareExprType = "leaf"
	HardwareExprAnd  HardwareExprType = "and"
	HardwareExprOr   HardwareExprType = "or"
	HardwareExprNot  HardwareExprType = "not"
)

// HardwareExpr is a discriminated union expressing hardware conditions.
// A leaf names a single predicate (e.g. "hw-nvidia-gpu") and may carry
// set-membership Values (e.g. cpu model numbers) or a string Match pattern.
// Combinator nodes (and/or/not) compose Operands recursively (SR-005).
type HardwareExpr struct {
	Type      HardwareExprType `json:"type" yaml:"type"`
	Predicate string           `json:"predicate,omitempty" yaml:"predicate,omitempty"`
	Values    []string         `json:"values,omitempty" yaml:"values,omitempty"`
	Match     string           `json:"match,omitempty" yaml:"match,omitempty"`
	Operands  []HardwareExpr   `json:"operands,omitempty" yaml:"operands,omitempty"`
}

// OpinionRef references another opinion, optionally constrained to a version.
type OpinionRef struct {
	ID      OpinionID `json:"id" yaml:"id"`
	Version string    `json:"version,omitempty" yaml:"version,omitempty"`
}

// PatchRef references a patch opinion that resolves a known conflict pair (SR-003).
type PatchRef struct {
	ID        OpinionID `json:"id" yaml:"id"`
	Resolves  OpinionID `json:"resolves,omitempty" yaml:"resolves,omitempty"`
	Rationale string    `json:"rationale,omitempty" yaml:"rationale,omitempty"`
}

// Ordering declares relative install ordering constraints (SR-006).
type Ordering struct {
	Before []OpinionRef `json:"before,omitempty" yaml:"before,omitempty"`
	After  []OpinionRef `json:"after,omitempty" yaml:"after,omitempty"`
}

// FileAsset is a file payload carried by an opinion (SR-008): theming assets,
// dotfiles, wallpapers. Src is registry-relative; Dst is target-relative.
type FileAsset struct {
	Src string `json:"src" yaml:"src"`
	Dst string `json:"dst" yaml:"dst"`
}

// RepoDecl registers a custom package repository (SR-009) with an explicit
// trust level and priority relative to the foundation's default repos.
type RepoDecl struct {
	Name     string   `json:"name" yaml:"name"`
	URL      string   `json:"url" yaml:"url"`
	SigLevel SigLevel `json:"sig_level" yaml:"sig_level"`
	Priority int      `json:"priority,omitempty" yaml:"priority,omitempty"`
	Keyring  string   `json:"keyring,omitempty" yaml:"keyring,omitempty"`
}

// RuntimeToolInstall installs packages through a language/runtime package
// manager rather than the OS package manager (SR-010).
type RuntimeToolInstall struct {
	Manager  string   `json:"manager" yaml:"manager"`
	Packages []string `json:"packages" yaml:"packages"`
}

// ScriptPayload is an arbitrary script carried as data (SR-012). Scripts are
// never executed by the resolver; Capabilities declares what the script needs
// from a translator so unsupported scripts break visibly at composition time.
type ScriptPayload struct {
	Script       string   `json:"script" yaml:"script"`
	Capabilities []string `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

// DisplayManagerConfig declares display manager intent (SR-013).
type DisplayManagerConfig struct {
	Name      string `json:"name" yaml:"name"`
	Greeter   string `json:"greeter,omitempty" yaml:"greeter,omitempty"`
	AutoLogin bool   `json:"auto_login,omitempty" yaml:"auto_login,omitempty"`
}

// BootloaderConfig declares bootloader intent (SR-014).
type BootloaderConfig struct {
	Name     string `json:"name" yaml:"name"`
	Timeout  int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Snapshot bool   `json:"snapshot,omitempty" yaml:"snapshot,omitempty"`
}

// ServiceDecl enables or disables a system service (SR-015). Deferred marks
// services that must be enabled at first boot rather than inside the
// installer environment.
type ServiceDecl struct {
	Name     string `json:"name" yaml:"name"`
	Enable   bool   `json:"enable" yaml:"enable"`
	Deferred bool   `json:"deferred,omitempty" yaml:"deferred,omitempty"`
}

// SysctlParam sets one kernel sysctl key via a drop-in file (SR-016).
type SysctlParam struct {
	Key        string `json:"key" yaml:"key"`
	Value      string `json:"value" yaml:"value"`
	DropInFile string `json:"drop_in_file,omitempty" yaml:"drop_in_file,omitempty"`
}

// KernelParam sets one kernel boot command-line parameter (SR-017).
type KernelParam struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// GroupMembership adds the installing user to a system group (SR-018).
type GroupMembership struct {
	Group string `json:"group" yaml:"group"`
}

// MimeAssoc associates a MIME pattern with an application (SR-019).
type MimeAssoc struct {
	MimePattern string `json:"mime_pattern" yaml:"mime_pattern"`
	AppID       string `json:"app_id" yaml:"app_id"`
}

// ThemeDecl carries a theme bundle (SR-020): a directory of assets plus the
// symlinks that activate it. IsDefault marks the initially active theme.
type ThemeDecl struct {
	BundleDir string            `json:"bundle_dir" yaml:"bundle_dir"`
	Symlinks  map[string]string `json:"symlinks,omitempty" yaml:"symlinks,omitempty"`
	IsDefault bool              `json:"is_default,omitempty" yaml:"is_default,omitempty"`
}

// Opinion is the atomic unit of DebateOS: one OS-agnostic configuration
// decision with the metadata the resolver needs (SR-001..SR-020). Intent is
// always expressed free of distro mechanics (invariant 1).
type Opinion struct {
	Schema   int       `json:"schema" yaml:"schema"`
	ID       OpinionID `json:"id" yaml:"id"`
	Name     string    `json:"name" yaml:"name"`
	Category string    `json:"category" yaml:"category"`
	Intent   string    `json:"intent,omitempty" yaml:"intent,omitempty"`
	Status   string    `json:"status" yaml:"status"`

	DependsOn    []OpinionRef `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Conflicts    []OpinionRef `json:"conflicts,omitempty" yaml:"conflicts,omitempty"`
	KnownPatches []PatchRef   `json:"known_patches,omitempty" yaml:"known_patches,omitempty"`

	HardwareCondition *HardwareExpr `json:"hardware_condition,omitempty" yaml:"hardware_condition,omitempty"`

	InstallPhase string    `json:"install_phase,omitempty" yaml:"install_phase,omitempty"`
	Ordering     *Ordering `json:"ordering,omitempty" yaml:"ordering,omitempty"`

	TranslatorCapabilities []string `json:"translator_capabilities,omitempty" yaml:"translator_capabilities,omitempty"`

	Packages            []string              `json:"packages,omitempty" yaml:"packages,omitempty"`
	RemovePackages      []string              `json:"remove_packages,omitempty" yaml:"remove_packages,omitempty"`
	FileAssets          []FileAsset           `json:"file_assets,omitempty" yaml:"file_assets,omitempty"`
	CustomRepos         []RepoDecl            `json:"custom_repos,omitempty" yaml:"custom_repos,omitempty"`
	RuntimeToolInstalls []RuntimeToolInstall  `json:"runtime_tool_installs,omitempty" yaml:"runtime_tool_installs,omitempty"`
	ExecutionPhase      string                `json:"execution_phase,omitempty" yaml:"execution_phase,omitempty"`
	ScriptPayload       *ScriptPayload        `json:"script_payload,omitempty" yaml:"script_payload,omitempty"`
	DisplayManager      *DisplayManagerConfig `json:"display_manager,omitempty" yaml:"display_manager,omitempty"`
	Bootloader          *BootloaderConfig     `json:"bootloader,omitempty" yaml:"bootloader,omitempty"`
	Services            []ServiceDecl         `json:"services,omitempty" yaml:"services,omitempty"`
	SysctlParams        []SysctlParam         `json:"sysctl_params,omitempty" yaml:"sysctl_params,omitempty"`
	KernelParams        []KernelParam         `json:"kernel_params,omitempty" yaml:"kernel_params,omitempty"`
	GroupMemberships    []GroupMembership     `json:"group_memberships,omitempty" yaml:"group_memberships,omitempty"`
	MimeAssociations    []MimeAssoc           `json:"mime_associations,omitempty" yaml:"mime_associations,omitempty"`
	Theme               *ThemeDecl            `json:"theme,omitempty" yaml:"theme,omitempty"`
}

// PointMember marks one opinion's role inside a point: every member is either
// required or nice-to-have within that point (SR-021, docs/04 rule 1).
type PointMember struct {
	ID     OpinionID `json:"id" yaml:"id"`
	Status string    `json:"status" yaml:"status"`
}

// Point is a curated, coherent bundle of opinions (SR-021). Points are
// foundation-agnostic and versioned/forkable as registry artifacts.
type Point struct {
	Schema  int           `json:"schema" yaml:"schema"`
	ID      string        `json:"id" yaml:"id"`
	Name    string        `json:"name" yaml:"name"`
	Intent  string        `json:"intent,omitempty" yaml:"intent,omitempty"`
	Curator string        `json:"curator,omitempty" yaml:"curator,omitempty"`
	Members []PointMember `json:"members" yaml:"members"`
}

// PointRef references a point included in a speech.
type PointRef struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// HardwareProfile declares the target machine's hardware facts used to
// evaluate hardware-conditional opinions at composition time.
type HardwareProfile struct {
	Predicates []string          `json:"predicates,omitempty" yaml:"predicates,omitempty"`
	Facts      map[string]string `json:"facts,omitempty" yaml:"facts,omitempty"`
}

// Speech is a user's complete composition (SR-022): a foundation target plus
// the points (and individually selected opinions) that compose the system.
type Speech struct {
	Schema     int              `json:"schema" yaml:"schema"`
	ID         string           `json:"id,omitempty" yaml:"id,omitempty"`
	Name       string           `json:"name,omitempty" yaml:"name,omitempty"`
	Foundation string           `json:"foundation" yaml:"foundation"`
	Points     []PointRef       `json:"points" yaml:"points"`
	Opinions   []OpinionRef     `json:"opinions,omitempty" yaml:"opinions,omitempty"`
	Hardware   *HardwareProfile `json:"hardware,omitempty" yaml:"hardware,omitempty"`
}
