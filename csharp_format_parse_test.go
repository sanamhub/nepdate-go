package nepdate

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

// C# catalogue rows for formatting, month names and parsing (see TestCSharp in csharp_test.go). C#'s SmartDateParser and
// the string extensions that call it map to ParseLenient (PARITY D-06); NepaliDate.Parse and
// TryParse map to Parse.

const (
	d06Guess = "D-06: no guessing of the field order, no 2-digit year expansion, no fuzzy month names"
	naNull   = "n/a: a Go string is never nil"
)

var csharpFormatParseRows = []csharpRow{
	// Abilities/NepaliDateFormattableTests, date 2081-04-15
	{id: "NepaliDateFormattableTests.ToString_NullFormat_ReturnsDefault", run: formats(nd(2081, 4, 15), "", "2081/04/15")},
	{id: "NepaliDateFormattableTests.ToString_G_SameAsDefault", run: formats(nd(2081, 4, 15), "G", nd(2081, 4, 15).String())},
	{id: "NepaliDateFormattableTests.ToString_g_SameAsDefault", run: formats(nd(2081, 4, 15), "g", nd(2081, 4, 15).String())},
	{id: "NepaliDateFormattableTests.ToString_d_SameAsDefault", run: formats(nd(2081, 4, 15), "d", "2081/04/15")},
	{id: "NepaliDateFormattableTests.ToString_D_ReturnsLongDateString", run: func(t *testing.T) {
		got := nd(2081, 4, 15).Format("D", English)
		check(t, strings.Contains(got, "2081") && strings.Contains(got, "Shrawan"), "D has the year and the month name")
	}},
	{id: "NepaliDateFormattableTests.ToString_s_ReturnsSortableIsoFormat", run: formats(nd(2081, 4, 15), "s", "2081-04-15")},
	{id: "NepaliDateFormattableTests.CustomFormat_yyyy_MM_dd_WithDash", run: formats(nd(2081, 4, 15), "yyyy-MM-dd", "2081-04-15")},
	{id: "NepaliDateFormattableTests.CustomFormat_dd_MM_yyyy_WithDot", run: formats(nd(2081, 4, 15), "dd.MM.yyyy", "15.04.2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_MM_slash_dd_slash_yyyy", run: formats(nd(2081, 4, 15), "MM/dd/yyyy", "04/15/2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_MMMxxx_ReturnsThreeCharMonthAbbreviation", run: formats(nd(2081, 4, 15), "MMM yyyy", "Shr 2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_MMMM_ReturnsFullMonthName", run: formats(nd(2081, 4, 15), "MMMM dd, yyyy", "Shrawan 15, 2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_yy_ReturnsTwoDigitYear", run: formats(nd(2081, 4, 15), "yy/MM/dd", "81/04/15")},
	{id: "NepaliDateFormattableTests.CustomFormat_M_WithoutLeadingZero_ReturnsUnpadded", run: formats(nd(2081, 4, 15), "M/d/yyyy", "4/15/2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_SingleDigitMonth_WithoutLeadingZero", run: formats(nd(2081, 1, 5), "M/d/yyyy", "1/5/2081")},
	{id: "NepaliDateFormattableTests.CustomFormat_LiteralInSingleQuotes", run: formats(nd(2081, 4, 15), "yyyy'BS'", "2081BS")},
	{id: "NepaliDateFormattableTests.CustomFormat_BackslashEscapedTokens_ProduceLiterals", run: formats(nd(2081, 4, 15), `yyyy\M\M-\d\d`, "2081MM-dd")},
	{id: "NepaliDateFormattableTests.CustomFormat_BackslashBeforeLiteralChar_IsPassedThrough", run: formats(nd(2081, 4, 15), `yyyy\-MM\-dd`, "2081-04-15")},
	// C# string interpolation and string.Format call IFormattable; Go calls Format directly.
	{id: "NepaliDateFormattableTests.StringInterpolation_WithFormatSpecifier_UsesIFormattable", run: formats(nd(2081, 4, 15), "s", "2081-04-15")},
	{id: "NepaliDateFormattableTests.StringFormatMethod_WithSpecifier", run: formats(nd(2081, 4, 15), "yyyy-MM-dd", "2081-04-15")},
	{id: "NepaliDateFormattableTests.SortableFormat_LexicographicOrderEqualsChronologicalOrder", run: func(t *testing.T) {
		check(t, nd(2080, 12, 30).Format("s", English) < nd(2081, 1, 1).Format("s", English), "s sorts like the dates")
	}},

	// Abilities/NepaliDateParsableTests: IParsable is Parse, ISpanParsable is the same generic
	// parser over []byte, which UnmarshalText uses.
	{id: "NepaliDateParsableTests.Parse_Via_GenericConstraint_Returns_CorrectDate", run: parsesTo("2081-04-15", nd(2081, 4, 15))},
	{id: "NepaliDateParsableTests.TryParse_Via_GenericConstraint_ValidDate_Returns_True", run: parsesTo("2081/04/15", nd(2081, 4, 15))},
	{id: "NepaliDateParsableTests.TryParse_Via_GenericConstraint_InvalidDate_Returns_False", run: parseFailsSyntax("garbage")},
	{id: "NepaliDateParsableTests.SpanParse_Via_GenericConstraint_Returns_CorrectDate", run: bytesParseTo("2081-04-15", nd(2081, 4, 15))},
	{id: "NepaliDateParsableTests.SpanTryParse_Via_GenericConstraint_ValidDate_Returns_True", run: bytesParseTo("2081/04/15", nd(2081, 4, 15))},

	// Core/NepaliDateMonthNameTests
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#1", run: monthIs(nd(2080, 1, 1), Baishakh)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#2", run: monthIs(nd(2080, 2, 1), Jestha)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#3", run: monthIs(nd(2080, 3, 1), Ashad)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#4", run: monthIs(nd(2080, 4, 1), Shrawan)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#5", run: monthIs(nd(2080, 5, 1), Bhadra)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#6", run: monthIs(nd(2080, 6, 1), Ashoj)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#7", run: monthIs(nd(2080, 7, 1), Kartik)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#8", run: monthIs(nd(2080, 8, 1), Mangsir)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#9", run: monthIs(nd(2080, 9, 1), Poush)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#10", run: monthIs(nd(2080, 10, 1), Magh)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#11", run: monthIs(nd(2080, 11, 1), Falgun)},
	{id: "NepaliDateMonthNameTests.MonthName_AllMonths_ReturnCorrectEnum#12", run: monthIs(nd(2080, 12, 1), Chaitra)},
	{id: "NepaliDateMonthNameTests.MonthName_SameMonthDifferentDays_AllReturnSameEnum", run: func(t *testing.T) {
		for _, d := range []Date{nd(2080, 4, 1), nd(2080, 4, 15), lastOf(2080, 4)} {
			check(t, d.Month() == Shrawan, "Shrawan on every day")
		}
	}},
	{id: "NepaliDateMonthNameTests.MonthName_SameMonthDifferentYears_AllReturnSameEnum#1", run: monthIs(nd(2070, 9, 1), Poush)},
	{id: "NepaliDateMonthNameTests.MonthName_SameMonthDifferentYears_AllReturnSameEnum#2", run: monthIs(nd(2080, 9, 1), Poush)},
	{id: "NepaliDateMonthNameTests.MonthName_SameMonthDifferentYears_AllReturnSameEnum#3", run: monthIs(nd(2090, 9, 1), Poush)},
	{id: "NepaliDateMonthNameTests.MonthName_SameMonthDifferentYears_AllReturnSameEnum#4", run: monthIs(nd(2100, 9, 1), Poush)},
	{id: "NepaliDateMonthNameTests.MonthName_AddingOneMonth_AdvancesEnum", run: func(t *testing.T) {
		next, err := nd(2080, 11, 1).AddMonths(1, Clamp)
		check(t, err == nil && next.Month() == Chaitra, "Falgun + 1 month is Chaitra")
	}},
	{id: "NepaliDateMonthNameTests.MonthName_YearBoundary_WrapsFromChaitraToBasikhakh", run: func(t *testing.T) {
		check(t, nd(2080, 12, 1).Month() == Chaitra && nd(2081, 1, 1).Month() == Baishakh, "Chaitra then Baishakh")
	}},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#1", run: monthNumber(Baishakh, 1)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#2", run: monthNumber(Jestha, 2)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#3", run: monthNumber(Ashad, 3)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#4", run: monthNumber(Shrawan, 4)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#5", run: monthNumber(Bhadra, 5)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#6", run: monthNumber(Ashoj, 6)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#7", run: monthNumber(Kartik, 7)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#8", run: monthNumber(Mangsir, 8)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#9", run: monthNumber(Poush, 9)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#10", run: monthNumber(Magh, 10)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#11", run: monthNumber(Falgun, 11)},
	{id: "NepaliDateMonthNameTests.NepaliMonthsEnum_IntegerValueMatchesMonthNumber#12", run: monthNumber(Chaitra, 12)},

	// Core/SmartDateParserTests, expected 2080-04-15
	{id: "SmartDateParserTests.Parse_StandardFormat_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "2080/04/15", "2080-04-15", "2080.04.15")},
	{id: "SmartDateParserTests.Parse_InvertedFormat_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "15/04/2080", "15-04-2080", "15.04.2080")},
	{id: "SmartDateParserTests.Parse_MonthDayYearFormat_ReturnsCorrectDate", skip: d06Guess},
	{id: "SmartDateParserTests.Parse_WithMonthNames_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080", "15 Sawan 2080", "15 Saun 2080", "Shrawan 15, 2080", "Shrawan 15 2080")},
	{id: "SmartDateParserTests.Parse_WithNepaliUnicode_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "२०८०/०४/१५", "१५/०४/२०८०", "१५ श्रावण २०८०", "श्रावण १५, २०८०")},
	{id: "SmartDateParserTests.Parse_MixedFormats_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "15 साउन 2080", "साउन 15, २०८०", "15 Shrawan २०८०")},
	{id: "SmartDateParserTests.Parse_WithSuffixes_ReturnsCorrectDate", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080 B.S.", "15 साउन 2080 BS", "15 Shrawan 2080 V.S.", "15 साउन, 2080 मिति", "15 साउन, 2080 गते")},
	{id: "SmartDateParserTests.Parse_WithTypos_ReturnsCorrectDate", skip: d06Guess},
	{id: "SmartDateParserTests.Parse_ShorterYearFormats_ReturnsCorrectDate", skip: d06Guess},
	{id: "SmartDateParserTests.Parse_InvalidFormat_ThrowsFormatException", run: lenientFailsAll("not a date", "15/13/2080", "32/03/2080")},
	{id: "SmartDateParserTests.TryParse_ValidFormat_ReturnsTrue", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080")},
	{id: "SmartDateParserTests.TryParse_InvalidFormat_ReturnsFalse", run: lenientFailsAll("not a date")},
	{id: "SmartDateParserTests.ExtensionMethod_ToNepaliDate_ParsesCorrectly", run: lenientAll(nd(2080, 4, 15), "2080/04/15", "२०८०/०४/१५", "15 Shrawan 2080")},
	{id: "SmartDateParserTests.ExtensionMethod_TryToNepaliDate_ParsesCorrectly", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080")},

	// Extensions/StringExtensionsTests
	{id: "StringExtensionsTests.ToNepaliDate_ValidDateString_ReturnsNepaliDate", run: lenientAll(nd(2080, 5, 15), "2080/05/15")},
	{id: "StringExtensionsTests.ToNepaliDate_NepaliDigits_ReturnsNepaliDate", run: lenientAll(nd(2080, 5, 15), "२०८०/०५/१५")},
	{id: "StringExtensionsTests.ToNepaliDate_InvalidDateString_ThrowsFormatException", run: lenientFailsAll("not a date")},
	{id: "StringExtensionsTests.TryToNepaliDate_ValidDateString_ReturnsTrue", run: lenientAll(nd(2080, 5, 15), "2080/05/15")},
	{id: "StringExtensionsTests.TryToNepaliDate_InvalidDateString_ReturnsFalse", run: lenientFailsAll("not a date")},

	// Core/OptimizationVerificationTests, phase 3 rows
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#1", run: stringIs(nd(2080, 5, 15), "2080/05/15")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#2", run: stringIs(nd(2080, 1, 1), "2080/01/01")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#3", run: stringIs(nd(2080, 12, 30), "2080/12/30")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#4", run: stringIs(nd(1901, 1, 1), "1901/01/01")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#5", run: stringIs(nd(2199, 12, 30), "2199/12/30")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#6", run: stringIs(nd(2000, 6, 9), "2000/06/09")},
	{id: "OptimizationVerificationTests.ToString_ProducesExactFormat#7", run: stringIs(nd(2100, 10, 29), "2100/10/29")},
	{id: "OptimizationVerificationTests.ToString_OutputIsExactly10Characters", run: func(t *testing.T) {
		s := nd(2080, 5, 15).String()
		check(t, len(s) == 10 && s[4] == '/' && s[7] == '/', "10 characters with slashes at 4 and 7")
	}},
	{id: "OptimizationVerificationTests.ToString_AllMonthsZeroPadded", run: func(t *testing.T) {
		for m := Baishakh; m <= Chaitra; m++ {
			check(t, nd(2080, m, 1).String()[5:7] == string(English.AppendDigits(nil, int(m), 2)), "month padded")
		}
	}},
	{id: "OptimizationVerificationTests.ToString_AllDaysZeroPadded", run: func(t *testing.T) {
		for day := 1; day <= nd(2080, 1, 1).MonthLength(); day++ {
			check(t, nd(2080, 1, day).String()[8:10] == string(English.AppendDigits(nil, day, 2)), "day padded")
		}
	}},
	{id: "OptimizationVerificationTests.ToString_RoundtripsWithConstructor", run: func(t *testing.T) {
		for y := 1901; y <= 2199; y += 50 {
			for m := Baishakh; m <= Chaitra; m++ {
				parsesTo(nd(y, m, 1).String(), nd(y, m, 1))(t)
			}
		}
	}},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#1", run: shortIs(YMD, Slash, true, "2080/05/15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#2", run: shortIs(YMD, Dash, true, "2080-05-15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#3", run: shortIs(YMD, Dot, true, "2080.05.15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#4", run: shortIs(YMD, Underscore, true, "2080_05_15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#5", run: shortIs(YMD, Space, true, "2080 05 15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#6", run: shortIs(YMD, Backslash, true, `2080\05\15`)},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#7", run: shortIs(DMY, Slash, true, "15/05/2080")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#8", run: shortIs(DMY, Dash, true, "15-05-2080")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#9", run: shortIs(MDY, Slash, true, "05/15/2080")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#10", run: shortIs(YDM, Slash, true, "2080/15/05")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#11", run: shortIs(MYD, Slash, true, "05/2080/15")},
	{id: "OptimizationVerificationTests.ToString_WithFormatAndSeparator_ProducesCorrectOutput#12", run: shortIs(DYM, Slash, true, "15/2080/05")},
	{id: "OptimizationVerificationTests.ToString_WithoutLeadingZeros_ProducesCorrectOutput#1", run: shortIs(YMD, Slash, false, "2080/5/15")},
	{id: "OptimizationVerificationTests.ToString_WithoutLeadingZeros_ProducesCorrectOutput#2", run: shortIs(DMY, Dash, false, "15-5-2080")},
	{id: "OptimizationVerificationTests.ToString_WithoutLeadingZeros_ProducesCorrectOutput#3", run: shortIs(MDY, Dot, false, "5.15.2080")},
	// C#'s ToUnicodeString is Short(YMD, Slash, true, Nepali) (FORMATTING §6).
	{id: "OptimizationVerificationTests.ToUnicodeString_DefaultFormat_ConvertsAllDigits", run: func(t *testing.T) {
		check(t, nd(2080, 5, 15).Short(YMD, Slash, true, Nepali) == "२०८०/०५/१५", "Devanagari digits")
	}},
	{id: "OptimizationVerificationTests.ToUnicodeString_EachDigitConvertsCorrectly", run: func(t *testing.T) {
		check(t, nd(1902, 3, 4).Short(YMD, Slash, true, Nepali) == "१९०२/०३/०४", "every digit of 1902/03/04")
	}},
	{id: "OptimizationVerificationTests.ToUnicodeString_SeparatorsPreserved", run: func(t *testing.T) {
		s := nd(2080, 5, 15).Short(YMD, Dash, true, Nepali)
		check(t, strings.Contains(s, "-") && !strings.Contains(s, "/"), "dash kept, no slash")
	}},
	{id: "OptimizationVerificationTests.ToUnicodeString_SameLengthAsToString", run: func(t *testing.T) {
		for y := 1901; y <= 2199; y += 100 {
			d := nd(y, 6, 15)
			check(t, utf8.RuneCountInString(d.Short(YMD, Slash, true, Nepali)) == len(d.String()), "same number of characters")
		}
	}},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliDigits_ParseCorrectly#1", run: lenientAll(nd(2080, 5, 15), "२०८०/०५/१५")},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliDigits_ParseCorrectly#2", run: lenientAll(nd(2080, 5, 15), "२०८०-०५-१५")},
	{id: "OptimizationVerificationTests.SmartDateParser_PureEnglishDigits_NoConversionNeeded", run: lenientAll(nd(2080, 5, 15), "2080/05/15")},
	{id: "OptimizationVerificationTests.SmartDateParser_MixedNepaliEnglishDigits_ParseCorrectly", run: lenientAll(nd(2080, 5, 15), "2०८0/0५/१5")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#1", run: lenientAll(nd(2080, 5, 15), "2080/05/15 B.S.")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#2", run: lenientAll(nd(2080, 5, 15), "2080/05/15 BS")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#3", run: lenientAll(nd(2080, 5, 15), "2080/05/15 V.S.")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#4", run: lenientAll(nd(2080, 5, 15), "2080/05/15 VS")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#5", run: lenientAll(nd(2080, 5, 15), "2080/05/15 bs")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithIndicators_RemovesAndParsesCorrectly#6", run: lenientAll(nd(2080, 5, 15), "2080/05/15 v.s.")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithNepaliKeywords_RemovesAndParses", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080 गते")},
	{id: "OptimizationVerificationTests.SmartDateParser_WithExtraSpaces_ParsesCorrectly", run: lenientAll(nd(2080, 5, 15), "  2080 / 05 / 15  ")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#1", run: lenientAll(nd(2080, 1, 15), "15 Baisakh 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#2", run: lenientAll(nd(2080, 2, 15), "15 Jestha 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#3", run: lenientAll(nd(2080, 3, 15), "15 Asar 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#4", run: lenientAll(nd(2080, 4, 15), "15 Shrawan 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#5", run: lenientAll(nd(2080, 5, 15), "15 Bhadra 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#6", run: lenientAll(nd(2080, 6, 15), "15 Ashwin 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#7", run: lenientAll(nd(2080, 7, 15), "15 Kartik 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#8", run: lenientAll(nd(2080, 8, 15), "15 Mangsir 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#9", run: lenientAll(nd(2080, 9, 15), "15 Poush 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#10", run: lenientAll(nd(2080, 10, 15), "15 Magh 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#11", run: lenientAll(nd(2080, 11, 15), "15 Falgun 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_AllMonthNames_ParseCorrectly#12", run: lenientAll(nd(2080, 12, 15), "15 Chaitra 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_MonthNameDifferentPositions_ParsesCorrectly#1", run: lenientAll(nd(2080, 1, 15), "Baisakh 15, 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_MonthNameDifferentPositions_ParsesCorrectly#2", run: lenientAll(nd(2080, 1, 15), "2080 Baisakh 15")},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliUnicodeMonthNames_ParseCorrectly#1", run: lenientAll(nd(2080, 1, 15), "15 वैशाख 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliUnicodeMonthNames_ParseCorrectly#2", run: lenientAll(nd(2080, 2, 15), "15 जेष्ठ 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliUnicodeMonthNames_ParseCorrectly#3", run: lenientAll(nd(2080, 3, 15), "15 असार 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_NepaliUnicodeMonthNames_ParseCorrectly#4", run: lenientAll(nd(2080, 4, 15), "15 श्रावण 2080")},
	{id: "OptimizationVerificationTests.SmartDateParser_TwoDigitYear_AdjustedTo2000s", skip: d06Guess},
	{id: "OptimizationVerificationTests.SmartDateParser_InvalidInput_ThrowsFormatException", run: lenientFailsAll("not a date")},
	{id: "OptimizationVerificationTests.SmartDateParser_NullInput_ThrowsException", skip: naNull},
	{id: "OptimizationVerificationTests.SmartDateParser_EmptyInput_ThrowsException", run: lenientFailsAll("")},
	{id: "OptimizationVerificationTests.SmartDateParser_WhitespaceInput_ThrowsException", run: lenientFailsAll("   ")},
	{id: "OptimizationVerificationTests.SmartDateParser_TryParse_ValidInput_ReturnsTrue", run: lenientAll(nd(2080, 5, 15), "2080/05/15")},
	{id: "OptimizationVerificationTests.SmartDateParser_TryParse_InvalidInput_ReturnsFalse", run: lenientFailsAll("garbage")},
	{id: "OptimizationVerificationTests.SmartDateParser_TryParse_NullInput_ReturnsFalse", skip: naNull},
	{id: "OptimizationVerificationTests.ToString_EveryMonthFirstDay_RoundtripsViaConstructor", run: func(t *testing.T) {
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				if got, err := Parse(nd(y, m, 1).String()); err != nil || got != nd(y, m, 1) {
					t.Fatalf("%d-%d-01 round-trips to %v, %v", y, m, got, err)
				}
			}
		}
	}},
	{id: "OptimizationVerificationTests.UnicodeConversion_AllDigits_RoundtripCorrectly", run: func(t *testing.T) {
		for y := 2070; y <= 2090; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				d := nd(y, m, 15)
				lenientAll(d, d.Short(YMD, Slash, true, Nepali))(t)
			}
		}
	}},
	// C#'s ToLongDateString is Long.
	{id: "OptimizationVerificationTests.ToLongDateString_Default_ContainsMonthNameAndDay", run: func(t *testing.T) {
		s := nd(2080, 5, 15).Long(LongOptions{}, English)
		check(t, strings.Contains(s, "15") && strings.Contains(s, "2080") && strings.Contains(s, ","), "day, year and comma")
	}},
	{id: "OptimizationVerificationTests.ToLongDateString_WithDayName_ContainsDayOfWeek", run: func(t *testing.T) {
		d := nd(2080, 5, 15)
		check(t, strings.Contains(d.Long(LongOptions{Weekday: true}, English), d.Weekday().String()), "weekday name")
	}},
	{id: "OptimizationVerificationTests.ToLongDateString_WithoutYear_DoesNotContainYear", run: func(t *testing.T) {
		check(t, !strings.Contains(nd(2080, 5, 15).Long(LongOptions{OmitYear: true}, English), "2080"), "no year")
	}},
}

func formats(d Date, pattern, want string) func(*testing.T) {
	return func(t *testing.T) {
		if got := d.Format(pattern, English); got != want {
			t.Errorf("Format(%q) = %q, want %q", pattern, got, want)
		}
	}
}

func stringIs(d Date, want string) func(*testing.T) {
	return func(t *testing.T) {
		if got := d.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	}
}

func shortIs(o DateOrder, sep Separator, pad bool, want string) func(*testing.T) {
	return func(t *testing.T) {
		if got := nd(2080, 5, 15).Short(o, sep, pad, English); got != want {
			t.Errorf("Short(%d, %d, %v) = %q, want %q", o, sep, pad, got, want)
		}
	}
}

func monthIs(d Date, want Month) func(*testing.T) {
	return func(t *testing.T) {
		check(t, d.Month() == want, "month of "+d.String())
	}
}

func monthNumber(m Month, want int) func(*testing.T) {
	return func(t *testing.T) {
		check(t, int(m) == want, "Month value of "+m.String())
	}
}

func bytesParseTo(input string, want Date) func(*testing.T) {
	return func(t *testing.T) {
		got, err := parseStrict([]byte(input))
		if err != nil || got != want {
			t.Errorf("parseStrict([]byte(%q)) = %v, %v; want %v", input, got, err, want)
		}
	}
}

func lenientAll(want Date, inputs ...string) func(*testing.T) {
	return func(t *testing.T) {
		for _, in := range inputs {
			if got, err := ParseLenient(in); err != nil || got != want {
				t.Errorf("ParseLenient(%q) = %v, %v; want %v", in, got, err, want)
			}
		}
	}
}

// lenientFailsAll expects any parse error and the zero Date (C#'s FormatException, or false
// from TryParse with the default date).
func lenientFailsAll(inputs ...string) func(*testing.T) {
	return func(t *testing.T) {
		for _, in := range inputs {
			got, err := ParseLenient(in)
			var pe *ParseError
			ok := errors.As(err, &pe) || errors.Is(err, ErrOutOfRange) || errors.Is(err, ErrInvalidMonth) || errors.Is(err, ErrInvalidDay)
			if !ok || !got.IsZero() {
				t.Errorf("ParseLenient(%q) = %v, %v; want an error", in, got, err)
			}
		}
	}
}
