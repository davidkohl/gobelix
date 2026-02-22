// dataitems/cat048/v117/common.go
// This file re-exports data items that are identical between v1.17 and v1.32
package v117

import (
	v132 "github.com/davidkohl/gobelix/cat/cat048/dataitems/v132"
)

// Data items that are identical to v1.32
// Note: TrackNumber and TargetAddress (AircraftAddress) are now in common package
type (
	// Fixed-length items (unchanged between versions)
	TimeOfDay                = v132.TimeOfDay
	MeasuredPosition         = v132.MeasuredPosition
	Mode3ACode               = v132.Mode3ACode
	FlightLevel              = v132.FlightLevel
	AircraftIdentification   = v132.AircraftIdentification
	CalculatedPosition       = v132.CalculatedPosition
	CalculatedTrackVelocity  = v132.CalculatedTrackVelocity
	TrackQuality             = v132.TrackQuality
	Mode3ACodeConfidence     = v132.Mode3ACodeConfidence
	ModeCCodeAndConfidence   = v132.ModeCCodeAndConfidence
	Height3D                 = v132.Height3D
	CommunicationsCapability = v132.CommunicationsCapability
	ACASResolutionAdvisory   = v132.ACASResolutionAdvisory
	Mode1Code                = v132.Mode1Code
	Mode2Code                = v132.Mode2Code
	Mode1CodeConfidence      = v132.Mode1CodeConfidence
	Mode2CodeConfidence      = v132.Mode2CodeConfidence

	// Compound/Variable items that are the same structure
	RadarPlotCharacteristics = v132.RadarPlotCharacteristics
	BDSRegisterData          = v132.BDSRegisterData
	TrackStatus              = v132.TrackStatus
	WarningErrorCondition    = v132.WarningErrorCondition

	// Special fields
	SpecialPurpose    = v132.SpecialPurpose
	ReservedExpansion = v132.ReservedExpansion
)
