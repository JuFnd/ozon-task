package graph

import (
	"context"
	"fmt"
	"log/slog"
	"ozon-task/pkg/variables"
	"ozon-task/services/posts/delivery/graph/model"
	"strconv"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type ICore interface {
	GetPosts(ctx context.Context, limit int, offset int) ([]*model.Post, error)
	GetPostByID(ctx context.Context, id int, limit int, offset int) (*model.Post, error)
	GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error)
	AddPost(ctx context.Context, data string, userId int, isCommented bool) (*model.Post, error)
	AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*model.Comment, error)
}

type Resolver struct {
	Core ICore
	Log  *slog.Logger
}

func (r *Resolver) GetPosts(ctx context.Context, limit, offset *int) ([]*model.Post, error) {
	if limit == nil || offset == nil {
		r.Log.Error("Paginator error:", "%s", "dont have limit or offset")
		return nil, fmt.Errorf("dont have limit or offset")
	}

	posts, err := r.Core.GetPosts(ctx, *limit, *offset)
	if err != nil {
		r.Log.Error("get posts error:", "error", err.Error())
		return nil, fmt.Errorf("get posts error:%w", err)
	}

	return posts, nil
}

func (r *Resolver) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	postId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		r.Log.Error("Parse parent error:", "error", err.Error())
		return nil, fmt.Errorf("Parse parent error:%w", err)
	}

	post, err := r.Core.GetPostByID(ctx, int(postId), 1, 1)
	if err != nil {
		r.Log.Error("get post error:", "error", err.Error())
		return nil, fmt.Errorf("get post error:%w", err)
	}

	return post, nil
}

func (r *Resolver) GetCommentsByPostID(ctx context.Context, postID string, limit, offset *int) ([]*model.Comment, error) {
	if limit == nil || offset == nil {
		r.Log.Error("Params error:", "%s", "%s", "dont have limit , offset ")
		return nil, fmt.Errorf("dont have limit, offset")
	}

	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		r.Log.Error("Parse id error:", "error", err.Error())
		return nil, fmt.Errorf("Parse id error:%w", err)
	}

	comments, err := r.Core.GetCommentsByPostID(ctx, int(id), *limit, *offset)
	if err != nil {
		r.Log.Error("get comments error:", "error", err.Error())
		return nil, fmt.Errorf("get comments error:%w", err)
	}

	return comments, nil
}

func (r *Resolver) AddPost(ctx context.Context, data string, isCommented bool) (*model.Post, error) {
	session := ctx.Value(variables.UserIDKey)
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	post, err := r.Core.AddPost(ctx, data, int(session.(int64)), isCommented)
	if err != nil {
		r.Log.Error("add post error:", "error", err.Error())
		return nil, fmt.Errorf("add post error:%w", err)
	}

	return post, nil
}

func (r *Resolver) AddComment(ctx context.Context, postID string, data string, parentID *string) (*model.Comment, error) {
	session := ctx.Value(variables.UserIDKey)
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	idConverted, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		r.Log.Error("Parse postID error:", "error", err.Error())
		return nil, fmt.Errorf("Parse postID error:%w", err)
	}

	parent, err := strconv.ParseInt(*parentID, 10, 64)
	if err != nil {
		r.Log.Error("Parse parent error:", "error", err.Error())
		return nil, fmt.Errorf("Parse parent error:%w", err)
	}

	comment, err := r.Core.AddComment(ctx, int(idConverted), int(session.(int64)), data, int(parent))
	if err != nil {
		r.Log.Error("add comment error:", "error", err.Error())
		return nil, fmt.Errorf("add comment error:%w", err)
	}

	return comment, nil
}
