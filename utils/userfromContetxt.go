package utils

import (
	"context"
	"naevis/globals"
)

func GetUserIDFromContext(ctx context.Context) string {
	requestingUserID, ok := ctx.Value(globals.UserIDKey).(string)
	if !ok || requestingUserID == "" {
		return ""
	}
	return requestingUserID
}
