// cat/cat001/version.go
package cat001

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat001/uap"
)

// Version constants
const (
	Version12 = "1.2"
)

// NewUAP returns the UAP for the specified version of CAT001
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version12:
		return uap.NewUAP12()
	default:
		return nil, fmt.Errorf("unsupported CAT001 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version12
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version12}
}
