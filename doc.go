// Package nepdate converts between the Bikram Sambat (Nepali) and Gregorian calendars, and
// formats, parses and does arithmetic on BS dates.
//
// The supported range is BS 1901-01-01 to 2199-12-30 (AD 1844-04-11 to 2143-04-15). Dates
// outside it return ErrOutOfRange. The zero Date is "no date": IsZero reports true and every
// operation that can fail returns ErrOutOfRange.
//
// Import it as
//
//	import nepdate "github.com/sanamhub/nepdate-go"
//
// Tithi, public holidays and events live in the separate package
// github.com/sanamhub/nepdate-go/patro. This package never imports it, so a program that only
// converts dates links none of the patro data.
//
// A Date marshals to JSON as "YYYY-MM-DD". Marshalling the zero Date returns ErrOutOfRange, so
// an optional date should be a *Date: nil marshals as null.
package nepdate

//go:generate go run ./internal/gen
