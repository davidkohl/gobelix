// Package cat011 provides ASTERIX Category 011 (A-SMGCS Data Reports) implementation.
package cat011

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat011/uap"
)

// Version constants
const (
	Version13 = "1.3"
)

// NewUAP returns the UAP for the specified version of CAT011
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version13:
		return uap.NewUAP13()
	default:
		return nil, fmt.Errorf("unsupported CAT011 version: %s", version)
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
