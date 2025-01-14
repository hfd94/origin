package log

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log/internal/encoder"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var levelMap map[zapcore.Level]Level

func init() {
	levelMap = map[zapcore.Level]Level{
		zap.DebugLevel:  DebugLevel,
		zap.InfoLevel:   InfoLevel,
		zap.WarnLevel:   WarnLevel,
		zap.ErrorLevel:  ErrorLevel,
		zap.FatalLevel:  FatalLevel,
		zap.DPanicLevel: PanicLevel,
	}
}

type Logger struct {
	logger *zap.SugaredLogger
	opts   *options
}

func NewLogger(opts ...Option) *Logger {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var (
		fileEncoder     zapcore.Encoder
		terminalEncoder zapcore.Encoder
	)
	switch o.format {
	case JsonFormat:
		fileEncoder = encoder.NewJsonEncoder(o.timeFormat, o.callerFullPath)
		terminalEncoder = fileEncoder
	default:
		fileEncoder = encoder.NewTextEncoder(o.timeFormat, o.callerFullPath, false)
		terminalEncoder = encoder.NewTextEncoder(o.timeFormat, o.callerFullPath, true)
	}

	options := make([]zap.Option, 0, 3)
	options = append(options, zap.AddCaller())
	switch o.stackLevel {
	case DebugLevel:
		options = append(options, zap.AddStacktrace(zapcore.DebugLevel), zap.AddCallerSkip(1+o.callerSkip))
	case InfoLevel:
		options = append(options, zap.AddStacktrace(zapcore.InfoLevel), zap.AddCallerSkip(1+o.callerSkip))
	case WarnLevel:
		options = append(options, zap.AddStacktrace(zapcore.WarnLevel), zap.AddCallerSkip(1+o.callerSkip))
	case ErrorLevel:
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(1+o.callerSkip))
	case FatalLevel:
		options = append(options, zap.AddStacktrace(zapcore.FatalLevel), zap.AddCallerSkip(1+o.callerSkip))
	case PanicLevel:
		options = append(options, zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(1+o.callerSkip))
	}

	l := &Logger{opts: o}

	var cores []zapcore.Core
	if o.file != "" {
		if o.classifiedStorage {
			cores = append(cores,
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(DebugLevel), l.buildLevelEnabler(DebugLevel)),
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(InfoLevel), l.buildLevelEnabler(InfoLevel)),
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(WarnLevel), l.buildLevelEnabler(WarnLevel)),
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(ErrorLevel), l.buildLevelEnabler(ErrorLevel)),
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(FatalLevel), l.buildLevelEnabler(FatalLevel)),
				zapcore.NewCore(fileEncoder, l.buildWriteSyncer(PanicLevel), l.buildLevelEnabler(PanicLevel)),
			)
		} else {
			cores = append(cores, zapcore.NewCore(fileEncoder, l.buildWriteSyncer(NoneLevel), l.buildLevelEnabler(NoneLevel)))
		}
	}

	if o.stdout {
		cores = append(cores, zapcore.NewCore(terminalEncoder, zapcore.AddSync(os.Stdout), l.buildLevelEnabler(NoneLevel)))
	}

	if len(cores) >= 0 {
		l.logger = zap.New(zapcore.NewTee(cores...), options...).Sugar()
	}

	return l
}

func (l *Logger) buildWriteSyncer(level Level) zapcore.WriteSyncer {
	writer, err := NewWriter(WriterOptions{
		Path:    l.opts.file,
		Level:   level,
		nodeId:  l.opts.nodeId,
		MaxAge:  l.opts.fileMaxAge,
		MaxSize: l.opts.fileMaxSize * 1024 * 1024,
		CutRule: l.opts.fileCutRule,
	})
	if err != nil {
		panic(err)
	}

	return zapcore.AddSync(writer)
}

func (l *Logger) buildLevelEnabler(level Level) zapcore.LevelEnabler {
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if v := levelMap[lvl]; l.opts.level != NoneLevel {
			return v >= l.opts.level && (level == NoneLevel || (level >= l.opts.level && v >= level))
		} else {
			return level == NoneLevel || v >= level
		}
	})
}

// 打印日志
func (l *Logger) print(level Level, stack bool, a ...interface{}) {
	if l.logger == nil {
		return
	}

	var msg string
	if len(a) == 1 {
		if str, ok := a[0].(string); ok {
			msg = str
		} else {
			msg = fmt.Sprint(a...)
		}
	} else {
		msg = fmt.Sprint(a...)
	}

	switch level {
	case DebugLevel:
		l.logger.Debugw(msg, encoder.StackFlag, stack)
	case InfoLevel:
		l.logger.Infow(msg, encoder.StackFlag, stack)
	case WarnLevel:
		l.logger.Warnw(msg, encoder.StackFlag, stack)
	case ErrorLevel:
		l.logger.Errorw(msg, encoder.StackFlag, stack)
	case FatalLevel:
		l.logger.Fatalw(msg, encoder.StackFlag, stack)
	case PanicLevel:
		l.logger.DPanicw(msg, encoder.StackFlag, stack)
	}
}

// Print 打印日志，不打印堆栈信息
func (l *Logger) Print(level Level, a ...interface{}) {
	l.print(level, false, a...)
}

// Printf 打印模板日志，不打印堆栈信息
func (l *Logger) Printf(level Level, format string, a ...interface{}) {
	l.print(level, false, fmt.Sprintf(format, a...))
}

// Debug 打印调试日志
func (l *Logger) Debug(a ...interface{}) {
	l.print(DebugLevel, true, a...)
}

// Debugf 打印调试模板日志
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.print(DebugLevel, true, fmt.Sprintf(format, a...))
}

// Info 打印信息日志
func (l *Logger) Info(a ...interface{}) {
	l.print(InfoLevel, true, a...)
}

// Infof 打印信息模板日志
func (l *Logger) Infof(format string, a ...interface{}) {
	l.print(InfoLevel, true, fmt.Sprintf(format, a...))
}

// Warn 打印警告日志
func (l *Logger) Warn(a ...interface{}) {
	l.print(WarnLevel, true, a...)
}

// Warnf 打印警告模板日志
func (l *Logger) Warnf(format string, a ...interface{}) {
	l.print(WarnLevel, true, fmt.Sprintf(format, a...))
}

// Error 打印错误日志
func (l *Logger) Error(a ...interface{}) {
	l.print(ErrorLevel, true, a...)
}

// Errorf 打印错误模板日志
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.print(ErrorLevel, true, fmt.Sprintf(format, a...))
}

// Fatal 打印致命错误日志
func (l *Logger) Fatal(a ...interface{}) {
	l.print(FatalLevel, true, a...)
}

// Fatalf 打印致命错误模板日志
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.print(FatalLevel, true, fmt.Sprintf(format, a...))
}

// Panic 打印Panic日志
func (l *Logger) Panic(a ...interface{}) {
	l.print(PanicLevel, true, a...)
}

// Panicf 打印Panic模板日志
func (l *Logger) Panicf(format string, a ...interface{}) {
	l.print(PanicLevel, true, fmt.Sprintf(format, a...))
}

// Sync 同步缓存中的日志
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Close 关闭日志
func (l *Logger) Close() error {
	return l.logger.Sync()
}
