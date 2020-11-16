package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var game string
var adminID string

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	err := s.UpdateStreamingStatus(0, "with Ebola-chan", game)
	if err != nil {
		fmt.Println(err)
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	fmt.Println(m.Author.Username + " " + m.Content)

	if strings.HasPrefix(m.Content, "!shutdown") && m.Author.ID == adminID {
		s.Close()
	} else if strings.HasPrefix(m.Content, "!covid top") {
		res, err := getTop()
		if err != nil {
			fmt.Println(err)
			return
		}
		for x := 0; x < 5; x++ {
			s.ChannelMessageSend(m.ChannelID, res[x])
		}
	} else if strings.HasPrefix(m.Content, "!covid stats") && len(strings.Split(m.Content, " ")) > 2 {
		cases, deaths, recovers, err := getCountry(strings.Split(m.Content, " ")[2])
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Country not found")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "```Cases: "+cases+"\nDeaths: "+deaths+"\nRecovers: "+recovers+"```")
	}
}

func main() {
	var token string
	fmt.Scan(&token)
	fmt.Scan(&adminID)
	fmt.Scan(&game)

	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.AddHandler(onReady)
	bot.AddHandler(onMessageCreate)

	err = bot.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Covid-chan is now running...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
