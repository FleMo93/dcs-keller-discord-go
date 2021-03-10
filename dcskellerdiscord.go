package dcskellerdiscordgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	serverstatus "github.com/FleMo93/dcs-kellergeschwader-serverstatus-go"
	"github.com/bwmarrin/discordgo"
)

func verboseMsg(msg string, verbose bool) {
	if verbose {
		fmt.Println(msg)
	}
}

func readServerStatusFile(filePath string) (serverstatus.DCSServerStatus, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	status := serverstatus.DCSServerStatus{}
	if err != nil {
		return status, err
	}

	err = json.Unmarshal(fileBytes, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

func secondsToTimeString(time int) string {
	hours := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60/60))
	minutes := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60%60))
	seconds := fmt.Sprintf("%02s", strconv.Itoa(int(time)%60))
	return hours + ":" + minutes + ":" + seconds
}

func getPlayerListString(serverStatus serverstatus.DCSServerStatus) string {
	// plane - player group
	players := make(map[string][]serverstatus.DCSServerStatusPlayer)

	for _, player := range serverStatus.Players {
		players[player.Role] = append(players[player.Role], player)
	}

	var planeKeys []string
	for k := range players {
		planeKeys = append(planeKeys, k)
	}
	sort.Strings(planeKeys)

	listString := ""
	for _, planeName := range planeKeys {
		listString += "**" + planeName + "**\n"

		sort.Slice(players[planeName], func(i int, j int) bool {
			return players[planeName][i].Name < players[planeName][j].Name
		})

		for _, player := range players[planeName] {
			listString += "‏‏‎ ‎‏‏‎ ‎**`" + player.Name + "`** _" + secondsToTimeString(int(player.OnlineTime)) + " h_\n"
		}
		listString += "\n"
	}

	return listString
}

var serverColorOffline int = 11878449 //b54031
var serverColorOnline int = 3388721   //33b531
func updateServerStatusMessage(session *discordgo.Session, botChannel string, serverStatusMessageID string, serverOnline bool, statusFile serverstatus.DCSServerStatus, dcsServer *serverstatus.DCSServer, verbose bool) error {
	verboseMsg("Update server status message", verbose)
	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/F16_drawing.svg/320px-F16_drawing.svg.png",
	}

	if serverOnline == true {
		embedMessage.Color = serverColorOnline
		embedMessage.Description += "**Online**\n"
		embedMessage.Description += "IP address: **" + dcsServer.IPADDRESS + ":" + dcsServer.PORT + "**\n"
		embedMessage.Description += "Mission: **" + dcsServer.MISSIONNAME + "**\n"

		embedMessage.Description += "Next mission: **" + secondsToTimeString(statusFile.MissionTimeLeft) + " h**\n"
		playerList := getPlayerListString(statusFile)
		embedMessage.Description += "Players online: **" + strconv.Itoa(len(statusFile.Players)) + "**"
		embedMessage.Description += "\n\n" + playerList
	} else {
		embedMessage.Color = serverColorOffline
		embedMessage.Description += "**Offline**\n"
	}

	embedMessage.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update:",
	}

	message, err := session.ChannelMessageEditEmbed(botChannel, serverStatusMessageID, &embedMessage)
	if err != nil {
		return err
	}

	verboseMsg("Edited message", verbose)

	if message.Content != "" {
		verboseMsg("Clear message content", verbose)
		_, err := session.ChannelMessageEdit(message.ChannelID, message.ID, "")
		if err != nil {
			return err
		}
		verboseMsg("Message content cleared", verbose)
	}

	verboseMsg("Server status update finished", verbose)
	return nil
}

var weatherColorSunny int = 9492192          //90d6e0
var weatherColorPartialCloudy int = 11519689 //afc6c9
var weatherColorCloudy int = 9211020         //8c8c8c

var weatherIconSunny = ":sunny:"
var weatherIconPartialCloudy = ":partly_sunny:"
var weatherIconPartialCloudyRainy = ":white_sun_rain_cloud:"
var weatherIconCloudy = ":cloud:"
var weatherIconCloudyRainy = ":cloud_rain:"
var weatherIconStormy = ":thunder_cloud_rain:"

func updateServerWeatherMessage(session *discordgo.Session, botChannel string, serverWeatherMessageID string, serverOnline bool, statusFile serverstatus.DCSServerStatus, verbose bool) error {
	verboseMsg("Update server weather message", verbose)
	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Current Weather"

	if serverOnline {
		cloudStatus := ""
		cloudStatusIcon := ""

		if statusFile.Weather.Clouds.Density <= 2 {
			cloudStatus = "Sunny"
			cloudStatusIcon = weatherIconSunny
			embedMessage.Color = weatherColorSunny
		} else if statusFile.Weather.Clouds.Density <= 5 {
			cloudStatus = "Partial overcast"
			cloudStatusIcon = weatherIconPartialCloudy
			embedMessage.Color = weatherColorPartialCloudy
		} else {
			cloudStatus = "Overcast"
			cloudStatusIcon = weatherIconCloudy
			embedMessage.Color = weatherColorCloudy
		}

		if statusFile.Weather.Clouds.Iprecptns == 1 {
			cloudStatus += " rainy"

			if statusFile.Weather.Clouds.Density <= 8 {
				cloudStatusIcon = weatherIconPartialCloudyRainy
			} else {
				cloudStatusIcon = weatherIconCloudyRainy
			}
		} else if statusFile.Weather.Clouds.Iprecptns == 2 {
			cloudStatus += " stormy"
			cloudStatusIcon = weatherIconStormy
		}

		embedMessage.Description += strconv.Itoa(statusFile.Weather.Season.Temperature) + "°C "
		embedMessage.Description += cloudStatusIcon + " "
		embedMessage.Description += cloudStatus + " at " + strconv.Itoa(statusFile.Weather.Clouds.Base) + " ft "
		embedMessage.Description += "for " + strconv.Itoa(statusFile.Weather.Clouds.Thickness) + " ft\n"
		embedMessage.Description += "\n"

		embedMessage.Description += "**Wind**\n"
		embedMessage.Description += " ‎‏‏‎ ‎At ground: ‎‏‏‎ ‎ ‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.AtGround.Dir)) + " - " + fmt.Sprintf("%02s", strconv.Itoa(statusFile.Weather.Wind.AtGround.Speed)) + " kn\n"
		embedMessage.Description += " ‎‏‏‎ ‎At 6,500 ft: ‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.At2000.Dir)) + " - " + fmt.Sprintf("%02s", strconv.Itoa(statusFile.Weather.Wind.At2000.Speed)) + " kn\n"
		embedMessage.Description += " ‎‏‏‎ ‎At 26,00 ft: ‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.At8000.Dir)) + " - " + fmt.Sprintf("%02s", strconv.Itoa(statusFile.Weather.Wind.At8000.Speed)) + " kn\n"
		embedMessage.Description += "\n"

		embedMessage.Description += "**Time**\n"
		embedMessage.Description += " ‎‏‏‎  ‎‏‏" + secondsToTimeString(statusFile.Time)
	} else {
		embedMessage.Color = serverColorOffline
		embedMessage.Description = "Server offline"
	}

	embedMessage.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update:",
	}

	message, err := session.ChannelMessageEditEmbed(botChannel, serverWeatherMessageID, &embedMessage)
	if err != nil {
		return err
	}
	verboseMsg("Edited message", verbose)

	if message.Content != "" {
		verboseMsg("Clear message content", verbose)
		_, err := session.ChannelMessageEdit(message.ChannelID, message.ID, "")
		if err != nil {
			return err
		}
		verboseMsg("Message content cleared", verbose)
	}

	verboseMsg("Server weather update finished", verbose)
	return nil
}

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, botChannel string, serverStatusMessageID string, weatherStatusMessageID string, username string, password string, serverName string, serverStatusFile string, verbose bool) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	verboseMsg("Bot created", verbose)

	serverStatus, err := serverstatus.GetServerStatus(username, password, serverName)
	serverOnline := true

	if err != nil {
		if err.Error() == "Server not found" {
			serverOnline = false
		} else {
			return err
		}
	}
	verboseMsg("Returned server status", verbose)

	status, err := readServerStatusFile(serverStatusFile)
	if err != nil {
		return err
	}

	err = updateServerStatusMessage(session, botChannel, serverStatusMessageID, serverOnline, status, &serverStatus, verbose)
	if err != nil {
		return err
	}

	err = updateServerWeatherMessage(session, botChannel, weatherStatusMessageID, serverOnline, status, verbose)
	if err != nil {
		return err
	}

	return nil
}

// CreateMessage Creates the messages needed for the bot status
func CreateMessage(token string, botChanngel string) ([]string, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return []string{}, err
	}

	msg, err := session.ChannelMessageSend(botChanngel, "Initial message")
	if err != nil {
		return []string{}, err
	}

	return []string{msg.ID}, nil
}
