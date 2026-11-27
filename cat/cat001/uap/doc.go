// Package uap provides User Application Profile definitions for ASTERIX Category 001.
//
// The UAP defines the structure and ordering of data items within an ASTERIX
// Category 001 record. This package provides UAP implementations for supported
// versions of the Category 001 specification.
//
// # Plot vs Track UAPs
//
// CAT001 defines two distinct UAPs with incompatible FRN-to-DataItem mappings:
//
//   - Plot UAP: For raw radar detections (15 FRNs)
//   - Track UAP: For processed track data from radars with built-in trackers (22 FRNs)
//
// The FRN mappings differ significantly between the two UAPs. For example:
//
//	FRN 3 in Plot UAP  = I001/040 (Measured Position Polar)
//	FRN 3 in Track UAP = I001/161 (Track Number)
//
// I001/020 TYP (bit 8 of its first octet) tells which UAP applies to the
// rest of the record. NewUAP12 reads TYP per record and switches between the
// two (it implements asterix.RecordUAPSelector).
//
// # Usage
//
//	// Plot or track per record, selected by I001/020 TYP:
//	uap, err := uap.NewUAP12()
//
//	// Force one layout:
//	uap, err := uap.NewUAP12Plot() // or uap.NewUAP12Track()
//
// # Typical ATC Data Flow
//
// In most Air Traffic Control systems:
//
//	Radar → CAT001 Plots → Tracker System → CAT062 Tracks → Display
//
// The tracking is performed by dedicated tracker systems, which output CAT062.
// Therefore, most CAT001 data you encounter will be plots, not tracks.
package uap
