// cat/cat034/version.go
package cat034

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat034/uap"
)

// Version constants
const (
	Version129 = "1.29"
)

// NewUAP returns the UAP for the specified version of CAT034
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version129:
		return uap.NewUAP129()
	default:
		return nil, fmt.Errorf("unsupported CAT034 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version129
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version129}
}
