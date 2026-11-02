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
// an optional date should be a *Date: nil marshals as null. Unmarshalling null is rejected with
// ErrOutOfRange for the same reason.
//
// The zero Date is the only invalid value the type can hold, and it means "no date". Its
// methods never invent a date:
//
//   - IsZero reports true, and the accessors Year, Month, Day, YearDay, MonthLength, YearLength,
//     YMD and Quarter return 0, Gregorian returns 0, 0, 0, Month returns 0, and String, Short,
//     Long and Format return "".
//   - Operations that return an error return ErrOutOfRange: AddDays, AddMonths, AddYears,
//     MarshalText, AppendText, MarshalJSON, UnmarshalJSON, and ToEnglishDates.
//   - IsToday, IsYesterday and IsTomorrow report false.
//   - The zero Range is empty, and the range methods return the zero of their result type.
//
// Use IsZero before an accessor when 0 would be ambiguous. A valid date is never zero, so a
// plain comparison d == nepdate.Date{} is also correct.
package nepdate

//go:generate go run ./internal/gen
