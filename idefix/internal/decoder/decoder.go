// internal/decoder/decoder.go
package decoder

import (
	"fmt"

	"github.com/davidkohl/gobelix/asterix"
	"github.com/davidkohl/gobelix/cat/cat021"
	"github.com/davidkohl/gobelix/cat/cat048"
	"github.com/davidkohl/gobelix/cat/cat062"
	"github.com/davidkohl/gobelix/cat/cat063"
	"github.com/davidkohl/gobelix/encoding"
)

// Config represents decoder configuration options
type Config struct {
	DumpAll    bool
	DumpCat021 bool
	DumpCat048 bool
	DumpCat062 bool
	DumpCat063 bool
}

// CreateDecoder creates and configures a decoder with the specified UAPs
func CreateDecoder(config Config) (*encoding.Decoder, error) {
	// Create a buffer pool for improved performance
	pool := encoding.NewBufferPool()

	// Create a decoder with improved options
	decoder := encoding.NewDecoder(
		encoding.WithDecoderBufferPool(pool),
		encoding.WithDecoderParallelism(2), // Use moderate parallelism
	)

	var uaps []asterix.UAP

	if config.DumpAll || config.DumpCat021 {
		uap021, err := cat021.NewUAP("2.6")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Cat021 UAP: %w", err)
		}
		decoder.RegisterUAP(uap021)
		uaps = append(uaps, uap021)
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
