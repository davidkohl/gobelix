// dataitems/cat048/warning_error_condition.go
package v132

import (
	"bytes"
	"fmt"
	"strings"
)

// WarningErrorCondition implements I048/030
// Warning/error conditions detected by a radar station for the target report involved
// and target classification information.
type WarningErrorCondition struct {
	Codes      []uint8 // List of warning/error/classification codes
	extensions uint8   // Number of extensions present
}

// Code values as defined in the specification
const (
	CodeMultipathReply                 uint8 = 1  // Multipath Reply (Reflection)
	CodeSidelobeReply                  uint8 = 2  // Reply due to sidelobe interrogation/reception
	CodeSplitPlot                      uint8 = 3  // Split plot
	CodeSecondTimeAround               uint8 = 4  // Second time around reply
	CodeAngel                          uint8 = 5  // Angel
	CodeTerrestrialVehicle             uint8 = 6  // Slow moving target correlated with road infrastructure
	CodeFixedPSRPlot                   uint8 = 7  // Fixed PSR plot
	CodeSlowPSRTarget                  uint8 = 8  // Slow PSR target
	CodeLowQualityPSRPlot              uint8 = 9  // Low quality PSR plot
	CodePhantomSSRPlot                 uint8 = 10 // Phantom SSR plot
	CodeNonMatchingMode3ACode          uint8 = 11 // Non-Matching Mode-3/A Code
	CodeAbnormalMode3AltitudeCode      uint8 = 12 // Mode C code / Mode S altitude code abnormal value
	CodeTargetInClutterArea            uint8 = 13 // Target in Clutter Area
	CodeMaxDopplerResponseInZeroFilter uint8 = 14 // Maximum Doppler Response in Zero Filter
	CodeTransponderAnomaly             uint8 = 15 // Transponder anomaly detected
	CodeDuplicatedIllegalModeS         uint8 = 16 // Duplicated or Illegal Mode S Aircraft Address
	CodeModeS_ErrorCorrection          uint8 = 17 // Mode S error correction applied
	CodeUndecodableMode3AltitudeCode   uint8 = 18 // Undecodable Mode C code / Mode S altitude code
	CodeBirds                          uint8 = 19 // Birds
	CodeFlockOfBirds                   uint8 = 20 // Flock of Birds
	CodeMode1Present                   uint8 = 21 // Mode-1 was present in original reply
	CodeMode2Present                   uint8 = 22 // Mode-2 was present in original reply
	CodeWindTurbine                    uint8 = 23 // Plot potentially caused by Wind Turbine
	CodeHelicopter                     uint8 = 24 // Helicopter
	CodeMaxReinterrogationsSurv        uint8 = 25 // Maximum number of re-interrogations reached (surveillance)
	CodeMaxReinterrogationsBDS         uint8 = 26 // Maximum number of re-interrogations reached (BDS Extractions)
	CodeBDSOverlayIncoherence          uint8 = 27 // BDS Overlay Incoherence
	CodePotentialBDSSwap               uint8 = 28 // Potential BDS Swap Detected
	CodeTrackUpdateZenithalGap         uint8 = 29 // Track Update in the Zenithal Gap
	CodeModeS_TrackReacquired          uint8 = 30 // Mode S Track re-acquired
	CodeDuplicatedMode5Pair            uint8 = 31 // Duplicated Mode 5 Pair NO/PIN detected
	CodeWrongDFReplyFormat             uint8 = 32 // Wrong DF reply format detected
	CodeTransponderAnomalyMSXPD        uint8 = 33 // Transponder anomaly (MS XPD replies with Mode A/C to Mode A/C only all-call)
	CodeTransponderAnomalySI           uint8 = 34 // Transponder anomaly (SI capability report wrong)
	CodePotentialICConflict            uint8 = 35 // Potential IC Conflict
	CodeICConflictDetection            uint8 = 36 // IC Conflict detection possible - no conflict currently detected
	CodeDuplicateMode5PIN              uint8 = 37 // Duplicate Mode 5 PIN
)

// Decode implements the DataItem interface
func (w *WarningErrorCondition) Decode(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	w.Codes = w.Codes[:0] // Reset existing codes

	// Keep reading extensions as long as FX bit is set
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return bytesRead, fmt.Errorf("reading warning/error condition: %w", err)
		}
		bytesRead++

		code := (b >> 1) & 0x7F // bits 8-2
		if code > 0 {           // Only add non-zero codes
			w.Codes = append(w.Codes, code)
		}

		fx := (b & 0x01) != 0 // bit 1 (FX)
		if !fx {
			break // No more extensions
		}
		w.extensions++

		// Safety check to prevent infinite loops
		if w.extensions > 10 { // Arbitrary limit
			return bytesRead, fmt.Errorf("too many extensions in warning/error condition")
		}
	}

	return bytesRead, w.Validate()
}

// Encode implements the DataItem interface
func (w *WarningErrorCondition) Encode(buf *bytes.Buffer) (int, error) {
	if err := w.Validate(); err != nil {
		return 0, err
	}

	if len(w.Codes) == 0 {
		// If no codes, encode a zero byte (no extension)
		err := buf.WriteByte(0)
		if err != nil {
			return 0, fmt.Errorf("writing empty warning/error condition: %w", err)
		}
		return 1, nil
	}

	bytesWritten := 0

	// Write out all codes
	for i, code := range w.Codes {
		b := (code & 0x7F) << 1 // bits 8-2
		if i < len(w.Codes)-1 {
			b |= 0x01 // bit 1 (FX) - set if not the last code
		}

		err := buf.WriteByte(b)
		if err != nil {
			return bytesWritten, fmt.Errorf("writing warning/error condition: %w", err)
		}
		bytesWritten++
	}

	return bytesWritten, nil
}

// Validate implements the DataItem interface
func (w *WarningErrorCondition) Validate() error {
	// Check that all codes are in valid range
	for i, code := range w.Codes {
		if code > 127 {
			return fmt.Errorf("code at position %d exceeds valid range [0,127]: %d", i, code)
		}
		if code == 0 {
			return fmt.Errorf("invalid zero code at position %d", i)
		}
	}
	return nil
}

// String returns a human-readable representation
func (w *WarningErrorCondition) String() string {
	if len(w.Codes) == 0 {
		return "No warnings/errors"
	}

	var descriptions []string
	for _, code := range w.Codes {
		desc := getCodeDescription(code)
		descriptions = append(descriptions, fmt.Sprintf("%d(%s)", code, desc))
	}

	return strings.Join(descriptions, ", ")
}

// getCodeDescription returns a description for a code
func getCodeDescription(code uint8) string {
	switch code {
	case CodeMultipathReply:
		return "Multipath Reply"
	case CodeSidelobeReply:
		return "Sidelobe Reply"
	case CodeSplitPlot:
		return "Split Plot"
	case CodeSecondTimeAround:
		return "Second Time Around"
	case CodeAngel:
		return "Angel"
	case CodeTerrestrialVehicle:
		return "Terrestrial Vehicle"
	case CodeFixedPSRPlot:
		return "Fixed PSR Plot"
	case CodeSlowPSRTarget:
		return "Slow PSR Target"
	case CodeLowQualityPSRPlot:
		return "Low Quality PSR Plot"
	case CodePhantomSSRPlot:
		return "Phantom SSR Plot"
	case CodeNonMatchingMode3ACode:
		return "Non-Matching Mode-3/A Code"
	case CodeAbnormalMode3AltitudeCode:
		return "Abnormal Mode C/Alt Value"
	case CodeTargetInClutterArea:
		return "Target In Clutter"
	case CodeMaxDopplerResponseInZeroFilter:
		return "Max Doppler in Zero Filter"
	case CodeTransponderAnomaly:
		return "Transponder Anomaly"
	case CodeDuplicatedIllegalModeS:
		return "Duplicate/Illegal Mode S Address"
	case CodeModeS_ErrorCorrection:
		return "Mode S Error Correction Applied"
	case CodeUndecodableMode3AltitudeCode:
		return "Undecodable Mode C/Alt Code"
	case CodeBirds:
		return "Birds"
	case CodeFlockOfBirds:
		return "Flock of Birds"
	case CodeMode1Present:
		return "Mode-1 Present"
	case CodeMode2Present:
		return "Mode-2 Present"
	case CodeWindTurbine:
		return "Wind Turbine"
	case CodeHelicopter:
		return "Helicopter"
	case CodeMaxReinterrogationsSurv:
		return "Max Re-interrogations (Surveillance)"
	case CodeMaxReinterrogationsBDS:
		return "Max Re-interrogations (BDS)"
	case CodeBDSOverlayIncoherence:
		return "BDS Overlay Incoherence"
	case CodePotentialBDSSwap:
		return "Potential BDS Swap"
	case CodeTrackUpdateZenithalGap:
		return "Track Update in Zenithal Gap"
	case CodeModeS_TrackReacquired:
		return "Mode S Track Re-acquired"
	case CodeDuplicatedMode5Pair:
		return "Duplicate Mode 5 Pair"
	case CodeWrongDFReplyFormat:
		return "Wrong DF Reply Format"
	case CodeTransponderAnomalyMSXPD:
		return "Transponder Anomaly (A/C All-call)"
	case CodeTransponderAnomalySI:
		return "Transponder Anomaly (SI)"
	case CodePotentialICConflict:
		return "Potential IC Conflict"
	case CodeICConflictDetection:
		return "IC Conflict Detection"
	case CodeDuplicateMode5PIN:
		return "Duplicate Mode 5 PIN"
	default:
		if code >= 64 && code <= 127 {
			return "Manufacturer Specific"
		}
		return "Unknown"
	}
}

// AddCode adds a warning/error code to the data item
func (w *WarningErrorCondition) AddCode(code uint8) {
	if code > 0 && code <= 127 {
		w.Codes = append(w.Codes, code)
	}
}
