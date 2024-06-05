package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"ozon-task/pkg/variables"
	"ozon-task/services/authorization/proto/authorization"
	"ozon-task/services/posts/delivery/graph/model"
	inmemory_repository "ozon-task/services/posts/repository/inMemory"
	relational_repository "ozon-task/services/posts/repository/relational"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IRepository interface {
	GetPosts(ctx context.Context, limit int, offset int) ([]*model.Post, error)
	GetPostByID(ctx context.Context, id int) (*model.Post, error)
	GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error)
	AddPost(ctx context.Context, data string, user *model.User, isCommented bool) (*model.Post, error)
	AddComment(ctx context.Context, post *model.Post, user *model.User, data string, parentID int) (*model.Comment, error)
}

type Core struct {
	postsRepository IRepository
	logger          *slog.Logger
	client          authorization.AuthorizationClient
}

func GetClient(address string) (authorization.AuthorizationClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc connect err: %w", err)
	}
	client := authorization.NewAuthorizationClient(conn)

	return client, nil
}

func GetCore(postsRelConfig *variables.RelationalDataBaseConfig, postsCacheConfig *variables.CacheDataBaseConfig, grpcCfg *variables.GrpcConfig, inMemory bool, logger *slog.Logger) (*Core, error) {
	var repository IRepository
	var err error
	if inMemory {
		repository, err = inmemory_repository.GetPostsRepository(postsCacheConfig, logger)
	} else {
		repository, err = relational_repository.GetPostsRepository(postsRelConfig, logger)
	}

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Repository can't create %v", err))
	}

	postsGrpcClient, err := GetClient(grpcCfg.Address + ":" + grpcCfg.Port)

	if err != nil {
		return nil, fmt.Errorf("grpc connect err: %w", err)
	}

	return &Core{
		postsRepository: repository,
		logger:          logger,
		client:          postsGrpcClient,
	}, nil
}

func (core *Core) GetPosts(ctx context.Context, limit int, offset int) ([]*model.Post, error) {
	posts, err := core.postsRepository.GetPosts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return posts, nil
}

func (core *Core) GetPostByID(ctx context.Context, id int, limit int, offset int) (*model.Post, error) {
	post, err := core.postsRepository.GetPostByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return post, nil
}

func (core *Core) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error) {
	comments, err := core.postsRepository.GetCommentsByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Comments Not Founded %v", err))
	}
	return comments, nil
}

func (core *Core) AddPost(ctx context.Context, data string, userId int, isCommented bool) (*model.Post, error) {
	user := model.User{ID: strconv.Itoa(userId)}
	post, err := core.postsRepository.AddPost(ctx, data, &user, isCommented)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return post, nil
}

func (core *Core) AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*model.Comment, error) {
	post, err := core.postsRepository.GetPostByID(ctx, postID)
	user := model.User{ID: strconv.Itoa(userId)}
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}

	if !*post.IsCommented {
		return nil, fmt.Errorf("Post can't be commented")
	}

	comment, err := core.postsRepository.AddComment(ctx, post, &user, data, parentID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Comments Not Founded %v", err))
	}
	return comment, nil
}

func (core *Core) GetUserRole(ctx context.Context, id int64) (string, error) {
	grpcRequest := authorization.RoleRequest{Id: id}

	grpcResponse, err := core.client.GetRole(ctx, &grpcRequest)
	if err != nil {
		core.logger.Error(variables.GrpcRecievError, err)
		return "", fmt.Errorf(variables.GrpcRecievError, err)
	}
	return grpcResponse.GetRole(), nil
}

func (core *Core) GetUserId(ctx context.Context, sid string) (int64, error) {
	grpcRequest := authorization.FindIdRequest{Sid: sid}

	grpcResponse, err := core.client.GetId(ctx, &grpcRequest)
	if err != nil {
		core.logger.Error(variables.GrpcRecievError, err)
		return 0, fmt.Errorf(variables.GrpcRecievError, err)
	}
	return grpcResponse.Value, nil
}
