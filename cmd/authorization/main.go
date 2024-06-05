package main

import (
	"fmt"
	"log/slog"
	"os"
	"ozon-task/configs"
	"ozon-task/pkg/variables"
	delivery_grpc "ozon-task/services/authorization/delivery/grpc"
	delivery "ozon-task/services/authorization/delivery/http"
	"ozon-task/services/authorization/usecase"
)

// @title Authorization service
// @version 1.0
// @description Avito banner service

// @host localhost
// @BasePath /

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	logFile, err := os.Create("authorization.log")
	if err != nil {
		fmt.Println("Error creating log file")
		return
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	authAppConfig, err := configs.ReadAuthAppConfig()
	if err != nil {
		logger.Error(variables.ReadAuthConfigError, err.Error())
		return
	}

	relationalDataBaseConfig, err := configs.ReadRelationalAuthDataBaseConfig()
	if err != nil {
		logger.Error(variables.ReadAuthSqlConfigError, err.Error())
		return
	}

	cacheDatabaseConfig, err := configs.ReadCacheDatabaseConfig()
	if err != nil {
		logger.Error(variables.ReadAuthCacheConfigError, err.Error())
		return
	}

	core, err := usecase.GetCore(relationalDataBaseConfig, cacheDatabaseConfig, logger)
	if err != nil {
		logger.Error(variables.CoreInitializeError, err)
		return
	}

	grpcServer, err := delivery_grpc.NewServer(relationalDataBaseConfig, cacheDatabaseConfig, logger)
	if err != nil {
		logger.Error(variables.ListenAndServeError)
		return
	}

	api := delivery.GetAuthorizationApi(core, logger)

	errs := make(chan error, 2)
	go func() {
		errs <- api.ListenAndServe(authAppConfig)
	}()

	go func() {
		errs <- grpcServer.ListenAndServeGrpc()
	}()

	err = <-errs
	if err != nil {
		logger.Error(variables.ListenAndServeError, err.Error())
	}
}
