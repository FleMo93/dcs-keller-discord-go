package dcskellerdiscordgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type dcsServer struct {
	ID          string `json:"ID"`
	NAME        string `json:"NAME"`
	IPADDRESS   string `json:"IP_ADDRESS"`
	PORT        string `json:"PORT"`
	MISSIONNAME string `json:"MISSION_NAME"`
	MISSIONTIME string `json:"MISSION_TIME"`
	PLAYERS     string `json:"PLAYERS"`
	PLAYERSMAX  string `json:"PLAYERS_MAX"`
	PASSWORD    string `json:"PASSWORD"`
	URLTODETAIL string `json:"URL_TO_DETAIL"`
}

type dcsServerList struct {
	SERVERSMAXCOUNT int         `json:"SERVERS_MAX_COUNT"`
	SERVERSMAXDATE  string      `json:"SERVERS_MAX_DATE"`
	PLAYERSCOUNT    int         `json:"PLAYERS_COUNT"`
	MYSERVERS       []dcsServer `json:"MY_SERVERS"`
	SERVERS         []struct {
		NAME                 string `json:"NAME"`
		IPADDRESS            string `json:"IP_ADDRESS"`
		PORT                 string `json:"PORT"`
		MISSIONNAME          string `json:"MISSION_NAME"`
		MISSIONTIME          string `json:"MISSION_TIME"`
		PLAYERS              string `json:"PLAYERS"`
		PLAYERSMAX           string `json:"PLAYERS_MAX"`
		PASSWORD             string `json:"PASSWORD"`
		DESCRIPTION          string `json:"DESCRIPTION"`
		UALIAS0              string `json:"UALIAS_0"`
		MISSIONTIMEFORMATTED string `json:"MISSION_TIME_FORMATTED"`
	} `json:"SERVERS"`
}

type dcsServerStatusPlayer struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type dcsServerStatus struct {
	Players       map[string]dcsServerStatusPlayer `json:"players"`
	MissionsNames []string                         `json:"missionsNames"`
}

func getServerStatus(username string, password string, serverName string) (dcsServer, error) {
	client := &http.Client{}

	url := "https://www.digitalcombatsimulator.com/en/personal/server/?ajax=y&_=" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return dcsServer{}, err
	}

	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return dcsServer{}, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)
	jsonStart := strings.Index(body, "{")
	serverList := body[jsonStart:]
	serverStatus := &dcsServerList{}
	err = json.Unmarshal([]byte(serverList), serverStatus)
	if err != nil || body == "" {
		return dcsServer{}, err
	}

	for _, server := range serverStatus.MYSERVERS {
		if server.NAME == serverName {
			return server, nil
		}
	}
	return dcsServer{}, errors.New("Server not found")
}

func verboseMsg(msg string, verbose bool) {
	if verbose {
		fmt.Println(msg)
	}
}

func readServerStatusFile(filePath string) (dcsServerStatus, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	status := dcsServerStatus{}
	if err != nil {
		return status, err
	}

	err = json.Unmarshal(fileBytes, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}

func getPlayerListString(serverStatus dcsServerStatus) string {
	players := make(map[string][]string)

	for _, player := range serverStatus.Players {
		if players[player.Role] == nil {
			players[player.Role] = []string{}
		}

		players[player.Role] = append(players[player.Role], player.Name)
	}

	var planeKeys []string
	for k := range players {
		planeKeys = append(planeKeys, k)
	}
	sort.Strings(planeKeys)

	listString := ""
	for _, planeName := range planeKeys {
		listString += "**" + planeName + "**\n"

		var playerKeys []string
		for _, playerName := range players[planeName] {
			playerKeys = append(playerKeys, playerName)
		}
		sort.Strings(playerKeys)

		for _, player := range playerKeys {
			listString += "‏‏‎ ‎‏‏‎ ‎" + player + "\n"
		}
		listString += "\n"
	}

	return listString
}

// RunBot starts the dcs kellergeschwader discord bot
func RunBot(token string, botChannel string, serverStatusMessageID string, username string, password string, serverName string, serverStatusFile string, verbose bool) error {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	verboseMsg("Bot created", verbose)

	colorOnline := 3388721   //33b531
	colorOffline := 11878449 //b54031
	serverStatus, err := getServerStatus(username, password, serverName)
	serverOnline := true

	if err != nil {
		if err.Error() == "Server not found" {
			serverOnline = false
		} else {
			return err
		}
	}
	verboseMsg("Returned server status", verbose)

	embedMessage := discordgo.MessageEmbed{}
	embedMessage.Title = "Server Status"
	embedMessage.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/F16_drawing.svg/320px-F16_drawing.svg.png",
	}

	if serverOnline == true {
		playersOnline, err := strconv.Atoi(serverStatus.PLAYERS)
		if err != nil {
			return err
		}
		playersOnline--

		embedMessage.Color = colorOnline
		embedMessage.Description += "**Online**\n"
		embedMessage.Description += "IP address: **" + serverStatus.IPADDRESS + ":" + serverStatus.PORT + "**\n"
		embedMessage.Description += "Mission: **" + serverStatus.MISSIONNAME + "**\n"
		embedMessage.Description += "Players online: **" + strconv.Itoa(playersOnline) + "**"

		if serverStatusFile != "" {
			status, err := readServerStatusFile(serverStatusFile)
			if err != nil {
				return err
			}

			playerList := getPlayerListString(status)
			embedMessage.Description += "\n\n" + playerList
		}
	} else {
		embedMessage.Color = colorOffline
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
