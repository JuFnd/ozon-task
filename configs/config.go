package configs

import (
	"flag"
	"fmt"
	"os"
	"ozon-task/pkg/variables"
	"syscall"

	"gopkg.in/yaml.v2"
)

func readYAMLFile[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	var config T
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	return &config, nil
}

func ParseFlagsAndReadYAMLFile[T any](fileName string, defaultFilePath string, flags *flag.FlagSet) (*T, error) {
	flag.Parse()
	var path string
	flag.StringVar(&path, fileName, defaultFilePath, "Путь к конфигу"+fileName)

	config, err := readYAMLFile[T](path)
	if err == syscall.ENOENT {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to parse %s from provided path: %v", fileName, err))
	}

	return config, nil
}

func ReadAuthAppConfig() (*variables.AppConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.AppConfig]("auth_config_path", "configs/AuthorizationAppConfig.yml", flag.CommandLine)
}

func ReadGrpcConfig() (*variables.GrpcConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.GrpcConfig]("grpc_config_path", "configs/GrpcConfig.yml", flag.CommandLine)
}

func ReadPostsAppConfig() (*variables.AppConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.AppConfig]("posts_config_path", "configs/PostsAppConfig.yml", flag.CommandLine)
}

func ReadRelationalAuthDataBaseConfig() (*variables.RelationalDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.RelationalDataBaseConfig]("sql_config_auth_path", "configs/AuthorizationSqlDataBaseConfig.yml", flag.CommandLine)
}

func ReadRelationalPostsDataBaseConfig() (*variables.RelationalDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.RelationalDataBaseConfig]("sql_config_films_path", "configs/PostsSqlDataBaseConfig.yml", flag.CommandLine)
}

func ReadCacheDatabaseConfig() (*variables.CacheDataBaseConfig, error) {
	return ParseFlagsAndReadYAMLFile[variables.CacheDataBaseConfig]("cache_config_path", "configs/AuthorizationCacheDataBaseConfig.yml", flag.CommandLine)
}
