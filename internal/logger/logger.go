package logger

import (
	"io"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// Init initializes the logger to log to the specified file and terminal.
func Init(logPath string) {
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(io.MultiWriter(f, os.Stdout))
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{"component", "category"},
		TimestampFormat: "15:04:05",
		NoColors:        true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}
