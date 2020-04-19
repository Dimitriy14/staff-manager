package logger

type elasticLoggerImpl struct {
	log Logger
}

func NewElasticLogger(log Logger) *elasticLoggerImpl {
	return &elasticLoggerImpl{log: log}
}

func (gl *elasticLoggerImpl) Printf(format string, v ...interface{}) {
	gl.log.Errorf("", format, v)
}
