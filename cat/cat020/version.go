// cat/cat020/version.go
package cat020

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat020/uap"
)

// Version constants
const (
	Version15 = "1.5"
)

// NewUAP returns the UAP for the specified version of CAT020
func NewUAP(version string) (asterix.UAP, error) {
	switch version {
	case Version15:
		return uap.NewUAP15()
	default:
		return nil, fmt.Errorf("unsupported CAT020 version: %s", version)
	}
}

// LatestVersion returns the latest available version
func LatestVersion() string {
	return Version15
}

// AvailableVersions returns all supported versions
func AvailableVersions() []string {
	return []string{Version15}
}
