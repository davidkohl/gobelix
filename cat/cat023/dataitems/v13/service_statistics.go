// cat/cat023/dataitems/v13/service_statistics.go
package v13

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davidkohl/gobelix/asterix"
)

// ServiceStatistics represents I023/120 - Service Statistics (CAT023 ed 1.3
// §5.2.8). Repetitive item: REP, then REP blocks of six octets:
//
//	TYPE (8 bits) | REF (1 bit) spare (7 bits) | COUNTER VALUE (32 bits)
type ServiceStatistics struct {
	// Entries holds one counter block per repetition.
	Entries []ServiceStatisticsEntry

	// Deprecated: TYPE, REF and CV are the pre-ed-1.3-layout view kept for
	// compatibility. After Decode, TYPE and REF are those of the first entry
	// and CV holds every counter. Encode uses them only when Entries is
	// empty: one block per CV, each with TYPE and REF. When both views are
	// set they must agree (TYPE and REF those of Entries[0], CV one counter
	// per entry), or Validate and Encode return an error.
	TYPE *TYPEValue
	REF  *REFValue
	CV   []CounterValue
}

// ServiceStatisticsEntry is one I023/120 counter block.
type ServiceStatisticsEntry struct {
	// Type of report counter: 0 unknown messages, 1 'too old' messages,
	// 2 failed conversions, 3 total received, 4 total transmitted,
	// 20 TIS-B management, 21 Basic, 22 High Dynamic, 23 Full Position,
	// 24 Basic Ground, 25 TCP, 26 UTC time, 27 Data, 28 High Resolution,
	// 29 Aircraft Target Airborne, 30 Aircraft Target Ground,
	// 31 Ground Vehicle Target, 32 2-slots TCP.
	Type uint8
	// Ref is the counting reference: false from midnight, true from the
	// last report.
	Ref bool
	// Counter is the 32-bit counter value.
	Counter uint32
}

// TYPEValue is the TYPE of the first counter block (deprecated view).
type TYPEValue struct {
	TYPE uint8
}

// REFValue is the REF of the first counter block (deprecated view).
type REFValue struct {
	REF bool // false: from midnight, true: from the last report
}

// CounterValue is one counter value (deprecated view).
type CounterValue struct {
	Counter uint32
}

// Decode decodes the Service Statistics from bytes
func (s *ServiceStatistics) Decode(buf *bytes.Buffer) (int, error) {
	*s = ServiceStatistics{}
	rep, err := buf.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("%w: reading I023/120 REP", asterix.ErrBufferTooShort)
	}
	n := 1
	if rep == 0 {
		// "followed by at least one block of 6 octets" (ed 1.3 §5.2.8)
		return n, fmt.Errorf("%w: I023/120 REP is 0, at least one counter is required", asterix.ErrInvalidMessage)
	}
	if buf.Len() < int(rep)*6 {
		return n, fmt.Errorf("%w: I023/120 REP %d needs %d octets, have %d", asterix.ErrBufferTooShort, rep, int(rep)*6, buf.Len())
	}
	s.Entries = make([]ServiceStatisticsEntry, rep)
	s.CV = make([]CounterValue, rep)
	for i := range s.Entries {
		b := buf.Next(6)
		n += 6
		e := ServiceStatisticsEntry{
			Type:    b[0],
			Ref:     b[1]&0x80 != 0,
			Counter: uint32(b[2])<<24 | uint32(b[3])<<16 | uint32(b[4])<<8 | uint32(b[5]),
		}
		s.Entries[i] = e
		s.CV[i] = CounterValue{Counter: e.Counter}
	}
	if rep > 0 {
		s.TYPE = &TYPEValue{TYPE: s.Entries[0].Type}
		s.REF = &REFValue{REF: s.Entries[0].Ref}
	}
	return n, nil
}

// entries returns the blocks to encode.
func (s *ServiceStatistics) entries() []ServiceStatisticsEntry {
	if len(s.Entries) > 0 {
		return s.Entries
	}
	out := make([]ServiceStatisticsEntry, len(s.CV))
	for i, cv := range s.CV {
		out[i].Counter = cv.Counter
		if s.TYPE != nil {
			out[i].Type = s.TYPE.TYPE
		}
		if s.REF != nil {
			out[i].Ref = s.REF.REF
		}
	}
	return out
}

// Encode encodes the Service Statistics to bytes
func (s *ServiceStatistics) Encode(buf *bytes.Buffer) (int, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	entries := s.entries()
	out := make([]byte, 1, 1+6*len(entries))
	out[0] = byte(len(entries))
	for _, e := range entries {
		var ref byte
		if e.Ref {
			ref = 0x80
		}
		out = append(out, e.Type, ref,
			byte(e.Counter>>24), byte(e.Counter>>16), byte(e.Counter>>8), byte(e.Counter))
	}
	n, err := buf.Write(out)
	if err != nil {
		return n, fmt.Errorf("writing I023/120: %w", err)
	}
	return n, nil
}

// Validate validates the Service Statistics
func (s *ServiceStatistics) Validate() error {
	n := len(s.entries())
	if n == 0 {
		return fmt.Errorf("%w: I023/120 needs at least one counter (ed 1.3 §5.2.8)", asterix.ErrInvalidMessage)
	}
	if n > 255 {
		return fmt.Errorf("%w: I023/120 holds at most 255 counters, got %d", asterix.ErrInvalidMessage, n)
	}
	if len(s.Entries) > 0 && !s.legacyAgrees() {
		return fmt.Errorf("%w: I023/120 deprecated TYPE/REF/CV disagree with Entries", asterix.ErrInvalidMessage)
	}
	return nil
}

// legacyAgrees reports whether the deprecated view, where set, matches
// Entries as Decode would fill it.
func (s *ServiceStatistics) legacyAgrees() bool {
	if s.TYPE != nil && s.TYPE.TYPE != s.Entries[0].Type {
		return false
	}
	if s.REF != nil && s.REF.REF != s.Entries[0].Ref {
		return false
	}
	if len(s.CV) == 0 {
		return true
	}
	if len(s.CV) != len(s.Entries) {
		return false
	}
	for i, cv := range s.CV {
		if cv.Counter != s.Entries[i].Counter {
			return false
		}
	}
	return true
}

// String returns a string representation
func (s *ServiceStatistics) String() string {
	entries := s.entries()
	if len(entries) == 0 {
		return "ServiceStatistics{}"
	}
	parts := make([]string, len(entries))
	for i, e := range entries {
		ref := "midnight"
		if e.Ref {
			ref = "last report"
		}
		parts[i] = fmt.Sprintf("type %d since %s: %d", e.Type, ref, e.Counter)
	}
	return fmt.Sprintf("ServiceStatistics{%s}", strings.Join(parts, ", "))
}
