// Package nepdate converts and formats Bikram Sambat (Nepali calendar) dates.
//
// The supported range is BS 1901-01-01 to 2199-12-30 (AD 1844-04-11 to 2143-04-15). Dates
// outside it return ErrOutOfRange. The zero Date is "no date": IsZero reports true and every
// operation that can fail returns ErrOutOfRange.
//
// Import it as
//
//	import nepdate "github.com/sanamhub/nepdate-go"
//
// The behaviour follows the shared NepDate specification in shared/spec/ (ALGORITHM.md,
// FORMATTING.md, PARSING.md and PARITY.md). Tithi, public holidays and events are in the
// separate package github.com/sanamhub/nepdate-go/patro, which this package never imports.
package nepdate

//go:generate go run ./internal/gen
