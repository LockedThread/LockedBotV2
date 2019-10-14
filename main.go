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

const (
	OWNER = "545743465267593216"
)

var (
	Token      string
	CommandMap map[string]*Command
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	CommandMap = make(map[string]*Command)
	discord, err := discordgo.New("Bot " + Token)
	checkErr(err)
	err = discord.Open()
	checkErr(err)

	err = discord.UpdateListeningStatus("LockedThread's Front Door")
	checkErr(err)

	discord.AddHandler(messageCreate)

	command := Command{
		Aliases: []string{"-setuser"},
		Execute: func(data CommandData) {
			fmt.Println("You executed " + data.Label)
			fmt.Println(data.toString())

			if isOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.sendMessage("-setuser [@mention] [role]")
					break
				case 2:
					//guild := getGuild(data.Session, data.GuildID)
					mentions := data.Message.Mentions

					guildMember, err := data.Session.GuildMember(data.GuildID, mentions[0].ID)
					checkErr(err)
					hasRole := hasRole(guildMember, data.Arguments[1])
					guild := getGuild(data.Session, data.GuildID)
					if !hasRole {
						role := getRole(guild, data.Arguments[1])
						err := data.Session.GuildMemberRoleAdd(guild.ID, data.User.ID, role.ID)
						checkErr(err)
						data.sendMessage()
					} else {
						data.sendMessage("You have set %s's ")
					}

					break
				}
			} else {

			}

		}}

	registerCommand(command.Aliases, &command)

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
	command := findCommand(splitMessage[0])

	if command != nil {
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
			GuildID:   m.GuildID,
			Message:   m.Message,
			User:      m.Author,
			Arguments: arguments,
			Channel:   channel,
			Session:   s,
		})
	}
}

func registerCommand(aliases []string, command *Command) {
	for aliasIndex := range aliases {
		CommandMap[strings.ToLower(aliases[aliasIndex])] = command
	}
}

func findCommand(label string) *Command {
	return CommandMap[strings.ToLower(label)]
}
