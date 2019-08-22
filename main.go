package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"github.com/xausky/hand-network-loginer/icon"
	"github.com/xausky/hand-network-loginer/loginer"

	"github.com/shibukawa/configdir"
	"golang.org/x/crypto/ssh/terminal"
)

type Authorization struct {
	Username string
	Password string
}

var authorization Authorization

func main() {
	config := flag.Bool("config", false, "Edit username and password.")
	flag.Parse()
	configDir := configdir.New("xausky", "hand-network-loginer").QueryFolders(configdir.Global)[0]
	configData, err := configDir.ReadFile("authorization.json")
	if *config || os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Username: ")
		username, err := reader.ReadString('\n')
		if err != nil {
			log.Panic(err)
		}
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Panic(err)
		}
		authorization.Username = strings.TrimSpace(username)
		authorization.Password = base64.StdEncoding.EncodeToString(bytePassword)
		fmt.Printf("\nConfig %s success, use 'hand-network-loginer' to launche.", authorization.Username)
		data, err := json.Marshal(&authorization)
		if err != nil {
			log.Panic(err)
		}
		configDir.WriteFile("authorization.json", data)
	} else {
		err := json.Unmarshal(configData, &authorization)
		if err != nil {
			log.Panic(err)
		}
		systray.Run(onReady, func() {
			log.Println("hand-network-loginer exit.")
		})
	}
}

func onReady() {
	systray.SetIcon(icon.HandIcon)
	systray.SetTooltip("Auto login hand network and monitor network state.")
	aboutMenuItem := systray.AddMenuItem("About", "About the software")
	quitMenuItem := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-quitMenuItem.ClickedCh:
				systray.Quit()
				return
			case <-aboutMenuItem.ClickedCh:
				open.Run("https://github.com/xausky/hand-network")
			}
		}
	}()
	log.Println("hand-network-loginer start.")
	go func() {
		login := loginer.Loginer{}
		login.Username = authorization.Username
		login.Password = authorization.Password
		login.Login()
		for {
			time.Sleep(time.Duration(float32(login.HeartBeatCyc)*0.8) * time.Millisecond)
			login.HeartBeat()
		}
	}()
}
