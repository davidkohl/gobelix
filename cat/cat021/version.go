// Package cat021 provides ASTERIX Category 021 (ADS-B Target Reports) implementation.
package cat021

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021/uap"
)

// Version constants
const (
	Version26 = "2.6"
)

// NewUAP returns the UAP for the specified version of CAT021
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version26:
		return uap.NewUAP26()
	default:
		return nil, fmt.Errorf("unsupported CAT021 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version26
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version26}
}
