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
	User      *discordgo.User
	Arguments []string
	Channel   *discordgo.Channel
}
