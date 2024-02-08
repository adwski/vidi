package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type RequestValidator struct {
	validator *validator.Validate
	logger    *zap.Logger
}

func NewRequestValidator(logger *zap.Logger) *RequestValidator {
	return &RequestValidator{
		validator: validator.New(),
		logger:    logger,
	}
}

func (cv *RequestValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		var vErrs validator.ValidationErrors
		if !errors.As(err, &vErrs) || len(vErrs) == 0 {
			cv.logger.Error("unknown validation error", zap.Error(err))
			return fmt.Errorf("unknown error")
		}
		return fmt.Errorf("missing required params: %s", JoinErrFields(vErrs))
	}
	return nil
}

func JoinErrFields(vErrs validator.ValidationErrors) string {
	result := make([]string, 0, len(vErrs))
	for _, v := range vErrs {
		result = append(result, strings.ToLower(v.Field()))
	}
	return strings.Join(result, ", ")
}
