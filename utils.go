package main

import (
	"bytes"
	"encoding/json"
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

func hasRole(member *discordgo.Member, roleId string) bool {
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

func getRole(guild *discordgo.Guild, roleString string) *discordgo.Role {
	for roleIndex := range guild.Roles {
		role := guild.Roles[roleIndex]
		if strings.ToLower(role.Name) == strings.ToLower(roleString) {
			return role
		}
	}
	return nil
}

func getPrettyPrinted(data interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(TEXT_EMPTY, TEXT_INDENT)

	err := encoder.Encode(data)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

func registerCommand(command *Command) {
	for aliasIndex := range command.Aliases {
		CommandMap[strings.ToLower(command.Aliases[aliasIndex])] = command
	}
}

func findCommand(label string) *Command {
	return CommandMap[strings.ToLower(label)]
}
