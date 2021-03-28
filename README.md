# Odile
Odile is a simple GUI for the [croc](https://github.com/schollz/croc) utility by Schollz. This program uses [Fyne](https://fyne.io/), a UI toolkit written in Go, as the graphical interface. Effort was made to keep the language in Go (what croc is written in) and the code in one file.

## Current Status
* Fyne works well, but the file selection currently doesn't fulfill all of croc's needs. I am looking into allowing multiple selection options and folders (perhaps like the [FileZilla interface](https://filezilla-project.org/))
* Additional features like utilizing a custom relay server are not implemented. 
* GUI is still lacking in error checking and communicating information to the user (such as the progress bar, what files will be accepted, etc.)
* Currently program has only been tested on Windows 10.

Any comments or suggestions are appreciated.

## Build
To compile without background terminal, add these flags.
```go build -ldflags -H=windowsgui gui.go```