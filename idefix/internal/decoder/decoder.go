// internal/decoder/decoder.go
package decoder

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat001"
	"github.com/davidkohl/gobelix/cat/cat002"
	"github.com/davidkohl/gobelix/cat/cat021"
	"github.com/davidkohl/gobelix/cat/cat034"
	"github.com/davidkohl/gobelix/cat/cat048"
	"github.com/davidkohl/gobelix/cat/cat062"
	"github.com/davidkohl/gobelix/cat/cat063"
	"github.com/davidkohl/gobelix/encoding"
)

// Config represents decoder configuration options
type Config struct {
	DumpAll    bool
	DumpCat001 bool
	DumpCat002 bool
	DumpCat021 bool
	DumpCat034 bool
	DumpCat048 bool
	DumpCat062 bool
	DumpCat063 bool
}

// CreateDecoder creates and configures a decoder with the specified UAPs
func CreateDecoder(config Config) (*asterix.Decoder, error) {
	// Initialize the default buffer pool if it doesn't exist
	if encoding.DefaultBufferPool == nil {
		encoding.DefaultBufferPool = encoding.NewBufferPool()
	}

	// Create a decoder without any special options
	// This avoids potential issues with nil buffer pools
	decoder := asterix.NewDecoder()

	var uaps []asterix.UAP

	if config.DumpAll || config.DumpCat001 {
		uap001, err := cat001.NewUAP(cat001.Version12)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat001 UAP: %w", err)
		}
		decoder.RegisterUAP(uap001)
		uaps = append(uaps, uap001)
	}

	if config.DumpAll || config.DumpCat002 {
		uap002, err := cat002.NewUAP(cat002.Version10)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat002 UAP: %w", err)
		}
		decoder.RegisterUAP(uap002)
		uaps = append(uaps, uap002)
	}

	if config.DumpAll || config.DumpCat021 {
		uap021, err := cat021.NewUAP(cat021.Version26)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat021 UAP: %w", err)
		}
		decoder.RegisterUAP(uap021)
		uaps = append(uaps, uap021)
	}

	if config.DumpAll || config.DumpCat034 {
		uap034, err := cat034.NewUAP(cat034.Version129)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat034 UAP: %w", err)
		}
		decoder.RegisterUAP(uap034)
		uaps = append(uaps, uap034)
	}

	if config.DumpAll || config.DumpCat048 {
		uap048, err := cat048.NewUAP("1.32")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat048 UAP: %w", err)
		}
		decoder.RegisterUAP(uap048)
		uaps = append(uaps, uap048)
	}

	if config.DumpAll || config.DumpCat062 {
		uap062, err := cat062.NewUAP("1.17")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat062 UAP: %w", err)
		}
		decoder.RegisterUAP(uap062)
		uaps = append(uaps, uap062)
	}

	if config.DumpAll || config.DumpCat063 {
		uap063, err := cat063.NewUAP("1.6")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat063 UAP: %w", err)
		}
		decoder.RegisterUAP(uap063)
		uaps = append(uaps, uap063)
	}

	if len(uaps) == 0 {
		return nil, fmt.Errorf("no categories selected, use --dumpAll or specify categories")
	}

	return decoder, nil
}
