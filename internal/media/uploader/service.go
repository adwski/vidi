package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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
	notFoundError = "not found"

	defaultMediaStoreArtifactName = "/artifact.mp4"
)

var (
	contentTypeMP4 = []byte("video/mp4")
	methodPOST     = []byte("POST")
)

// Service is a media file uploader service. It implements fasthttp handler that
// reads uploaded file and stores it in media store.
// Every request is also checked for valid "upload"-session.
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
	Logger         *zap.Logger
	Notificator    *notificator.Notificator
	SessionStorage *sessionStore.Store
	URIPathPrefix  string
	S3PathPrefix   string
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3Bucket       string
	S3SSL          bool
}

func New(cfg *Config) (*Service, error) {
	s3Store, errS3 := s3.NewStore(&s3.StoreConfig{
		Logger:       cfg.Logger,
		Endpoint:     cfg.S3Endpoint,
		AccessKey:    cfg.S3AccessKey,
		SecretKey:    cfg.S3SecretKey,
		Bucket:       cfg.S3Bucket,
		SSL:          cfg.S3SSL,
		CreateBucket: true,
	})
	if errS3 != nil {
		return nil, fmt.Errorf("cannot configure s3 media store: %w", errS3)
	}
	return &Service{
		uriPrefixLen: len(cfg.URIPathPrefix),
		s3pathPrefix: []byte(fmt.Sprintf("%s/", strings.TrimSuffix(cfg.S3PathPrefix, "/"))),
		s3pathSuffix: []byte(defaultMediaStoreArtifactName),
		logger:       cfg.Logger.With(zap.String("component", "uploader")),
		sessS:        cfg.SessionStorage,
		mediaS:       s3Store,
		notificator:  cfg.Notificator,
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
		if errors.Is(errSess, sessionStore.ErrNotFound) {
			ctx.Error(notFoundError, fasthttp.StatusNotFound)
			return
		}
		svc.logger.Error("cannot get session", zap.Error(errSess))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
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

	if err = svc.pipeBodyToS3(ctx, sess, size); err != nil {
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
	// --------------------------------------------------
	// Postprocessing phase
	// --------------------------------------------------
	go svc.postProcess(sess)
}

func (svc *Service) pipeBodyToS3(ctx *fasthttp.RequestCtx, sess *session.Session, size int) error {
	defer func() {
		if errC := ctx.Request.CloseBodyStream(); errC != nil {
			svc.logger.Error("error while closing body stream", zap.Error(errC))
		}
	}()

	// Manually pipe data streams because there's no native way
	// to do it using fasthttp and s3 client together.
	var (
		r, w                        = io.Pipe()
		done                        = make(chan struct{})
		errPut, errBody, errR, errW error
	)
	go func() {
		if errPut = svc.mediaS.Put(ctx, svc.getUploadArtifactName(sess), r, int64(size)); errPut != nil {
			svc.logger.Error("error while uploading artifact", zap.Error(errPut))
		}
		if errR = r.Close(); errR != nil {
			svc.logger.Error("error closing pipe reader", zap.Error(errR))
		}
		done <- struct{}{}
	}()
	go func() {
		if errBody = ctx.Request.BodyWriteTo(w); errBody != nil {
			svc.logger.Error("error while reading body", zap.Error(errBody))
		}
		if errW = w.Close(); errW != nil {
			svc.logger.Error("error closing pipe writer", zap.Error(errW))
		}
		done <- struct{}{}
	}()
	<-done
	<-done
	if errPut != nil || errBody != nil || errR != nil || errW != nil {
		return errors.New("pipe error")
	}
	return nil
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
	return string(append(append(append(b, svc.s3pathPrefix...), []byte(sess.ID)...), svc.s3pathSuffix...))
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
