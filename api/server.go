package api

import (
	"fmt"

	db "github.com/KamisAyaka/simplebank/db/sqlc"
	"github.com/KamisAyaka/simplebank/token"
	"github.com/KamisAyaka/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker:%w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
	server.setipRouter()

	return server, nil
}

func (server *Server) setipRouter() {
	router := gin.Default()
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.DELETE("/accounts/:id", server.deleteAccount)
	authRoutes.GET("/accounts", server.listAccount)

	authRoutes.GET("/entries/:id", server.getEntry)
	authRoutes.GET("/entries", server.listEntries)

	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.GET("/transfers/:id", server.getTransfer)
	authRoutes.GET("/transfers", server.listTransfers)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
