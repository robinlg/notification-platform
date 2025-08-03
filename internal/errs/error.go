package errs

import "errors"

// 定义统一的错误类型
var (
	// 业务错误
	ErrBizNotFound = errors.New("BizID不存在")
)
