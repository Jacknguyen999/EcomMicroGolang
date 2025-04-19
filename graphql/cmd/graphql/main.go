package main

import (
	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/thomas/EcommerceAPI/graphql/graph"
	"github.com/thomas/EcommerceAPI/pkg/auth"
	"github.com/thomas/EcommerceAPI/pkg/middleware"
)

type AppConfig struct {
	AccountUrl    string `envconfig:"ACCOUNT_SERVICE_URL"`
	ProductUrl    string `envconfig:"PRODUCT_SERVICE_URL"`
	OrderUrl      string `envconfig:"ORDER_SERVICE_URL"`
	RecommenderUrl string `envconfig:"RECOMMENDER_SERVICE_URL"`
	SecretKey     string `envconfig:"SECRET_KEY"`
	Issuer        string `envconfig:"ISSUER"`
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	server, err := graph.NewGraphQLServer(cfg.AccountUrl, cfg.ProductUrl, cfg.OrderUrl, cfg.RecommenderUrl)
	if err != nil {
		log.Fatal(err)
	}

	srv := handler.New(server.ToExecutableSchema())
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	// srv.AddTransport(transport.Options{})
	// srv.AddTransport(transport.GET{})

	engine := gin.Default()

	engine.Use(middleware.GinContextToContextMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "It works",
		})
	})
	// Create a JWT service
	jwtService := auth.NewJwtService(cfg.SecretKey, cfg.Issuer)

	// Main GraphQL endpoint with authentication
	engine.POST("/graphql",
		middleware.AuthorizeJWT(jwtService),
		gin.WrapH(srv),
	)

	// Create a separate endpoint for the playground to use
	engine.POST("/graphql-playground", gin.WrapH(srv))

	// Set up the playground to use the unauthenticated endpoint
	// engine.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql-playground")))

	engine.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))

	log.Fatal(engine.Run(":8080"))
}
