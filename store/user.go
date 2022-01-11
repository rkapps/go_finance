package store

import "context"

type ctxToken int

const (
	UserContextUID ctxToken = iota
)

//User holds meta data for the firebase user
type User struct {
	UID string
}

func UserFromCtx(ctx context.Context) User {
	return ctx.Value(UserContextUID).(User)
}
