package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	Token    string
	Commands []Command
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	discord, err := discordgo.New("Bot " + Token)
	checkErr(err)
	err = discord.Open()
	checkErr(err)

	err = discord.UpdateListeningStatus("LockedThread's Front Door")
	checkErr(err)

	discord.AddHandler(messageCreate)

	command := Command{
		Aliases: []string{"-shit", "-bitch"},
		Execute: func(data CommandData) {
			fmt.Println("You executed " + data.Label)
			data.sendMessage(data.toString())
		}}

	Commands = append(Commands, command)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	splitMessage := strings.Split(m.Message.Content, " ")
	for commandIndex := range Commands {
		command := Commands[commandIndex]
		for aliasIndex := range command.Aliases {
			alias := command.Aliases[aliasIndex]
			if strings.ToLower(alias) == strings.ToLower(splitMessage[0]) {
				channel, err := s.Channel(m.ChannelID)
				checkErr(err)

				var arguments []string
				if len(splitMessage) >= 2 {
					arguments = splitMessage[1:]
				} else {
					arguments = []string{}
				}

				command.execute(CommandData{
					Label:     splitMessage[0],
					User:      m.Author,
					Arguments: arguments,
					Channel:   channel,
					Session:   s,
				})
			}
		}
	}
}
