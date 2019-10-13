package main

import "github.com/bwmarrin/discordgo"

type Command struct {
	Aliases []string
	Execute func(data CommandData)
}

func (c Command) execute(data CommandData) {
	c.Execute(data)
}

type CommandData struct {
	Label     string
	Arguments []string
	Session   *discordgo.Session
	User      *discordgo.User
	Channel   *discordgo.Channel
}

func (data CommandData) sendMessage(message string) *discordgo.Message {
	send, err := data.Session.ChannelMessageSend(data.Channel.ID, message)
	checkErr(err)
	return send
}
