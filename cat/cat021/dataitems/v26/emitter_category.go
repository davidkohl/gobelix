// dataitems/cat021/emitter_category.go
package v26

import (
	"bytes"
	"fmt"
)

// EmitterCategoryType represents the type of emitter
type EmitterCategoryType uint8

// Emitter Category values
const (
	EmitterLight            EmitterCategoryType = 1  // Light aircraft (< 7000 kg)
	EmitterSmall            EmitterCategoryType = 2  // Small aircraft (7000 to 34000 kg)
	EmitterMedium           EmitterCategoryType = 3  // Medium aircraft (34000 to 136000 kg)
	EmitterHigh             EmitterCategoryType = 4  // High Vortex aircraft (> 136000 kg)
	EmitterHeavy            EmitterCategoryType = 5  // Heavy aircraft
	EmitterHighPerformance  EmitterCategoryType = 6  // High performance aircraft
	EmitterRotorcraft       EmitterCategoryType = 7  // Rotorcraft
	EmitterUnassigned8      EmitterCategoryType = 8  // Unassigned
	EmitterGlider           EmitterCategoryType = 9  // Glider / sailplane
	EmitterLighterThanAir   EmitterCategoryType = 10 // Lighter than air
	EmitterParachutist      EmitterCategoryType = 11 // Parachutist / skydiver
	EmitterUltraLight       EmitterCategoryType = 12 // Ultra light / hang glider / paraglider
	EmitterUnassigned13     EmitterCategoryType = 13 // Unassigned
	EmitterUAV              EmitterCategoryType = 14 // Unmanned Aerial Vehicle
	EmitterSpace            EmitterCategoryType = 15 // Space / transatmospheric vehicle
	EmitterUnassigned16     EmitterCategoryType = 16 // Unassigned
	EmitterSurfaceEmergency EmitterCategoryType = 17 // Surface emergency vehicle
	EmitterSurfaceService   EmitterCategoryType = 18 // Surface service vehicle
	EmitterPointObstacle    EmitterCategoryType = 19 // Fixed ground or tethered obstruction
	EmitterClusterObstacle  EmitterCategoryType = 20 // Cluster obstacle
	EmitterLineObstacle     EmitterCategoryType = 21 // Line obstacle
	EmitterUnassigned22     EmitterCategoryType = 22 // Unassigned
	EmitterUnassigned23     EmitterCategoryType = 23 // Unassigned
	EmitterUnassigned24     EmitterCategoryType = 24 // Unassigned
)

// EmitterCategory implements I021/020
// This data item defines the type of emitter from which the information in the track is derived
type EmitterCategory struct {
	ECAT EmitterCategoryType // Emitter category
}

// Decode reads the EmitterCategory data from the buffer
func (e *EmitterCategory) Decode(buf *bytes.Buffer) (int, error) {
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading emitter category: %w", err)
	}
	if n != 1 {
		return n, fmt.Errorf("insufficient data for emitter category: got %d bytes, want 1", n)
	}

	e.ECAT = EmitterCategoryType(data[0])
	return n, e.Validate()
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
	if e.ECAT < 1 || e.ECAT > 24 {
		return fmt.Errorf("invalid emitter category value: %d", e.ECAT)
	}
	return nil
}

// String returns a human-readable representation of the EmitterCategory
func (e *EmitterCategory) String() string {
	switch e.ECAT {
	case EmitterLight:
		return "Light Aircraft"
	case EmitterSmall:
		return "Small Aircraft"
	case EmitterMedium:
		return "Medium Aircraft"
	case EmitterHigh:
		return "High Vortex Aircraft"
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
