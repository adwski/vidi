// Package uploader contains app that handles media part uploads.
package uploader

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/adwski/vidi/internal/event"
	"github.com/adwski/vidi/internal/event/notificator"
	"github.com/adwski/vidi/internal/media/store/s3"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	internalError = "internal error"
	sizeError     = "incorrect size"
	notFoundError = "not found"
)

var (
	contentTypeVidiMediapart = []byte("application/x-vidi-mediapart")
	methodPOST               = []byte("POST")
)

// Service is a media file uploader service. It implements fasthttp handler that
// reads uploaded part and stores it in media store.
// Every request is also checked for valid "upload"-session.
//
// After each successful part upload, uploader calculates sha256 checksum
// and asynchronously notifies videoapi.
type Service struct {
	logger       *zap.Logger
	sessS        *sessionStore.Store
	mediaS       *s3.Store
	notificator  *notificator.Notificator
	s3pathPrefix []byte
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

	sessID, partNum, err := svc.getParamsFromURI(ctx.Request.RequestURI())
	if err != nil {
		svc.logger.Debug("cannot params from uri", zap.Error(err))
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
	sess, errSess := svc.sessS.Get(ctx, string(sessID))
	if errSess != nil {
		if errors.Is(errSess, sessionStore.ErrNotFound) {
			ctx.Error(notFoundError, fasthttp.StatusNotFound)
			return
		}
		svc.logger.Error("cannot get session", zap.Error(errSess))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}

	// validate size
	if uint64(size) > sess.PartSize {
		svc.logger.Error("content size is greater than part size",
			zap.Int("size", size),
			zap.Uint64("partSize", sess.PartSize))
		ctx.Error(sizeError, fasthttp.StatusBadRequest)
		return
	}

	// --------------------------------------------------
	// Request is valid and session exists
	// Proceed with upload
	// --------------------------------------------------
	artifactName := svc.getUploadArtifactName(sessID, partNum)
	buf := bytes.NewBuffer(ctx.Request.Body())
	err = svc.mediaS.Put(ctx, artifactName, buf, int64(size))
	if err != nil {
		svc.logger.Error("error while uploading artifact",
			zap.Int("size", size),
			zap.Error(err),
			zap.String("artifactName", artifactName),
			zap.Uint64("partSize", sess.PartSize))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}
	checksum, err := svc.mediaS.CalcSha256(ctx, artifactName)
	if err != nil {
		svc.logger.Error("unable to calc sha256",
			zap.String("artifactName", artifactName))
		ctx.Error(internalError, fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)

	// --------------------------------------------------
	// Postprocessing phase
	// --------------------------------------------------
	go svc.notificator.Send(&event.Event{
		PartInfo: &event.PartInfo{
			VideoID:  sess.VideoID,
			Checksum: checksum,           // this is already base64 encoded
			Num:      parseUint(partNum), // getParamsFromURI ensures that partNum contains valid number
		},
		Kind: event.KindVideoPartUploaded,
	})
}

func parseUint(b []byte) (num uint) {
	for _, ch := range b {
		num = num*10 + uint(ch-'0') //nolint:mnd // no magic here
	}
	return
}

func checkHeader(ctx *fasthttp.RequestCtx) (int, error) {
	cType := ctx.Request.Header.ContentType()
	if !bytes.Equal(cType, contentTypeVidiMediapart) {
		return 0, errors.New("wrong content type")
	}

	cLength := ctx.Request.Header.ContentLength()
	if cLength <= 0 {
		return 0, errors.New("wrong or missing content length")
	}
	return cLength, nil
}

func (svc *Service) getUploadArtifactName(sessID, partNum []byte) string {
	b := make([]byte, 0, len(svc.s3pathPrefix)+len(sessID)+len(partNum)+1)
	b = append(b, svc.s3pathPrefix...)
	b = append(b, sessID...)
	b = append(b, '/')
	b = append(b, partNum...)
	return string(b)
}

func (svc *Service) getParamsFromURI(uri []byte) ([]byte, []byte, error) {
	// URI: /prefix/<upload-session-id>/<partNum>
	if svc.uriPrefixLen >= len(uri) {
		return nil, nil, errors.New("request uri is less than configured prefix")
	}
	sessAndNum := uri[svc.uriPrefixLen+1:]
	idx := bytes.IndexByte(sessAndNum, '/')
	if idx <= 0 {
		return nil, nil, errors.New("invalid uri")
	}
	sessID := sessAndNum[:idx]
	partNum := sessAndNum[idx+1:]
	if !isNumber(partNum) {
		return nil, nil, errors.New("num is not a number")
	}
	return sessID, partNum, nil
}

func isNumber(b []byte) bool {
	for _, ch := range b {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
