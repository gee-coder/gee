package error

type Func func(geeError *GeeError)

type GeeError struct {
	err    error
	ErrFuc Func
}

func Default() *GeeError {
	return &GeeError{}
}

func (e *GeeError) Put(err error) {
	if err != nil {
		e.err = err
		panic(e)
	}
}

func (e *GeeError) Error() string {
	return e.err.Error()
}

// 暴露一个方法 让用户自定义
func (e *GeeError) Result(errFuc Func) {
	e.ErrFuc = errFuc
}

func (e *GeeError) ExecResult() {
	e.ErrFuc(e)
}
