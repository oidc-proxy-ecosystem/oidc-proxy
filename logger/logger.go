package logger

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type ILogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Write(b []byte) (int, error)
}

type logger struct {
	logTpl   *template.Template
	timeFmt  string
	writer   io.Writer
	logLevel LogLevel
	skip     int
}

type LogTemplate struct {
	Time       string
	Funcname   string
	Filename   string
	Level      string
	Goroutine  string
	LineNumber string
	Message    string
}

var loggerContextKey LoggerKey

type LoggerKey struct{}

func NewContext(ctx context.Context, log ILogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) ILogger {
	l, ok := ctx.Value(loggerContextKey).(ILogger)
	if !ok {
		return Log
	}
	return l
}

type LogFormatType string

func (f LogFormatType) String() string {
	return string(f)
}

func ConvertLogFmt(fmt string) LogFormatType {
	switch strings.ToLower(fmt) {
	case "short":
		return FormatShort
	case "standard":
		return FormatStandard
	case "long":
		return FormatLong
	default:
		return FormatStandard
	}
}

const (
	// FormatShort time:{{.Time}}\tlevel:{{.Level}}\tmessage:{{.Message}}\n
	FormatShort LogFormatType = "time:{{.Time}}\tlevel:{{.Level}}\tmessage:{{.Message}}\n"
	// FormatStandard time:{{.Time}}\tlevel:{{.Level}}\tfilename:{{.Filename}}:{{.LineNumber}}\tmessage:{{.Message}}\n
	FormatStandard LogFormatType = "time:{{.Time}}\tlevel:{{.Level}}\tfilename:{{.Filename}}:{{.LineNumber}}\tmessage:{{.Message}}\n"
	// FormatLong time:{{.Time}}\tlevel:{{.Level}}\tfilename:{{.Filename}}:{{.LineNumber}}\tfuncname:{{.Funcname}}\tmessage:{{.Message}}\n
	FormatLong LogFormatType = "time:{{.Time}}\tlevel:{{.Level}}\tfilename:{{.Filename}}:{{.LineNumber}}\tfuncname:{{.Funcname}}\tmessage:{{.Message}}\n"
)

type TimeFormatType string

func (t TimeFormatType) String() string {
	return string(t)
}

func ConvertTimeFmt(fmt string) TimeFormatType {
	switch strings.ToLower(fmt) {
	case "date":
		return FormatDate
	case "datetime":
		return FormatDatetime
	case "millisec":
		return FormatMillisec
	default:
		return FormatDatetime
	}
}

const (
	// FormatDate 2006/01/02
	FormatDate TimeFormatType = "2006/01/02"
	// FormatDatetime 2006/01/02 15:04:05
	FormatDatetime TimeFormatType = "2006/01/02 15:04:05"
	// FormatMillisec 2006/01/02 15:04:05.000
	FormatMillisec TimeFormatType = "2006/01/02 15:04:05.000"
)

var Log ILogger = &logger{}
var _ io.Writer = &logger{}

func New(writer io.Writer, logLevel LogLevel, formatType LogFormatType, timeFmtTyp TimeFormatType) ILogger {
	t, err := template.New("log").Parse(formatType.String())
	if err != nil {
		panic(err)
	}
	return &logger{
		logTpl:   t,
		timeFmt:  timeFmtTyp.String(),
		writer:   writer,
		logLevel: logLevel,
		skip:     2,
	}
}

type LogLevel int

func (l LogLevel) String() string {
	switch l {
	case Critical:
		return "[CRIT]"
	case Error:
		return "[ERROR]"
	case Warn:
		return "[WARN]"
	case Info:
		return "[INFO]"
	case Debug:
		return "[DEBUG]"
	}
	return "[!!PANIC]"
}

func Convert(level string) LogLevel {
	switch strings.ToLower(level) {
	case "critical":
		return Critical
	case "error", "err":
		return Error
	case "warn", "warnign":
		return Warn
	case "info", "prod":
		return Info
	case "debug", "dev":
		return Debug
	default:
		return Info
	}
}

const (
	Critical LogLevel = iota
	Error
	Warn
	Info
	Debug
)

func (l *logger) isEnabledLevel(level LogLevel) bool {
	return level <= l.logLevel
}

func (l *logger) print(loglevel LogLevel, v ...interface{}) {
	if l.isEnabledLevel(loglevel) {
		pc, fileName, lineNumber, ok := runtime.Caller(2)
		if !ok {
			return
		}
		funcName := runtime.FuncForPC(pc).Name()
		funcName = funcName[strings.LastIndex(funcName, ".")+1:]
		fileName = fileName[strings.LastIndex(fileName, "/")+1:]
		now := time.Now().Format(l.timeFmt)
		print := fmt.Sprint(v...)
		prints := strings.Split(print, "\n")
		for _, p := range prints {
			if strings.TrimSpace(p) != "" {
				d := &LogTemplate{
					Time:       now,
					Funcname:   funcName,
					Filename:   fileName,
					Level:      loglevel.String(),
					LineNumber: strconv.Itoa(lineNumber),
					Goroutine:  strconv.Itoa(runtime.NumGoroutine()),
					Message:    p,
				}
				l.logTpl.Execute(l.writer, d)
			}
		}
	}
}

func (l *logger) Debug(v ...interface{})      { l.print(Debug, v...) }
func (l *logger) Info(v ...interface{})       { l.print(Info, v...) }
func (l *logger) Warning(v ...interface{})    { l.print(Warn, v...) }
func (l *logger) Error(v ...interface{})      { l.print(Error, v...) }
func (l *logger) Critical(v ...interface{})   { l.print(Critical, v...) }
func (l *logger) Write(b []byte) (int, error) { return l.writer.Write(b) }
func (l *logger) Copy() ILogger {
	return &logger{
		logTpl:   l.logTpl,
		timeFmt:  l.timeFmt,
		writer:   l.writer,
		logLevel: l.logLevel,
		skip:     l.skip + 1,
	}
}
