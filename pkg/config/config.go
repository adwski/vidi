package config

import (
	"errors"
	"net/url"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// ViperEC is a Viper Error Catcher. It provides casting error checks in addition
// to usual viper functionality.
type ViperEC struct {
	*viper.Viper
	errs map[string]error
}

func NewViperEC() *ViperEC {
	return &ViperEC{
		Viper: viper.New(),
		errs:  make(map[string]error),
	}
}

func (vec *ViperEC) GetDuration(key string) time.Duration {
	d, err := cast.ToDurationE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
	}
	if d == 0 {
		vec.errs[key] = errors.New("cannot be zero")
	}
	return d
}

func (vec *ViperEC) GetStringAllowEmpty(key string) string {
	s, err := cast.ToStringE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
	}
	return s
}

func (vec *ViperEC) GetBool(key string) bool {
	s, err := cast.ToBoolE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
	}
	return s
}

func (vec *ViperEC) GetBoolNoError(key string) bool {
	return cast.ToBool(vec.Get(key))
}

func (vec *ViperEC) GetString(key string) string {
	s, err := cast.ToStringE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
		return ""
	}
	if len(s) == 0 {
		vec.errs[key] = errors.New("cannot be empty")
	}
	return s
}

func (vec *ViperEC) GetURL(key string) string {
	s, err := cast.ToStringE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
		return ""
	}
	_, err = url.Parse(s)
	if err != nil {
		vec.errs[key] = errors.New("is not valid url")
		return ""
	}
	return s
}

func (vec *ViperEC) GetURIPrefix(key string) string {
	s, err := cast.ToStringE(vec.Get(key))
	if err != nil {
		vec.errs[key] = err
		return ""
	}
	if len(s) == 0 {
		vec.errs[key] = errors.New("cannot be empty")
		return ""
	}
	if s[0] != '/' {
		vec.errs[key] = errors.New("must start with /")
		return ""
	}
	if len(s) > 1 && s[len(s)-1] == '/' {
		vec.errs[key] = errors.New("must not end with /")
		return ""
	}
	return s
}

func (vec *ViperEC) HasErrors() bool {
	return len(vec.errs) != 0
}

func (vec *ViperEC) Errors() map[string]error {
	return vec.errs
}
