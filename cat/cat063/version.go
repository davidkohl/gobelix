// cat/cat063/version.go
package cat063

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat063/uap"
)

// Version constants
const (
	Version16 = "1.6"
)

// NewUAP returns the UAP for the specified version of CAT063
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version16:
		return uap.NewUAP063()
	default:
		return nil, fmt.Errorf("unsupported CAT063 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version16
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version16}
}
