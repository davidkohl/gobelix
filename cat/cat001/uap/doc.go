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
// This means you cannot use a unified UAP - the correct UAP must be selected
// based on what your radar system transmits.
//
// # Choosing the Right UAP
//
// Use Plot UAP (default) when:
//   - Receiving data from primary surveillance radar (PSR)
//   - Receiving data from secondary surveillance radar (SSR)
//   - Your radar sends raw detections without track processing
//   - You're unsure which format your radar uses (most send plots)
//
// Use Track UAP when:
//   - Your radar has a built-in tracker that outputs track data
//   - You specifically know your radar transmits CAT001 tracks
//   - The data includes track-specific items like I001/161 (Track Number),
//     I001/042 (Cartesian Position), I001/200 (Track Velocity)
//
// # Usage
//
//	// For plot data (most common - this is the default):
//	uap, err := uap.NewUAP12()      // or uap.NewUAP12Plot()
//
//	// For track data (only if your radar has built-in tracking):
//	uap, err := uap.NewUAP12Track()
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
