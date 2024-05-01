package tool

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"math/rand"
)

const (
	greetMessage = "// Welcome to Vidi terminal tool //"
)

var (
	defaultHuhTheme = huh.ThemeDracula()

	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4).Bold(true).Faint(true)

	errStyle = defaultHuhTheme.Focused.ErrorIndicator.
			Padding(0, 0, 1, 4).
			Bold(true).
			Border(lipgloss.RoundedBorder())

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
	vidiSplashes   = []string{
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
		`.##..##..######..#####...######.
.##..##....##....##..##....##...
.##..##....##....##..##....##...
..####.....##....##..##....##...
...##....######..#####...######.
................................`,
		` ___      ___ ___  ________  ___     
|\  \    /  /|\  \|\   ___ \|\  \    
\ \  \  /  / | \  \ \  \_|\ \ \  \   
 \ \  \/  / / \ \  \ \  \ \\ \ \  \  
  \ \    / /   \ \  \ \  \_\\ \ \  \ 
   \ \__/ /     \ \__\ \_______\ \__\
    \|__|/       \|__|\|_______|\|__|`}
)

func init() {
	vidiSplashText = vidiSplashes[rand.Intn(5)]
}
