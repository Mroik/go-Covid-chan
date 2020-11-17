package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var game string
var adminID string
var musicBuffer = make([][]byte, 0)
var musicFile string //Music file has to be in .dca format

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
		cases, deaths, recovers, err := getCountry(m.Content[13:])
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Country not found")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "```Cases: "+cases+"\nDeaths: "+deaths+"\nRecovers: "+recovers+"```")
	} else if strings.HasPrefix(m.Content, "!covid reminder") {
		//Give as channelID the channel in which the user is in
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "There was an internal problem")
			return
		}
		for _, vs := range guild.VoiceStates {
			if vs.UserID == m.Author.ID {
				err := playMusicBuffer(s, m.GuildID, vs.ChannelID)
				if err != nil {
					fmt.Println(err)
					s.ChannelMessageSend(m.ChannelID, "There was an internal problem")
					return
				}
				return
			}
		}
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "There was an internal problem")
		}
	}
}

func loadMusicBuffer() error {
	file, err := os.Open(musicFile)
	if err != nil {
		return err
	}
	var fileLen uint16
	for {
		err = binary.Read(file, binary.LittleEndian, &fileLen)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}
		inBuff := make([]byte, fileLen)
		err = binary.Read(file, binary.LittleEndian, &inBuff)
		if err != nil {
			return err
		}
		musicBuffer = append(musicBuffer, inBuff)
	}
}

func playMusicBuffer(s *discordgo.Session, guildID string, channelID string) error {
	//TODO on connect
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)
	for _, buffer := range musicBuffer {
		vc.OpusSend <- buffer
	}
	vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)
	vc.Disconnect()
	return nil
}

func main() {
	var token string
	fmt.Scan(&token)
	fmt.Scan(&adminID)
	fmt.Scan(&game)
	fmt.Scan(&musicFile)

	err := loadMusicBuffer()
	if err != nil {
		fmt.Println(err)
	}

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
