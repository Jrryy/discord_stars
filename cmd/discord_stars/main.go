package main

import (
	"flag"
	"fmt"
	dgo "github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

func sendHelp(session *dgo.Session, channel string) error {
	helpString := "```\n" +
		"Usage of this bot (all commands are preceded by \";\"):\n" +
		"\t- ;h[elp]: Display this message.\n" +
		"```"
	_, e := session.ChannelMessageSend(channel, helpString)
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
		default:
			log.Printf("The command %s was invalid ", m.Content)
		}
		if e != nil {
			log.Print(e)
		}
	}
}

func getToken() (token string, e error) {
	token = ""
	e = nil
	tokenFlag := flag.String("token", "", "Token found in https://discordapp.com/developers/applications/<bot_id>/bot")
	flag.Parse()
	tokenEnv, found := syscall.Getenv("DISCORD_STARS_TOKEN")
	if !found {
		if *tokenFlag == "" {
			e = fmt.Errorf("token was not found")
			flag.PrintDefaults()
		} else {
			token = *tokenFlag
		}
		return
	} else {
		token = tokenEnv
	}
	return
}

func main() {
	token, e := getToken()
	if e != nil {
		fmt.Println("An error occurred: ", e)
		return
	}
	session, e := dgo.New("Bot " + token)
	if e != nil {
		fmt.Println("An error occurred: ", e)
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
