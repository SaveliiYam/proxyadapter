package jwt

import "context"

// AccessParser проверяет Bearer access JWT (реализация — вызов jwt_api, локальный parse и т.д.).
type AccessParser interface {
	ParseAccess(ctx context.Context, token string) (userUUID string, err error)
}
