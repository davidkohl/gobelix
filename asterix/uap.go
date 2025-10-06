// asterix/uap.go
package asterix

import "fmt"

// UAP (User Application Profile) defines the structure of an ASTERIX category
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

// NewBaseUAP creates a new BaseUAP with the given parameters
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

// Category returns the ASTERIX category number
func (u *BaseUAP) Category() Category {
	return u.category
}

// Version returns the specification version implemented
func (u *BaseUAP) Version() string {
	return u.version
}

// Fields returns the data field definitions
// The returned slice is the internal slice. Do not modify it.
// UAP definitions are meant to be immutable after creation.
func (u *BaseUAP) Fields() []DataField {
	return u.fields
}

// Validate implements basic validation checking mandatory fields
func (u *BaseUAP) Validate(items map[string]DataItem) error {
	// Check mandatory fields
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

// GetFieldByDataItem returns the field definition for a data item
func (u *BaseUAP) GetFieldByDataItem(id string) (DataField, bool) {
	for _, field := range u.fields {
		if field.DataItem == id {
			return field, true
		}
	}
	return DataField{}, false
}

// GetFieldByFRN returns the field definition for a field reference number
func (u *BaseUAP) GetFieldByFRN(frn uint8) (DataField, bool) {
	for _, field := range u.fields {
		if field.FRN == frn {
			return field, true
		}
	}
	return DataField{}, false
}

// IsMandatory checks if a data item is mandatory
func (u *BaseUAP) IsMandatory(id string) bool {
	for _, mandatoryID := range u.mandatoryIDs {
		if mandatoryID == id {
			return true
		}
	}
	return false
}

// MaxFRN returns the maximum FRN in this UAP
func (u *BaseUAP) MaxFRN() uint8 {
	var max uint8
	for _, field := range u.fields {
		if field.FRN > max {
			max = field.FRN
		}
	}
	return max
}

// MessageTypeValidator is a function that validates message types
type MessageTypeValidator func(messageType uint8, items map[string]DataItem) error

// TypedUAP extends BaseUAP with message type validation
type TypedUAP struct {
	*BaseUAP
	messageTypes           map[uint8]string               // Maps message type values to descriptions
	typeValidators         map[uint8]MessageTypeValidator // Message type-specific validation
	messageTypeItemID      string                         // ID of the item containing the message type
	extractMessageTypeFunc func(DataItem) (uint8, error)  // Function to extract message type
}

// NewTypedUAP creates a new TypedUAP for categories with message types
func NewTypedUAP(baseUAP *BaseUAP, messageTypeItemID string) *TypedUAP {
	return &TypedUAP{
		BaseUAP:           baseUAP,
		messageTypes:      make(map[uint8]string),
		typeValidators:    make(map[uint8]MessageTypeValidator),
		messageTypeItemID: messageTypeItemID,
	}
}

// RegisterMessageType adds a message type to the UAP
func (u *TypedUAP) RegisterMessageType(typeValue uint8, description string, validator MessageTypeValidator) {
	u.messageTypes[typeValue] = description
	if validator != nil {
		u.typeValidators[typeValue] = validator
	}
}

// MessageTypes returns a map of message types
func (u *TypedUAP) MessageTypes() map[uint8]string {
	// Return a copy to prevent modification
	types := make(map[uint8]string, len(u.messageTypes))
	for k, v := range u.messageTypes {
		types[k] = v
	}
	return types
}

// Validate overrides the basic validation to include message type validation
func (u *TypedUAP) Validate(items map[string]DataItem) error {
	// First do base validation
	if err := u.BaseUAP.Validate(items); err != nil {
		return err
	}

	// If no message type item, just return success
	if u.messageTypeItemID == "" {
		return nil
	}

	// Get the message type
	typeItem, exists := items[u.messageTypeItemID]
	if !exists {
		return nil // No type validation if type item not present
	}

	// Extract the message type value
	// This is category-specific and would be implemented by the concrete UAP
	typeValue, err := u.extractMessageType(typeItem)
	if err != nil {
		return err
	}

	// Check if message type is registered
	if _, exists := u.messageTypes[typeValue]; !exists {
		return fmt.Errorf("%w: unknown message type %d", ErrInvalidField, typeValue)
	}

	// Run type-specific validation if defined
	if validator, exists := u.typeValidators[typeValue]; exists {
		return validator(typeValue, items)
	}

	return nil
}

// extractMessageType extracts the message type value from a data item
func (u *TypedUAP) extractMessageType(item DataItem) (uint8, error) {
	if u.extractMessageTypeFunc != nil {
		return u.extractMessageTypeFunc(item)
	}

	// Default implementation assumes the item has a simple byte value
	// Real implementation would depend on the category
	return 0, fmt.Errorf("extractMessageType not implemented for this UAP")
}
