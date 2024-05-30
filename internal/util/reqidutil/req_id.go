package reqidutil

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

func GetOrNewRequestId(ctx context.Context) string {
	if ctx != nil {
		if curr, ok := ctx.Value(ContextKeyRequestId).(string); ok {
			return curr
		}
	}
	return newRequestId()
}

func newRequestId() string {
	uid, _ := uuid.NewRandom()
	xreqid := uid.String()
	xreqid = strings.ReplaceAll(xreqid, "-", "")
	return xreqid
}
