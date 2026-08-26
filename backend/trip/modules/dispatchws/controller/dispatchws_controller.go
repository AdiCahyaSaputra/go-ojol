package controller

import (
	"log"
	"net/http"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type (
	DispatchWSController interface {
		ServeWS(ctx *gin.Context)
		ServeCustomerWS(ctx *gin.Context)
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
		},
	}
}

func (c *dispatchWSController) ServeWS(ctx *gin.Context) {
	c.serve(ctx, c.dispatchWSService.HandleConn)
}

func (c *dispatchWSController) ServeCustomerWS(ctx *gin.Context) {
	c.serve(ctx, c.dispatchWSService.HandleCustomerConn)
}

func (c *dispatchWSController) serve(ctx *gin.Context, handle func(userID string, conn *websocket.Conn)) {
	userID, _ := ctx.Get("user_id")
	id, _ := userID.(string)
	if id == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var respHeader http.Header
	if proto := ctx.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
		respHeader = http.Header{"Sec-WebSocket-Protocol": []string{proto}}
	}

	c.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true // Blindly trust it(?)
	}
	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, respHeader)
	if err != nil {
		log.Print(err)
		return
	}

	handle(id, conn)
}
