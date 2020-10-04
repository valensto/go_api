package session

import (
	"context"
	"fmt"
)

type key int

var idKey key = 0

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, idKey, id)
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(idKey).(string)
	if !ok {
		return "", fmt.Errorf("no user find")
	}
	return userID, nil
}
