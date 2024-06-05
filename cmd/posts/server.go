package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"ozon-task/configs"
	"ozon-task/pkg/middleware"
	"ozon-task/pkg/variables"
	"ozon-task/services/posts/delivery/graph"
	"ozon-task/services/posts/usecase"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8081"

func main() {
	logFile, err := os.Create("posts.log")
	if err != nil {
		fmt.Println("Error creating log file")
		return
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	postsAppConfig, err := configs.ReadPostsAppConfig()
	if err != nil {
		logger.Error(variables.ReadAuthConfigError, err.Error())
		return
	}

	relationalDataBaseConfig, err := configs.ReadRelationalPostsDataBaseConfig()
	if err != nil {
		logger.Error(variables.ReadAuthSqlConfigError, err.Error())
		return
	}

	cacheDatabaseConfig, err := configs.ReadCacheDatabaseConfig()
	if err != nil {
		logger.Error(variables.ReadAuthCacheConfigError, err.Error())
		return
	}
	grpcCfg, err := configs.ReadGrpcConfig()
	if err != nil {
		logger.Error("failed to parse grpc configs file: %s", err.Error())
		return
	}
	core, err := usecase.GetCore(relationalDataBaseConfig, cacheDatabaseConfig, grpcCfg, postsAppConfig.InMemory, logger)
	if err != nil {
		logger.Error(variables.CoreInitializeError, err)
		return
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	resolver := &graph.Resolver{
		Core: core,
		Log:  logger,
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", middleware.AuthorizationMiddleware(srv, core, logger))

	log.Printf("Server Post with GraphQL running on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
