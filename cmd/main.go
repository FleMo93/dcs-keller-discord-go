package main

import (
	m "dcskellerdiscordgo"
	"log"
	"os"
	"strings"
)

func main() {
	arg := os.Args
	token := ""
	botChannel := ""
	serverStatusMessageID := ""

	for _, ele := range arg {
		if strings.Index(ele, "-t ") == 0 {
			token = ele[3:]
		} else if strings.Index(ele, "-c ") == 0 {
			botChannel = ele[3:]
		} else if strings.Index(ele, "--serverStatusMessageId ") == 0 {
			serverStatusMessageID = ele[24:]
		}
	}
	if token == "" {
		log.Fatal("Missing \"-t \" parameter")
	} else if botChannel == "" {
		log.Fatal("Missing \"-c \" parameter")
	}

	err := m.RunBot(token, botChannel, serverStatusMessageID)
	if err != nil {
		log.Fatal(err)
	}
}
