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

	// GetFieldByDataItem returns the field definition for a data item (O(1) lookup)
	GetFieldByDataItem(id string) (DataField, bool)

	// GetFieldByFRN returns the field definition for a FRN (O(1) lookup)
	GetFieldByFRN(frn uint8) (DataField, bool)

	// CreateDataItem creates a new instance of a data item by its ID
	CreateDataItem(id string) (DataItem, error)

	// Validate validates a complete record against category-specific rules
	Validate(items map[string]DataItem) error
}

// BaseUAP provides common UAP functionality
type BaseUAP struct {
	category         Category
	version          string
	fields           []DataField
	mandatoryIDs     []string             // Pre-computed list of mandatory item IDs
	fieldsByDataItem map[string]DataField // O(1) lookup by data item ID
	fieldsByFRN      map[uint8]DataField  // O(1) lookup by FRN
	layout           decodeLayout         // FSPEC bit -> field, for Record.Decode
}

// maxFSPECOctets is the longest FSPEC that FSPEC.Decode accepts, and
// maxFSPECFRN the highest FRN it can signal.
const (
	maxFSPECOctets = 8
	maxFSPECFRN    = maxFSPECOctets * 7
)

// decodeLayout maps each FRN that an FSPEC can signal to the field that
// Record.Decode reads for it. field[frn] is -1 for an FRN without a data
// item (spare, reserved, beyond the UAP); rfsFRN is the RS bit, 0 if none;
// bit frn of notInRFS is set for an FRN that the RFS field must not carry.
type decodeLayout struct {
	field    [maxFSPECFRN + 1]int16
	rfsFRN   uint8
	notInRFS uint64
}

// RandomFieldSequencer is implemented by a UAP whose FSPEC reserves an RS
// bit for a Random Field Sequencing (RFS) organised field (EUROCONTROL
// ASTERIX Part 1; CAT001 ed 1.4 §6, CAT002 ed 1.2 §6). RFSFRN returns the
// FRN of that bit, or 0 when the UAP has none. BaseUAP implements it; see
// BaseUAP.EnableRFS.
type RandomFieldSequencer interface {
	RFSFRN() uint8
}

// layoutUAP gives Record.Decode the precomputed decodeLayout of a BaseUAP
// (or a UAP embedding one).
type layoutUAP interface {
	decodeLayout() *decodeLayout
}

// NewBaseUAP creates a new BaseUAP with the given parameters
func NewBaseUAP(cat Category, version string, fields []DataField) (*BaseUAP, error) {
	if !cat.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: no fields defined", ErrInvalidMessage)
	}

	// Validate field definitions, detect conflicts, and build lookup maps
	fieldsByDataItem := make(map[string]DataField, len(fields))
	fieldsByFRN := make(map[uint8]DataField, len(fields))
	var mandatoryIDs []string

	for _, field := range fields {
		if field.FRN == 0 {
			return nil, fmt.Errorf("%w: FRN cannot be 0 for %s",
				ErrInvalidField, field.DataItem)
		}
		if existing, exists := fieldsByFRN[field.FRN]; exists {
			return nil, fmt.Errorf("%w: duplicate FRN %d for items %s and %s",
				ErrInvalidField, field.FRN, existing.DataItem, field.DataItem)
		}
		fieldsByFRN[field.FRN] = field
		fieldsByDataItem[field.DataItem] = field

		if field.Mandatory {
			mandatoryIDs = append(mandatoryIDs, field.DataItem)
		}
	}

	u := &BaseUAP{
		category:         cat,
		version:          version,
		fields:           fields,
		mandatoryIDs:     mandatoryIDs,
		fieldsByDataItem: fieldsByDataItem,
		fieldsByFRN:      fieldsByFRN,
	}
	for i := range u.layout.field {
		u.layout.field[i] = -1
	}
	for i, field := range fields {
		if field.DataItem != "" && field.FRN <= maxFSPECFRN {
			u.layout.field[field.FRN] = int16(i)
		}
	}
	return u, nil
}

func (u *BaseUAP) decodeLayout() *decodeLayout {
	return &u.layout
}

// EnableRFS declares frn as the RFS indicator (RS bit) of this UAP, so that
// Record.Decode reads the RFS organised field at that FSPEC position:
// one octet N, then N times a one-octet FRN followed by the data item of
// that FRN. frn must not be used by a field. notInRFS lists the FRNs that
// the RFS field must not carry (an RFS entry naming one is a decode
// error), for example the SP field (CAT001 ed 1.4 §5.3.2.16 note 4,
// CAT002 ed 1.2 §5.3.2.13 note 4: non-standard information goes in the SP
// field "and not the RFS field(s)"). Each call replaces the previous one.
// Call it only while building the UAP, before it is shared.
func (u *BaseUAP) EnableRFS(frn uint8, notInRFS ...uint8) error {
	if frn == 0 || frn > maxFSPECFRN {
		return fmt.Errorf("%w: RFS FRN %d out of range", ErrInvalidField, frn)
	}
	if f, exists := u.fieldsByFRN[frn]; exists && f.DataItem != "" {
		return fmt.Errorf("%w: RFS FRN %d is used by %s", ErrInvalidField, frn, f.DataItem)
	}
	var mask uint64
	for _, x := range notInRFS {
		if x == 0 || x > maxFSPECFRN || x == frn {
			return fmt.Errorf("%w: FRN %d excluded from RFS out of range or the RS FRN", ErrInvalidField, x)
		}
		mask |= 1 << x
	}
	u.layout.rfsFRN = frn
	u.layout.notInRFS = mask
	return nil
}

// RFSFRN implements RandomFieldSequencer: the FRN of the RS bit, or 0.
func (u *BaseUAP) RFSFRN() uint8 {
	return u.layout.rfsFRN
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
	field, exists := u.fieldsByDataItem[id]
	return field, exists
}

// GetFieldByFRN returns the field definition for a field reference number
func (u *BaseUAP) GetFieldByFRN(frn uint8) (DataField, bool) {
	field, exists := u.fieldsByFRN[frn]
	return field, exists
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
