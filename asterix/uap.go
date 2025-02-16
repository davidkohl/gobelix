// asterix/uap.go
package asterix

import "fmt"

// UAP defines the User Application Profile for an ASTERIX category
type UAP interface {
	// Category returns the ASTERIX category number
	Category() Category

	// Version returns the specification version implemented
	Version() string

	// Fields returns the data field definitions for this category
	Fields() []DataField

	// CreateDataItem creates a new instance of a data item by its ID
	CreateDataItem(id string) (DataItem, error)

	// FRNByID returns the Field Reference Number for a given data item ID
	FRNByID(id string) uint8
}

// BaseUAP provides common UAP functionality
type BaseUAP struct {
	category Category
	version  string
	fields   []DataField
}

// NewBaseUAP creates a new base UAP implementation
func NewBaseUAP(cat Category, version string, fields []DataField) (*BaseUAP, error) {
	if !cat.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: no fields defined", ErrInvalidMessage)
	}

	// Validate field definitions
	seenFRNs := make(map[uint8]string)
	for _, field := range fields {
		if field.FRN == 0 {
			return nil, fmt.Errorf("%w: FRN cannot be 0 for %s",
				ErrInvalidField, field.DataItem)
		}
		if existing, exists := seenFRNs[field.FRN]; exists {
			return nil, fmt.Errorf("%w: duplicate FRN %d for items %s and %s",
				ErrInvalidField, field.FRN, existing, field.DataItem)
		}
		seenFRNs[field.FRN] = field.DataItem
	}

	return &BaseUAP{
		category: cat,
		version:  version,
		fields:   fields,
	}, nil
}

func (u *BaseUAP) Category() Category {
	return u.category
}

func (u *BaseUAP) Version() string {
	return u.version
}

func (u *BaseUAP) Fields() []DataField {
	// Return a copy to prevent modification
	fields := make([]DataField, len(u.fields))
	copy(fields, u.fields)
	return fields
}

func (u *BaseUAP) FRNByID(id string) uint8 {
	for _, field := range u.fields {
		if field.DataItem == id {
			return field.FRN
		}
	}
	return 0
}

// CreateDataItem must be implemented by specific UAP implementations
func (u *BaseUAP) CreateDataItem(id string) (DataItem, error) {
	return nil, fmt.Errorf("%w: CreateDataItem must be implemented by specific UAP",
		ErrUAPNotDefined)
}
