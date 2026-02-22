// cat/cat062/fusion_real_data_test.go
package cat062_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// TestFusionRealData tests decoding real CAT062 messages from ATLAS fusion
func TestFusionRealData(t *testing.T) {
	tests := []struct {
		name    string
		hexData string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Real fusion message #1 - I062/220 EOF error",
			hexData: "3E0056BFFFED066401015D1A2E0083AAA30012C7C5F803CEF036F5039FFDFBFDFF0A9600154237C31820C101010004012B154237C318201B5283010130B0000000000005C705C70002C400C800C6FF6F4D4B84640528",
			wantErr: true,
			errMsg:  "I062/220",
		},
		{
			name:    "Real fusion message #2 - I062/500 ATV buffer too short",
			hexData: "3E0056BFFFED066401015D1A2F008E9F4B001D3BC0009AE7FD1579FCDBFF830E000200005161B2D995A0C101010039DE475161B2D995A0CB2483010130B12000040400000005C805C8000084001400142A2484622528",
			wantErr: true,
			errMsg:  "I062/500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex to bytes
			data := make([]byte, len(tt.hexData)/2)
			for i := 0; i < len(data); i++ {
				var b byte
				_, err := fmt.Sscanf(tt.hexData[i*2:i*2+2], "%02X", &b)
				if err != nil {
					t.Fatalf("Failed to parse hex: %v", err)
				}
				data[i] = b
			}

			// Decode
			decoder := asterix.NewDecoder()
			uap062, err := uap.NewUAP120()
			if err != nil {
				t.Fatalf("Failed to create UAP: %v", err)
			}
			decoder.RegisterUAP(uap062)

			dataBlocks, err := decoder.DecodeAll(data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got none", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				} else {
					t.Logf("Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					t.Logf("Successfully decoded %d data blocks", len(dataBlocks))
				}
			}
		})
	}
}
