package main

//go get "github.com/hegedustibor/htgo-tts"
import "github.com/hegedustibor/htgo-tts"
import "time"

type Speech struct {
	Folder   string
	Language string
}

func main() {
    speech := htgotts.Speech{Folder: "audio", Language: "de"}
    speech.Speak(time.Now().Format("2006-01-02 15:04:05"))
}
