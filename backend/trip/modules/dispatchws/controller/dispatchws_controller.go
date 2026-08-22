package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type (
	DispatchWSController interface {
		ServeWS(ctx *gin.Context)
	}

	dispatchWSController struct {
		dispatchWSService service.DispatchWSService
		upgrader          websocket.Upgrader
	}
)

func NewDispatchWSController(s service.DispatchWSService) DispatchWSController {
	return &dispatchWSController{
		dispatchWSService: s,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     checkOrigin,
		},
	}
}

func (c *dispatchWSController) ServeWS(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	id, _ := userID.(string)
	if id == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	c.dispatchWSService.HandleConn(id, conn)
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	for allowed := range strings.SplitSeq(raw, ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
