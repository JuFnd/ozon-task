package inmemory_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"ozon-task/pkg/models"
	"ozon-task/pkg/variables"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type PostsCacheRepository struct {
	postsRedisClient *redis.Client
}

func (postsRedisRepository *PostsCacheRepository) reconnectRedis() error {
	err := postsRedisRepository.postsRedisClient.Close()
	if err != nil {
		return err
	}

	newClient := redis.NewClient(&redis.Options{
		Addr:     postsRedisRepository.postsRedisClient.Options().Addr,
		Password: postsRedisRepository.postsRedisClient.Options().Password,
		DB:       postsRedisRepository.postsRedisClient.Options().DB,
	})

	postsRedisRepository.postsRedisClient = newClient

	return nil
}

func (postsRedisRepository *PostsCacheRepository) pingRedis(timer int, logger *slog.Logger) error {
	var pingErrString string
	var reconnectErrString string
	var retries int

	for retries < variables.MaxRetries {
		_, pingErr := postsRedisRepository.postsRedisClient.Ping().Result()
		if pingErr == nil {
			return nil
		}
		pingErrString = pingErr.Error()

		reconnectErr := postsRedisRepository.reconnectRedis()
		if reconnectErr == nil {
			return nil
		}
		reconnectErrString = reconnectErr.Error()

		retries++
		logger.Error(variables.AuthorizationCachePingRetryError, pingErr.Error(), reconnectErr.Error())
		time.Sleep(time.Duration(timer) * time.Second)
	}

	return fmt.Errorf(fmt.Sprintf(variables.AuthorizationCachePingMaxRetriesError + pingErrString + reconnectErrString))
}

func GetPostsRepository(postsConfig *variables.CacheDataBaseConfig, logger *slog.Logger) (*PostsCacheRepository, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     postsConfig.Host,
		Password: postsConfig.Password,
		DB:       postsConfig.DbNumber,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	postsRedisRepository := &PostsCacheRepository{
		postsRedisClient: redisClient,
	}

	errs := make(chan error)

	go func() {
		errs <- postsRedisRepository.pingRedis(postsConfig.Timer, logger)
	}()

	if err := <-errs; err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return postsRedisRepository, nil
}

func (repo *PostsCacheRepository) GetPosts(ctx context.Context, limit int, offset int) ([]models.Post, error) {
	keys, err := repo.postsRedisClient.Keys("post:*").Result()
	if err != nil {
		return nil, err
	}

	var posts []models.Post
	for _, key := range keys {
		val, err := repo.postsRedisClient.Get(key).Result()
		if err != nil {
			return nil, err
		}

		var post models.Post
		err = json.Unmarshal([]byte(val), &post)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (repo *PostsCacheRepository) GetPostByID(ctx context.Context, id int) (*models.Post, error) {
	key := "post:" + strconv.Itoa(id)
	val, err := repo.postsRedisClient.Get(key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("post not found")
	} else if err != nil {
		return nil, err
	}

	var post models.Post
	err = json.Unmarshal([]byte(val), &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (repo *PostsCacheRepository) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]models.Comment, error) {
	keys, err := repo.postsRedisClient.Keys("comment:" + strconv.Itoa(postID) + ":*").Result()
	if err != nil {
		return nil, err
	}

	var comments []models.Comment
	for _, key := range keys {
		val, err := repo.postsRedisClient.Get(key).Result()
		if err != nil {
			return nil, err
		}

		var comment models.Comment
		err = json.Unmarshal([]byte(val), &comment)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (repo *PostsCacheRepository) AddPost(ctx context.Context, data string, userId int, isCommented bool) (*models.Post, error) {
	post := &models.Post{
		UserID:          userId,
		Content:         data,
		CommentsAllowed: isCommented,
		CreatedAt:       time.Now(),
	}

	postBytes, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}

	key := "post:" + strconv.Itoa(post.ID)
	err = repo.postsRedisClient.Set(key, postBytes, time.Duration(time.Hour*24)).Err()
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsCacheRepository) AddComment(ctx context.Context, postID int, userId int, data string, parentID int) (*models.Comment, error) {
	comment := &models.Comment{
		UserID:    userId,
		PostID:    postID,
		ParentID:  parentID,
		Content:   data,
		CreatedAt: time.Now(),
	}

	commentBytes, err := json.Marshal(comment)
	if err != nil {
		return nil, err
	}

	key := "comment:" + strconv.Itoa(postID) + ":" + strconv.Itoa(comment.ID)
	err = repo.postsRedisClient.Set(key, commentBytes, time.Duration(time.Hour*24)).Err()
	if err != nil {
		return nil, err
	}

	return comment, nil
}
