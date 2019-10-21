package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

type Ticket struct {
	Author      *discordgo.User
	TimeCreated *time.Time
	TextChannel *discordgo.Channel
}
