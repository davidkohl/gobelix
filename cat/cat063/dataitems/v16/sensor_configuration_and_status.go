// cat/cat063/dataitems/v16/sensor_configuration_and_status.go
package v16

import (
	"bytes"
	"fmt"
	"strings"
)

// SensorConnectionStatus defines the connection status values
type SensorConnectionStatus uint8

const (
	StatusOperational SensorConnectionStatus = iota
	StatusDegraded
	StatusInitialization
	StatusNotConnected
)

// SensorConfigurationAndStatus implements I063/060
// Configuration and status of the sensor
type SensorConfigurationAndStatus struct {
	// First part
	CON SensorConnectionStatus // Connection status
	PSR bool                   // PSR NOGO
	SSR bool                   // SSR NOGO
	MDS bool                   // Mode S NOGO
	ADS bool                   // ADS NOGO
	MLT bool                   // MLT NOGO

	// First extent (only present if HasFirstExtent is true)
	HasFirstExtent bool // Indicates if first extent is present
	OPS            bool // System inhibited for operational use
	ODP            bool // Data Processor Overload
	OXT            bool // Transmission Subsystem Overload
	MSC            bool // Monitoring System Disconnected
	TSV            bool // Time Source Invalid
	NPW            bool // No Plot Warning
}

func (s *SensorConfigurationAndStatus) Decode(buf *bytes.Buffer) (int, error) {
	// Read first part
	data := make([]byte, 1)
	n, err := buf.Read(data)
	if err != nil {
		return n, fmt.Errorf("reading sensor configuration first part: %w", err)
	}

	bytesRead := n
	firstPart := data[0]

	// Decode first part
	s.CON = SensorConnectionStatus((firstPart >> 6) & 0x03)
	s.PSR = (firstPart & 0x20) != 0
	s.SSR = (firstPart & 0x10) != 0
	s.MDS = (firstPart & 0x08) != 0
	s.ADS = (firstPart & 0x04) != 0
	s.MLT = (firstPart & 0x02) != 0

	// Check if first extent is present
	fx := (firstPart & 0x01) != 0
	s.HasFirstExtent = fx

	// If first extent is present, decode it
	if fx {
		data := make([]byte, 1)
		n, err := buf.Read(data)
		if err != nil {
			return bytesRead, fmt.Errorf("reading sensor configuration first extent: %w", err)
		}

		bytesRead += n
		firstExtent := data[0]

		s.OPS = (firstExtent & 0x80) != 0
		s.ODP = (firstExtent & 0x40) != 0
		s.OXT = (firstExtent & 0x20) != 0
		s.MSC = (firstExtent & 0x10) != 0
		s.TSV = (firstExtent & 0x08) != 0
		s.NPW = (firstExtent & 0x04) != 0

		// Check for future extensions (not defined in v1.6)
		fx = (firstExtent & 0x01) != 0
		if fx {
			// Read and skip any additional extensions
			for fx {
				data := make([]byte, 1)
				n, err := buf.Read(data)
				if err != nil {
					return bytesRead, fmt.Errorf("reading sensor configuration additional extent: %w", err)
				}

				bytesRead += n
				fx = (data[0] & 0x01) != 0
			}
		}
	}

	return bytesRead, s.Validate()
}

func (s *SensorConfigurationAndStatus) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}

	bytesWritten := 0

	// Encode first part
	var firstPart byte
	firstPart |= byte(s.CON) << 6

	if s.PSR {
		firstPart |= 0x20
	}
	if s.SSR {
		firstPart |= 0x10
	}
	if s.MDS {
		firstPart |= 0x08
	}
	if s.ADS {
		firstPart |= 0x04
	}
	if s.MLT {
		firstPart |= 0x02
	}

	// Add extension bit if we have the first extent
	if s.HasFirstExtent {
		firstPart |= 0x01
	}

	n, err := buf.Write([]byte{firstPart})
	if err != nil {
		return bytesWritten, fmt.Errorf("writing sensor configuration first part: %w", err)
	}
	bytesWritten += n

	// Encode first extent if present
	if s.HasFirstExtent {
		var firstExtent byte

		if s.OPS {
			firstExtent |= 0x80
		}
		if s.ODP {
			firstExtent |= 0x40
		}
		if s.OXT {
			firstExtent |= 0x20
		}
		if s.MSC {
			firstExtent |= 0x10
		}
		if s.TSV {
			firstExtent |= 0x08
		}
		if s.NPW {
			firstExtent |= 0x04
		}

		// Bit 2 is spare and should be 0
		// No further extensions in v1.6, so FX bit is 0

		n, err := buf.Write([]byte{firstExtent})
		if err != nil {
			return bytesWritten, fmt.Errorf("writing sensor configuration first extent: %w", err)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

func (s *SensorConfigurationAndStatus) Validate() error {
	if s.CON > StatusNotConnected {
		return fmt.Errorf("invalid sensor connection status: %d", s.CON)
	}
	return nil
}

func (s *SensorConfigurationAndStatus) String() string {
	var parts []string

	// First part
	switch s.CON {
	case StatusOperational:
		parts = append(parts, "Operational")
	case StatusDegraded:
		parts = append(parts, "Degraded")
	case StatusInitialization:
		parts = append(parts, "Initialization")
	case StatusNotConnected:
		parts = append(parts, "Not Connected")
	}

	if s.PSR {
		parts = append(parts, "PSR NOGO")
	}
	if s.SSR {
		parts = append(parts, "SSR NOGO")
	}
	if s.MDS {
		parts = append(parts, "Mode S NOGO")
	}
	if s.ADS {
		parts = append(parts, "ADS NOGO")
	}
	if s.MLT {
		parts = append(parts, "MLT NOGO")
	}

	// First extent
	if s.HasFirstExtent {
		if s.OPS {
			parts = append(parts, "Operational use inhibited")
		}
		if s.ODP {
			parts = append(parts, "DP Overload")
		}
		if s.OXT {
			parts = append(parts, "Transmission Overload")
		}
		if s.MSC {
			parts = append(parts, "Monitoring disconnected")
		}
		if s.TSV {
			parts = append(parts, "Time Source Invalid")
		}
		if s.NPW {
			parts = append(parts, "No Plots Warning")
		}
	}

	return strings.Join(parts, ", ")
}
