package main

import (
	"bytes"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func CheckErr(err error) {
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

func GetGuild(session *discordgo.Session, guildID string) *discordgo.Guild {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		guild, err = session.Guild(guildID)
		if err != nil {
			return nil
		}
	}
	return guild
}

func IsOwner(user *discordgo.User) bool {
	return user.ID == Owner
}

func HasRole(member *discordgo.Member, roleId string) bool {
	println("roleId=", roleId)
	for roleIndex := range member.Roles {
		role := member.Roles[roleIndex]
		println("roleIndex=", roleIndex)
		println("role=", role)
		if role == roleId {
			return true
		}
	}
	return false
}

func GetRole(guild *discordgo.Guild, roleString string) *discordgo.Role {
	for roleIndex := range guild.Roles {
		role := guild.Roles[roleIndex]
		if strings.ToLower(role.Name) == strings.ToLower(roleString) {
			return role
		}
	}
	return nil
}

func RegisterCommand(command *Command) {
	for aliasIndex := range command.Aliases {
		commandMap[strings.ToLower(command.Aliases[aliasIndex])] = command
	}
}

func FindCommand(label string) *Command {
	return commandMap[strings.ToLower(label)]
}

func GetResources(user *discordgo.User) (resources []string) {
	var resourceString string
	err := stmtFindResourceColumn.QueryRow(user.ID).Scan(&resourceString)
	CheckErr(err)
	if len(resourceString) == 0 {
		return resources
	} else {
		err = json.Unmarshal([]byte(resourceString), &resources)
		CheckErr(err)
	}
	return resources
}
