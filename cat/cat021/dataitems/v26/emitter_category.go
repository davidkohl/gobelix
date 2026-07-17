// dataitems/cat021/emitter_category.go
package v26

import (
	"bytes"
	"fmt"
)

// EmitterCategoryType represents the type of emitter
type EmitterCategoryType uint8

// Emitter Category values per ASTERIX CAT021 I021/020.
// NOTE: this enumeration is NOT the DO-260/1090ES "emitter category set A"
// coding (where e.g. rotorcraft = 7) - ASTERIX renumbers the categories.
const (
	EmitterLight           EmitterCategoryType = 1 // light aircraft <= 15500 lbs (~7000 kg)
	EmitterSmall           EmitterCategoryType = 2 // 15500 lbs < small < 75000 lbs
	EmitterMedium          EmitterCategoryType = 3 // 75000 lbs < medium < 300000 lbs
	EmitterHighVortexLarge EmitterCategoryType = 4 // high vortex large
	EmitterHeavy           EmitterCategoryType = 5 // heavy aircraft >= 300000 lbs
	EmitterHighPerformance EmitterCategoryType = 6 // >5g capability and >400 kt cruise
	// 7 to 9 reserved
	EmitterRotorcraft     EmitterCategoryType = 10 // rotorcraft
	EmitterGlider         EmitterCategoryType = 11 // glider / sailplane
	EmitterLighterThanAir EmitterCategoryType = 12 // lighter-than-air
	EmitterUAV            EmitterCategoryType = 13 // unmanned aerial vehicle
	EmitterSpace          EmitterCategoryType = 14 // space / transatmospheric vehicle
	EmitterUltraLight     EmitterCategoryType = 15 // ultralight / hang glider / paraglider
	EmitterParachutist    EmitterCategoryType = 16 // parachutist / skydiver
	// 17 to 19 reserved
	EmitterSurfaceEmergency EmitterCategoryType = 20 // surface emergency vehicle
	EmitterSurfaceService   EmitterCategoryType = 21 // surface service vehicle
	EmitterPointObstacle    EmitterCategoryType = 22 // fixed ground or tethered obstruction
	EmitterClusterObstacle  EmitterCategoryType = 23 // cluster obstacle
	EmitterLineObstacle     EmitterCategoryType = 24 // line obstacle
)

// EmitterCategory implements I021/020
// This data item defines the type of emitter from which the information in the track is derived
type EmitterCategory struct {
	ECAT EmitterCategoryType // Emitter category
}

// Decode reads the EmitterCategory data from the buffer
func (e *EmitterCategory) Decode(buf *bytes.Buffer) (int, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("reading emitter category: %w", err)
	}

	e.ECAT = EmitterCategoryType(b)
	return 1, e.Validate()
}

// Encode writes the EmitterCategory data to the buffer
func (e *EmitterCategory) Encode(buf *bytes.Buffer) (int, error) {
	if err := e.Validate(); err != nil {
		return 0, err
	}

	err := buf.WriteByte(byte(e.ECAT))
	if err != nil {
		return 0, fmt.Errorf("writing emitter category: %w", err)
	}
	return 1, nil
}

// Validate checks if the EmitterCategory contains valid data
func (e *EmitterCategory) Validate() error {
	// Value 0 is valid (unknown/not set), values 1-24 are defined
	if e.ECAT > 24 {
		return fmt.Errorf("invalid emitter category value: %d", e.ECAT)
	}
	return nil
}

// String returns a human-readable representation of the EmitterCategory
func (e *EmitterCategory) String() string {
	switch e.ECAT {
	case 0:
		return "Unknown"
	case EmitterLight:
		return "Light Aircraft"
	case EmitterSmall:
		return "Small Aircraft"
	case EmitterMedium:
		return "Medium Aircraft"
	case EmitterHighVortexLarge:
		return "High Vortex Large Aircraft"
	case EmitterHeavy:
		return "Heavy Aircraft"
	case EmitterHighPerformance:
		return "High Performance Aircraft"
	case EmitterRotorcraft:
		return "Rotorcraft"
	case EmitterGlider:
		return "Glider/Sailplane"
	case EmitterLighterThanAir:
		return "Lighter Than Air"
	case EmitterParachutist:
		return "Parachutist/Skydiver"
	case EmitterUltraLight:
		return "Ultra Light/Hang Glider/Paraglider"
	case EmitterUAV:
		return "UAV"
	case EmitterSpace:
		return "Space/Transatmospheric Vehicle"
	case EmitterSurfaceEmergency:
		return "Surface Emergency Vehicle"
	case EmitterSurfaceService:
		return "Surface Service Vehicle"
	case EmitterPointObstacle:
		return "Point Obstacle"
	case EmitterClusterObstacle:
		return "Cluster Obstacle"
	case EmitterLineObstacle:
		return "Line Obstacle"
	default:
		return fmt.Sprintf("Unassigned(%d)", e.ECAT)
	}
}
