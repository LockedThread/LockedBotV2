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
	return fmt.Sprintf("%d, %s, %s, %s, %s", user.ID, user.Token, user.DiscordID, strings.Join(user.Resources, ","), strings.Join(user.IPAddresses, ","))
}

type GetUserError struct {
	message string
}

func (err GetUserError) Error() string {
	return fmt.Sprintf("%s", err.message)
}

type Resource struct {
	ID               int
	Name             string
	ResponseData     string
	DiscordChannelID string
}

func (response Resource) String() string {
	return fmt.Sprintf("%d, %s, %s, %s", response.ID, response.Name, response.ResponseData, response.DiscordChannelID)
}

type GetResourceError struct {
	message string
}

func (err GetResourceError) Error() string {
	return fmt.Sprintf("%s", err.message)
}
