package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"ozon-task/pkg/models"
	"ozon-task/pkg/variables"
	"ozon-task/services/authorization/proto/authorization"
	inmemory_repository "ozon-task/services/posts/repository/inMemory"
	relational_repository "ozon-task/services/posts/repository/relational"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IRepository interface {
	GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error)
	GetPostByID(ctx context.Context, id int) (*models.Post, error)
	GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error)
	AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error)
	AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error)
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

func (core *Core) GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error) {
	posts, err := core.postsRepository.GetPosts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return posts, nil
}

func (core *Core) GetPostByID(ctx context.Context, id int, limit int, offset int) (*models.Post, error) {
	post, err := core.postsRepository.GetPostByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return post, nil
}

func (core *Core) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error) {
	comments, err := core.postsRepository.GetCommentsByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Comments Not Founded %v", err))
	}
	return comments, nil
}

func (core *Core) AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error) {
	post, err := core.postsRepository.AddPost(ctx, data, userId, isCommented)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}
	return post, nil
}

func (core *Core) AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error) {
	post, err := core.postsRepository.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Posts Not Founded %v", err))
	}

	if !post.CommentsAllowed {
		return nil, fmt.Errorf("Post can't be commented")
	}

	comment, err := core.postsRepository.AddComment(ctx, postID, userId, data, parentID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Comments Not Founded %v", err))
	}
	return comment, nil
}
