Output tapes/showcase.gif
Output tapes/frames/

Set Theme Atom
Set Shell zsh
Set Width 1600
Set Height 900
Set TypingSpeed 250ms

Set Margin 10
Set MarginFill "#6B50FF"
Set BorderRadius 10
Set Framerate 24

Set PlaybackSpeed 1.0

# prepare
Enter@2
Type "echo 'Generating certificates'"
Enter@1

Type "make tls"
Sleep 1
Enter@4
Enter
Type "echo 'Starting compose project'"
Enter@1

# images better be pre-built
Type "make docker-dev"
Sleep 1
Enter@5
Type "make docker-ps"
Enter@5

# start vidit
Type "clear"
Enter@1
Type "echo 'Starting vidit tool'"
Enter@1
Type "echo '-d sets starting directory for file picker'"
Enter@1
Type "go run cmd/vidit/main.go -d $(pwd)/testfiles"
Enter@5

# config screen
Type "http://localhost"
Enter@3

# new user screen
# select register
Down@1
Enter@2

# enter creds
Type "JaneDoe"
Enter@2
Type "password"
Enter@2
Left@1
Enter@5

# show empty videos screen
Enter@3
Backspace@1

# show quotas screen
Down@1
Down@1
Enter@1
Down@1
Down@3
Backspace@2

# goto uploads screen
Down@2
Enter@2
# select file
Down@1
Enter@1

# enter name
Type "test video"
Sleep 1
Enter@1
# confirm
Left@1
Enter@7
Enter@2

# show quotas screen
Down@1
Down@1
Enter@1
Down@1
Down@3
Backspace@2

# show videos screen
Enter@2
Type@6 w
Backspace@5
Ctrl+c
Sleep 3

# show state
Type "echo 'Vidit persists its state into home dir'"
Enter@1
Type "ls $HOME/.vidi"
Enter@2

Type "cat $HOME/.vidi/state.json"
Enter@2

Type "echo 'Next run will reuse endpoint config and user info'"
Enter@1

# start again
Type "go run cmd/vidit/main.go"
Enter@5

# show videos screen
Enter@3
Backspace@5
Ctrl+c
Sleep 3
