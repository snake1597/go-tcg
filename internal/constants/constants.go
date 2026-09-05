package constants

type PlayerID string

const (
	PlayerOne PlayerID = "player-1"
	PlayerTwo PlayerID = "player-2"
)

type InputKind string

const (
	InputConcede InputKind = "concede"
)

const (
	CardDataSchemaVersion = 1
	CardDataSourcePattern = "./card/*.json"
	ReplayFormatVersion   = 1
	FixedDeckVersion      = "standard-fire-v2"
	FixedCardDataVersion  = "card-data-v3"
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
