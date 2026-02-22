// cat/cat048/dataitems/v132/aircraft_address.go
package v132

import (
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// AircraftAddress is an alias for common.TargetAddress.
// It implements I048/220 - Aircraft Address (24-bit Mode S address).
// Deprecated: Use common.TargetAddress directly.
type AircraftAddress = common.TargetAddress
