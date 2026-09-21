package constants

const (
	ViewHandleSubjectActionConcede         = "action:concede"
	ViewHandleSubjectActionPass            = "action:pass"
	ViewHandleSubjectActionActivatePrefix  = "action:activate:"
	ViewHandleSubjectActionAttackPrefix    = "action:attack:"
	ViewHandleSubjectActionWieldPrefix     = "action:wield:"
	ViewHandleSubjectActionCardistryPrefix = "action:cardistry:"
	ViewHandleSubjectActionAbilityPrefix   = "action:ability:"
	ViewHandleSubjectMaterializePrefix     = "action:materialize:"
	ViewHandleSubjectSkipMaterialize       = "action:skip-materialize"
	ViewHandleSubjectCardPrefix            = "card:"
	ViewHandleSubjectChoicePrefix          = "choice:"
	ViewHandleSubjectTriggerOrderPrefix    = "trigger-order:"
)

type ActionKind string

const (
	ActionConcede         ActionKind = "concede"
	ActionMaterialize     ActionKind = "materialize"
	ActionPass            ActionKind = "pass"
	ActionSkipMaterialize ActionKind = "skip_materialize"
	ActionActivate        ActionKind = "activate"
	ActionAttack          ActionKind = "attack"
	ActionWield           ActionKind = "wield"
)

type Phase string

const (
	PhaseWakeUp       Phase = "wake_up"
	PhaseMaterialize  Phase = "materialize"
	PhaseRecollection Phase = "recollection"
	PhaseDraw         Phase = "draw"
	PhaseMain         Phase = "main"
	PhaseEnd          Phase = "end"
)

const (
	CardDataSchemaVersion = 1
	// CanonicalStateSchemaVersion 標示 StateHash 使用的 canonical state 結構版本。
	CanonicalStateSchemaVersion = 4
	CardDataSourcePattern       = "./card/*.json"
	// ReplayFormatVersion 標示包含 Input.Reserve 的 canonical replay 格式版本。
	ReplayFormatVersion  = 4
	FixedDeckVersion     = "standard-fire-v2"
	FixedCardDataVersion = "card-data-v3"
)

type ReplayFailure string

const (
	ReplayVersionMismatch   ReplayFailure = "version_mismatch"
	ReplayInputRejected     ReplayFailure = "input_rejected"
	ReplayStateHashMismatch ReplayFailure = "state_hash_mismatch"
)

type SupportStatus string

const (
	Supported   SupportStatus = "supported"
	Unsupported SupportStatus = "unsupported"
)

type RulingStatus string

const (
	RulingResolved RulingStatus = "resolved"
	RulingApproved RulingStatus = "approved-project-ruling"
	RulingPending  RulingStatus = "pending"
)

type GateKind string

const (
	GateAbility   GateKind = "ability-slot"
	GateContent   GateKind = "content"
	GateMechanism GateKind = "mechanism"
	GateOperation GateKind = "operation"
	GateRegistry  GateKind = "registry"
	GateRuling    GateKind = "ruling"
)
