// Package tool contains TUI client that can be used together with ViDi.
// It supports most of user-operations (excluding video watch).
//
// The tool (called 'vidit') is build using bubbletea ecosystem and
// has corresponding design patterns.
//
// Vidit consists of bubbletea program and number of screens, each of them
// responsible for displaying particular data and interacting
// with user about this data if necessary (e.g. menu screen, upload screen).
//
// At each particular time only one screen is active, and this is
// determined by vidit's state and user input (e.g. Vidit has no initial
// config and user should provide it).
//
// Bubbletea message flow is redirected to the active screen which is
// responsible for reacting to messages and rendering info using
// same bubbletea Update()-View() pattern.
//
// Vidit keeps internal state in order to track active user,
// login info, upload info, endpoint configuration. This allows
// to auto login during startup and resume interrupted upload.
//
// Each screen internally also has its own state which in most cases
// CANNOT BE REVERTED (main example is huh forms), which is why
// when vidit changes screens it instantiates new one and not reuses
// exising.
//
// During message processing screen may send outerControl message
// back to vidit Update() method. And this means vidit should decide
// what to do next (usual case is that user finished interacting with
// current screen, and it should be changed to something else).
//
// Vidit also has 'world-event' channel that is 'external world to active screen'
// communication channel. This channel is used when screen should track something
// happening asynchronously (independent from screen itself). This is mainly
// used for upload progress.
//
//nolint:godot // false positives
package tool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"

	userapi "github.com/adwski/vidi/internal/api/user/client"
	videoapi "github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"github.com/adwski/vidi/internal/logging"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/enescakir/emoji"
	"github.com/go-resty/resty/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const exitCodeErrAlreadyStarted = 2

var ErrAlreadyStarted = errors.New("already started")

type (
	// Tool is a vidit tui client side tool.
	// It has clients to interact with APIs,
	// screens to display info to user and
	// react to user input, and state that is
	// persisted between restarts.
	Tool struct { // a.k.a 'Vidit'
		userapi  *userapi.Client
		videoapi videoapi.UsersideapiClient
		httpC    *resty.Client
		logger   *zap.Logger
		prog     *tea.Program

		// feedback channel to turn any world event into tea.Msg
		fb chan tea.Msg

		err error

		// tool's persistent state
		state *State

		// current screen
		screen screen

		// tool's homeDir
		dir string
		// file picker dir
		filePickerDir string

		// flag indicating that user selected to enter credentials
		enterCreds bool
		// flag indicating that user selected to resume upload
		resumingUpload bool
		// quit screen flag
		quitting bool

		// started flag, to enforce tool's ability to be started once.
		started bool

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

	// Config is Tool's config. It is optional and used together with NewWithConfig()
	Config struct {
		EnforceHomeDir string
		FilePickerDir  string
		EarlyInit      bool
	}
)

// New creates ViDi tui tool instance.
func New() (*Tool, error) {
	return NewWithConfig(Config{})
}

// NewWithConfig creates ViDi tui tool instance using specified config.
// Logger is instantiated with file output with hardcoded debug level.
// It could also run early initialization if specified (mainly used by tests).
func NewWithConfig(cfg Config) (*Tool, error) {
	fs := pflag.NewFlagSet("main", pflag.ContinueOnError)
	defFPDir := fs.StringP("default-file-picker-dir", "d", "", "default upload file picker dir")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("cannot parse command line arguments: %w", err)
	}

	dir, err := initStateDir(cfg.EnforceHomeDir)
	if err != nil {
		return nil, fmt.Errorf("cannot create state dir: %w", err)
	}

	logger, err := logging.GetZapLoggerFile(dir + logFile)
	if err != nil {
		return nil, fmt.Errorf("cannot configure logger: %w", err)
	}
	t := &Tool{
		logger: logger,
		dir:    dir,
		httpC:  resty.New(),
		fb:     make(chan tea.Msg),
	}
	if t.filePickerDir, err = getFilePickerDefaultDir(*defFPDir, cfg.EnforceHomeDir); err != nil {
		return nil, err
	}

	if cfg.EarlyInit {
		t.initialize()
	}
	return t, nil
}

func getFilePickerDefaultDir(dirFromArgs, dirFromConfig string) (string, error) {
	if dirFromArgs != "" {
		if info, err := os.Stat(dirFromArgs); err == nil && info.IsDir() {
			// valid dir
			return dirFromArgs, nil
		}
		// not valid dir, move silently to other options
	}

	if dirFromConfig != "" {
		// caller should take care of validity of this
		return dirFromConfig, nil
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user home directory: %w", err)
	}
	return dir, nil
}

// Run starts tool. It returns only on interrupt.
func (t *Tool) Run() int {
	return t.RunWithContext(context.Background())
}

func (t *Tool) RunWithContext(ctx context.Context) int {
	if t.started {
		t.logger.Error("cannot start tool more than once", zap.Error(ErrAlreadyStarted))
		return exitCodeErrAlreadyStarted
	}
	var code int
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := t.run(ctx, wg); err != nil {
		code = 1
	}
	cancel()
	wg.Wait()
	return code
}

func (t *Tool) RunWithProgram(ctx context.Context, wg *sync.WaitGroup, errc chan<- error, prog *tea.Program) {
	if t.started {
		errc <- ErrAlreadyStarted
		return
	}
	t.prog = prog
	t.listenForEvents(ctx, wg)
}

// run spawns tea program and world event loop.
func (t *Tool) run(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	t.initialize()
	var teaOpts = []tea.ProgramOption{tea.WithContext(ctx)}
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		// workaround for missing tty when running in CI
		// https://github.com/charmbracelet/bubbletea/issues/761
		teaOpts = append(teaOpts, tea.WithInput(nil))
	}
	t.prog = tea.NewProgram(t, teaOpts...)
	wg.Add(1)
	go t.listenForEvents(ctx, wg)
	if _, err := t.prog.Run(); err != nil {
		if !errors.Is(err, tea.ErrProgramKilled) { // ErrProgramKilled happens when context is canceled
			t.logger.Debug("runtime error", zap.Error(err), zap.Stack("stack"))
			return fmt.Errorf("runtime error: %w", err)
		}
	}
	t.logger.Debug("program exited")
	return nil
}

// listenForEvents proxies world events to tea message flow.
func (t *Tool) listenForEvents(ctx context.Context, wg *sync.WaitGroup) {
Loop:
	for {
		select {
		case msg := <-t.fb:
			t.prog.Send(msg)
		case <-ctx.Done():
			break Loop
		}
	}
	wg.Done()
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
		// catch quit combination
		if m.String() == "ctrl+c" {
			t.quitting = true
			return t, tea.Quit
		}
	}
	// send message to active screen
	cmd, oc := t.screen.update(msg)
	t.logger.Debug("updating screen",
		zap.Any("screen", t.screen.name()),
		zap.Any("in-msg", msg),
		zap.Any("out-oc", oc))
	if oc == nil {
		// no outerControl means current screen doesn't need outside help
		return t, cmd
	}

	cycle := true

	// -----------------------------------------------------------
	// Here we catch & process control messages from active screen
	// -----------------------------------------------------------
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
		t.err = t.processLoginExistingUser(t.state.activeUserUnsafe().Name, dta.password)
	case mainMenuControl:
		switch dta.option {
		case mainMenuOptionSwitchUser:
			t.state.CurrentUser = -1
		case mainMenuOptionQuotas:
			t.state.activeUserUnsafe().QuotaUsage, t.err = t.getQuotas()
			if t.err == nil {
				t.mainFlowScreen = mainFlowScreenQuotas
			}
		case mainMenuOptionVideos:
			t.state.activeUserUnsafe().Videos, t.err = t.getVideos()
			t.mainFlowScreen = mainFlowScreenVideos
		case mainFlowScreenUpload:
			t.mainFlowScreen = mainFlowScreenUpload
		case mainMenuOptionResumeUpload:
			t.resumingUpload = true
			t.mainFlowScreen = mainFlowScreenUpload
			// Spawn resume upload goroutine, it will produce world events.
			// TODO Should add context (at the moment not so obvious where to take it from)
			go t.resumeUploadFileNotify(t.state.activeUserUnsafe().CurrentUpload)

		default:
		}
	case videosControl:
		switch {
		case dta.vid == "":
			t.mainFlowScreen = mainFlowScreenMainMenu
		case dta.delete:
			t.err = t.deleteVideo(dta.vid)
			t.mainFlowScreen = mainFlowScreenMainMenu
		case dta.watch:
			go t.getWatchURL(dta.vid)
		}
	case quotasControl:
		t.mainFlowScreen = mainFlowScreenMainMenu
	case uploadControl:
		switch dta.msg {
		default:
			t.mainFlowScreen = mainFlowScreenMainMenu
		case uploadControlMsgFileSelected:
			// Spawn upload goroutine, it will produce world events.
			go t.uploadFileNotify(dta.name, dta.path)
			// We're not switching screens here.
			cycle = false
		case uploadControlMsgDone:
			t.mainFlowScreen = mainFlowScreenMainMenu
			t.state.activeUserUnsafe().CurrentUpload = nil
		}
	}
	if cycle {
		// reconcile screen changes
		t.cycleViews()
		return t, t.screen.init()
	}
	return t, nil
}

// View delegates rendering to active screen or renders quit screen.
// It's part of bubbletea Model interface.
func (t *Tool) View() string {
	if t.quitting {
		return quitTextStyle.Render("bye" + emoji.WavingHand.String())
	}
	return t.screen.view()
}

// cycleViews switches views according state changes.
// For the same state it should always fall into the same screen.
func (t *Tool) cycleViews() {
	// check basic stuff first
	switch {
	case t.err != nil:
		// in case of error, show it to user
		t.logger.Debug("error occurred, switching to err screen", zap.Error(t.err))
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
				t.screen = newReLogScreen(t.state.activeUserUnsafe().Name)
			} else {
				// display user selection options
				t.screen = newUserSelect(t.state.Users, t.state.CurrentUser)
			}
		}
	}
	t.enterCreds = false     // reset creds flag
	t.resumingUpload = false // reset resume flag
	t.logger.Debug("cycling screens", zap.Any("screen", t.screen.name()))
}

func (t *Tool) mainFlow() {
	switch t.mainFlowScreen {
	case mainMenuOptionQuotas:
		// quota usage screen
		t.screen = newQuotasScreen(t.state.activeUserUnsafe().QuotaUsage)
	case mainFlowScreenVideos:
		// videos screen
		t.screen = newVideosScreen(t.state.activeUserUnsafe().Videos)
	case mainFlowScreenUpload:
		// upload screen
		t.screen = newUploadScreen(t.filePickerDir, t.resumingUpload)
	default: // mainFlowScreenMainMenu
		// main menu
		u := t.state.activeUserUnsafe()
		t.screen = newMainMenuScreen(u.Name, u.CurrentUpload != nil)
	}
}

const (
	mainFlowScreenMainMenu = iota
	mainFlowScreenVideos
	mainFlowScreenUpload
	mainFlowScreenQuotas
)

type (
	// screen is responsible for rendering set of elements
	// for a particular purpose, i.e 'main menu screen' or 'new user screen'.
	screen interface {
		init() tea.Cmd
		update(msg tea.Msg) (tea.Cmd, *outerControl)
		view() string
		name() string
	}

	// outerControl is a control structure returned by screen.
	// It contains data necessary to continue screen cycle.
	outerControl struct {
		data interface{} // TODO may be just 'outerControl any' ?
	}
)

func (oc *outerControl) String() string {
	var msg string
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
