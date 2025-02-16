// asterix/category.go
package asterix

import "fmt"

// Category represents an ASTERIX category number
type Category uint8

// Define known categories
const (
	Cat021 Category = 21
	Cat062 Category = 62
	Cat063 Category = 63
)

func (c Category) String() string {
	return fmt.Sprintf("CAT%03d", c)
}

func (c Category) IsValid() bool {
	switch c {
	case Cat021, Cat062, Cat063:
		return true
	default:
		return false
	}
}
