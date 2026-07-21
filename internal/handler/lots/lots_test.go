package lots

import (
	"auction-house-lotTrio/internal/middleware"
	"auction-house-lotTrio/internal/service/auth"
	"testing"

	"github.com/gin-gonic/gin"
)

func testRouter(svc auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	r := gin.New()
	mw := middleware.NewMiddleware(svc)

}

func TestHandler_CloseExpiredLot(t *testing.T) {

}
