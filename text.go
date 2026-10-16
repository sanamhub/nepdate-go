package nepdate

// MarshalText returns d as "YYYY-MM-DD" (PARITY D-11), so encoding/json and other encoders
// write a Date as a string and accept it as a map key. It returns ErrOutOfRange for the zero
// Date.
func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, ErrOutOfRange
	}
	return d.appendYMD(make([]byte, 0, 10), '-'), nil
}

// AppendText appends "YYYY-MM-DD" to b (encoding.TextAppender). It does not allocate when b
// has room, and returns b unchanged with ErrOutOfRange for the zero Date.
func (d Date) AppendText(b []byte) ([]byte, error) {
	if d.IsZero() {
		return b, ErrOutOfRange
	}
	return d.appendYMD(b, '-'), nil
}

// UnmarshalText sets *d from text in the strict form Parse accepts. On error *d is unchanged.
// It parses the bytes directly, so it does not allocate on success.
func (d *Date) UnmarshalText(text []byte) error {
	parsed, err := parseStrict(text)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
