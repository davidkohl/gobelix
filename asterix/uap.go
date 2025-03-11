// asterix/uap.go
package asterix

import "fmt"

// UAP defines the User Application Profile for an ASTERIX category
type UAP interface {
	// Category returns the ASTERIX category number
	Category() Category

	// Version returns the specification version implemented
	Version() string

	// Fields returns the data field definitions
	Fields() []DataField

	// CreateDataItem creates a new instance of a data item by its ID
	CreateDataItem(id string) (DataItem, error)

	// Validate validates a complete record against category-specific rules
	Validate(items map[string]DataItem) error
}

// BaseUAP provides common UAP functionality
type BaseUAP struct {
	category     Category
	version      string
	fields       []DataField
	mandatoryIDs []string // Pre-computed list of mandatory item IDs
}

func NewBaseUAP(cat Category, version string, fields []DataField) (*BaseUAP, error) {
	if !cat.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: no fields defined", ErrInvalidMessage)
	}

	// Validate field definitions and detect conflicts
	seenFRNs := make(map[uint8]string)
	var mandatoryIDs []string

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

		if field.Mandatory {
			mandatoryIDs = append(mandatoryIDs, field.DataItem)
		}
	}

	return &BaseUAP{
		category:     cat,
		version:      version,
		fields:       fields,
		mandatoryIDs: mandatoryIDs,
	}, nil
}

func (u *BaseUAP) Category() Category {
	return u.category
}

func (u *BaseUAP) Version() string {
	return u.version
}

func (u *BaseUAP) Fields() []DataField {
	fields := make([]DataField, len(u.fields))
	copy(fields, u.fields)
	return fields
}

// Validate implements basic validation checking mandatory fields
func (u *BaseUAP) Validate(items map[string]DataItem) error {
	for _, id := range u.mandatoryIDs {
		if _, exists := items[id]; !exists {
			return fmt.Errorf("%w: %s", ErrMandatoryField, id)
		}
	}
	return nil
}

// CreateDataItem must be implemented by specific UAPs
func (u *BaseUAP) CreateDataItem(id string) (DataItem, error) {
	return nil, fmt.Errorf("%w: CreateDataItem must be implemented by specific UAP",
		ErrUAPNotDefined)
}
