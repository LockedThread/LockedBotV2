package main

import (
	"bytes"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func SplitSubN(s string, n int) []string {
	sub := ""
	var subs []string

	runes := bytes.Runes([]byte(s))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%n == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}
	return subs
}

func getGuild(session *discordgo.Session, guildID string) *discordgo.Guild {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		guild, err = session.Guild(guildID)
		if err != nil {
			return nil
		}
	}
	return guild
}

func isOwner(user *discordgo.User) bool {
	return user.ID == OWNER
}

func hasRole(member *discordgo.Member, roleString string) bool {
	for roleIndex := range member.Roles {
		role := member.Roles[roleIndex]
		if strings.ToLower(role) == roleString {
			return true
		}
	}
	return false
}

func getRole(guild *discordgo.Guild, roleString string) *discordgo.Role {
	for roleIndex := range guild.Roles {
		role := guild.Roles[roleIndex]
		if strings.ToLower(role.Name) == strings.ToLower(roleString) {
			return role
		}
	}
	return nil
}
