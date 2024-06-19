package gee

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	geeError "github.com/gee-coder/gee/error"
)

func detailMsg(err any) string {
	var pcs [32]uintptr
	// 栈帧从当前函数的上上上一层开始
	n := runtime.Callers(3, pcs[:])
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v", err))
	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

func Recovery(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				err2 := err.(error)
				if err2 != nil {
					var gError *geeError.GeeError
					if errors.As(err2, &gError) {
						gError.ExecResult()
						return
					}
				}
				ctx.Logger.Error(detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		next(ctx)
	}
}
