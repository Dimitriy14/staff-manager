package logger

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	UseFile  bool   `json:"UseFile"`
	FileName string `json:"FileName"`
	LogLevel string `json:"LogLevel"`
}

type Logger interface {
	Debug(txID string, v ...interface{})
	Debugf(txID, format string, v ...interface{})

	Info(txID string, v ...interface{})
	Infof(txID, format string, v ...interface{})

	Warn(txID string, v ...interface{})
	Warnf(txID, format string, v ...interface{})

	Error(txID string, v ...interface{})
	Errorf(txID, format string, v ...interface{})

	Close() error
}

// Load loads logger
func Load(cfg Config) (Logger, error) {
	var (
		output io.WriteCloser
		err    error
	)
	output = os.Stdout

	if cfg.UseFile {
		output, err = os.OpenFile(cfg.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.Wrap(err, "opening log file")
		}
	}

	logLvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "parsing log level")
	}

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(output)
	log.SetLevel(logLvl)

	return &loggerImpl{
		f:   output,
		log: log,
	}, err
}

type loggerImpl struct {
	f   io.Closer
	log *logrus.Logger
}

func (l *loggerImpl) Debug(txID string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Debug(v...)
}

func (l *loggerImpl) Debugf(txID, format string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Debugf(format, v...)
}

func (l *loggerImpl) Info(txID string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Info(v...)
}

func (l *loggerImpl) Infof(txID, format string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Infof(format, v...)
}

func (l *loggerImpl) Warn(txID string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Warn(v...)
}

func (l *loggerImpl) Warnf(txID, format string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Warnf(format, v...)
}

func (l *loggerImpl) Error(txID string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Error(v...)
}

func (l *loggerImpl) Errorf(txID, format string, v ...interface{}) {
	l.log.WithFields(logrus.Fields{"txID": txID}).Errorf(format, v...)
}

func (l *loggerImpl) Close() error {
	return l.f.Close()
}
