package main

import (
	m "dcskellerdiscordgo"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	arg := os.Args
	token := ""
	botChannel := ""
	serverStatusMessageID := ""
	username := ""
	password := ""
	serverName := ""
	createMessage := false

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

		err := m.RunBot(token, botChannel, serverStatusMessageID, username, password, serverName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
