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

type serverConfigJSON struct {
	ServerName       string `json:"serverName"`
	DiscordChannelID string `json:"discordChannelId"`
	DiscordMessageID string `json:"discordMessageId"`
	ServerStatusFile string `json:"serverStatusFile"`
	ThumbnailUrl     string `json:"thumbnailUrl"`
}

type configJSON struct {
	Discord struct {
		Token   string `json:"token"`
		Account struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"account"`
	} `json:"discord"`
	Server []serverConfigJSON
}

func main() {
	arg := os.Args
	createMessage := false
	verbose := false

	for _, ele := range arg {
		if strings.Index(ele, "--createMessage") == 0 {
			createMessage = true
		} else if strings.Index(ele, "--verbose") == 0 {
			verbose = true
		}
	}

	fileBytes, err := ioutil.ReadFile("./config.json")

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	config := configJSON{}
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	if createMessage {
		log.Print("Create bot messages")
		for _, server := range config.Server {
			msgIDs, err := m.CreateMessage(config.Discord.Token, server.DiscordChannelID)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Message IDs created:")
			for _, msgID := range msgIDs {
				fmt.Println(msgID)
			}
		}
	} else {
		server := []m.DCSServer{}
		for _, serverConfig := range config.Server {
			server = append(server, m.DCSServer{
				ServerName:           serverConfig.ServerName,
				ServerStatusFilePath: serverConfig.ServerStatusFile,
				DiscordChannelId:     serverConfig.DiscordChannelID,
				DiscordMessageId:     serverConfig.DiscordMessageID,
				ThumbnailURL:         serverConfig.ThumbnailUrl,
			})
		}
		err := m.RunBot(
			config.Discord.Token,
			config.Discord.Account.Username,
			config.Discord.Account.Password,
			server,
			verbose,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}
