package nepdate

// A serial is the number of days since BS 1901-01-01 (ALGORITHM §2): 0..maxSerial.

// toSerial returns the serial of a valid date.
func toSerial(d Date) int32 {
	return bsMonthStart[monthIndex(int(d.year), Month(d.month))] + int32(d.day) - 1
}

// fromSerial returns the date of serial s, which must be 0..maxSerial. The estimate is the
// month containing s or the one before it (checked by internal/gen), so one comparison fixes it
// (ADR-0002 §3).
func fromSerial(s int32) Date {
	k := int((int64(s)*bsEstMul + bsEstAdd) >> estShift)
	if bsMonthStart[k+1] <= s {
		k++
	}
	return Date{
		year:  uint16(minYear + k/12),         //nolint:gosec // k is 0..3587
		month: uint8(k%12 + 1),                //nolint:gosec // 1..12
		day:   uint8(s - bsMonthStart[k] + 1), //nolint:gosec // 1..32
	}
}
