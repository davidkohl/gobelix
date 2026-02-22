// cat/common/dataitems/icao_charset.go
package common

// SixBitToASCII implements the ICAO Annex 10 Vol IV character set mapping
// Used for encoding/decoding aircraft identification (callsign) in ASTERIX
// Each 6-bit code (0-63) maps to a character in this 64-character array
// Space (0x20) represents code 0 and undefined/reserved codes
var SixBitToASCII = [64]byte{
	' ', // 0: Space
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', // 1-13
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', // 14-26
	' ', ' ', ' ', ' ', ' ', // 27-31: Reserved
	' ', // 32: Space
	' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', // 33-47: Reserved
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', // 48-57
	' ', ' ', ' ', ' ', ' ', ' ', // 58-63: Reserved
}

// ASCIIToSixBit maps ASCII characters to 6-bit ICAO codes
// Returns 0 (space) for invalid characters
var ASCIIToSixBit = func() [128]byte {
	var table [128]byte
	// Initialize all to space (0)
	for i := range table {
		table[i] = 0
	}
	// Map valid characters
	for code, char := range SixBitToASCII {
		if char != ' ' || code == 0 || code == 32 {
			table[char] = byte(code)
		}
	}
	// Map space explicitly
	table[' '] = 32
	return table
}()

// DecodeICAOChar converts a 6-bit code to ASCII character
func DecodeICAOChar(code byte) byte {
	if code > 63 {
		return ' '
	}
	return SixBitToASCII[code]
}

// EncodeICAOChar converts an ASCII character to 6-bit code
func EncodeICAOChar(char byte) byte {
	if char > 127 {
		return 32 // Space for invalid chars
	}
	return ASCIIToSixBit[char]
}

// DecodeICAOString decodes 8 characters from 6 bytes (48 bits)
// This is the format used in CAT021/170 (Target Identification)
func DecodeICAOString6Bytes(data []byte) string {
	if len(data) < 6 {
		return ""
	}

	var chars [8]byte
	chars[0] = DecodeICAOChar((data[0] & 0xFC) >> 2)
	chars[1] = DecodeICAOChar(((data[0] & 0x03) << 4) | ((data[1] & 0xF0) >> 4))
	chars[2] = DecodeICAOChar(((data[1] & 0x0F) << 2) | ((data[2] & 0xC0) >> 6))
	chars[3] = DecodeICAOChar(data[2] & 0x3F)
	chars[4] = DecodeICAOChar((data[3] & 0xFC) >> 2)
	chars[5] = DecodeICAOChar(((data[3] & 0x03) << 4) | ((data[4] & 0xF0) >> 4))
	chars[6] = DecodeICAOChar(((data[4] & 0x0F) << 2) | ((data[5] & 0xC0) >> 6))
	chars[7] = DecodeICAOChar(data[5] & 0x3F)

	// Trim trailing spaces
	end := 8
	for end > 0 && chars[end-1] == ' ' {
		end--
	}

	return string(chars[:end])
}

// EncodeICAOString6Bytes encodes up to 8 characters into 6 bytes (48 bits)
func EncodeICAOString6Bytes(s string) []byte {
	// Pad to 8 characters with spaces
	padded := make([]byte, 8)
	for i := range padded {
		padded[i] = ' '
	}
	copy(padded, s)

	// Convert to 6-bit codes
	var codes [8]byte
	for i, c := range padded {
		codes[i] = EncodeICAOChar(c)
	}

	// Pack into 6 bytes
	data := make([]byte, 6)
	data[0] = (codes[0] << 2) | (codes[1] >> 4)
	data[1] = (codes[1] << 4) | (codes[2] >> 2)
	data[2] = (codes[2] << 6) | codes[3]
	data[3] = (codes[4] << 2) | (codes[5] >> 4)
	data[4] = (codes[5] << 4) | (codes[6] >> 2)
	data[5] = (codes[6] << 6) | codes[7]

	return data
}
