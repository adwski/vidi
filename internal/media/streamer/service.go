package streamer

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/adwski/vidi/internal/media/store"

	"github.com/adwski/vidi/internal/media/store/s3"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	internalError = "internal error"
	notFoundError = "not found"

	contentTypeSegment  = "video/iso.segment"
	contentTypeVideoMP4 = "video/mp4"
	contentTypeAudioMP4 = "audio/mp4"
	contentTypeMPD      = "application/dash+xml"
)

var (
	methodGET      = []byte("GET")
	objTypeSegment = []byte(".m4s")
	objTypeMP4     = []byte(".mp4")
	objTypeMPD     = []byte(".mpd")

	trackTypeAudio = []byte("soun")
	trackTypeVideo = []byte("vide")
)

// Service is a streaming service. It implements fasthttp handler that
// serves MPEG-DASH segments.
// Segments are taken from media store.
// Every request is also checked for valid "watch"-session.
type Service struct {
	logger       *zap.Logger
	sessS        *sessionStore.Store
	mediaS       *s3.Store
	cors         *CORSConfig
	s3PathPrefix []byte
	uriPrefixLen int
}

type CORSConfig struct {
	AllowOrigin string
}

type Config struct {
	Logger        *zap.Logger
	SessionStore  *sessionStore.Store
	CORSConfig    *CORSConfig
	URIPathPrefix string
	S3PathPrefix  string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3SSL         bool
}

func New(cfg *Config) (*Service, error) {
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
		sessS:        cfg.SessionStore,
		cors:         cfg.CORSConfig,
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
		if errors.Is(errSess, sessionStore.ErrNotFound) {
			ctx.Error(notFoundError, fasthttp.StatusNotFound)
			return
		}
		svc.logger.Debug("cannot get session", zap.Error(errSess))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}

	// --------------------------------------------------
	// Request is valid and session exists
	// Proceed with segment handling
	// --------------------------------------------------
	// Get segment reader
	rc, size, errS3 := svc.mediaS.Get(ctx, svc.getSegmentName(sess, path))
	if errS3 != nil {
		if errors.Is(errS3, store.ErrNotFount) {
			ctx.Error(notFoundError, fasthttp.StatusNotFound)
			return
		}
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

	// --------------------------------------------------
	// Set headers and body
	// --------------------------------------------------
	if svc.cors != nil {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", svc.cors.AllowOrigin)
	}
	ctx.Response.Header.Set("Content-Type", cType)
	// Set body reader, fasthttp will handle the rest
	ctx.SetBodyStream(rc, int(size)) // reader will be closed by fasthttp
}

func (svc *Service) getSegmentName(sess *session.Session, path []byte) string {
	var b []byte
	return string(append(append(append(b, svc.s3PathPrefix...), []byte(sess.Location)...), path...))
}

func (svc *Service) getSessionIDAndSegmentPathFromURI(uri []byte) (string, []byte, string, error) {
	// URI: /prefix/<session-id>/<segment>
	if svc.uriPrefixLen >= len(uri) {
		return "", nil, "", errors.New("request uri is less than configured prefix")
	}
	sessionAndPath := uri[svc.uriPrefixLen+1:]
	idx := bytes.IndexByte(sessionAndPath, '/')
	if idx == -1 || idx+1 == len(sessionAndPath) {
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
	case bytes.HasSuffix(path, objTypeMP4): // for init segments
		switch {
		case bytes.HasPrefix(path[1:], trackTypeAudio):
			cType = contentTypeAudioMP4
		case bytes.HasPrefix(path[1:], trackTypeVideo):
			cType = contentTypeVideoMP4
		default:
			return "", nil, "", fmt.Errorf("cannot determine mp4 track type")
		}
	case bytes.HasSuffix(path, objTypeMPD):
		// TODO MPD is also served as segment at the moment.
		//  In the future it should be moved to video api.
		cType = contentTypeMPD
	default:
		return "", nil, "", fmt.Errorf("invalid segment type")
	}
	return string(sess), path, cType, nil
}
