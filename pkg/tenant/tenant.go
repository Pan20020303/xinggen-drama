package tenant

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const contextUserIDKey = "user_id"

func GetUserID(c *gin.Context) (uint, error) {
	v, ok := c.Get(contextUserIDKey)
	if !ok {
		return 0, fmt.Errorf("missing user id in context")
	}
	uid, ok := v.(uint)
	if !ok || uid == 0 {
		return 0, fmt.Errorf("invalid user id in context")
	}
	return uid, nil
}
