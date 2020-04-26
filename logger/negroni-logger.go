package logger

type negroniLoggerImpl struct {
	log Logger
}

func NewNegroniLogger(log Logger) *negroniLoggerImpl {
	return &negroniLoggerImpl{log: log}
}

func (gl *negroniLoggerImpl) Println(v ...interface{}) {
	gl.log.Debug("", v...)
}

func (gl *negroniLoggerImpl) Printf(format string, v ...interface{}) {
	gl.log.Debugf("", format, v...)
}
