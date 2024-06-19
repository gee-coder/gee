package gee

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
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

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

type LoggerFormatter = func(params *LogFormatterParams) string

var defaultFormatter = func(params *LogFormatterParams) string {
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	if params.IsDisplayColor {
		return fmt.Sprintf("%s[gee]%s|%s%v%s|%s%3d%s| %s%13v%s | %15s | %s %-7s %s %s %#v %s \n",
			yellow, reset, blue, params.TimeStamp.Format("2006-01-02 15:04:05"), reset,
			params.StatusCodeColor(), params.StatusCode, reset,
			red, params.Latency, reset,
			params.ClientIP,
			magenta, params.Method, reset,
			cyan, params.Path, reset,
		)
	}
	return fmt.Sprintf("[gee] %v | %3d | %13v | %15s |%-7s %#v",
		params.TimeStamp.Format("2006-01-02 15:04:05"),
		params.StatusCode,
		params.Latency, params.ClientIP, params.Method, params.Path,
	)
}

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
	IsColor   bool
}

func LoggingWithConfig(conf LoggingConfig, next HandlerFunc) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}
	out := conf.out
	displayColor := false
	if out == nil {
		out = os.Stdout
		displayColor = true
	}
	return func(ctx *Context) {
		r := ctx.R
		param := &LogFormatterParams{
			Request:        r,
			IsDisplayColor: displayColor,
		}
		start := time.Now()
		param.Path = r.URL.Path
		raw := r.URL.RawQuery
		if raw != "" {
			param.Path = param.Path + "?" + raw
		}
		next(ctx)
		param.TimeStamp = time.Now()
		param.Latency = time.Now().Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		param.ClientIP = net.ParseIP(ip)
		param.Method = r.Method
		param.StatusCode = ctx.StatusCode
		_, err := fmt.Fprint(out, formatter(param))
		if err != nil {
			log.Println(err)
		}
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
