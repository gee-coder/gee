package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gee-coder/gee/internal/msstrings"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

// 级别
type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

type Fields map[string]any

type LoggerWriter struct {
	Level LoggerLevel
	Out   io.Writer
}

type LoggingFormatParam struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
	Msg          any
}

type LoggingFormatter interface {
	Format(param *LoggingFormatParam) string
}

type Logger struct {
	Formatter    LoggingFormatter
	Level        LoggerLevel
	Outs         []*LoggerWriter
	LoggerFields Fields
	logPath      string
	LogFileSize  int64
}

func New() *Logger {
	return &Logger{}
}

func Default() *Logger {
	logger := New()
	logger.Formatter = &TextFormatter{}
	logger.Level = LevelDebug
	w := &LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
	return logger
}

func FileWriter(name string) io.Writer {
	w, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return w
}

func (l *Logger) CloseWriter() {
	for _, out := range l.Outs {
		file := out.Out.(*os.File)
		if file != nil {
			_ = file.Close()
		}
	}
}

func (l *Logger) CheckFileSize(w *LoggerWriter) {
	// 判断对应的文件大小
	logFile := w.Out.(*os.File)
	if logFile != nil {
		stat, err := logFile.Stat()
		if err != nil {
			log.Println(err)
			return
		}
		size := stat.Size()
		// 这里要检查大小，如果满足条件 就重新创建文件，并且更换logger中的输出
		if l.LogFileSize <= 0 {
			// 默认100M
			l.LogFileSize = 100 << 20
		}
		if size >= l.LogFileSize {
			_, name := path.Split(stat.Name())
			fileName := name[0:strings.Index(name, ".")]
			writer := FileWriter(path.Join(l.logPath, msstrings.JoinStrings(fileName, ".", time.Now().UnixMilli(), ".log")))
			w.Out = writer
		}
	}

}

func (l *Logger) Print(level LoggerLevel, msg any) {
	if l.Level > level {
		// 级别不满足 不打印日志
		return
	}
	param := &LoggingFormatParam{
		Level:        level,
		Msg:          msg,
		LoggerFields: l.LoggerFields,
	}
	for _, out := range l.Outs {
		if out.Out == os.Stdout {
			param.IsColor = true
			l.print(param, out)
		}
		if out.Level == -1 || out.Level == level {
			param.IsColor = false
			l.print(param, out)
			l.CheckFileSize(out)
		}
	}
}

func (l *Logger) print(param *LoggingFormatParam, out *LoggerWriter) {
	formatter := l.Formatter.Format(param)
	fmt.Fprintln(out.Out, formatter)
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		Level:        l.Level,
		LoggerFields: fields,
	}
}

func (l *Logger) SetLogPath(logPath string) {
	l.logPath = logPath
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: -1,
		Out:   FileWriter(path.Join(logPath, "all.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelDebug,
		Out:   FileWriter(path.Join(logPath, "debug.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelInfo,
		Out:   FileWriter(path.Join(logPath, "info.log")),
	})
	l.Outs = append(l.Outs, &LoggerWriter{
		Level: LevelError,
		Out:   FileWriter(path.Join(logPath, "error.log")),
	})
}

type LoggerFormatter struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
}

func (f *LoggerFormatter) LevelColor() string {
	switch f.Level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *LoggerFormatter) MsgColor() string {
	switch f.Level {
	case LevelError:
		return red
	default:
		return ""
	}
}

func (f *LoggerFormatter) format(msg any) string {
	now := time.Now()
	if f.IsColor {
		// 要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [gee] %s %s%v%s | level= %s %s %s | msg=%s %#v %s | fields=%v ",
			yellow, reset, blue, now.Format("2006-01-02 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset, f.LoggerFields,
		)
	}
	return fmt.Sprintf("[gee] %v | level=%s | msg=%#v | fields=%#v",
		now.Format("2006-01-02 15:04:05"),
		f.Level.Level(), msg, f.LoggerFields)
}
