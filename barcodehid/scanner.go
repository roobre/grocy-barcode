package barcodehid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	// Ref: https://gist.github.com/ekaitz-zarraga/2b25b94b711684ba4e969e5a5723969b
	kcA       = 0x04
	kcZ       = 0x1d
	asciiZero = byte('0')

	kcOne  = 0x1e
	kcZero = 0x27
	asciiA = byte('a')

	kcEnter = 0x28
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

		switch {
		case kc >= kcA && kc <= kcZ:
			code += string(kc - kcA + asciiA)
		case kc >= kcOne && kc <= kcZero:
			// Keycodes for 1..9 start at kcOne, continue up to 9, and then jump back to zero.
			// To get the number, we subtract the first keycode (kcOne), and add 1 to make it 1. We then take the modulus so
			// the last keycode, 10, becomes zero. Finally, we add the ascii value for '0'.
			code += string((kc-kcOne+1)%10 + asciiZero)
		}
	}

	return code, nil
}
