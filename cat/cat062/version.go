// cat/cat062/uap.go
package cat062

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat062/uap"
)

// Version constants
const (
	Version117 = "1.17"
	Version120 = "1.20"
)

// GetUAP returns the UAP for the specified version of CAT062
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version117:
		return uap.NewUAP117()
	case Version120:
		return uap.NewUAP120()
	default:
		return nil, fmt.Errorf("unsupported CAT062 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version120
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version117, Version120}
}
