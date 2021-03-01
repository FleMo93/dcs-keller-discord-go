package dcskellerdiscordgo

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, botChannel string, serverStatusMessageID string) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	colorOnline := 3388721   //33b531
	colorOffline := 11878449 //b54031
	serverOnline := true
	playerOnline := 0

	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/F16_drawing.svg/320px-F16_drawing.svg.png",
	}
	embedMessage.Timestamp = time.Now().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update",
	}

	if serverOnline == true {
		embedMessage.Description += "Online\n"
		embedMessage.Color = colorOnline
	} else {
		embedMessage.Description += "Offline\n"
		embedMessage.Color = colorOffline
	}

	embedMessage.Description += strconv.Itoa(playerOnline) + " players online"

	session.ChannelMessageEditEmbed(botChannel, serverStatusMessageID, &embedMessage)
	return nil
}
