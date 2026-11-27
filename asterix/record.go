// asterix/record.go
package asterix

import (
	"bytes"
	"fmt"
	"io"
	"math/bits"
)

// Record represents a single ASTERIX record.
//
// Thread Safety: Record is NOT safe for concurrent use.
// Each Record instance should be accessed by only one goroutine at a time,
// or protected by external synchronization.
// Methods that modify the Record (SetDataItem, Reset) should not be called
// concurrently with any other methods.
type Record struct {
	category Category            // Category of this record
	fspec    *FSPEC              // Field Specification
	items    map[string]DataItem // Data items indexed by their reference (e.g., "I021/010")
	uap      UAP                 // User Application Profile defining the structure
}

// NewRecord creates a new record for a specific category
func NewRecord(cat Category, uap UAP) (*Record, error) {
	if !cat.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCategory, cat)
	}
	if uap == nil {
		return nil, fmt.Errorf("%w: UAP cannot be nil", ErrInvalidMessage)
	}
	if uap.Category() != cat {
		return nil, fmt.Errorf("%w: UAP category %d does not match record category %d",
			ErrInvalidMessage, uap.Category(), cat)
	}

	return &Record{
		category: cat,
		fspec:    NewFSPEC(),
		items:    make(map[string]DataItem),
		uap:      uap,
	}, nil
}

// SetDataItem adds or updates a data item
func (r *Record) SetDataItem(id string, item DataItem) error {
	if item == nil {
		return fmt.Errorf("%w: data item cannot be nil", ErrInvalidMessage)
	}

	// Find FRN for this item using O(1) lookup
	field, exists := r.uap.GetFieldByDataItem(id)
	if !exists {
		return fmt.Errorf("%w: %s", ErrUnknownDataItem, id)
	}

	if err := item.Validate(); err != nil {
		return WrapError(err, "validating %s", id)
	}

	r.items[id] = item
	return r.fspec.SetFRN(field.FRN)
}

// GetDataItem retrieves a data item by its ID
func (r *Record) GetDataItem(id string) (DataItem, bool) {
	item, exists := r.items[id]
	return item, exists
}

// RecordUAPSelector is implemented by a UAP that stands for several
// record layouts of one category edition, where the layout of each record is
// chosen by the content of the record itself. CAT001 is the case in point:
// I001/020 TYP selects the plot UAP or the track UAP (CAT001 ed 1.4 §4.2.2).
//
// Record.Decode asks SelectDecodeUAP for the layout before reading the items;
// Record.Encode asks SelectEncodeUAP and rebuilds the FSPEC against the
// selected layout, so FRNs are always those of the selected UAP.
type RecordUAPSelector interface {
	UAP
	// SelectDecodeUAP returns the UAP for the record whose encoding starts at
	// data[0] (the first FSPEC octet). It must not retain data.
	SelectDecodeUAP(data []byte) (UAP, error)
	// SelectEncodeUAP returns the UAP for a record holding these items.
	SelectEncodeUAP(items map[string]DataItem) (UAP, error)
}

// encodeUAP returns the UAP that lays out this record when it is encoded.
func (r *Record) encodeUAP() (UAP, error) {
	if sel, ok := r.uap.(RecordUAPSelector); ok {
		return sel.SelectEncodeUAP(r.items)
	}
	return r.uap, nil
}

// Encode writes the record to a buffer
func (r *Record) Encode(buf *bytes.Buffer) (int, error) {
	uap, err := r.encodeUAP()
	if err != nil {
		return 0, err
	}

	// Validate the record against UAP rules
	if err := uap.Validate(r.items); err != nil {
		return 0, err
	}

	if uap != r.uap {
		// FRNs depend on the selected layout: rebuild the FSPEC against it.
		fspec := NewFSPEC()
		for id := range r.items {
			field, ok := uap.GetFieldByDataItem(id)
			if !ok {
				return 0, NewEncodingError(r.category, id,
					"item not defined in the UAP selected for this record", ErrUnknownDataItem)
			}
			if err := fspec.SetFRN(field.FRN); err != nil {
				return 0, err
			}
		}
		r.fspec = fspec
	}

	bytesWritten := 0

	// Write FSPEC
	n, err := r.fspec.Encode(buf)
	if err != nil {
		return bytesWritten, fmt.Errorf("encoding FSPEC: %w", err)
	}
	bytesWritten += n

	// Write items in FRN order
	for _, field := range uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		item, exists := r.items[field.DataItem]
		if !exists {
			return bytesWritten, NewEncodingError(
				r.category,
				field.DataItem,
				fmt.Sprintf("FRN %d marked in FSPEC but item not present", field.FRN),
				ErrInvalidMessage,
			)
		}

		n, err := item.Encode(buf)
		if err != nil {
			return bytesWritten, NewEncodingError(
				r.category,
				field.DataItem,
				"failed to encode item",
				err,
			)
		}
		bytesWritten += n
	}

	return bytesWritten, nil
}

// Decode reads a record from a buffer
func (r *Record) Decode(buf *bytes.Buffer) (int, error) {
	if buf.Len() == 0 {
		return 0, io.EOF
	}

	bytesRead := 0

	uap := r.uap
	if sel, ok := r.uap.(RecordUAPSelector); ok {
		selected, err := sel.SelectDecodeUAP(buf.Bytes())
		if err != nil {
			return 0, NewDecodeError(r.category, "", "selecting record UAP", err)
		}
		uap = selected
	}

	// Read FSPEC
	n, err := r.fspec.Decode(buf)
	if err != nil {
		return bytesRead, NewDecodeError(
			r.category,
			"",
			"decoding FSPEC",
			err,
		)
	}
	bytesRead += n

	// Clear existing items (reuse map allocation)
	for k := range r.items {
		delete(r.items, k)
	}

	// Every FSPEC bit must name something this UAP can read. A bit for an
	// FRN without a data item (spare, beyond the UAP, or an RS bit the UAP
	// does not support) stands for data of unknown length: skipping it
	// would shift every later item, so the record is rejected.
	var rfsSent bool
	if l, ok := uap.(layoutUAP); ok {
		n, rfsSent, err = r.decodeItemsByLayout(buf, uap, l.decodeLayout())
	} else {
		n, rfsSent, err = r.decodeItemsByFields(buf, uap)
	}
	bytesRead += n
	if err != nil {
		return bytesRead, err
	}
	if rfsSent {
		if err := r.standardFSPEC(uap); err != nil {
			return bytesRead, err
		}
	}

	// Validate the record
	if err := uap.Validate(r.items); err != nil {
		return bytesRead, err
	}

	return bytesRead, nil
}

// decodeItemsByLayout reads the items of the set FSPEC bits in FRN order,
// using the precomputed layout of a BaseUAP. It reports whether an RFS
// field was read.
func (r *Record) decodeItemsByLayout(buf *bytes.Buffer, uap UAP, l *decodeLayout) (int, bool, error) {
	fields := uap.Fields()
	read := 0
	rfsSent := false
	for i, b := range r.fspec.data() {
		b &= 0xFE // FX
		for b != 0 {
			bit := bits.LeadingZeros8(b)
			b &^= 0x80 >> bit
			frn := uint8(i*7 + bit + 1)
			if frn == l.rfsFRN {
				n, err := r.decodeRFS(buf, uap, frn, read)
				read += n
				if err != nil {
					return read, rfsSent, err
				}
				rfsSent = true
				continue
			}
			idx := l.field[frn]
			if idx < 0 {
				return read, rfsSent, NewDecodeError(r.category, "", "checking FSPEC against the UAP", unknownFRNError(frn))
			}
			n, err := r.decodeField(buf, uap, &fields[idx], read)
			read += n
			if err != nil {
				return read, rfsSent, err
			}
		}
	}
	return read, rfsSent, nil
}

// decodeItemsByFields reads the items of any UAP in the order of its
// Fields, after checking that every set FSPEC bit has a data item.
func (r *Record) decodeItemsByFields(buf *bytes.Buffer, uap UAP) (int, bool, error) {
	rfsFRN := uapRFSFRN(uap)
	if err := checkFSPEC(uap, r.fspec, rfsFRN); err != nil {
		return 0, false, NewDecodeError(r.category, "", "checking FSPEC against the UAP", err)
	}
	rfsPending := rfsFRN != 0 && r.fspec.GetFRN(rfsFRN)
	rfsSent := rfsPending
	read := 0
	fields := uap.Fields()
	for i := range fields {
		field := &fields[i]
		if rfsPending && field.FRN > rfsFRN {
			n, err := r.decodeRFS(buf, uap, rfsFRN, read)
			read += n
			if err != nil {
				return read, rfsSent, err
			}
			rfsPending = false
		}
		if field.FRN == rfsFRN || !r.fspec.GetFRN(field.FRN) {
			continue
		}
		n, err := r.decodeField(buf, uap, field, read)
		read += n
		if err != nil {
			return read, rfsSent, err
		}
	}
	if rfsPending {
		n, err := r.decodeRFS(buf, uap, rfsFRN, read)
		read += n
		if err != nil {
			return read, rfsSent, err
		}
	}
	return read, rfsSent, nil
}

// decodeField reads the data item of field into r.items. pos is the
// record offset of the item, for error reporting.
func (r *Record) decodeField(buf *bytes.Buffer, uap UAP, field *DataField, pos int) (int, error) {
	item, err := uap.CreateDataItem(field.DataItem)
	if err != nil {
		// A fixed-length item the UAP lists but cannot build is skipped by
		// its length from the UAP, so later items stay aligned.
		if field.Type == Fixed {
			if buf.Len() < int(field.Length) {
				return 0, NewDecodeError(
					r.category,
					field.DataItem,
					fmt.Sprintf("buffer too short to skip field: need %d bytes, have %d", field.Length, buf.Len()),
					ErrBufferTooShort,
				)
			}
			buf.Next(int(field.Length))
			return int(field.Length), nil
		}
		return 0, NewDecodeError(r.category, field.DataItem, "creating data item", err)
	}

	n, err := item.Decode(buf)
	if err != nil {
		return 0, NewDecodeError(
			r.category,
			field.DataItem,
			"decoding data item",
			err,
		).WithPosition(pos, buf.Len()+pos)
	}
	r.items[field.DataItem] = item
	return n, nil
}

// decodeRFS reads a Random Field Sequencing organised field: one octet N,
// then N times a one-octet FRN followed by the data item of that FRN
// (CAT001 ed 1.4 §6, CAT002 ed 1.2 §6). The items join r.items as if they
// had been sent in FSPEC order; Decode then rebuilds the FSPEC to that
// standard form (RS bit cleared), so encoding the record writes no RFS field.
func (r *Record) decodeRFS(buf *bytes.Buffer, uap UAP, rfsFRN uint8, pos int) (int, error) {
	read := 0
	count, err := buf.ReadByte()
	if err != nil {
		return read, NewDecodeError(r.category, "RFS", "reading RFS field count", ErrBufferTooShort)
	}
	read++
	for i := 0; i < int(count); i++ {
		frn, err := buf.ReadByte()
		if err != nil {
			return read, NewDecodeError(r.category, "RFS",
				fmt.Sprintf("reading FRN of RFS field %d of %d", i+1, count), ErrBufferTooShort)
		}
		read++
		field, ok := uap.GetFieldByFRN(frn)
		if !ok || field.DataItem == "" || frn == rfsFRN {
			return read, NewDecodeError(r.category, "RFS",
				fmt.Sprintf("RFS field %d names FRN %d, which has no data item in the UAP", i+1, frn),
				ErrInvalidFSPEC)
		}
		if _, dup := r.items[field.DataItem]; dup || r.fspec.GetFRN(frn) {
			return read, NewDecodeError(r.category, field.DataItem,
				fmt.Sprintf("FRN %d sent twice (FSPEC or RFS field)", frn), ErrInvalidFSPEC)
		}
		n, err := r.decodeField(buf, uap, &field, pos+read)
		read += n
		if err != nil {
			return read, err
		}
	}

	return read, nil
}

// standardFSPEC rebuilds r.fspec from the items, against uap: the
// standard form of a record whose items arrived (partly) in an RFS field.
func (r *Record) standardFSPEC(uap UAP) error {
	fspec := NewFSPEC()
	for id := range r.items {
		f, ok := uap.GetFieldByDataItem(id)
		if !ok {
			return NewDecodeError(r.category, id, "rebuilding FSPEC after RFS", ErrUnknownDataItem)
		}
		if err := fspec.SetFRN(f.FRN); err != nil {
			return NewDecodeError(r.category, id, "rebuilding FSPEC after RFS", err)
		}
	}
	r.fspec = fspec
	return nil
}

// uapRFSFRN returns the FRN of the RS bit of uap, or 0.
func uapRFSFRN(uap UAP) uint8 {
	if rfs, ok := uap.(RandomFieldSequencer); ok {
		return rfs.RFSFRN()
	}
	return 0
}

// checkFSPEC returns an ErrInvalidFSPEC error naming the first FRN that is
// set in f but has no data item in uap and is not its RS bit rfsFRN.
func checkFSPEC(uap UAP, f *FSPEC, rfsFRN uint8) error {
	for i, b := range f.data() {
		for bit := 0; bit < 7; bit++ {
			if b&(0x80>>bit) == 0 {
				continue
			}
			frn := uint8(i*7 + bit + 1)
			if frn == rfsFRN {
				continue
			}
			if field, ok := uap.GetFieldByFRN(frn); !ok || field.DataItem == "" {
				return unknownFRNError(frn)
			}
		}
	}
	return nil
}

func unknownFRNError(frn uint8) error {
	return fmt.Errorf("%w: FRN %d is set but has no data item in the UAP (spare, reserved or not supported)",
		ErrInvalidFSPEC, frn)
}

// Category returns the category of this record
func (r *Record) Category() Category {
	return r.category
}

// UAP returns the UAP used by this record
func (r *Record) UAP() UAP {
	return r.uap
}

// FSPEC returns the field specification of this record
func (r *Record) FSPEC() *FSPEC {
	return r.fspec
}

// Items returns a map of all data items in this record
// WARNING: The returned map contains pointers to the original data items.
// Modifying the items will affect the record. This is intentional for performance.
// If you need a deep copy, use Clone() instead.
func (r *Record) Items() map[string]DataItem {
	return r.items
}

// ItemCount returns the number of data items in this record
func (r *Record) ItemCount() int {
	return len(r.items)
}

// HasDataItem checks if a specific data item is present
func (r *Record) HasDataItem(id string) bool {
	_, exists := r.items[id]
	return exists
}

// EstimateSize estimates the encoded size of this record in bytes
func (r *Record) EstimateSize() int {
	size := r.fspec.Size() // FSPEC size

	uap, err := r.encodeUAP()
	if err != nil {
		uap = r.uap
	}

	// Add size of each data item
	for _, field := range uap.Fields() {
		if !r.fspec.GetFRN(field.FRN) {
			continue
		}

		if field.Type == Fixed {
			size += int(field.Length)
			continue
		}

		// For non-fixed items, we need a rough estimate
		// This is just an estimate, the actual size may differ
		size += 4 // Reasonable default size
	}

	return size
}

// Clone creates a deep copy of this record
//
// PERFORMANCE NOTE: This uses encode/decode for cloning, which is slower than
// field-by-field copy but guarantees correctness. To optimize, each DataItem
// implementation would need its own Clone() method. For most use cases, the
// current implementation is sufficient. If you need high-performance cloning,
// consider caching decoded records or using copy-on-write patterns.
func (r *Record) Clone() (*Record, error) {
	newRecord, err := NewRecord(r.category, r.uap)
	if err != nil {
		return nil, err
	}

	// Encode to buffer and decode to create a deep copy
	// This ensures all data items are properly deep copied
	buf := new(bytes.Buffer)
	_, err = r.Encode(buf)
	if err != nil {
		return nil, err
	}

	_, err = newRecord.Decode(buf)
	if err != nil {
		return nil, err
	}

	return newRecord, nil
}

// Reset clears all data items but keeps the category and UAP
func (r *Record) Reset() {
	r.fspec.Reset()
	// Clear existing items (reuse map allocation)
	for k := range r.items {
		delete(r.items, k)
	}
}

// Validate checks if the record is valid according to the UAP
func (r *Record) Validate() error {
	return r.uap.Validate(r.items)
}
