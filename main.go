package main

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var game string
var adminID string
var musicBuffer = make([][]byte, 0)
var musicFile string //Music file has to be in .dca format
var musicInUse map[string]bool = make(map[string]bool)
var database *sql.DB
var user, pass, dbname string

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

	if m.GuildID != "" {
		temp_guild, _ := s.Guild(m.GuildID)
		if !insertGuild(database, m.GuildID, temp_guild.Name) {
			fmt.Println("There was an error with inserting the guild")
		}
		temp_channel, _ := s.Channel(m.ChannelID)
		if !insertChannel(database, m.ChannelID, temp_channel.Name, m.GuildID) {
			fmt.Println("There was an error with inserting the channel")
		}
		if !(strings.TrimSpace(m.Content) == "") {
			if !insertMessage(database, m.Author.ID, m.Content, m.ChannelID) {
				fmt.Println("There was an error with inserting the message")
			}
		}
		for _, x := range m.Attachments {
			if !insertAttachment(database, m.Author.ID, x.URL, m.ChannelID) {
				fmt.Println("There was an error with inserting the attchment")
			}
		}
	}

	if strings.HasPrefix(m.Content, "!shutdown") && m.Author.ID == adminID {
		s.Close()
		database.Close()
	} else if strings.HasPrefix(m.Content, "!covid top") {
		s.ChannelTyping(m.ChannelID)
		res, err := getTop()
		if err != nil {
			fmt.Println(err)
			return
		}
		for x := 0; x < 5; x++ {
			s.ChannelMessageSend(m.ChannelID, res[x])
		}
	} else if strings.HasPrefix(m.Content, "!covid stats") && len(strings.Split(m.Content, " ")) > 2 {
		s.ChannelTyping(m.ChannelID)
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
					//TODO fix this to recognize different type of errors
					s.ChannelMessageSend(m.ChannelID, "Can't use that right now")
				}
				return
			}
		}
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "There was an internal problem")
		}
	} else if strings.HasPrefix(m.Content, "!count_guilds") {
		s.ChannelTyping(m.ChannelID)
		guilds, err := s.UserGuilds(100, "", "")
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, "Internal error")
			return
		}
		s.ChannelMessageSend(m.ChannelID, strconv.Itoa(len(guilds)))
	} else if strings.HasPrefix(m.Content, "!covid") {
		s.ChannelTyping(m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "```!covid stats <country>\n!covid top```")
	} else if strings.HasPrefix(m.Content, "!assignRole") && len(strings.Split(m.Content, " ")) > 3 {
		temp := strings.Split(m.Content, " ")
		err := s.GuildMemberRoleAdd(temp[1], temp[3], temp[2])
		if err != nil {
			fmt.Println(err)
		} else {
			s.ChannelTyping(m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "Role assigned")
		}
	} else if strings.HasPrefix(m.Content, "!removeRole") && len(strings.Split(m.Content, " ")) > 3 {
		temp := strings.Split(m.Content, " ")
		err := s.GuildMemberRoleRemove(temp[1], temp[3], temp[2])
		if err != nil {
			fmt.Println(err)
		} else {
			s.ChannelTyping(m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "Role removed")
		}
	} else if strings.HasPrefix(m.Content, "!showRoles") && len(strings.Split(m.Content, " ")) > 1 {
		roles, err := s.GuildRoles(strings.Split(m.Content, " ")[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Couldn't show roles")
			return
		}
		for _, x := range roles {
			s.ChannelMessageSend(m.ChannelID, x.ID+" "+x.Name)
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
	if musicInUse[guildID] {
		return errors.New("Bot already in use")
	}
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		//Temporary fix
		if _, ok := s.VoiceConnections[guildID]; ok {
			vc = s.VoiceConnections[guildID]
		} else {
			return err
		}
	}
	setMusic(true, guildID)
	defer setMusic(false, guildID)
	time.Sleep(250 * time.Millisecond)
	vc.Speaking(true)
	for _, buffer := range musicBuffer {
		vc.OpusSend <- buffer
	}
	vc.Speaking(false)
	time.Sleep(250 * time.Millisecond)
	err = vc.Disconnect()
	if err != nil {
		musicInUse[guildID] = false
		return err
	}
	return nil
}

func setMusic(x bool, guildID string) {
	musicInUse[guildID] = x
}

func main() {
	var token string
	fmt.Scan(&token)
	fmt.Scan(&adminID)
	fmt.Scan(&game)
	fmt.Scan(&musicFile)
	fmt.Scan(&user)
	fmt.Scan(&pass)
	fmt.Scan(&dbname)

	var err error
	database, err = createDatabase(user, pass, "", dbname)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = loadMusicBuffer()
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
