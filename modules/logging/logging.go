package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func Setup(level, path string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
	return nil
}
