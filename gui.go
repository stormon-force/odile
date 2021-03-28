package main

import (
	"log"
	"flag"
	"strings"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"

	"github.com/schollz/croc/v8/src/croc"
	"github.com/schollz/croc/v8/src/models"
	"github.com/schollz/croc/v8/src/utils"
)

const (
	VERSION = "1.0.0"
)

// TO DO : Better error checking
// TO DO : Multiple files, include folders

// Mimic func send(c *cli.Context) located in github.com/schollz/croc/v8/src/croc
// Setting defaults for mostly everything
func Send(paths []string) (secret string, err error) {
	//setDebugLevel(c)
	//comm.Socks5Proxy = c.String("socks5")
	crocOptions := croc.Options{
		//SharedSecret:   c.String("code"),
		IsSender:       true,
		Debug:          false,
		//NoPrompt:       c.Bool("yes"),
		RelayAddress:   models.DEFAULT_RELAY,
		RelayAddress6:  models.DEFAULT_RELAY6,
		//Stdout:         c.Bool("stdout"),
		//DisableLocal:   c.Bool("no-local"),
		//OnlyLocal:      c.Bool("local"),
		//IgnoreStdin:    c.Bool("ignore-stdin"),
		RelayPorts:     strings.Split("9009,9010,9011,9012,9013", ","),
		//Ask:            c.Bool("ask"),
		//NoMultiplexing: c.Bool("no-multi"),
		RelayPassword:  models.DEFAULT_PASSPHRASE,
		//SendingText:    c.String("text") != "",
		//NoCompress:     c.Bool("no-compress"),
	}
	if crocOptions.RelayAddress != models.DEFAULT_RELAY {
		crocOptions.RelayAddress6 = ""
	} else if crocOptions.RelayAddress6 != models.DEFAULT_RELAY6 {
		crocOptions.RelayAddress = ""
	}

	if len(crocOptions.SharedSecret) == 0 {
		// generate code phrase
		crocOptions.SharedSecret = utils.GetRandomName()
	}

	cr, err := croc.New(crocOptions)
	if err != nil {
		return
	}

	// File paths will already be choosen by the GUI portion
	// Run in seperate go routine
	go cr.Send(croc.TransferOptions{
		PathToFiles:      paths,
		KeepPathInRemote: false,
	})

	return crocOptions.SharedSecret, nil
}

// Mimic func receive(c *cli.Context) located in github.com/schollz/croc/v8/src/croc
// Setting defaults for mostly everything
func Recv(secret string) (err error) {
	//comm.Socks5Proxy = c.String("socks5")
	crocOptions := croc.Options{
		SharedSecret:  secret, // passed in from GUI
		IsSender:      false,
		//Debug:         c.Bool("debug"),
		NoPrompt:      true, // Disable yes/no files
		RelayAddress:   models.DEFAULT_RELAY,
		RelayAddress6:  models.DEFAULT_RELAY6,
		//Stdout:        c.Bool("stdout"),
		//Ask:           c.Bool("ask"),
		RelayPassword:  models.DEFAULT_PASSPHRASE,
		//OnlyLocal:     c.Bool("local"),
		//IP:            c.String("ip"),
	}
	if crocOptions.RelayAddress != models.DEFAULT_RELAY {
		crocOptions.RelayAddress6 = ""
	} else if crocOptions.RelayAddress6 != models.DEFAULT_RELAY6 {
		crocOptions.RelayAddress = ""
	}

	cr, err := croc.New(crocOptions)
	if err != nil {
		return
	}

	err = cr.Receive()
	return
}

func CombineWords(word1 string, word2 string, word3 string) string {
	if(word1 == "" || word2 == "" || word3 == "") {
		return ""
	}
	return strings.Join(
		[]string{word1, word2, word3},
		"-",
	)
}

func FormatFileChoice(fileChoice fyne.URIReadCloser) string {
	// TO DO : This probably won't work for non UTF-8 or UNIX filepaths
	// Remove 'file://' portion
	return fileChoice.URI().String()[7:]
}

type OdileGUI struct {
	// Application member variables
	// Currently only supporting one file
	FileList []string 
	// Where files will be placed
	OutputPath string 

	// GUI member variables
	// File selection window
	FileSelect  	*dialog.FileDialog

	// File selection window button
	FileOpenButton 	*widget.Button

	// Send and received button
	SendButton 	*widget.Button
	RecvButton 	*widget.Button

	// Password for receive
	Input1 	*widget.Entry
	Input2 	*widget.Entry
	Input3 	*widget.Entry

	// Display password when pressing send
	SendPasswordLabel *widget.Label
	FileChoiceLabel *widget.Label

	// GUI Application
	App 	fyne.App
	Window 	fyne.Window
	Content *fyne.Container
}

// TO DO : Move action logic (send, receive, choose file) out of GUI declaration
func (g *OdileGUI) Init(){
	g.FileList = make([]string, 1)

	g.App = app.New()
	g.Window = g.App.NewWindow("croc-Odile")

	g.FileSelect = dialog.NewFileOpen(
		func (fileChoice fyne.URIReadCloser, err error) {
			log.Println("File selection result", fileChoice, err)
			if(fileChoice != nil){
				log.Println("\tFile:", fileChoice.URI())
				g.FileList[0] = FormatFileChoice(fileChoice)
			}
			g.FileChoiceLabel.SetText(g.FileList[0])
		},
		g.Window,
	)

	g.FileOpenButton = widget.NewButton("Choose File", func() {
		g.FileSelect.Show()
	})

	g.Input1 = widget.NewEntry()
	g.Input2 = widget.NewEntry()
	g.Input3 = widget.NewEntry()

	g.SendButton = widget.NewButton("Send", func() {
		log.Println(g.FileList)
		secret, err := Send(g.FileList)
		log.Println("Send Function:", secret, err)
		g.SendPasswordLabel.SetText(secret)
	})

	g.RecvButton = widget.NewButton("Receive", func() {
		secret := CombineWords(
			g.Input1.Text,
			g.Input2.Text,
			g.Input3.Text,
		)
		// A word was incorrectly formatted
		if(secret == ""){ 
			log.Println("Error in one of input words, not receiving")
			return
		}
		var err error
		if err = os.Chdir(g.OutputPath); err != nil {
			log.Panicln("Failed to change to proper directory, ending program", err)
		}
		err = Recv(secret)
		log.Println("Receive Function:", secret, err)
	})

	g.SendPasswordLabel = widget.NewLabel("")
	g.FileChoiceLabel = widget.NewLabel("")

	g.Content = container.NewVBox(
		g.FileOpenButton,
		g.FileChoiceLabel,
		g.SendButton,
		g.SendPasswordLabel,
		g.RecvButton,
		g.Input1,
		g.Input2,
		g.Input3,
	)
}

func Start(debugOption bool) {
	// Create received files directory
	path := "./output"
	if _, err := os.Stat(path); os.IsNotExist(err) {
	    os.Mkdir(path, os.ModeDir)
	}

	// Create GUI 
	Gui := OdileGUI{
		OutputPath: path,
	}
	Gui.Init()

	// Set content and window size
	Gui.Window.SetContent(Gui.Content)
	Gui.Window.Resize(fyne.NewSize(500, 500))

	// Display and run
	Gui.Window.ShowAndRun()
}

func main() {
	log.Println("Starting Odile", VERSION)
	debugOption := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	Start(*debugOption)
}