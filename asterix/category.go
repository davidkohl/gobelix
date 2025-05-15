// asterix/category.go
package asterix

import "fmt"

// Category represents an ASTERIX category number
type Category uint8

// Define known categories
const (
	Cat021 Category = 21  // ADS-B Reports
	Cat048 Category = 48  // Monoradar Target Reports
	Cat062 Category = 62  // System Track Data
	Cat063 Category = 63  // Sensor Status Messages
	Cat065 Category = 65  // SDPS Service Status Messages
	Cat247 Category = 247 // Mode S Reports
	Cat252 Category = 252 // ASTERIX Summary Records
)

// String returns a string representation of the category
func (c Category) String() string {
	return fmt.Sprintf("CAT%03d", c)
}

// IsValid returns true if this is a recognized ASTERIX category
func (c Category) IsValid() bool {
	switch c {
	case Cat021, Cat048, Cat062, Cat063, Cat065, Cat247, Cat252:
		return true
	default:
		// Additional check for other valid categories
		return c > 0
	}
}

// IsBlockable returns true if this category supports blocking
// Categories created before Edition 2.2 of Part 1 of ASTERIX can use blocking
// Newer categories must not use blocking
func (c Category) IsBlockable() bool {
	switch c {
	case Cat021, Cat048, Cat062, Cat063, Cat065, Cat247:
		return true
	default:
		return false
	}
}

// Byte converts the category to a byte
func (c Category) Byte() byte {
	return byte(c)
}

// CategoryFromByte creates a Category from a byte
func CategoryFromByte(b byte) Category {
	return Category(b)
}

// CategoryInfo provides additional metadata about a category
type CategoryInfo struct {
	Number      Category
	Name        string
	Description string
	Version     string
	Blockable   bool
}

// GetCategoryInfo returns detailed information about a category
func GetCategoryInfo(cat Category) CategoryInfo {
	switch cat {
	case Cat021:
		return CategoryInfo{
			Number:      Cat021,
			Name:        "CAT021",
			Description: "ADS-B Target Reports",
			Version:     "2.6",
			Blockable:   true,
		}
	case Cat048:
		return CategoryInfo{
			Number:      Cat048,
			Name:        "CAT048",
			Description: "Monoradar Target Reports",
			Version:     "1.30",
			Blockable:   true,
		}
	case Cat062:
		return CategoryInfo{
			Number:      Cat062,
			Name:        "CAT062",
			Description: "System Track Data",
			Version:     "1.19",
			Blockable:   true,
		}
	case Cat063:
		return CategoryInfo{
			Number:      Cat063,
			Name:        "CAT063",
			Description: "Sensor Status Messages",
			Version:     "1.7",
			Blockable:   true,
		}
	case Cat065:
		return CategoryInfo{
			Number:      Cat065,
			Name:        "CAT065",
			Description: "SDPS Service Status Messages",
			Version:     "1.4",
			Blockable:   true,
		}
	case Cat247:
		return CategoryInfo{
			Number:      Cat247,
			Name:        "CAT247",
			Description: "Mode S Reports",
			Version:     "1.2",
			Blockable:   true,
		}
	case Cat252:
		return CategoryInfo{
			Number:      Cat252,
			Name:        "CAT252",
			Description: "ASTERIX Summary Records",
			Version:     "7.0",
			Blockable:   false,
		}
	default:
		return CategoryInfo{
			Number:      cat,
			Name:        fmt.Sprintf("CAT%03d", cat),
			Description: "Unknown category",
			Version:     "",
			Blockable:   false,
		}
	}
}
