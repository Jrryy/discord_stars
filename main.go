package main

import (
	"fmt"
	dgo "github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	s "strings"
	"syscall"
)

func sendHelp(session *dgo.Session, channel string) error {
	helpString := ";h[elp]: Display this message.\n"
	_, e := session.ChannelMessageSend(channel, helpString)
	return e
}

func messageHandler(session *dgo.Session, m *dgo.MessageCreate) {
	if m.Author.ID == session.State.User.ID {
		return
	}
	if s.HasPrefix(m.Content, "!help") || s.HasPrefix(m.Content, "!h") {
		e := sendHelp(session, m.ChannelID)
		if e != nil {
			log.Print(e)
		}
	}
	log.Printf("The command %s was invalid ", m.Content)
}

func main() {
	session, e := dgo.New("Bot " + "MjMwMzU2MjE1OTI0OTE2MjI2.XmQbaw.WEhgE3ScSw1yTaRmHamICRbZO6E")
	if e != nil {
		fmt.Println("An error occurred: ", e)
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
