package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/url"
	"os"
	"time"
)

const (
	stateDir  = "/.vidi"
	stateFile = "/state.json"
	logFile   = "/log.json"

	configURLPath = "/config.json"
)

type (
	State struct {
		Endpoint    string `json:"endpoint"`
		CurrentUser int    `json:"current_user"`
		Users       []User `json:"users"`

		dir string
		p   *jwt.Parser
	}

	User struct {
		Name           string       `json:"name"`
		Token          string       `json:"token"`
		TokenExpiresAt int64        `json:"expires_at"`
		CurrentUpload  *Upload      `json:"upload"`
		Videos         []Video      `json:"-"`
		QuotaUsage     []QuotaParam `json:"-"`
	}

	Video struct {
		ID        string
		Name      string
		Status    string
		Size      string
		CreatedAt string
	}

	Upload struct {
		ID       string `json:"id"`
		Filename string `json:"filename"`
		Parts    []Part `json:"parts"`
	}

	Part struct {
		Num      uint   `json:"num"`
		Size     uint   `json:"size"`
		Checksum string `json:"checksum"`
	}

	QuotaParam struct {
		Name  string
		Value string
	}
)

func newState(dir string) *State {
	return &State{
		CurrentUser: -1,
		dir:         dir,
		p:           jwt.NewParser(),
	}
}

func (s *State) getCurrentUserUnsafe() *User {
	if s.noUser() {
		return nil
	}
	return &s.Users[s.CurrentUser]
}

func (s *State) noUser() bool {
	return s.CurrentUser == -1
}

func (s *State) noUsers() bool {
	return len(s.Users) == 0
}

func (s *State) userID(name string) int {
	for i, u := range s.Users {
		if u.Name == name {
			return i
		}
	}
	return -1
}

func (s *State) noEndpoint() bool {
	return s.Endpoint == ""
}

func (s *State) load() error {
	f, err := os.Open(s.dir + stateFile)
	if err != nil {
		return fmt.Errorf("cannot open state file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("cannot read state file: %w", err)
	}
	if err = json.Unmarshal(b, s); err != nil {
		return fmt.Errorf("cannot unmarshal state file: %w", err)
	}
	if len(s.Users) == 0 {
		s.CurrentUser = -1
	}
	s.sanitize()
	return nil
}

func (s *State) persist() error {
	f, err := os.OpenFile(s.dir+stateFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open state file for writing: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal state: %w", err)
	}
	if _, err = f.Write(b); err != nil {
		return fmt.Errorf("cannot write state file: %w", err)
	}
	return nil
}

func (s *State) sanitize() {
	// check user idx validity
	if len(s.Users) <= s.CurrentUser {
		s.CurrentUser = 0
	}
	// check usernames
	ln := len(s.Users)
	for i := 0; i < ln; {
		if len(s.Users[i].Name) < 3 {
			// invalid user, remove it
			if i < ln-1 {
				s.Users[i], s.Users[ln-1] = s.Users[ln-1], s.Users[i]
			}
			ln--
			s.Users = s.Users[:ln]
			// reset current user idx for safety
			s.CurrentUser = 0
		} else {
			i++
		}
	}
	if len(s.Users) == 0 {
		s.CurrentUser = -1
	}
	// check endpoint
	if _, err := url.Parse(s.Endpoint); err != nil {
		s.Endpoint = ""
	}
}

func (s *State) checkToken() error {
	if s.noUser() {
		return errors.New("no user selected")
	}
	t := s.Users[s.CurrentUser].Token
	if t == "" {
		return errors.New("token is empty")
	}
	token, _, err := s.p.ParseUnverified(t, jwt.MapClaims{})
	if err != nil {
		s.Users[s.CurrentUser].Token = ""
		return fmt.Errorf("cannot parse token: %w", err)
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return fmt.Errorf("unable to get epiration time from token: %w", err)
	}
	if exp.Before(time.Now()) {
		return fmt.Errorf("token is expired at %s", exp.Format(time.RFC3339))
	}
	s.Users[s.CurrentUser].TokenExpiresAt = exp.Unix()
	return s.persist()
}
