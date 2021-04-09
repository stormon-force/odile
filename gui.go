package main

import (
	"log"
	"flag"
	"strings"
	"time"
	"os"
	"fmt"

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
	VERSION = "0.1.0"
)

// TO DO : Better error checking
// TO DO : include folders as selection process

// Utility functions
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

// Wrap croc logic
type CrocWrapper struct {
	Client *croc.Client // Current croc client
	Transmitting bool 	// Transmission active
}

// Mimic func send(c *cli.Context) located in github.com/schollz/croc/v8/src/croc
// Setting defaults for mostly everything
func (cw *CrocWrapper) Send(paths []string) (secret string, err error) {
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

	cw.Client, err = croc.New(crocOptions)
	if err != nil {
		return
	}

	// Before starting alert the progess bar graphic
	cw.Transmitting = true

	// File paths will already be choosen by the GUI portion
	// Run in seperate go routine
	go cw.Client.Send(croc.TransferOptions{
		PathToFiles:      paths,
		KeepPathInRemote: false,
	})

	return crocOptions.SharedSecret, nil
}

// Mimic func receive(c *cli.Context) located in github.com/schollz/croc/v8/src/croc
// Setting defaults for mostly everything
func (cw *CrocWrapper) Recv(secret string) (err error) {
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

	cw.Client, err = croc.New(crocOptions)
	if err != nil {
		return
	}

	// Before starting alert the progess bar graphic
	cw.Transmitting = true

	err = cw.Client.Receive()
	return
}

type OdileGUI struct {
	// Application member variables
	// Handle logic for croc operations (send, receive)
	Croc 	*CrocWrapper
	// Support multiple files
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

	// Progress Bar
	ProgressBar *widget.ProgressBar

	// Display password when pressing send
	SendPasswordLabel *widget.Label
	FileChoiceLabel *widget.Label

	// GUI Application
	App 	fyne.App
	Window 	fyne.Window
	Content *fyne.Container
}

func (g *OdileGUI) AddPathList(path string, pathList []string) error{
	pathPresent := false
	for _, p := range pathList{
		if p == path{
			pathPresent = true
		}
	}
	if !pathPresent{
		g.FileList = append(g.FileList, path)
		log.Println("Added %v to file path list", path)
	} else {
		return fmt.Errorf("Could not add %v, already in list", path)
	}

	return nil
}

func (g *OdileGUI) RefreshFileList(pathList []string){
	pathString := strings.Join(
		pathList,
		"\n",
	)
	g.FileChoiceLabel.SetText(pathString)
}

func (g *OdileGUI) RunProgressBar(){
	// Spin locks ahoy!	
	// Wait until send is ready
	for !g.Croc.Transmitting{
	}
	log.Println("Croc client is ready, waiting for file info transfer")
	// Spin lock until other side is ready
	for !g.Croc.Client.Step2FileInfoTransfered{
	}
	log.Println("File info transfer is ready")

	var TotalSize int64
	TotalSize = 0
	for _, file := range g.Croc.Client.FilesToTransfer{
		TotalSize += file.Size
	}
	log.Printf("Total size of all file: %v\n", TotalSize)

	for !g.Croc.Client.SuccessfulTransfer{
		g.ProgressBar.SetValue(float64(g.Croc.Client.TotalSent) / float64(TotalSize))
		time.Sleep(time.Millisecond * 10)
	}

	g.ProgressBar.SetValue(1.0)
	time.Sleep(time.Millisecond * 1000)
	g.ProgressBar.SetValue(0.0)
}

// TO DO : Move action logic (send, receive, choose file) out of GUI declaration?
func (g *OdileGUI) Init(){
	g.Croc = &CrocWrapper{}

	g.FileList = []string{}

	g.App = app.New()
	g.Window = g.App.NewWindow("croc-Odile")

	g.ProgressBar = widget.NewProgressBar()

	g.FileSelect = dialog.NewFileOpen(
		func (fileChoice fyne.URIReadCloser, err error) {
			log.Println("File selection result", fileChoice, err)
			if(fileChoice != nil){
				log.Println("\tFile:", fileChoice.URI())
				errA := g.AddPathList(FormatFileChoice(fileChoice), g.FileList)
				if(errA != nil){
					log.Println(errA)
				} else {
					g.RefreshFileList(g.FileList)
				}
				log.Println(g.FileList)
			}
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
		go g.RunProgressBar()
		secret, err := g.Croc.Send(g.FileList)

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
		go g.RunProgressBar()
		err = g.Croc.Recv(secret)
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
		g.ProgressBar,
	)
}

func Run(debugOption bool) {
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

func SetLogOutput() {
	path := "./logs"
	if _, err := os.Stat(path); os.IsNotExist(err) {
	    os.Mkdir(path, os.ModeDir)
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	logfileName := "./logs/odile_" + timestamp + ".txt"
	
	// Create if not there
	f, err := os.Create(logfileName)
	// Why do we need OpenFile? Here: https://stackoverflow.com/questions/19965795/how-to-write-log-to-file
	file, err := os.OpenFile(logfileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(file)
	f.Close()
}

func main() {
	debugOption := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	//if(*debugOption){
	//SetLogOutput()
	//}
	log.Println("Starting Odile", VERSION)

	Run(*debugOption)
}