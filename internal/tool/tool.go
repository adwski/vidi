package tool

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	userapi "github.com/adwski/vidi/internal/api/user/client"
	videoapi "github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"github.com/adwski/vidi/internal/logging"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/enescakir/emoji"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type (
	// Tool is a vidit tui client side tool.
	Tool struct {
		userapi  *userapi.Client
		videoapi videoapi.UsersideapiClient
		httpC    *resty.Client

		logger *zap.Logger

		err error

		// tool's persistent state
		state *State

		// current screen
		screen screen

		// flag indicating that user should enter credentials
		enterCreds bool
		// quit screen flag
		quitting bool

		// homeDir
		dir string

		// main menu transitions
		mainFlowScreen int
	}

	// RemoteCFG is config that is dynamically retrieved from ViDi when tool starts.
	RemoteCFG struct {
		UserAPIURL  string `json:"user_api_url"`
		VideoAPIURL string `json:"video_api_url"`
		VidiCAB64   string `json:"vidi_ca"`
		vidiCA      []byte
	}
)

const (
	mainFlowScreenMainMenu = iota
	mainFlowScreenVideos
	mainFlowScreenUpload
	mainFlowScreenQuotas
)

// New creates ViDi tui tool instance.
func New() (*Tool, error) {
	dir, err := initStateDir()
	if err != nil {
		return nil, fmt.Errorf("cannot create state dir: %w", err)
	}

	logger, err := logging.GetZapLoggerFile(dir + logFile)
	if err != nil {
		return nil, fmt.Errorf("cannot configure logger: %w", err)
	}
	return &Tool{
		logger: logger,
		dir:    dir,
		httpC:  resty.New(),
	}, nil
}

// Run starts tool. It returns only on interrupt.
func (t *Tool) Run() int {
	t.initialize()
	if _, err := tea.NewProgram(t).Run(); err != nil {
		t.logger.Error("runtime error", zap.Error(err), zap.Stack("stack"))
		fmt.Println("runtime error:", err)
		return 1
	}
	return 0
}

// Init initializes current screen.
// It's part of bubbletea Model interface.
func (t *Tool) Init() tea.Cmd {
	return t.screen.init()
}

// Update routes bubbletea messages to active screen and processes
// state changes according to control data returned by views.
// It's part of bubbletea Model interface.
func (t *Tool) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(tea.KeyMsg); ok {
		if m.String() == "ctrl+c" {
			t.quitting = true
			return t, tea.Quit
		}
	}
	cmd, oc := t.screen.update(msg)
	t.logger.Debug("updating screen",
		zap.Any("screen", t.screen.name()),
		zap.Any("in-msg", msg),
		zap.Any("out-oc", oc))
	if oc == nil {
		// no outerControl means current screen is still active
		return t, cmd
	}

	// -----------------------------------------------------
	// Here we catch & process control messages from screens
	// -----------------------------------------------------
	switch dta := oc.data.(type) {
	case msgErrorScreenDone:
		t.err = nil
	case msgViDiURL:
		t.err = t.initClients(string(dta))
	case userControl:
		if t.state.userID(dta.username) == -1 {
			t.err = t.processNewUser(dta)
		} else {
			t.err = fmt.Errorf("user '%s' already exisits", dta.username)
		}
	case userSelectControl:
		t.enterCreds = true
		switch {
		case dta.option == optionUserSelectCurrent:
		case dta.option == optionUserLogInNew:
			t.state.CurrentUser = -1
		default:
			t.state.CurrentUser = dta.option
		}
	case reLogControl:
		t.err = t.processLoginExistingUser(t.state.getCurrentUserUnsafe().Name, dta.password)
	case mainMenuControl:
		switch dta.option {
		case mainMenuOptionSwitchUser:
			t.state.CurrentUser = -1
		case mainMenuOptionQuotas:
			t.state.getCurrentUserUnsafe().QuotaUsage, t.err = t.getQuotas()
			if t.err == nil {
				t.mainFlowScreen = mainFlowScreenQuotas
			}
		case mainMenuOptionVideos:
			t.state.getCurrentUserUnsafe().Videos, t.err = t.getVideos()
			t.mainFlowScreen = mainFlowScreenVideos
		case mainFlowScreenUpload:
			t.mainFlowScreen = mainFlowScreenUpload
		default:
		}
	case videosControl:
		if dta.vid == "" {
			t.mainFlowScreen = mainFlowScreenMainMenu
		} else {
			// TODO: switch to watch video screen
		}
	case quotasControl:
		t.mainFlowScreen = mainFlowScreenMainMenu
	case uploadControl:
		switch dta.msg {
		default:
			t.mainFlowScreen = mainFlowScreenMainMenu
		case uploadControlMsgFileSelected:
			t.err = t.prepareUpload(dta.path)
		}
	}
	// reconcile screen changes
	t.cycleViews()
	return t, t.screen.init()
}

// View delegates rendering to active screen or renders quit screen.
// It's part of bubbletea Model interface.
func (t *Tool) View() string {
	if t.quitting {
		return quitTextStyle.Render("bye" + emoji.WavingHand.String())
	}
	return t.screen.view()
}

// initialize initializes application up to the point
// that is possible with existing state (config).
func (t *Tool) initialize() {
	if t.state == nil {
		t.state = newState(t.dir)
	}
	if err := t.state.load(); err != nil {
		t.logger.Warn("cannot load state", zap.Error(err))
		// avoid side effects
		t.state = newState(t.dir)
	}
	if !t.state.noEndpoint() {
		t.err = t.initClients(t.state.Endpoint)
	}
	t.cycleViews()
}

func (t *Tool) noClients() bool {
	return t.userapi == nil || t.videoapi == nil
}

// cycleViews switches views according state changes.
// For the same state it should always fall into the same screen.
func (t *Tool) cycleViews() {
	// check basic stuff first
	switch {
	case t.err != nil:
		// in case of error, show it to user
		t.screen = newErrorScreen(t.err)
	case t.state.noEndpoint(), t.noClients():
		// invalid endpoint, should configure it again
		t.screen = newConfigScreen()
	case t.state.noUsers() || t.state.noUser() && t.enterCreds:
		// no users OR new user selected,
		// should register or login
		t.screen = newUserScreen()
	default:
		// Some users exist. From now on we can:
		// - use current user from state if its token is valid
		// - ask to re-enter password if token is invalid or expired
		// - provide option to select another saved user
		// - provide option to register or login as new user
		err := t.state.checkToken()
		if err == nil {
			// token is valid
			// proceed to main flow screens
			t.mainFlow()
		} else {
			// token is invalid
			t.logger.Debug("token check error", zap.Error(err))

			if t.enterCreds {
				// user previously chose to enter credentials,
				// proceed to reLog screen with selected user
				t.screen = newReLogScreen(t.state.getCurrentUserUnsafe().Name)
			} else {
				// display user selection options
				t.screen = newUserSelect(t.state.Users, t.state.CurrentUser)
			}
		}
	}
	t.enterCreds = false // reset creds flag
	t.logger.Debug("cycling screens", zap.Any("screen", t.screen.name()))
}

func (t *Tool) mainFlow() {
	switch t.mainFlowScreen {
	case mainMenuOptionQuotas:
		// quota usage screen
		t.screen = newQuotasScreen(t.state.getCurrentUserUnsafe().QuotaUsage)
	case mainFlowScreenVideos:
		// videos screen
		t.screen = newVideosScreen(t.state.getCurrentUserUnsafe().Videos)
	case mainFlowScreenUpload:
		// upload screen
		t.screen = newUploadScreen()
	default: // mainFlowScreenMainMenu
		// main menu
		t.screen = newMainMenuScreen(t.state.getCurrentUserUnsafe().Name)
	}
}

// getRemoteConfig retrieves json config from ViDi endpoint.
func (t *Tool) getRemoteConfig(ep string) (*RemoteCFG, error) {
	var rCfg RemoteCFG
	resp, err := t.httpC.NewRequest().
		SetHeader("Accept", "application/json").
		SetResult(&rCfg).Get(ep + configURLPath)
	if err != nil {
		return nil, fmt.Errorf("cannot contact ViDi endpoint: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("ViDi endpoint returned err status: %s", resp.Status())
	}
	rCfg.vidiCA, err = base64.StdEncoding.DecodeString(rCfg.VidiCAB64)
	if err != nil {
		return nil, fmt.Errorf("cannot decode vidi ca: %w", err)
	}
	t.logger.Debug("vidi ca decoded", zap.String("vidi_ca", rCfg.VidiCAB64))
	return &rCfg, nil
}

// initClients initializes video api and user api clients using provided ViDi endpoint.
func (t *Tool) initClients(ep string) error {
	rCfg, err := t.getRemoteConfig(ep)
	if err != nil {
		return err
	}

	// GRPC client is always spawned with tls creds
	// We're using CA from remote config
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(rCfg.vidiCA) {
		return fmt.Errorf("credentials: failed to append certificates")
	}
	creds := credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    cp,
	})
	cc, err := grpc.Dial(rCfg.VideoAPIURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("cannot create vidi connection: %w", err)
	}
	t.videoapi = videoapi.NewUsersideapiClient(cc)

	// Persist endpoint, since no more errors can be caught
	// until we start making actual requests
	t.state.Endpoint = ep
	if err = t.state.persist(); err != nil {
		return fmt.Errorf("cannot persist state: %w", err)
	}

	t.userapi = userapi.New(&userapi.Config{
		Endpoint: rCfg.UserAPIURL,
	})

	t.logger.Debug("successfully configured ViDi clients",
		zap.String("videoAPI", rCfg.VideoAPIURL),
		zap.String("userAPI", rCfg.UserAPIURL))
	return nil
}

// initStateDir creates state dir if it not exists.
func initStateDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot identify home dir: %w", err)
	}
	if err = os.MkdirAll(dir+stateDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create state directory: %w", err)
	}
	return dir + stateDir, nil
}

type (
	// screen is responsible for rendering set of elements
	// for particular purpose, i.e 'main menu screen' or 'new user screen'.
	screen interface {
		init() tea.Cmd
		update(msg tea.Msg) (tea.Cmd, *outerControl)
		view() string
		name() string
	}

	// outerControl is control structure returned by finished screen.
	// It contains data necessary to continue screen cycle.
	outerControl struct {
		data interface{}
	}
)

func (oc *outerControl) String() string {
	msg := ""
	switch t := oc.data.(type) {
	case msgViDiURL:
		msg = "vidi url: " + string(t)
	case msgErrorScreenDone:
		msg = "err screen done"
	case userControl:
		msg = "user control: " + t.String()
	case reLogControl:
		msg = "reLog control"
	case userSelectControl:
		msg = "userSelectControl: " + t.String()
	case mainMenuControl:
		msg = "mainMenuControl: " + t.String()
	case uploadControl:
		msg = "uploadControl"
	case quotasControl:
		msg = "quotasControl"
	}
	return msg
}
