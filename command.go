package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

var (
	NoPermissionEmbed = NewEmbed().
		SetTitle("No Permission").
		SetFooter("Bot by LockedThread#5691").
		SetDescription("You don't have permission to do this command.").
		SetColor(Red)
)

type Command struct {
	Aliases []string
	Execute func(data CommandData)
}

func (c Command) execute(data CommandData) {
	c.Execute(data)
}

type CommandData struct {
	Label     string
	GuildID   string
	Arguments []string
	Message   *discordgo.Message
	Session   *discordgo.Session
	User      *discordgo.User
	Channel   *discordgo.Channel
}

func (cd CommandData) String() string {
	return fmt.Sprintf("[%s, %s, %s, %s, %s, %s]", cd.Label, cd.GuildID, cd.Arguments, cd.Session, cd.User.String(), cd.Channel)
}

func (cd CommandData) SendMessage(message string, a ...interface{}) *discordgo.Message {
	message = fmt.Sprintf(message, a...)
	if len(message) <= 2000 {
		send, err := cd.Session.ChannelMessageSend(cd.Channel.ID, message)
		CheckErr(err)
		return send
	} else {
		var messageInstance *discordgo.Message

		strings := SplitSubN(message, 2000)
		for messageIndex := range strings {
			send, err := cd.Session.ChannelMessageSend(cd.Channel.ID, strings[messageIndex])
			CheckErr(err)
			messageInstance = send
		}

		return messageInstance
	}
}

func (cd CommandData) SendEmbed(embed *Embed) *discordgo.Message {
	message, err := cd.Session.ChannelMessageSendEmbed(cd.Channel.ID, embed.MessageEmbed)
	CheckErr(err)
	return message
}

func (cd CommandData) SendNoPermission() {
	cd.SendEmbed(NoPermissionEmbed)
}
