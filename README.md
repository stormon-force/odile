# Odile
Odile is a simple GUI for the [croc](https://github.com/schollz/croc) utility by Schollz. This program uses [Fyne](https://fyne.io/), a UI toolkit written in Go, as the graphical interface. Effort was made to keep the language in Go (what croc is written in) and the code in one file.

## Current Status
* Fyne works well, but the file selection doesn't fulfill all of croc's needs. Features are ongoing, changes are possible. (perhaps like the [FileZilla interface](https://filezilla-project.org/))
* Additional features like utilizing a custom relay server are not implemented. 
* GUI is still lacking in error checking and communicating information to the user
* Currently program has only been tested on Windows 10.

Any comments or suggestions are appreciated.

## Build
To compile without background terminal, add these flags.
```go build -ldflags -H=windowsgui gui.go```

*Windows Firewall keeps asking about this program, what gives?*
https://stackoverflow.com/questions/55201561/golang-run-on-windows-without-deal-with-the-firewall