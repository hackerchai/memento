package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hackerchai/memento/pkg/xlog"
)

type HelloService struct {
	logger *xlog.Logger
}

func NewHelloService(logger *xlog.Logger) *HelloService {
	return &HelloService{
		logger: logger,
	}
}

func (s *HelloService) Greet(ctx context.Context, name string) string {
	//no need to get reqID from context, xlog will get it from context automatically
	//reqID := s.logger.ReqIDFromContext(ctx)

	s.logger.InfoX(ctx).Str("user", name).Msg("Service: Greeting user")
	return "Hello " + name
}

func main() {
	defaultLogger := xlog.NewDefaultLogger()

	r := gin.Default()

	r.Use(loggerMiddleware(defaultLogger))

	helloService := NewHelloService(defaultLogger)

	r.GET("/hello", func(c *gin.Context) {
		helloHandler(c, helloService)
	})

	r.Run(":8080")
}

func loggerMiddleware(logger *xlog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get reqID from header or generate new one
		reqID := c.GetHeader(string(xlog.DefaultReqIDKey))
		if reqID == "" {
			reqID = xlog.GenReqId()
		}

		// Set reqID to context
		ctx := logger.SetReqIDToContext(c.Request.Context(), reqID)
		// Set reqID to gin context for other middleware to use
		c.Set(string(xlog.DefaultReqIDKey), reqID)

		// Set context to request
		c.Request = c.Request.WithContext(logger.WithContext(ctx))

		c.Next()
	}
}

func helloHandler(c *gin.Context, service *HelloService) {
	ctx := c.Request.Context()

	// get logger from context
	logger := xlog.Ctx(ctx)
	logger.InfoX(ctx).Msg("helloHandler called")

	name := c.DefaultQuery("name", "Guest")
	greeting := service.Greet(ctx, name)

	reqID := logger.ReqIDFromContext(ctx)
	// Or get reqID from gin context
	//reqID = c.GetString(string(xlog.DefaultReqIDKey))

	c.JSON(http.StatusOK, gin.H{
		"message": greeting,
		"reqID":   reqID,
	})
}
