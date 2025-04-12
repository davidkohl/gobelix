// cat/cat063/dataitems/v16/sensor_configuration_and_status_test.go
package v16_test

import (
	"bytes"
	"testing"

	v16 "github.com/davidkohl/gobelix/cat/cat063/dataitems/v16"
)

func TestSensorConfigurationAndStatus_EncodeDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    v16.SensorConfigurationAndStatus
		expected []byte
	}{
		{
			name: "Only first part (no extensions)",
			input: v16.SensorConfigurationAndStatus{
				CON:            v16.StatusOperational,
				PSR:            true,
				SSR:            false,
				MDS:            true,
				ADS:            false,
				MLT:            true,
				HasFirstExtent: false,
			},
			expected: []byte{0x28}, // 00101000
		},
		{
			name: "First part and first extent",
			input: v16.SensorConfigurationAndStatus{
				CON:            v16.StatusDegraded,
				PSR:            false,
				SSR:            true,
				MDS:            false,
				ADS:            true,
				MLT:            false,
				HasFirstExtent: true,
				OPS:            true,
				ODP:            false,
				OXT:            true,
				MSC:            false,
				TSV:            true,
				NPW:            false,
			},
			expected: []byte{0x45, 0xA8}, // 01000101, 10101000
		},
		{
			name: "Status not connected with all NOGO and all warnings",
			input: v16.SensorConfigurationAndStatus{
				CON:            v16.StatusNotConnected,
				PSR:            true,
				SSR:            true,
				MDS:            true,
				ADS:            true,
				MLT:            true,
				HasFirstExtent: true,
				OPS:            true,
				ODP:            true,
				OXT:            true,
				MSC:            true,
				TSV:            true,
				NPW:            true,
			},
			expected: []byte{0xBF, 0xFC}, // 10111111, 11111100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Encode
			buf := new(bytes.Buffer)
			n, err := tt.input.Encode(buf)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if n != len(tt.expected) {
				t.Errorf("Encode() returned n = %v, want %v", n, len(tt.expected))
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("Encode() = %v, want %v", buf.Bytes(), tt.expected)
			}

			// Test Decode
			decodedItem := &v16.SensorConfigurationAndStatus{}
			encodedBuf := bytes.NewBuffer(tt.expected)
			n, err = decodedItem.Decode(encodedBuf)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if n != len(tt.expected) {
				t.Errorf("Decode() returned n = %v, want %v", n, len(tt.expected))
			}

			// Compare decoded value with original input
			if decodedItem.CON != tt.input.CON {
				t.Errorf("Decode() CON = %v, want %v", decodedItem.CON, tt.input.CON)
			}
			if decodedItem.PSR != tt.input.PSR {
				t.Errorf("Decode() PSR = %v, want %v", decodedItem.PSR, tt.input.PSR)
			}
			if decodedItem.SSR != tt.input.SSR {
				t.Errorf("Decode() SSR = %v, want %v", decodedItem.SSR, tt.input.SSR)
			}
			if decodedItem.MDS != tt.input.MDS {
				t.Errorf("Decode() MDS = %v, want %v", decodedItem.MDS, tt.input.MDS)
			}
			if decodedItem.ADS != tt.input.ADS {
				t.Errorf("Decode() ADS = %v, want %v", decodedItem.ADS, tt.input.ADS)
			}
			if decodedItem.MLT != tt.input.MLT {
				t.Errorf("Decode() MLT = %v, want %v", decodedItem.MLT, tt.input.MLT)
			}
			if decodedItem.HasFirstExtent != tt.input.HasFirstExtent {
				t.Errorf("Decode() HasFirstExtent = %v, want %v", decodedItem.HasFirstExtent, tt.input.HasFirstExtent)
			}

			// Only check the extension fields if it has extensions
			if tt.input.HasFirstExtent {
				if decodedItem.OPS != tt.input.OPS {
					t.Errorf("Decode() OPS = %v, want %v", decodedItem.OPS, tt.input.OPS)
				}
				if decodedItem.ODP != tt.input.ODP {
					t.Errorf("Decode() ODP = %v, want %v", decodedItem.ODP, tt.input.ODP)
				}
				if decodedItem.OXT != tt.input.OXT {
					t.Errorf("Decode() OXT = %v, want %v", decodedItem.OXT, tt.input.OXT)
				}
				if decodedItem.MSC != tt.input.MSC {
					t.Errorf("Decode() MSC = %v, want %v", decodedItem.MSC, tt.input.MSC)
				}
				if decodedItem.TSV != tt.input.TSV {
					t.Errorf("Decode() TSV = %v, want %v", decodedItem.TSV, tt.input.TSV)
				}
				if decodedItem.NPW != tt.input.NPW {
					t.Errorf("Decode() NPW = %v, want %v", decodedItem.NPW, tt.input.NPW)
				}
			}
		})
	}
}

func TestSensorConfigurationAndStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		input    v16.SensorConfigurationAndStatus
		expected string
	}{
		{
			name: "Basic operational with warnings",
			input: v16.SensorConfigurationAndStatus{
				CON:            v16.StatusOperational,
				PSR:            true, // NOGO
				SSR:            false,
				MDS:            false,
				ADS:            false,
				MLT:            false,
				HasFirstExtent: false,
			},
			expected: "Operational, PSR NOGO",
		},
		{
			name: "With extensions",
			input: v16.SensorConfigurationAndStatus{
				CON:            v16.StatusDegraded,
				PSR:            true,
				SSR:            true,
				MDS:            false,
				ADS:            false,
				MLT:            false,
				HasFirstExtent: true,
				OPS:            true,
				ODP:            false,
				OXT:            false,
				MSC:            false,
				TSV:            true,
				NPW:            true,
			},
			expected: "Degraded, PSR NOGO, SSR NOGO, Operational use inhibited, Time Source Invalid, No Plots Warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
