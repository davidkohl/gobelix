// asterix/fspec.go
package asterix

import (
	"fmt"
	"io"
)

// FSPEC represents the Field Specification of an ASTERIX record.
// It efficiently stores the presence bits for data fields and provides
// fast operations for field presence checking.
//
// Performance optimization: Uses inline storage for 1-2 bytes (most common case)
// to avoid heap allocation. Most ASTERIX messages have < 14 fields (fit in 2 bytes).
//
// Thread Safety: FSPEC is NOT safe for concurrent use.
// Each FSPEC instance should be accessed by only one goroutine at a time.
// Methods that modify the FSPEC (SetFRN, Decode, Reset) should not be called
// concurrently with any other methods on the same instance.
type FSPEC struct {
	inline [2]byte // Inline storage for common case (1-2 bytes)
	heap   []byte  // Heap storage for larger FSPECs (3+ bytes)
	size   int     // Number of valid bytes
}

// NewFSPEC creates a new empty FSPEC
func NewFSPEC() *FSPEC {
	return &FSPEC{
		size: 0,
		// inline and heap are zero-initialized
	}
}

// data returns a slice pointing to the current storage (inline or heap)
func (f *FSPEC) data() []byte {
	if f.size <= 2 {
		return f.inline[:f.size]
	}
	return f.heap[:f.size]
}

// dataPtr returns a pointer to the byte at index i
// IMPORTANT: Only call this after ensureCapacity(i+1)
func (f *FSPEC) dataPtr(i int) *byte {
	if i < 2 && f.heap == nil {
		// Still using inline storage
		return &f.inline[i]
	}
	// Using heap storage
	return &f.heap[i]
}

// ensureCapacity ensures we have capacity for n bytes
func (f *FSPEC) ensureCapacity(n int) {
	if n <= 2 {
		// Use inline storage
		return
	}

	// Need heap storage
	if f.heap == nil {
		// First time switching to heap - copy inline data
		f.heap = make([]byte, n)
		if f.size > 0 {
			copy(f.heap, f.inline[:f.size])
		}
	} else if len(f.heap) < n {
		// Grow heap storage
		newHeap := make([]byte, n)
		copy(newHeap, f.heap)
		f.heap = newHeap
	}
}

// SetFRN marks a Field Reference Number as present in the FSPEC
func (f *FSPEC) SetFRN(frn uint8) error {
	if frn == 0 {
		return fmt.Errorf("invalid FRN: FRN cannot be 0")
	}

	byteIndex := (frn - 1) / 7 // 7 bits per byte (last bit is FX)
	bitPosition := (frn - 1) % 7

	// Ensure we have capacity for this byte
	f.ensureCapacity(int(byteIndex) + 1)

	// Ensure capacity for this byte and set previous FX bits
	if int(byteIndex) >= f.size {
		// Set FX bit in all preceding bytes
		for i := f.size; i < int(byteIndex); i++ {
			*f.dataPtr(i) |= 0x01 // Set FX bit
		}
		f.size = int(byteIndex) + 1
	}

	// Set the specific bit
	*f.dataPtr(int(byteIndex)) |= 0x80 >> bitPosition

	// Ensure FX bits are set in all but the last byte
	for i := 0; i < f.size-1; i++ {
		*f.dataPtr(i) |= 0x01
	}

	return nil
}

// GetFRN checks if a Field Reference Number is present
func (f *FSPEC) GetFRN(frn uint8) bool {
	if frn == 0 || f.size == 0 {
		return false
	}

	byteIndex := (frn - 1) / 7
	bitPosition := (frn - 1) % 7

	if int(byteIndex) >= f.size {
		return false
	}

	return (*f.dataPtr(int(byteIndex)) & (0x80 >> bitPosition)) != 0
}

// Encode writes the FSPEC to an io.Writer
func (f *FSPEC) Encode(w io.Writer) (int, error) {
	if f.size == 0 {
		return 0, fmt.Errorf("invalid FSPEC: no bits set")
	}

	return w.Write(f.data())
}

// Decode reads the FSPEC from an io.Reader
func (f *FSPEC) Decode(r io.Reader) (int, error) {
	f.size = 0

	// Read the first byte into inline storage
	if _, err := io.ReadFull(r, f.inline[:1]); err != nil {
		return 0, fmt.Errorf("reading FSPEC: %w", err)
	}
	f.size = 1

	// Read extension bytes as needed
	for *f.dataPtr(f.size-1)&0x01 != 0 {
		// Safety check to prevent malformed data causing excessive reads
		if f.size >= 8 {
			return f.size, fmt.Errorf("invalid FSPEC: too many extension bytes")
		}

		// Ensure capacity for next byte
		f.ensureCapacity(f.size + 1)

		// Read the next byte
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return f.size, fmt.Errorf("reading FSPEC extension: %w", err)
		}
		*f.dataPtr(f.size) = b[0]
		f.size++
	}

	return f.size, nil
}

// DecodeFromBytes decodes an FSPEC from a byte slice starting at the specified offset.
// Returns the number of bytes read and any error encountered.
func (f *FSPEC) DecodeFromBytes(data []byte, offset int) (int, error) {
	if offset >= len(data) {
		return 0, io.EOF
	}

	f.size = 0

	// Read the first byte
	f.inline[0] = data[offset]
	f.size = 1
	bytesRead := 1

	// Read extension bytes as needed
	for *f.dataPtr(f.size-1)&0x01 != 0 {
		// Check if we have more data
		if offset+bytesRead >= len(data) {
			return bytesRead, fmt.Errorf("unexpected end of data in FSPEC")
		}

		// Safety check to prevent malformed data causing excessive reads
		if f.size >= 8 {
			return bytesRead, fmt.Errorf("invalid FSPEC: too many extension bytes")
		}

		// Ensure capacity
		f.ensureCapacity(f.size + 1)

		// Get the next byte
		*f.dataPtr(f.size) = data[offset+bytesRead]
		f.size++
		bytesRead++

		// Stop if this byte doesn't have the FX bit set
		if *f.dataPtr(f.size-1)&0x01 == 0 {
			break
		}
	}

	return bytesRead, nil
}

// EncodeToBytes encodes the FSPEC to a byte slice starting at the specified offset.
// The slice must have sufficient capacity. Returns the number of bytes written.
func (f *FSPEC) EncodeToBytes(data []byte, offset int) (int, error) {
	if f.size == 0 {
		return 0, fmt.Errorf("invalid FSPEC: no bits set")
	}

	if offset+f.size > len(data) {
		return 0, fmt.Errorf("buffer too small for FSPEC")
	}

	copy(data[offset:], f.data())
	return f.size, nil
}

// Size returns the size of the FSPEC in bytes
func (f *FSPEC) Size() int {
	return f.size
}

// Reset resets the FSPEC to empty state, keeping allocated memory
func (f *FSPEC) Reset() {
	// Clear inline storage
	f.inline[0] = 0
	f.inline[1] = 0

	// Clear heap storage if used (keep allocation)
	if f.heap != nil {
		for i := range f.heap {
			f.heap[i] = 0
		}
	}

	f.size = 0
}

// Copy creates a copy of the FSPEC
func (f *FSPEC) Copy() *FSPEC {
	newFSPEC := &FSPEC{
		size: f.size,
	}

	// Copy inline storage
	newFSPEC.inline = f.inline

	// Copy heap storage if present
	if f.heap != nil {
		newFSPEC.heap = make([]byte, len(f.heap))
		copy(newFSPEC.heap, f.heap)
	}

	return newFSPEC
}

// FSPECBitCount returns the number of data bits set (excluding FX bits)
func (f *FSPEC) FSPECBitCount() int {
	count := 0
	for i := 0; i < f.size; i++ {
		// Check bits 8-2 (exclude FX bit)
		b := *f.dataPtr(i)
		for j := uint(0); j < 7; j++ {
			if b&(0x80>>j) != 0 {
				count++
			}
		}
	}
	return count
}

// HasDataBits returns true if any data bits are set (excluding FX bits)
func (f *FSPEC) HasDataBits() bool {
	for i := 0; i < f.size; i++ {
		// Check if any of bits 8-2 are set (exclude FX bit)
		if *f.dataPtr(i)&0xFE != 0 {
			return true
		}
	}
	return false
}

// String returns a string representation of the FSPEC
func (f *FSPEC) String() string {
	if f.size == 0 {
		return "FSPEC{empty}"
	}

	return fmt.Sprintf("FSPEC{size:%d, bits:%08b}", f.size, f.data())
}

// FSPECFromUint64 creates an FSPEC from a uint64 value for simple test cases
// The uint64 is interpreted as a bit field where bit 0 represents FRN 1, etc.
func FSPECFromUint64(bits uint64) *FSPEC {
	fspec := NewFSPEC()

	// Set each bit that is 1 in the input
	for i := uint8(1); i <= 64; i++ {
		if bits&(1<<(i-1)) != 0 {
			fspec.SetFRN(i)
		}
	}

	return fspec
}

// ToUint64 converts the FSPEC to a uint64 bit field for simple test cases
// Only works for FSPECs with fields up to FRN 64
func (f *FSPEC) ToUint64() uint64 {
	var result uint64

	// For each possible FRN up to 64
	for i := uint8(1); i <= 64; i++ {
		if f.GetFRN(i) {
			result |= 1 << (i - 1)
		}
	}

	return result
}
