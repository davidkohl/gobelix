// cat/cat001/uap/uap_v12.go
package uap

import (
	"github.com/davidkohl/gobelix/asterix"
	v12 "github.com/davidkohl/gobelix/cat/cat001/dataitems/v12"
)

// UAP12 is the CAT001 v1.2 (ed 1.4) UAP that selects the plot or the track
// layout per record from I001/020 TYP, as ed 1.4 §4.2.2 and §5.3.2.2
// require ("TYP ... to apply the proper UAP to decode the remaining
// information within the record").
//
// It implements asterix.RecordUAPSelector: Record.Decode reads TYP from the
// record bytes before decoding the items, and Record.Encode lays the record
// out with the plot or track UAP according to the TYP of the I001/020 item
// (plot when I001/020 is absent). Fields and FRN lookups of UAP12 itself are
// those of the plot UAP; GetFieldByDataItem also accepts track-only items so
// they can be set on a record.
type UAP12 struct {
	*asterix.BaseUAP
	plot  *UAP12Plot
	track *UAP12Track
}

// NewUAP12 creates the CAT001 v1.2 UAP that selects plot or track per
// record from I001/020 TYP. Use NewUAP12Plot or NewUAP12Track to force one
// layout.
func NewUAP12() (*UAP12, error) {
	plotUAP, err := NewUAP12Plot()
	if err != nil {
		return nil, err
	}
	trackUAP, err := NewUAP12Track()
	if err != nil {
		return nil, err
	}
	// SelectDecodeUAP reads TYP from I001/020 at its FSPEC position (FRN 2).
	// Sent inside the RFS field, I001/020 would come too late to select the
	// layout, so these selector-owned instances refuse it there (as well as
	// SP, FRN 20). I001/020 is compulsory (ed 1.4 §5.3.2.2, §5.4.2.2).
	for _, b := range []*asterix.BaseUAP{plotUAP.BaseUAP, trackUAP.BaseUAP} {
		if err := b.EnableRFS(21, 20, 2); err != nil {
			return nil, err
		}
	}
	return &UAP12{BaseUAP: plotUAP.BaseUAP, plot: plotUAP, track: trackUAP}, nil
}

// SelectDecodeUAP implements asterix.RecordUAPSelector. data starts at the
// record FSPEC. I001/020 is FRN 2 in both layouts, preceded only by
// I001/010 (FRN 1, two octets), so TYP can be read before decoding.
func (u *UAP12) SelectDecodeUAP(data []byte) (asterix.UAP, error) {
	if len(data) == 0 {
		return u.plot, nil
	}
	off := 0
	for off < len(data) && data[off]&0x01 != 0 {
		off++
	}
	off++ // last FSPEC octet
	if data[0]&0x40 == 0 {
		// No I001/020: compulsory per spec; fall back to the plot layout.
		// An I001/020 in the RFS field is refused there (see NewUAP12).
		return u.plot, nil
	}
	if data[0]&0x80 != 0 {
		off += 2 // I001/010
	}
	if off >= len(data) {
		// Truncated record: the plot decoder reports the short buffer.
		return u.plot, nil
	}
	if data[off]&0x80 != 0 {
		return u.track, nil
	}
	return u.plot, nil
}

// SelectEncodeUAP implements asterix.RecordUAPSelector.
func (u *UAP12) SelectEncodeUAP(items map[string]asterix.DataItem) (asterix.UAP, error) {
	if trd, ok := items["I001/020"].(*v12.TargetReportDescriptor); ok && trd.TYP == v12.TYPTrack {
		return u.track, nil
	}
	return u.plot, nil
}

// GetFieldByDataItem returns the plot field for id, or the track field for
// a track-only item.
func (u *UAP12) GetFieldByDataItem(id string) (asterix.DataField, bool) {
	if f, ok := u.plot.GetFieldByDataItem(id); ok {
		return f, true
	}
	return u.track.GetFieldByDataItem(id)
}

// CreateDataItem creates a new instance of a Cat001 data item.
// Supports all data items from both Plot and Track UAPs.
func (u *UAP12) CreateDataItem(id string) (asterix.DataItem, error) {
	return u.track.CreateDataItem(id)
}

// Validate implements validations for Cat001
func (u *UAP12) Validate(items map[string]asterix.DataItem) error {
	return u.BaseUAP.Validate(items)
}
