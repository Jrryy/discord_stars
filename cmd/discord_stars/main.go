package main

import (
	"encoding/json"
	"flag"
	"fmt"
	dgo "github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

var discordToken string
var brawlStarsAPIToken string
var client *http.Client

func addApiHeaders(request *http.Request) {
	request.Header.Add("Accept", "application/json")
	request.Header.Add("authorization", "Bearer "+brawlStarsAPIToken)
}

func sendHelp(session *dgo.Session, channel string) error {
	helpString := "```\n" +
		"Usage of this bot (all commands are preceded by \";\"):\n" +
		"\t- ;h[elp]: Display this message.\n" +
		"\t- ;info <player tag>: Display data about a player with tag <player tag>. If no tag is provided, or there isn't any player with it, no data will be shown." +
		"```"
	_, e := session.ChannelMessageSend(channel, helpString)
	return e
}

func registerPlayer(session *dgo.Session, channel string) error {
	return nil
}

func showPlayerData(session *dgo.Session, channel string, player string) error {
	playerUrl := url.PathEscape(player)
	request, e := http.NewRequest("GET", "https://api.brawlstars.com/v1/players/"+playerUrl, nil)
	if e != nil {
		return e
	}
	addApiHeaders(request)
	response, e := client.Do(request)
	if e != nil {
		return e
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("player with id %s not found", player)
	}
	defer response.Body.Close()
	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return e
	}
	var playerDataDict map[string]interface{}
	e = json.Unmarshal(body, &playerDataDict)
	playerDataString := fmt.Sprintf(
		"```\n"+
			"Player name: %v\n"+
			"Trophies: %v\n"+
			"Victories: %v\n"+
			"```",
		playerDataDict["name"].(string),
		playerDataDict["trophies"].(float64),
		playerDataDict["3vs3Victories"].(float64))

	_, e = session.ChannelMessageSend(channel, playerDataString)
	return e
}

func messageHandler(session *dgo.Session, m *dgo.MessageCreate) {
	if m.Author.ID == session.State.User.ID {
		return
	}
	re, _ := regexp.Compile(";(\\w+).*")
	var e error
	matchedCommand := re.FindStringSubmatch(m.Content)
	if len(matchedCommand) > 1 {
		switch matchedCommand[1] {
		case "h", "help":
			e = sendHelp(session, m.ChannelID)
		case "r", "register":
			e = registerPlayer(session, m.ChannelID)
		case "info":
			matchedLastIndex := re.FindStringSubmatchIndex(m.Content)[3]
			playerId := strings.TrimSpace(m.Content[matchedLastIndex:])
			e = showPlayerData(session, m.ChannelID, playerId)
		default:
			log.Printf("The command %s was invalid ", m.Content)
		}
		if e != nil {
			log.Print(e)
		}
	}
}

func getTokens() (e error) {
	tokenFlag := flag.String("token", "", "Token found in https://discordapp.com/developers/applications/<bot_id>/bot")
	apiToken := flag.String("apiToken", "", "Token obtained in the Brawl Stars developers portal")
	flag.Parse()
	tokenEnv, foundToken := syscall.Getenv("DISCORD_STARS_TOKEN")
	apiTokenEnv, foundApiToken := syscall.Getenv("BRAWL_STARS_API_TOKEN")
	if !foundToken {
		if *tokenFlag == "" {
			e = fmt.Errorf("discord token was not found")
			flag.PrintDefaults()
		} else {
			discordToken = *tokenFlag
		}
	} else {
		discordToken = tokenEnv
	}
	if !foundApiToken {
		if *apiToken == "" {
			e = fmt.Errorf("brawl stars api token not found")
		} else {
			brawlStarsAPIToken = *apiToken
		}
	} else {
		brawlStarsAPIToken = apiTokenEnv
	}
	return
}

func testApi() (e error) {
	request, e := http.NewRequest("GET", "https://api.brawlstars.com/v1/players/%239UG88U0RJ", nil)
	if e != nil {
		return
	}
	addApiHeaders(request)
	response, e := client.Do(request)
	if e != nil {
		return
	}
	if response.StatusCode != 200 {
		e = fmt.Errorf("the api returned status code %v", response.StatusCode)
	}
	return
}

func main() {
	e := getTokens()
	if e != nil {
		fmt.Println("An error occurred when obtaining the tokens: ", e)
		return
	}
	session, e := dgo.New("Bot " + discordToken)
	if e != nil {
		fmt.Println("An error occurred when opening a connection to Discord: ", e)
		return
	}
	client = &http.Client{}
	e = testApi()
	if e != nil {
		fmt.Println("An error occurred when testing the API: ", e)
		return
	}
	// Register the messageCreate func as a callback for MessageCreate events.
	session.AddHandler(messageHandler)

	// Open a websocket connection to Discord and begin listening.
	e = session.Open()
	if e != nil {
		fmt.Println("error opening connection, ", e)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	e = session.Close()
	if e != nil {
		fmt.Println(e)
	}
}
