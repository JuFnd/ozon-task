package relational_repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"ozon-task/pkg/models"
	"ozon-task/pkg/variables"
	"time"
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
func (repository *ProfileRelationalRepository) GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error) {
	query := "SELECT id, user_id, content, created_at, comments_allowed FROM posts LIMIT $1 OFFSET $2"
	rows, err := repository.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.CommentsAllowed)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (repository *ProfileRelationalRepository) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	query := "SELECT id, user_id, content, created_at, comments_allowed FROM posts WHERE id = $1"
	row := repository.db.QueryRow(query, id)

	var post models.Post
	err := row.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.CommentsAllowed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (repository *ProfileRelationalRepository) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error) {
	query := "SELECT id, user_id, post_id, parent_id, content, created_at FROM comments WHERE post_id = $1 LIMIT $2 OFFSET $3"
	rows, err := repository.db.Query(query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (repository *ProfileRelationalRepository) AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error) {
	query := "INSERT INTO posts (user_id, content, created_at, comments_allowed) VALUES ($1, $2, $3, $4) RETURNING id"
	var postID int
	err := repository.db.QueryRow(query, 0, data, time.Now(), isCommented).Scan(&postID)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		ID:              postID,
		UserID:          userId,
		Content:         data,
		CreatedAt:       time.Now(),
		CommentsAllowed: isCommented,
	}

	return post, nil
}

func (repository *ProfileRelationalRepository) AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error) {
	query := "INSERT INTO comments (user_id, post_id, parent_id, content, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var commentID int
	err := repository.db.QueryRow(query, userId, postID, parentID, data, time.Now()).Scan(&commentID)
	if err != nil {
		return nil, err
	}

	comment := &models.Comment{
		ID:        commentID,
		UserID:    userId,
		PostID:    postID,
		ParentID:  parentID,
		Content:   data,
		CreatedAt: time.Now(),
	}

	return comment, nil
}
