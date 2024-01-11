package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/event"
	"github.com/adwski/vidi/internal/event/notificator"
	"github.com/adwski/vidi/internal/media/store/s3"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	internalError = "internal error"

	defaultMediaStoreArtifactName = "/artifact.mp4"

	uploadSessionDefaultTTL = 300 * time.Second
)

var (
	contentTypeMP4 = []byte("video/mp4")
	methodPOST     = []byte("POST")
)

type Service struct {
	logger       *zap.Logger
	sessS        *sessionStore.Store
	mediaS       *s3.Store
	notificator  *notificator.Notificator
	s3pathPrefix []byte
	s3pathSuffix []byte
	uriPrefixLen int
}

type Config struct {
	Logger        *zap.Logger
	RedisDSN      string
	URIPathPrefix string
	S3PathPrefix  string
	VideoAPIURL   string
	VideoAPIToken string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3SSL         bool
}

func New(cfg *Config) (*Service, error) {
	rUpload, errReU := sessionStore.NewStore(&sessionStore.Config{
		Name:     session.KindUpload,
		RedisDSN: cfg.RedisDSN,
		TTL:      uploadSessionDefaultTTL,
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
		SSL:       false,
	})
	if errS3 != nil {
		return nil, fmt.Errorf("cannot configure s3 media store: %w", errS3)
	}
	return &Service{
		uriPrefixLen: len(cfg.URIPathPrefix),
		s3pathPrefix: []byte(fmt.Sprintf("%s/", strings.TrimSuffix(cfg.S3PathPrefix, "/"))),
		s3pathSuffix: []byte(defaultMediaStoreArtifactName),
		logger:       cfg.Logger.With(zap.String("component", "uploader")),
		sessS:        rUpload,
		mediaS:       s3Store,
		notificator: notificator.New(&notificator.Config{
			Logger:        cfg.Logger,
			VideoAPIURL:   cfg.VideoAPIURL,
			VideoAPIToken: cfg.VideoAPIToken,
		}),
	}, nil
}

func (svc *Service) Handler() func(*fasthttp.RequestCtx) {
	return svc.handleUpload
}

func (svc *Service) handleUpload(ctx *fasthttp.RequestCtx) {
	// --------------------------------------------------
	// Get and check request params
	// --------------------------------------------------
	if !bytes.Equal(ctx.Method(), methodPOST) {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}

	sessID, err := svc.getSessionIDFromRequestURI(ctx.Request.RequestURI())
	if err != nil {
		svc.logger.Debug("cannot get session from uri", zap.Error(err))
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	size, errHeader := checkHeader(ctx)
	if errHeader != nil {
		svc.logger.Debug("request header error", zap.Error(errHeader))
		ctx.Error(errHeader.Error(), fasthttp.StatusBadRequest)
		return
	}

	// Retrieve session
	// TODO should we also update session TTL if upload size is significant?
	sess, errSess := svc.sessS.Get(ctx, sessID)
	if errSess != nil {
		svc.logger.Debug("cannot get session", zap.Error(errSess))
		ctx.Error(internalError, fasthttp.StatusNotFound)
		return
	}

	// --------------------------------------------------
	// Request is valid and session exists
	// Proceed with upload
	// --------------------------------------------------
	svc.notificator.Send(&event.Event{ // send uploading event
		Video: model.Video{
			ID:     sess.VideoID,
			Status: model.VideoStatusUploading,
		},
		Kind: event.KindUpdateStatus,
	})

	defer func() {
		if errC := ctx.Request.CloseBodyStream(); errC != nil {
			svc.logger.Error("error while closing body stream", zap.Error(errC))
		}
	}()
	// TODO Double check if this is true pipe style read/write and file would never be fully stored in heap.
	//  For whole mp4 this is important
	if err = svc.mediaS.Put(ctx, svc.getUploadArtifactName(sess), ctx.Request.BodyStream(), int64(size)); err != nil {
		svc.logger.Error("error while uploading artifact", zap.Error(err))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}

	// --------------------------------------------------
	// Postprocessing phase
	// --------------------------------------------------
	go svc.postProcess(sess)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (svc *Service) postProcess(sess *session.Session) {
	// Send uploaded event
	// TODO Should we also send error status event in case of upload error
	//      or allow user to retry upload?
	svc.notificator.Send(&event.Event{
		Video: model.Video{
			ID:       sess.VideoID,
			Status:   model.VideoStatusUploaded,
			Location: sess.ID,
		},
		Kind: event.KindUpdateStatusAndLocation,
	})
	if err := svc.sessS.Delete(context.Background(), sess.ID); err != nil {
		svc.logger.Error("error while deleting session",
			zap.Any("session", sess),
			zap.Error(err))
		return
	}
}

func checkHeader(ctx *fasthttp.RequestCtx) (int, error) {
	cType := ctx.Request.Header.ContentType()
	if !bytes.Equal(cType, contentTypeMP4) {
		return 0, errors.New("wrong content type")
	}

	cLength := ctx.Request.Header.ContentLength()
	if cLength == 0 {
		return 0, errors.New("wrong or missing content length")
	}
	return cLength, nil
}

func (svc *Service) getUploadArtifactName(sess *session.Session) string {
	var b []byte
	return string(append(append(append(b, svc.s3pathPrefix...), []byte(sess.VideoID)...), svc.s3pathSuffix...))
}

func (svc *Service) getSessionIDFromRequestURI(uri []byte) (string, error) {
	// URI: /prefix/<session-id>
	if svc.uriPrefixLen >= len(uri) {
		return "", errors.New("request uri is less than configured prefix")
	}
	sessID := uri[svc.uriPrefixLen+1:]
	idx := bytes.IndexByte(sessID, '/')
	if idx != -1 {
		return "", errors.New("invalid uri")
	}
	return string(sessID), nil
}
