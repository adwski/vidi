package tool

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"math/rand"
)

const (
	greetMessageTxt = "Welcome to Vidi terminal tool"
)

var (
	defaultHuhTheme = huh.ThemeDracula()

	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4).Bold(true).Faint(true)

	errStyle = defaultHuhTheme.Focused.ErrorIndicator.
			Padding(0, 0, 1, 4).
			Bold(true).
			Border(lipgloss.RoundedBorder())

	containerWithBorder = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Border(lipgloss.RoundedBorder()).
				Margin(1, 1, 0, 0).
				Padding(0, 1, 0, 1)

	greetStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(defaultHuhTheme.Focused.Title.GetForeground()).
			Background(defaultHuhTheme.Focused.Title.GetBackground()).
			Border(lipgloss.RoundedBorder()).
			Margin(1, 3, 0, 0).
			Padding(1, 2).
			Height(9).
			Width(50)

	vidiSplashText = ""
	greetMessage   = ""
	greetMessages  = []string{
		" ░▒▓█ " + greetMessageTxt + " █▓▒░ ",
		greetMessageTxt,
		" +#+ " + greetMessageTxt + " +#+ ",
		" ## " + greetMessageTxt + " ## ",
		` \\ ` + greetMessageTxt + ` \\ `,
	}
	vidiSplashes = []string{
		`░  ░░░░  ░░        ░░       ░░░        ░
▒  ▒▒▒▒  ▒▒▒▒▒  ▒▒▒▒▒  ▒▒▒▒  ▒▒▒▒▒  ▒▒▒▒
▓▓  ▓▓  ▓▓▓▓▓▓  ▓▓▓▓▓  ▓▓▓▓  ▓▓▓▓▓  ▓▓▓▓
███    ███████  █████  ████  █████  ████
████  █████        ██       ███        █`,
		`██╗   ██╗██╗██████╗ ██╗
██║   ██║██║██╔══██╗██║
██║   ██║██║██║  ██║██║
╚██╗ ██╔╝██║██║  ██║██║
 ╚████╔╝ ██║██████╔╝██║
  ╚═══╝  ╚═╝╚═════╝ ╚═╝`,
		`:::     ::: ::::::::::: ::::::::: ::::::::::: 
:+:     :+:     :+:     :+:    :+:    :+:     
+:+     +:+     +:+     +:+    +:+    +:+     
+#+     +:+     +#+     +#+    +:+    +#+     
 +#+   +#+      +#+     +#+    +#+    +#+     
  #+#+#+#       #+#     #+#    #+#    #+#     
    ###     ########### ######### ###########`,
		`.....##..##..######..#####...######.
.....##..##....##....##..##....##...
.....##..##....##....##..##....##...
......####.....##....##..##....##...
.......##....######..#####...######.
....................................`,
		` ___      ___ ___  ________  ___     
|\  \    /  /|\  \|\   ___ \|\  \    
\ \  \  /  / | \  \ \  \_|\ \ \  \   
 \ \  \/  / / \ \  \ \  \ \\ \ \  \  
  \ \    / /   \ \  \ \  \_\\ \ \  \ 
   \ \__/ /     \ \__\ \_______\ \__\
    \|__|/       \|__|\|_______|\|__|`}
)

func init() {
	n := rand.Intn(5)
	vidiSplashText = vidiSplashes[n]
	greetMessage = greetMessages[n]
}
