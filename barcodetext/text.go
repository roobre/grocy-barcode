package barcodetext

import (
	"bufio"
	"io"
	"strings"
)

type Scanner struct {
	reader *bufio.Reader
}

func New(reader io.Reader) Scanner {
	return Scanner{reader: bufio.NewReader(reader)}
}

func (s Scanner) Read() (string, error) {
	code, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(code), nil
}
