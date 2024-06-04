package delivery

import (
	"context"
	"errors"
	"log/slog"
	"ozon-task/pkg/models"
)

type ICore interface {
	GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error)
	GetPostByID(ctx context.Context, id int) (*models.Post, error)
	GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error)
	AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error)
	AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error)
}

type Resolver struct {
	core   ICore
	logger *slog.Logger
}

func (r *Resolver) GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error) {
	// Реализуйте логику для получения постов
	return nil, errors.New("Not implemented")
}

func (r *Resolver) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	// Реализуйте логику для получения поста по ID
	return nil, errors.New("Not implemented")
}

func (r *Resolver) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error) {
	// Реализуйте логику для получения комментариев по ID поста
	return nil, errors.New("Not implemented")
}

func (r *Resolver) AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error) {
	// Реализуйте логику для добавления поста
	return nil, errors.New("Not implemented")
}

func (r *Resolver) AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error) {
	// Реализуйте логику для добавления комментария
	return nil, errors.New("Not implemented")
}
