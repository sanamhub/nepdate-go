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

// MarshalJSON returns d as the JSON string "YYYY-MM-DD" (PARITY D-11). It returns ErrOutOfRange
// for the zero Date, so an optional date should be a *Date, whose nil value marshals as null.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, ErrOutOfRange
	}
	b := make([]byte, 0, 12)
	b = append(b, '"')
	b = d.appendYMD(b, '-')
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON sets *d from a JSON string in the strict form Parse accepts. It rejects null with
// ErrOutOfRange (PARITY D-01) and any non-string value with a *ParseError. On error *d is
// unchanged.
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return ErrOutOfRange
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return &ParseError{Input: string(data), Kind: InvalidCharacter}
	}
	parsed, err := parseStrict(data[1 : len(data)-1])
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
