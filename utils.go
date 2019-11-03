package main

import (
	"bytes"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
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
	for roleIndex := range member.Roles {
		role := member.Roles[roleIndex]
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
	if err != nil {
		return resources
	}
	if len(resourceString) == 0 {
		return resources
	} else {
		err = json.Unmarshal([]byte(resourceString), &resources)
		CheckErr(err)
	}
	return resources
}

//noinspection SpellCheckingInspection
func GetUser(discordUser *discordgo.User) (user User, err error) {
	row := stmtFindUserRow.QueryRow(discordUser.ID)
	var resourceString, ipAddressesString string

	err = row.Scan(&user.ID, &user.Token, &user.DiscordID, &resourceString, &ipAddressesString)
	if err != nil {
		return user, GetUserError{"Unable to find user with DiscordID " + discordUser.ID + " E: " + err.Error()}
	}
	err = json.Unmarshal([]byte(resourceString), &user.Resources)
	CheckErr(err)
	err = json.Unmarshal([]byte(ipAddressesString), &user.IPAddresses)
	CheckErr(err)
	return user, err
}

var ipAddressesPattern, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

func IsValidIP4(ipAddress string) bool {
	return ipAddressesPattern.MatchString(strings.Trim(ipAddress, " "))
}

func JoinArray(array []string) string {
	if len(array) == 0 {
		return ""
	} else if len(array) == 1 {
		return array[0]
	} else if len(array) == 2 {
		return array[0] + " and " + array[1]
	} else {
		result := ""

		for e := range array {
			part := array[e]
			if e == len(array)-1 {
				result += part
			} else if e == len(array)-2 {
				result += part + ", and "
			} else {
				result += part + ", "
			}
		}
		return result
	}
}

var patternChannels = regexp.MustCompile("<#[^>]*>")

func GetChannelMentions(message *discordgo.Message, searches int) (mentions []string) {
	mentions = patternChannels.FindAllString(message.Content, searches)

	for e := range mentions {
		mention := mentions[e]
		mentions[e] = mention[2:20]
	}
	return mentions
}

func GetResource(resourceName string) (resource Resource, err error) {
	err = stmtFindResourceRow.QueryRow(resourceName).Scan(&resource.ID, &resource.Name, &resource.ResponseData, &resource.DiscordChannelID)
	if err != nil {
		return resource, GetResourceError{"Unable to find resource with name " + resourceName + " E: " + err.Error()}
	}
	return resource, err
}

func GetAllResources() (resources []Resource) {
	rows, err := stmtGetAllResources.Query()
	CheckErr(err)
	for rows.Next() {
		var resource Resource
		err = rows.Scan(&resource.ID, &resource.Name, &resource.ResponseData, &resource.DiscordChannelID)
		resources = append(resources, resource)
	}
	return resources
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
	"0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GetRoleFromRoleID(guild *discordgo.Guild, roleID string) *discordgo.Role {
	for e := range guild.Roles {
		role := guild.Roles[e]
		if role.ID == roleID {
			return role
		}
	}
	return nil
}

func GetResourcesFromRoles(session *discordgo.Session, member *discordgo.Member) (resources []string) {
	guild := GetGuild(session, member.GuildID)
	for e := range member.Roles {

		resource, err := GetResource(GetRoleFromRoleID(guild, member.Roles[e]).Name)
		if err == nil {
			resources = append(resources, resource.Name)
		}
	}
	return resources
}
