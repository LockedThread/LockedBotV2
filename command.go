package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
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

func (data CommandData) toString() string {
	return fmt.Sprintf("[%s, %s, %s, %s, %s, %s]", data.Label, data.GuildID, data.Arguments, data.Session, data.User.String(), data.Channel)
}

func (data CommandData) sendMessage(message string, a ...interface{}) *discordgo.Message {
	message = fmt.Sprintf(message, a)
	if len(message) <= 2000 {
		send, err := data.Session.ChannelMessageSend(data.Channel.ID, message)
		checkErr(err)
		return send
	} else {
		var messageInstance *discordgo.Message

		strings := SplitSubN(message, 2000)
		for messageIndex := range strings {
			send, err := data.Session.ChannelMessageSend(data.Channel.ID, strings[messageIndex])
			checkErr(err)
			messageInstance = send
		}

		return messageInstance
	}
}
