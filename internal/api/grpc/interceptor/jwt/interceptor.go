package jwt

import (
	"context"

	"github.com/robinlg/notification-platform/internal/errs"
)

func GetBizIDFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(BizIDName)
	if val == nil {
		return 0, errs.ErrBizNotFound
	}
	v, ok := val.(int64)
	if !ok {
		return 0, errs.ErrBizNotFound
	}
	return v, nil
}
