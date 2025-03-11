// cat/cat062/uap.go
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

// GetUAP returns the UAP for the specified version of CAT062
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version26:
		return uap.NewUAP26()
	default:
		return nil, fmt.Errorf("unsupported CAT062 version: %s", version)
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
