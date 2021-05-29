package dcskellerdiscordgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"time"

	serverstatus "github.com/FleMo93/dcs-kellergeschwader-serverstatus-go"
	"github.com/bwmarrin/discordgo"
)

type DCSServer struct {
	ServerName           string
	ServerStatusFilePath string
	DiscordChannelId     string
	DiscordMessageId     string
	ThumbnailURL         string
	DcsInstallDir        string
}

type autoupdateCfg struct {
	Branch  string `json:"branch"`
	Version string `json:"version"`
}

func verboseMsg(msg string, verbose bool) {
	if verbose {
		fmt.Println(msg)
	}
}

func secondsToTimeString(time int) string {
	hours := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60/60))
	minutes := fmt.Sprintf("%02s", strconv.Itoa(int(time)/60%60))
	seconds := fmt.Sprintf("%02s", strconv.Itoa(int(time)%60))
	return hours + ":" + minutes + ":" + seconds
}

func getPlayerListString(serverStatus *serverstatus.DCSServerStatus) string {
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

func readAutoupdateCfg(dcsDir string) (autoupdateCfg, error) {
	fileBytes, err := ioutil.ReadFile(path.Join(dcsDir, "autoupdate.cfg"))

	if err != nil {
		return autoupdateCfg{}, err
	}

	config := autoupdateCfg{}
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return autoupdateCfg{}, err
	}

	return config, nil
}

var serverColorOffline int = 11878449 //b54031
var serverColorOnline int = 3388721   //33b531

var weatherIconSunny = ":sunny:"
var weatherIconPartialCloudy = ":partly_sunny:"
var weatherIconPartialCloudyRainy = ":white_sun_rain_cloud:"
var weatherIconCloudy = ":cloud:"
var weatherIconCloudyRainy = ":cloud_rain:"
var weatherIconStormy = ":thunder_cloud_rain:"

func setServerStatusDescription(embedMessage *discordgo.MessageEmbed, statusFile *serverstatus.DCSServerStatus, dcsServerWebInfo *serverstatus.DCSServer, autoupdateCfg *autoupdateCfg, verbose bool) {
	playerList := getPlayerListString(statusFile)

	embedMessage.Description += "Name: **" + dcsServerWebInfo.NAME + "**\n"
	embedMessage.Description += "IP address: **" + dcsServerWebInfo.IPADDRESS + ":" + dcsServerWebInfo.PORT + "**\n"
	embedMessage.Description += "Version: **" + autoupdateCfg.Version + " - " + autoupdateCfg.Branch + "**\n"
	embedMessage.Description += "Mission: **" + dcsServerWebInfo.MISSIONNAME + "**\n"
	embedMessage.Description += "Next mission: **" + secondsToTimeString(statusFile.MissionTimeLeft) + " h**\n"
	embedMessage.Description += "Players online: **" + strconv.Itoa(len(statusFile.Players)) + "**\n"
	embedMessage.Description += "\n" + playerList
	embedMessage.Description += "\n"
	embedMessage.Description += "**Current Weather**\n"
	cloudStatus := ""
	cloudStatusIcon := ""

	if statusFile.Weather.Clouds.Density <= 2 {
		cloudStatus = "Sunny"
		cloudStatusIcon = weatherIconSunny
	} else if statusFile.Weather.Clouds.Density <= 5 {
		cloudStatus = "Partial overcast"
		cloudStatusIcon = weatherIconPartialCloudy
	} else {
		cloudStatus = "Overcast"
		cloudStatusIcon = weatherIconCloudy
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
	embedMessage.Description += "\n\n"

	speedAtGround := fmt.Sprintf("%02s", strconv.FormatFloat(statusFile.Weather.Wind.AtGround.Speed, 'f', -1, 64))
	speedAt2000 := fmt.Sprintf("%02s", strconv.FormatFloat(statusFile.Weather.Wind.At2000.Speed, 'f', -1, 64))
	speedAt8000 := fmt.Sprintf("%02s", strconv.FormatFloat(statusFile.Weather.Wind.At8000.Speed, 'f', -1, 64))

	embedMessage.Description += "**Wind**\n"
	embedMessage.Description += " ‎‏‏‎ ‎At ground: ‎‏‏‎ ‎ ‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.AtGround.Dir)) + " - " + speedAtGround + " kn\n"
	embedMessage.Description += " ‎‏‏‎ ‎At 6,500 ft: ‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.At2000.Dir)) + " - " + speedAt2000 + " kn\n"
	embedMessage.Description += " ‎‏‏‎ ‎At 26,000 ft:‎‏‏‎ ‎hdg " + fmt.Sprintf("%03s", strconv.Itoa(statusFile.Weather.Wind.At8000.Dir)) + " - " + speedAt8000 + " kn\n"
	embedMessage.Description += "\n"

	embedMessage.Description += "**Time**\n"
	embedMessage.Description += " ‎‏‏‎  ‎‏‏" + secondsToTimeString(statusFile.Time)
}

func updateServerStatusMessage(session *discordgo.Session, discordChannelId string, discordMessageId string, serverOnline bool, serverOptions *DCSServer, serverStatusFileInfo *serverstatus.DCSServerStatus, dcsServerWebInfo *serverstatus.DCSServer, autoupdateCfg *autoupdateCfg, verbose bool) error {
	verboseMsg("Update server status message", verbose)
	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: serverOptions.ThumbnailURL,
	}

	if serverOnline == true {
		embedMessage.Color = serverColorOnline
		setServerStatusDescription(&embedMessage, serverStatusFileInfo, dcsServerWebInfo, autoupdateCfg, verbose)
	} else {
		embedMessage.Color = serverColorOffline
		embedMessage.Description += "**Offline**\n"
	}

	embedMessage.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	embedMessage.Footer = &discordgo.MessageEmbedFooter{
		Text: "Last update:",
	}

	message, err := session.ChannelMessageEditEmbed(discordChannelId, discordMessageId, &embedMessage)
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

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, username string, password string, dcsServer []DCSServer, verbose bool) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	verboseMsg("Bot created", verbose)

	serverNames := []string{}
	for _, server := range dcsServer {
		serverNames = append(serverNames, server.ServerName)
	}

	serverStatus, err := serverstatus.GetServerStatus(username, password, serverNames)

	for index, server := range dcsServer {
		serverOnline := true

		if err != nil {
			if err.Error() == "Server not found" {
				serverOnline = false
			} else {
				return err
			}
		}

		verboseMsg("Returned server status", verbose)

		status, err := serverstatus.ReadServerStatusFile(server.ServerStatusFilePath)
		if err != nil {
			return err
		}

		autoupdateCfg, err := readAutoupdateCfg(server.DcsInstallDir)
		if err != nil {
			return err
		}

		err = updateServerStatusMessage(session, server.DiscordChannelId, server.DiscordMessageId, serverOnline, &server, &status, &serverStatus[index], &autoupdateCfg, verbose)
		if err != nil {
			return err
		}
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
