package log

func SetLogger(logger ILogger) {
	if logger == nil {
		return
	}
	_logger = logger
}

func GetZap() {

}

func GetLogger() ILogger { return _logger }

// Print 打印日志，不含堆栈信息
func Print(level Level, a ...interface{}) {
	if _logger != nil {
		_logger.Print(level, a...)
	}
}

// Printf 打印模板日志，不含堆栈信息
func Printf(level Level, format string, a ...interface{}) {
	if _logger != nil {
		_logger.Printf(level, format, a...)
	}
}

// Debug 打印调试日志
func Debug(a ...interface{}) {
	if _logger != nil {
		_logger.Debug(a...)
	}
}

// Debugf 打印调试模板日志
func Debugf(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Debugf(format, a...)
	}
}

// Info 打印信息日志
func Info(a ...interface{}) {
	if _logger != nil {
		_logger.Info(a...)
	}
}

// Infof 打印信息模板日志
func Infof(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Infof(format, a...)
	}
}

// Warn 打印警告日志
func Warn(a ...interface{}) {
	if _logger != nil {
		_logger.Warn(a...)
	}
}

// Warnf 打印警告模板日志
func Warnf(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Warnf(format, a...)
	}
}

// Error 打印错误日志
func Error(a ...interface{}) {
	if _logger != nil {
		_logger.Error(a...)
	}
}

// Errorf 打印错误模板日志
func Errorf(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Errorf(format, a...)
	}
}

// Fatal 打印致命错误日志
func Fatal(a ...interface{}) {
	if _logger != nil {
		_logger.Fatal(a...)
	}
}

// Fatalf 打印致命错误模板日志
func Fatalf(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Fatalf(format, a...)
	}
}

// Panic 打印Panic日志
func Panic(a ...interface{}) {
	if _logger != nil {
		_logger.Panic(a...)
	}
}

// Panicf 打印Panic模板日志
func Panicf(format string, a ...interface{}) {
	if _logger != nil {
		_logger.Panicf(format, a...)
	}
}

// Close 关闭日志
func Close() {
	if _logger != nil {
		_ = _logger.Close()
	}
}
