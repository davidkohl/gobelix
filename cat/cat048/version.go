// cat/cat048/version.go
package cat048

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat048/uap"
)

// Version constants
const (
	Version132 = "1.32"
)

// NewUAP returns the UAP for the specified version of CAT048
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version132:
		return uap.NewUAP132()
	default:
		return nil, fmt.Errorf("unsupported CAT048 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version132
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version132}
}
