// cat/cat023/cat023_test.go
package cat023_test

import (
	"testing"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat023"
)

func TestCat023Version(t *testing.T) {
	if cat023.LatestVersion() != "1.3" {
		t.Errorf("expected version 1.3, got %s", cat023.LatestVersion())
	}

	versions := cat023.AvailableVersions()
	if len(versions) != 1 || versions[0] != "1.3" {
		t.Errorf("expected [1.3], got %v", versions)
	}
}

func TestCat023UAP(t *testing.T) {
	uap, err := cat023.NewUAP("1.3")
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	if uap.Category() != asterix.Cat023 {
		t.Errorf("expected category 023, got %v", uap.Category())
	}

	if uap.Version() != "1.3" {
		t.Errorf("expected version 1.3, got %s", uap.Version())
	}
}

func TestCat023InvalidVersion(t *testing.T) {
	_, err := cat023.NewUAP("9.9")
	if err == nil {
		t.Error("expected error for invalid version, got nil")
	}
}

func TestCat023DataItemCreation(t *testing.T) {
	uap, err := cat023.NewUAP("1.3")
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	testCases := []string{
		"I023/000", // Report Type
		"I023/010", // Data Source Identifier
		"I023/015", // Service Type and Identification
		"I023/070", // Time of Day
		"I023/100", // Ground Station Status
		"I023/101", // Service Configuration
		"I023/110", // Service Status
		"I023/120", // Service Statistics
		"I023/200", // Operational Range
		"RE023",    // Reserved Expansion
		"SP023",    // Special Purpose
	}

	for _, itemID := range testCases {
		t.Run(itemID, func(t *testing.T) {
			item, err := uap.CreateDataItem(itemID)
			if err != nil {
				t.Errorf("failed to create data item %s: %v", itemID, err)
			}
			if item == nil {
				t.Errorf("CreateDataItem returned nil for %s", itemID)
			}
		})
	}
}

func TestCat023UnknownDataItem(t *testing.T) {
	uap, err := cat023.NewUAP("1.3")
	if err != nil {
		t.Fatalf("failed to create UAP: %v", err)
	}

	_, err = uap.CreateDataItem("I023/999")
	if err == nil {
		t.Error("expected error for unknown data item, got nil")
	}
}
