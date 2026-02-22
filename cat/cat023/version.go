// Package cat023 provides ASTERIX Category 023 (CNS/ATM Ground Station Messages) implementation.
package cat023

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat023/uap"
)

// Version constants
const (
	Version13 = "1.3"
)

// NewUAP returns the UAP for the specified version of CAT023
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version13:
		return uap.NewUAP13()
	default:
		return nil, fmt.Errorf("unsupported CAT023 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version13
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version13}
}
