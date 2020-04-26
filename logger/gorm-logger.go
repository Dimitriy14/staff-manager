package logger

type gormLoggerImpl struct {
	log Logger
}

func NewGORMLogger(log Logger) *gormLoggerImpl {
	return &gormLoggerImpl{log: log}
}

func (gl *gormLoggerImpl) Print(v ...interface{}) {
	gl.log.Debug("", v...)
}
