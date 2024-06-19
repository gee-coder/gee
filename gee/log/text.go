package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (f *TextFormatter) Format(param *LoggingFormatParam) string {
	now := time.Now()
	var builderField strings.Builder
	var fieldsDisplay = ""
	if param.LoggerFields != nil {
		fieldsDisplay = "| fields: "
		num := len(param.LoggerFields)
		count := 0
		for k, v := range param.LoggerFields {
			fmt.Fprintf(&builderField, "%s=%v", k, v)
			if count < num-1 {
				fmt.Fprintf(&builderField, ",")
				count++
			}
		}
	}
	msgKey := "\n msg: "
	var sb strings.Builder
	if param.Level == LevelError {
		msgKey = "\n Error Cause By: "
		var pcs [32]uintptr
		n := runtime.Callers(5, pcs[:])
		for _, pc := range pcs[:n] {
			fn := runtime.FuncForPC(pc)
			line, l := fn.FileLine(pc)
			sb.WriteString(fmt.Sprintf("\n\t%s:%d", line, l))
		}
	}
	if param.IsColor {
		// 要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := f.LevelColor(param.Level)
		msgColor := f.MsgColor(param.Level)
		return fmt.Sprintf("%s [gee] %s %s%v%s | level= %s %s %s %s%s %v %s %s %s%s \n",
			yellow, reset, blue, now.Format("2006-01-02 15:04:05"), reset,
			levelColor, param.Level.Level(), reset, msgColor, msgKey, param.Msg, reset, fieldsDisplay,
			builderField.String(), sb.String(),
		)
	}
	return fmt.Sprintf("[gee] %v | level=%s  %s  %v %s %s%s\n",
		now.Format("2006-01-02 15:04:05"),
		param.Level.Level(), msgKey, param.Msg, fieldsDisplay, builderField.String(), sb.String(),
	)
}

func (f *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
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

func (f *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
	case LevelError:
		return red
	default:
		return ""
	}
}
