package log

var (
	DefaultLogger Logger
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

func SetLogger(l Logger) {
	DefaultLogger = l
}

func Debugf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debugf(format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Infof(format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warnf(format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Errorf(format, v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Fatalf(format, v...)
	}
}
