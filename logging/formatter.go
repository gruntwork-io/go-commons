package logging

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

// TextFormatterWithBinName is a logrus formatter that prefixes the binary name to the log output.
type TextFormatterWithBinName struct {
	Name          string
	TextFormatter logrus.TextFormatter
}

func (formatter *TextFormatterWithBinName) Format(entry *logrus.Entry) ([]byte, error) {
	logTxt, err := formatter.TextFormatter.Format(entry)
	if err != nil {
		return logTxt, err
	}

	outBuf := &bytes.Buffer{}
	fmt.Fprintf(outBuf, "[%s] ", formatter.Name)
	if _, err := outBuf.Write(logTxt); err != nil {
		return logTxt, err
	}
	return outBuf.Bytes(), nil
}
