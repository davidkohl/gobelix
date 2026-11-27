# Changelog

Behaviour changes that callers can notice. Breaking changes are marked
**breaking**; where a field changed its unit or meaning, it was renamed so
that old code fails to compile instead of reading other values.

## Unreleased

### Core (`asterix`)

- **breaking:** `Record.Decode` returns a `DecodeError` wrapping
  `ErrInvalidFSPEC` when the FSPEC sets a bit for an FRN that has no data
  item in the UAP (spare, reserved, beyond the UAP, or an RS bit the UAP
  does not support). Before, the bit was ignored and every later item was
  read from the wrong offset without an error. Valid input decodes exactly
  as before, except for the changes listed below: REP = 0 in I002/070 and
  I023/120 is now a decode error, and the RS bit of CAT001/CAT002 is now
  decoded as an RFS field.
- `Record.Decode` walks the set FSPEC bits in FRN order through a table
  precomputed by `BaseUAP` instead of scanning every UAP field (faster; a
  UAP that does not embed `BaseUAP` keeps the field scan).
- Random Field Sequencing: `BaseUAP.EnableRFS(frn)` declares the RS bit,
  and `RandomFieldSequencer` reports it. The RFS field (N, then N times FRN
  + item) is read at the FSPEC position of the RS bit; its items join the
  record and the FSPEC is rebuilt to the standard form, so `Encode` writes
  no RFS field. An RFS entry with an unknown FRN, the RS FRN itself or an
  item sent twice is a decode error.
- `BaseUAP.EnableRFS(frn, notInRFS...)` also lists the FRNs that the RFS
  field must not carry; an RFS entry naming one is a decode error
  (`ErrInvalidFSPEC`). CAT001 and CAT002 use it for the SP field.
- For a `RecordUAPSelector` UAP (CAT001), `SetDataItem` keeps the FSPEC in
  the layout the items select, so `FSPEC()`, `EstimateSize` and
  `DataBlock.String` agree with the encoded record before `Encode` runs.
- `Record.Encode` no longer compares UAP interface values (that panicked
  for a UAP whose dynamic type is not comparable).
- `AsterixMessage.String` no longer sorts the UAP's shared `Fields()`
  slice in place, and for a `RecordUAPSelector` UAP (CAT001) it lists each
  record's items with the FRNs of the layout the selector picks for that
  record (track FRNs for a track record).
- `RecordUAPSelector` (new in 60ee4b9): a UAP can choose the record layout
  per record; `Decode` asks it from the record bytes, `Encode` from the
  items, and FRNs follow the selected layout.
- `common.ExplicitField` (new in 519a759) for SP/RE fields: LEN counts
  itself; LEN 0 or beyond the buffer is a decode error; an empty payload
  decodes as a non-nil slice.

### CAT001 (ed 1.4)

- **breaking:** `uap.NewUAP12` selects the plot or the track UAP per record
  from I001/020 TYP; track records no longer decode with plot FRNs. Use
  `NewUAP12Plot` or `NewUAP12Track` to force one layout.
- **breaking:** I001/020 TYP is one bit (bit 8; 0 plot, 1 track) and
  `Validate` refuses TYP other than 0 or 1. SIM moved to bit 7, SSR/PSR to
  6-5, ANT 4, SPI 3, RAB 2; the first extent carries TST, DS1/DS2, ME, MI.
  All were one or two bits off.
- **breaking:** I001/120 `MeasuredRadialDopplerSpeed.DopplerSpeed` (m/s,
  LSB 1 m/s) is now `DopplerSpeedKt` in knots, LSB 2^(-14+f) NM/s with the
  default f = 6 (14.0625 kt).
- **breaking:** I001/130 `RadarPlotCharacteristics` keeps only
  `Indicators` (every 7-bit indicator octet; at least one). The SRL..APD
  fields were a guess at the I048/130 layout; the content is application
  dependent (§5.2.13).
- **breaking:** I001/100 bits 28-17 are the Mode-C reply in Gray
  (Gillham) code: `GrayCode` holds it raw, `FlightLevel` is the decoded
  altitude (before, the bits were read as a binary flight level), and
  `AltitudeValid` says whether the code is a valid Gillham altitude.
  `GillhamToAltitude` / `AltitudeToGillham` are exported.
- I001/050 Mode-2: V, G, L at bits 16-14 and the full 12-bit code (the
  flags were read from the code nibble, and the code lost its top four
  bits). L is "smoothed Mode-2 code from a local tracker", not "late".
- I001/030 keeps every W/E value in `Values`; `W` is the first.
- I001/170 first extent TRE; SP001 at FRN 20 of both UAPs; RFS at FRN 21
  of both UAPs is decoded.
- SP001 (FRN 20) inside the RFS field is a decode error in both UAPs
  (§5.3.2.16 note 4, §5.4.2.20 note 3: non-standard information goes in
  the SP field "and not the RFS field(s)").
- With `NewUAP12`, I001/020 inside the RFS field is a decode error: TYP
  selects the layout and is read from I001/020 at FRN 2, so an I001/020
  sent in the RFS field decoded a track record with the plot layout.
  I001/020 is compulsory (§5.3.2.2, §5.4.2.2) and belongs at FRN 2.
  `NewUAP12Plot` and `NewUAP12Track` do not select by TYP and still accept
  it there.
- Legacy/new field pairs must agree when both are set, or `Validate` and
  `Encode` return an error: I001/030 `W` vs `Values[0]`; I001/100
  `GrayCode` vs `FlightLevel`/`AltitudeValid` (the code must be the
  Gillham code of the level). A zero legacy value (`W` 0, `FlightLevel` 0
  with `AltitudeValid` false) counts as unset. After `Decode` both views
  are filled and agree; change both, or reset the one you do not use.
- A set spare FSPEC bit (plot FRN 16-19, plot FRN above 21, track FRN
  above 22) is a decode error.

### CAT002 (ed 1.2)

- **breaking:** I002/070 IDENT is bits 15-11 (it was read one bit low, so
  every plot category came out wrong).
- **breaking:** I002/070 needs at least one counter: REP = 0 is a decode
  error and an empty `Values` does not encode (§5.2.8).
- I002/080 Warning/Error Conditions, I002/090 Collimation Error, I002/100
  Dynamic Window Type 1 and SP002 (FRN 13) are decoded (a record with
  I002/080 failed to decode before). RFS at FRN 14 is decoded; SP002
  inside the RFS field is a decode error (§5.3.2.13 note 4).
- I002/050 and I002/060 keep every extent in `Parts`; `Status` / `Mode`
  is the first. When both are set they must agree (`Status`/`Mode` ==
  `Parts[0]`, zero counts as unset), or `Validate` and `Encode` fail.
- A set FSPEC bit for the spare FRN 12 is a decode error.

### CAT023 (ed 1.3)

- **breaking:** UAP FRNs per ed 1.3 Table 3: I023/110 FRN 8, I023/120
  FRN 9, RE FRN 13, SP FRN 14 (they were at 14, 15, 21, 22, so no ed 1.3
  service status, statistics, RE or SP record decoded).
- **breaking:** I023/110 is a one-octet extensible item with STAT at bits
  4-2 (was read as a compound item with STAT at bits 8-6); STAT 5
  (initialisation) is accepted.
- **breaking:** I023/120 is repetitive: REP, then six-octet blocks (TYPE,
  REF + spare, 32-bit counter) in `Entries`. `TYPE`, `REF` and `CV` are
  deprecated; they are still filled on decode and still encode when
  `Entries` is empty. When both views are set they must agree (TYPE and
  REF those of `Entries[0]`, one `CV` per entry), or `Validate` and
  `Encode` fail. REP = 0 is a decode error and an empty item does not
  encode (§5.2.8).
- RE023/SP023: LEN counts itself (the decoder read one octet too many).

### CAT034 (ed 1.29)

- **breaking:** I034/120 `DataSourcePosition.Height` (ft, LSB 25 ft) is
  now `HeightM` in metres (LSB 1 m, §5.2.12). `Validate` rejects NaN and
  Inf in height, latitude and longitude.
- I034/050 and I034/060 keep every subfield, with bit-field accessors; a
  spare subfield bit or an extension announcing a subfield of undefined
  length is a decode error; I034/050 can omit COM (`OmitCOM`).
- I034/090 Collimation Error, I034/100 Generic Polar Window, I034/110 Data
  Filter, RE and SP are decoded (a record with RE or SP failed before).
- I034/000 accepts message types 1-7 (types 3-4 were misnamed, 5-7
  refused).
- I034/041 returns an error on a short buffer instead of `0, nil`.

### CAT062 (ed 1.17, 1.20)

- SP062 (1.20) and RE062 (1.17, 1.20): LEN counts itself (ASTERIX Part 1
  §4.3.5). The decoders read LEN octets after the length octet, one too
  many, so the next item or record was misaligned. LEN = 0 and a LEN
  beyond the data are decode errors; an empty `Data` encodes as LEN = 1
  (it was LEN = 0); `Validate` and `Encode` refuse a `Data` whose first
  octet is not its length. `Data` holds the LEN octet, as before. SP062
  of 1.17 already counted LEN correctly; its `Data` holds only the
  payload.
- A set FSPEC bit for a spare FRN (for example FRN 2) or an FRN beyond the
  UAP is a decode error (FSPEC check above).

### CAT048, CAT021

- Decoded values are unchanged. The only new behaviour is the FSPEC check
  above: a record that sets a spare FSPEC bit or an FRN beyond the UAP is
  now a decode error.
