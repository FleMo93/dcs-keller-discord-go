package main

import (
	m "dcskellerdiscordgo"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type configJson struct {
	Discord struct {
		Token                 string `json:"token"`
		Channel               string `json:"channel"`
		ServerStatusMessageID string `json:"serverStatusMessageId"`
	} `json:"discord"`
	Dcs struct {
		ServerName string `json:"serverName"`
		Account    struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"account"`
	} `json:"dcs"`
}

func main() {
	arg := os.Args
	token := ""
	botChannel := ""
	serverStatusMessageID := ""
	username := ""
	password := ""
	serverName := ""
	createMessage := false
	verbose := false

	fileBytes, err := ioutil.ReadFile("./config.json")

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	if err == nil {
		config := configJson{}
		err := json.Unmarshal(fileBytes, &config)
		if err != nil {
			log.Fatal(err)
		}

		token = config.Discord.Token
		botChannel = config.Discord.Channel
		serverStatusMessageID = config.Discord.ServerStatusMessageID
		username = config.Dcs.Account.Username
		password = config.Dcs.Account.Password
		serverName = config.Dcs.ServerName
	}

	for _, ele := range arg {
		if strings.Index(ele, "--token ") == 0 {
			token = ele[8:]
		} else if strings.Index(ele, "--channel ") == 0 {
			botChannel = ele[10:]
		} else if strings.Index(ele, "--serverStatusMessageId ") == 0 {
			serverStatusMessageID = ele[24:]
		} else if strings.Index(ele, "--username ") == 0 {
			username = ele[11:]
		} else if strings.Index(ele, "--password ") == 0 {
			password = ele[11:]
		} else if strings.Index(ele, "--serverName ") == 0 {
			serverName = ele[13:]
		} else if strings.Index(ele, "--createMessage") == 0 {
			createMessage = true
		} else if strings.Index(ele, "--verbose") == 0 {
			verbose = true
		}
	}

	if createMessage {
		if token == "" || botChannel == "" {
			log.Fatal("Missing parameter")
		}

		log.Print("Create bot messages")
		msgIDs, err := m.CreateMessage(token, botChannel)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Message IDs created:")
		for _, msgID := range msgIDs {
			fmt.Println(msgID)
		}
	} else {
		if token == "" || botChannel == "" || serverStatusMessageID == "" || username == "" || password == "" || serverName == "" {
			log.Fatal("Missing parameter")
		}

		err := m.RunBot(token, botChannel, serverStatusMessageID, username, password, serverName, verbose)
		if err != nil {
			log.Fatal(err)
		}
	}
}
