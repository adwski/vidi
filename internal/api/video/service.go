package video

import (
	"context"
	"fmt"
	"strings"

	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/adwski/vidi/internal/session"
	"go.uber.org/zap"
)

// Store is used by videoapi service to store Video and Part objects.
type Store interface {
	Create(ctx context.Context, vi *model.Video) error
	Get(ctx context.Context, id string, userID string) (*model.Video, error)
	GetAll(ctx context.Context, userID string) ([]*model.Video, error)
	Delete(ctx context.Context, id string, userID string) error
	Usage(ctx context.Context, userID string) (*model.UserUsage, error)

	GetListByStatus(ctx context.Context, status model.Status) ([]*model.Video, error)
	Update(ctx context.Context, vi *model.Video) error
	UpdateStatus(ctx context.Context, vi *model.Video) error

	UpdatePart(ctx context.Context, vid string, part *model.Part) error
	DeleteUploadedParts(ctx context.Context, vid string) error
}

// SessionStore stores upload and watch sessions.
type SessionStore interface {
	Set(ctx context.Context, session *session.Session) error
	Get(ctx context.Context, id string) (*session.Session, error)
}

// Service is a Video API service. It has two "realms": user-side API and service-side API.
//
// User-side API provides CRUD operations with Video objects for a single user.
// While service-side API provides handlers for video updates by media processing services.
//
// Besides different API handlers they also differs in authentication approach:
// user-side only checks for valid user id in jwt cookie, while service-API
// looks up jwt in Bearer token and checks for valid service role.
//
// In production environment only user-side API should be exposed to public.
type Service struct {
	logger          *zap.Logger
	idGen           *generators.ID
	watchSessions   SessionStore
	uploadSessions  SessionStore
	s               Store
	watchURLPrefix  string
	uploadURLPrefix string
	quotas          Quotas
}

type Quotas struct {
	VideosPerUser uint
	MaxTotalSize  uint64
}

type ServiceConfig struct {
	Logger             *zap.Logger
	Store              Store
	UploadSessionStore SessionStore
	WatchSessionStore  SessionStore
	WatchURLPrefix     string
	UploadURLPrefix    string
	Quotas             Quotas
}

func NewService(cfg *ServiceConfig) *Service {
	return &Service{
		logger:          cfg.Logger,
		s:               cfg.Store,
		uploadSessions:  cfg.UploadSessionStore,
		watchSessions:   cfg.WatchSessionStore,
		quotas:          cfg.Quotas,
		idGen:           generators.NewID(),
		watchURLPrefix:  strings.TrimRight(cfg.WatchURLPrefix, "/"),
		uploadURLPrefix: strings.TrimRight(cfg.UploadURLPrefix, "/"),
	}
}

func (svc *Service) getUploadURL(sessID string) string {
	return fmt.Sprintf("%s/%s", svc.uploadURLPrefix, sessID)
}

func (svc *Service) getWatchBaseURL(sessID string) string {
	return fmt.Sprintf("%s/%s/", svc.watchURLPrefix, sessID) // trailing / is important!
}

func (svc *Service) getWatchURL(sessID string) string {
	return fmt.Sprintf("%s/%s/%s", svc.watchURLPrefix, sessID, mp4.MPDSuffix)
}
