// encoding/doc.go
package encoding

/*
Package encoding provides optimized encoding and decoding functionality for ASTERIX messages.

It includes buffer pooling mechanisms and specialized encoding/decoding functions
designed for high-performance processing of ASTERIX data. This package aims to
minimize allocations and processing overhead while maintaining strict adherence
to the ASTERIX specification.

Main components:
  - BufferPool: Reusable memory pool to reduce GC pressure
  - Encoder: High-performance serialization of ASTERIX structures
  - Decoder: Efficient parsing of ASTERIX messages

The encoding package works with the structures defined in the asterix package to
provide a complete solution for processing ASTERIX data in high-throughput environments.
*/
