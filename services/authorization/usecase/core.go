package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"ozon-task/pkg/models"
	"ozon-task/pkg/util"
	"ozon-task/pkg/variables"
	"ozon-task/services/authorization/repository/profile"
	"ozon-task/services/authorization/repository/session"
	"regexp"
	"sync"
	"time"
)

type IProfileRelationalRepository interface {
	CreateUser(login string, password []byte) error
	FindUser(login string) (bool, error)
	GetUser(login string, password []byte) (*models.UserItem, bool, error)
	GetUserProfileId(login string) (int64, error)
	GetUserRole(id int64) (string, error)
}

type ISessionCacheRepository interface {
	SaveSessionCache(ctx context.Context, createdSessionObject models.Session, logger *slog.Logger) (bool, error)
	GetSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error)
	DeleteSessionCache(ctx context.Context, sid string, logger *slog.Logger) (bool, error)
	GetUserLogin(ctx context.Context, sid string, logger *slog.Logger) (string, error)
}

type Core struct {
	sessions ISessionCacheRepository
	logger   *slog.Logger
	mutex    sync.RWMutex
	profiles IProfileRelationalRepository
}

func GetCore(profileConfig *variables.RelationalDataBaseConfig, sessionConfig *variables.CacheDataBaseConfig, logger *slog.Logger) (*Core, error) {
	sessionRepository, err := session.GetSessionRepository(sessionConfig, logger)
	if err != nil {
		logger.Error(variables.SessionRepositoryNotActiveError)
		return nil, err
	}

	profileRepository, err := profile.GetProfileRepository(profileConfig, logger)
	if err != nil {
		logger.Error(variables.ProfileRepositoryNotActiveError)
		return nil, err
	}

	core := Core{
		sessions: sessionRepository,
		logger:   logger.With(variables.ModuleLogger, variables.CoreModuleLogger),
		profiles: profileRepository,
	}

	return &core, nil
}

func (core *Core) CreateSession(ctx context.Context, login string) (models.Session, error) {
	sid := util.RandStringRunes(32)

	newSession := models.Session{
		Login:     login,
		SID:       sid,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	core.mutex.Lock()
	sessionAdded, err := core.sessions.SaveSessionCache(ctx, newSession, core.logger)
	defer core.mutex.Unlock()

	if !sessionAdded && err != nil {
		return models.Session{}, err
	}

	if !sessionAdded {
		return models.Session{}, nil
	}

	return newSession, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	core.mutex.Lock()
	_, err := core.sessions.DeleteSessionCache(ctx, sid, core.logger)
	defer core.mutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (core *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	core.mutex.RLock()
	found, err := core.sessions.GetSessionCache(ctx, sid, core.logger)
	defer core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, nil
}

func (core *Core) CreateUserAccount(login string, password string) error {
	matched, err := regexp.MatchString(variables.LoginRegexp, login)
	if err != nil {
		core.logger.Error(variables.StatusInternalServerError+" %w", "core", "err", err)
		return fmt.Errorf(fmt.Sprintf(variables.StatusInternalServerError+" %s", err.Error()))
	}

	if !matched {
		core.logger.Error(variables.InvalidLoginOrPasswordError)
		return fmt.Errorf(fmt.Sprintf(variables.InvalidLoginOrPasswordError))
	}

	hashPassword := util.HashPassword(password)
	err = core.profiles.CreateUser(login, hashPassword)
	if err != nil {
		core.logger.Error(variables.CreateProfileError+" %w", "core", "err", err)
		return err
	}

	return nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.profiles.FindUser(login)
	if err != nil {
		core.logger.Error(variables.ProfileNotFoundError+" %w", err)
		return false, err
	}

	return found, nil
}

func (core *Core) FindUserAccount(login string, password string) (*models.UserItem, bool, error) {
	hashPassword := util.HashPassword(password)
	user, found, err := core.profiles.GetUser(login, hashPassword)
	if err != nil {
		core.logger.Error(variables.ProfileNotFoundError+" %w", "core", "err", err)
		return nil, false, err
	}
	return user, found, nil
}

func (core *Core) GetUserId(ctx context.Context, sid string) (int64, error) {
	login, err := core.sessions.GetUserLogin(ctx, sid, core.logger)
	if err != nil {
		return 0, err
	}

	id, err := core.profiles.GetUserProfileId(login)
	if err != nil {
		core.logger.Error(variables.GetProfileError, " id: %v", err)
		return 0, err
	}
	return id, nil
}

func (core *Core) GetUserRole(ctx context.Context, id int64) (string, error) {
	role, err := core.profiles.GetUserRole(id)
	if err != nil {
		core.logger.Error(variables.GetProfileRoleError+" %w", "core", "err", err)
		return "", fmt.Errorf(fmt.Sprintf(variables.GetProfileRoleError+" %v", err))
	}

	return role, nil
}
