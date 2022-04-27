package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// Will be run before or after any other test functions.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
