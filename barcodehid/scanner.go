package barcodehid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	kcEnter   = 0x28
	kcOne     = 0x1e
	asciiZero = 48
)

type Scanner struct {
	reader *bufio.Reader
}

func New(reader io.Reader) Scanner {
	return Scanner{reader: bufio.NewReader(reader)}
}

func (s Scanner) Read() (string, error) {
	data, err := s.reader.ReadBytes(kcEnter)
	if err != nil {
		return "", fmt.Errorf("reading keystrokes: %w", err)
	}

	data = bytes.TrimSuffix(data, []byte{kcEnter})

	var code string
	for _, kc := range data {
		if kc == 0 {
			// Data is full of zeroes for unknown reasons.
			continue
		}

		if kc < kcOne || kc > kcEnter {
			// Keycode is not a number keycode.
			continue
		}

		// Keycodes for 1..9 start at kcOne, continue up to 9, and then jump back to zero.
		// To get the number, we subtract the first keycode (kcOne), and add 1 to make it 1. We then take the modulus so
		// the last keycode, 10, becomes zero. Finally, we add the ascii value for '0'.
		code += string((kc-kcOne+1)%10 + asciiZero)
	}

	return code, nil
}
