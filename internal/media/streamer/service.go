package streamer

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adwski/vidi/internal/media/store/s3"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	internalError = "internal error"

	watchSessionDefaultTTL = 300 * time.Second

	contentTypeSegment = "video/iso.segment"
	contentTypeMPD     = "application/dash+xml"
)

var (
	methodGET      = []byte("GET")
	objTypeSegment = []byte(".m4s")
	objTypeMPD     = []byte(".mpd")
)

type Service struct {
	logger       *zap.Logger
	sessS        *sessionStore.Store
	mediaS       *s3.Store
	s3PathPrefix []byte
	uriPrefixLen int
}

type Config struct {
	Logger        *zap.Logger
	RedisDSN      string
	URIPathPrefix string
	S3PathPrefix  string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3SSL         bool
}

func New(cfg *Config) (*Service, error) {
	rWatch, errReU := sessionStore.NewStore(&sessionStore.Config{
		Name:     session.KindWatch,
		RedisDSN: cfg.RedisDSN,
		TTL:      watchSessionDefaultTTL,
	})
	if errReU != nil {
		return nil, fmt.Errorf("cannot configure upload session store: %w", errReU)
	}
	s3Store, errS3 := s3.NewStore(&s3.StoreConfig{
		Logger:    cfg.Logger,
		Endpoint:  cfg.S3Endpoint,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		Bucket:    cfg.S3Bucket,
		SSL:       cfg.S3SSL,
	})
	if errS3 != nil {
		return nil, fmt.Errorf("cannot configure s3 media store: %w", errS3)
	}
	return &Service{
		logger:       cfg.Logger,
		sessS:        rWatch,
		mediaS:       s3Store,
		s3PathPrefix: []byte(fmt.Sprintf("%s/", strings.TrimSuffix(cfg.S3PathPrefix, "/"))),
		uriPrefixLen: len(cfg.URIPathPrefix),
	}, nil
}

func (svc *Service) Handler() func(*fasthttp.RequestCtx) {
	return svc.handleWatch
}

func (svc *Service) handleWatch(ctx *fasthttp.RequestCtx) {
	// --------------------------------------------------
	// Get and check request params
	// --------------------------------------------------
	if !bytes.Equal(ctx.Method(), methodGET) {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}
	// Get all necessary params from request URI
	sessID, path, cType, err := svc.getSessionIDAndSegmentPathFromURI(ctx.Request.RequestURI())
	if err != nil {
		svc.logger.Debug("uri is not valid", zap.Error(err))
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	// Retrieve session
	sess, errSess := svc.sessS.GetExpireCached(ctx, sessID)
	if errSess != nil {
		svc.logger.Debug("cannot get session", zap.Error(errSess))
		ctx.Error(internalError, fasthttp.StatusNotFound)
		return
	}

	// --------------------------------------------------
	// Request is valid and session exists
	// Proceed with segment handling
	// --------------------------------------------------
	rc, size, errS3 := svc.mediaS.Get(ctx, svc.getSegmentName(sess, path))
	if errS3 != nil {
		svc.logger.Error("error while retrieving segment", zap.Error(errS3))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}

	svc.logger.Debug("serving segment",
		zap.String("video_id", sess.VideoID),
		zap.String("session_id", sess.ID),
		zap.String("path", string(path)),
		zap.Int64("size", size),
		zap.String("type", cType))

	ctx.Response.Header.Set("Content-Type", cType)
	ctx.SetBodyStream(rc, int(size)) // reader will be closed by fasthttp
}

func (svc *Service) getSegmentName(sess *session.Session, path []byte) string {
	var b []byte
	return string(append(append(append(b, svc.s3PathPrefix...), []byte(sess.VideoID)...), path...))
}

func (svc *Service) getSessionIDAndSegmentPathFromURI(uri []byte) (string, []byte, string, error) {
	// URI: /prefix/<session-id>/<segment>
	if svc.uriPrefixLen >= len(uri) {
		return "", nil, "", errors.New("request uri is less than configured prefix")
	}
	sessionAndPath := uri[svc.uriPrefixLen+1:]
	idx := bytes.IndexByte(sessionAndPath, '/')
	if idx != -1 || idx+1 == len(sessionAndPath) {
		return "", nil, "", errors.New("invalid uri")
	}
	var (
		sess  = sessionAndPath[:idx]
		path  = sessionAndPath[idx:]
		cType string
	)
	switch {
	case bytes.HasSuffix(path, objTypeSegment):
		cType = contentTypeSegment
	case bytes.HasSuffix(path, objTypeMPD):
		cType = contentTypeMPD
	default:
		return "", nil, "", fmt.Errorf("invalid segment type")
	}
	return string(sess), path, cType, nil
}
