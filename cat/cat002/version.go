// cat/cat002/version.go
package cat002

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat002/uap"
)

// Version constants
const (
	Version10 = "1.0"
)

// NewUAP returns the UAP for the specified version of CAT002
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version10:
		return uap.NewUAP10()
	default:
		return nil, fmt.Errorf("unsupported CAT002 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version10
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version10}
}
