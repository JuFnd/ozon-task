package relational_repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"ozon-task/pkg/variables"
	"ozon-task/services/posts/delivery/graph/model"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/stdlib"
)

type ProfileRelationalRepository struct {
	db *sql.DB
}

func GetPostsRepository(configDatabase *variables.RelationalDataBaseConfig, logger *slog.Logger) (*ProfileRelationalRepository, error) {
	dsn := fmt.Sprintf("user=%s dbname=%s password= %s host=%s port=%d sslmode=%s",
		configDatabase.User, configDatabase.DbName, configDatabase.Password, configDatabase.Host, configDatabase.Port, configDatabase.Sslmode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error(variables.SqlOpenError+"%w", "repo", "err", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Error(variables.SqlPingError+"%w", "repo", "err", err)
		return nil, err
	}

	db.SetMaxOpenConns(configDatabase.MaxOpenConns)

	profileDb := ProfileRelationalRepository{
		db: db,
	}

	errs := make(chan error)
	go func() {
		errs <- profileDb.pingDb(configDatabase.Timer, logger)
	}()

	if err := <-errs; err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return &profileDb, nil
}

func (repository *ProfileRelationalRepository) pingDb(timer uint32, logger *slog.Logger) error {
	var err error
	var retries int

	for retries < variables.MaxRetries {
		err = repository.db.Ping()
		if err == nil {
			return nil
		}

		retries++
		logger.Error(variables.SqlPingError+"%w", "repo", "err", err)
		time.Sleep(time.Duration(timer) * time.Second)
	}

	logger.Error(variables.SqlMaxPingRetriesError, err)
	return fmt.Errorf(fmt.Sprintf(variables.SqlMaxPingRetriesError+" %v", err))
}
func (repository *ProfileRelationalRepository) GetPosts(ctx context.Context, limit int, offset int) ([]*model.Post, error) {
	query := "SELECT id, user_id, content, created_at, comments_allowed FROM posts LIMIT $1 OFFSET $2"
	rows, err := repository.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Content, &post.CreatedAt, &post.IsCommented)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (repository *ProfileRelationalRepository) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	query := "SELECT id, user_id, content, created_at, comments_allowed FROM posts WHERE id = $1"
	row := repository.db.QueryRow(query, id)

	var post model.Post
	var user model.User
	var userId int
	err := row.Scan(&post.ID, &userId, &post.Content, &post.CreatedAt, &post.IsCommented)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	user.ID = strconv.Itoa(userId)
	post.Author = &user
	return &post, nil
}

func (repository *ProfileRelationalRepository) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error) {
	query := "SELECT id, user_id, post_id, parent_id, content, created_at FROM comments WHERE post_id = $1 LIMIT $2 OFFSET $3"
	rows, err := repository.db.Query(query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		var user model.User
		var post model.Post
		var postId int
		var userId int
		err := rows.Scan(&comment.ID, &userId, &postId, &comment.ParentID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		user.ID = strconv.Itoa(userId)
		post.ID = strconv.Itoa(postId)
		comment.Author = &user
		comment.Post = &post
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (repository *ProfileRelationalRepository) AddPost(ctx context.Context, data string, user *model.User, isCommented bool) (*model.Post, error) {
	query := "INSERT INTO posts (user_id, content, created_at, comments_allowed) VALUES ($1, $2, $3, $4) RETURNING id"
	var postID int
	err := repository.db.QueryRow(query, user.ID, data, time.Now(), isCommented).Scan(&postID)
	if err != nil {
		return nil, err
	}

	post := &model.Post{
		ID:          strconv.Itoa(postID),
		Content:     data,
		CreatedAt:   time.Now().String(),
		IsCommented: &isCommented,
	}

	return post, nil
}

func (repository *ProfileRelationalRepository) AddComment(ctx context.Context, post *model.Post, user *model.User, data string, parentID int) (*model.Comment, error) {
	query := "INSERT INTO comments (user_id, post_id, parent_id, content, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var commentID int
	err := repository.db.QueryRow(query, user.ID, post.ID, parentID, data, time.Now()).Scan(&commentID)
	if err != nil {
		return nil, err
	}
	fmt.Println(user.ID, post.ID, parentID, data, commentID)

	comment := &model.Comment{
		ID:        strconv.Itoa(commentID),
		ParentID:  strconv.Itoa(parentID),
		Author:    user,
		Post:      post,
		Content:   data,
		CreatedAt: time.Now().String(),
	}

	return comment, nil
}
