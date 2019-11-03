package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

type Ticket struct {
	Author      *discordgo.User
	TimeCreated *time.Time
	TextChannel *discordgo.Channel
}

type User struct {
	ID          int
	Token       string
	DiscordID   string
	Resources   []string
	IPAddresses []string
}

func (user User) String() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s ", user.ID, user.Token, user.DiscordID, strings.Join(user.Resources, ","), strings.Join(user.IPAddresses, ","))
}

type GetUserError struct {
	message string
}

func (err GetUserError) Error() string {
	return fmt.Sprintf("%s", err.message)
}
