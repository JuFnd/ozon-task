package inmemory_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"ozon-task/pkg/util"
	"ozon-task/pkg/variables"
	"ozon-task/services/posts/delivery/graph/model"
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

func (repo *PostsCacheRepository) GetPosts(ctx context.Context, limit int, offset int) ([]*model.Post, error) {
	keys, err := repo.postsRedisClient.Keys("post:*").Result()
	if err != nil {
		return nil, err
	}

	var posts []*model.Post
	for _, key := range keys {
		val, err := repo.postsRedisClient.Get(key).Result()
		if err != nil {
			return nil, err
		}

		var post model.Post
		err = json.Unmarshal([]byte(val), &post)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (repo *PostsCacheRepository) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	key := "post:" + strconv.Itoa(id)
	val, err := repo.postsRedisClient.Get(key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("post not found")
	} else if err != nil {
		return nil, err
	}

	var post model.Post
	err = json.Unmarshal([]byte(val), &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (repo *PostsCacheRepository) GetCommentsByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error) {
	keys, err := repo.postsRedisClient.Keys("comment:" + strconv.Itoa(postID) + ":*").Result()
	if err != nil {
		return nil, err
	}

	var comments []*model.Comment
	for _, key := range keys {
		val, err := repo.postsRedisClient.Get(key).Result()
		if err != nil {
			return nil, err
		}

		var comment model.Comment
		err = json.Unmarshal([]byte(val), &comment)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (repo *PostsCacheRepository) AddPost(ctx context.Context, data string, user *model.User, isCommented bool) (*model.Post, error) {
	post := &model.Post{
		ID:          strconv.Itoa(util.RandInt()),
		Author:      user,
		Content:     data,
		IsCommented: &isCommented,
		CreatedAt:   time.Now().String(),
	}

	postBytes, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}

	key := "post:" + post.ID
	err = repo.postsRedisClient.Set(key, postBytes, time.Duration(time.Hour*24)).Err()
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *PostsCacheRepository) AddComment(ctx context.Context, post *model.Post, user *model.User, data string, parentID int) (*model.Comment, error) {
	comment := &model.Comment{
		ID:        strconv.Itoa(util.RandInt()),
		Author:    user,
		Post:      post,
		ParentID:  strconv.Itoa(parentID),
		Content:   data,
		CreatedAt: time.Now().String(),
	}

	commentBytes, err := json.Marshal(comment)
	if err != nil {
		return nil, err
	}

	key := "comment:" + post.ID + ":" + comment.ID
	err = repo.postsRedisClient.Set(key, commentBytes, time.Duration(time.Hour*24)).Err()
	if err != nil {
		return nil, err
	}

	return comment, nil
}
