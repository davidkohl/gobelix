// cat/cat034/dataitems/v129/sector_number.go
package v129

import (
	common "github.com/davidkohl/gobelix/cat/common/dataitems"
)

// SectorNumber is an alias for common.SectorNumber.
// It implements I034/020 - Sector Number.
// Deprecated: Use common.SectorNumber directly.
type SectorNumber = common.SectorNumber

// NewSectorNumber creates a new Sector Number data item.
// Deprecated: Use &common.SectorNumber{} directly.
func NewSectorNumber() *common.SectorNumber {
	return &common.SectorNumber{}
}
