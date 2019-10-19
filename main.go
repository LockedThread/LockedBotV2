package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

import _ "github.com/go-sql-driver/mysql"

var (
	MySQL              *sql.DB = nil
	Token              string
	CommandMap         map[string]*Command
	UserMap            map[string]*User
	Config             *Configuration
	AvailableResources []Resource

	StmtInsertResource *sql.Stmt
	StmtFindResource   *sql.Stmt
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	Config = Config.SetupConfig()

	db, err := sql.Open("mysql", Config.User+":"+Config.Password+"@tcp("+Config.Host+")/"+Config.DbName)
	checkErr(err)

	MySQL = db
	initResourceFile()
	initPreparedStatements()

	CommandMap = make(map[string]*Command)
	discord, err := discordgo.New("Bot " + Token)
	checkErr(err)
	err = discord.Open()
	checkErr(err)

	err = discord.UpdateListeningStatus("LockedThread's Front Door")
	checkErr(err)

	discord.AddHandler(messageCreate)

	registerCommand(&Command{
		[]string{"-addresource"},
		func(data CommandData) {
			if isOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.sendMessage("-addresource [@mention] [resource/role]")
					break
				case 2:
					mentions := data.Message.Mentions

					guildMember, err := data.Session.GuildMember(data.GuildID, mentions[0].ID)
					checkErr(err)
					guild := getGuild(data.Session, data.GuildID)
					role := getRole(guild, data.Arguments[1])
					if role == nil {
						data.sendMessage("Unable to add role %[1]s to %[2]s because that role doesn't exist!", data.Arguments[1], guildMember.Mention())
						return
					}
					hasRole := hasRole(guildMember, role.ID)
					if !hasRole {
						err := data.Session.GuildMemberRoleAdd(guild.ID, guildMember.User.ID, role.ID)
						if err != nil {
							data.sendMessage("An error occured report this to LockedThread now!")
						} else {
							data.sendMessage("You have added the resource %[1]s to %[2]s.", role.Name, guildMember.Mention())
						}
					} else {
						data.sendMessage("%s already has that role but we will update their resource list in the database.", guildMember.Mention())
					}
					break
				}
			} else {
				data.sendNoPermission()
			}
		},
	})

	registerCommand(&Command{
		[]string{"-createresource"},
		func(data CommandData) {
			if isOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
					data.sendMessage("-createresource [resource/rolename]")
					break
				case 1:
					guild := getGuild(data.Session, data.GuildID)
					role := getRole(guild, data.Arguments[0])
					if role == nil {
						role, err = data.Session.GuildRoleCreate(guild.ID)
						checkErr(err)
						role, err = data.Session.GuildRoleEdit(guild.ID, role.ID, data.Arguments[0], 0xdb7c23, role.Hoist, 3263553, false)
						checkErr(err)
						data.sendMessage("Create role & resource with name %s", role.Name)
					} else {
						data.sendMessage("Resource already found with name %s", role.Name)
					}
					resource := findResource(role.Name)
					if resource == nil {
						resource = &Resource{
							RoleID:   role.ID,
							RoleName: role.Name,
						}
						AvailableResources = append(AvailableResources, *resource)
					}
					rows, err := StmtFindResource.Query(resource.RoleName)
					checkErr(err)

					next := rows.Next()
					if next == false {
						_, err := StmtInsertResource.Exec(resource.RoleName, "")
						checkErr(err)
					}
					err = rows.Close()
					checkErr(err)
				}
			} else {
				data.sendNoPermission()
			}
		},
	})

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = MySQL.Close()
	_ = discord.Close()
	saveResourceFile()
}

func initPreparedStatements() {
	stmt, err := MySQL.Prepare("INSERT INTO " + Config.Tables.ResourcesTable + " (resource_name, response_data) VALUES(?,?)")
	checkErr(err)
	StmtInsertResource = stmt
	stmt, err = MySQL.Prepare("SELECT * FROM " + Config.Tables.ResourcesTable + " WHERE resource_name = ?")
	checkErr(err)
	StmtFindResource = stmt
}

func initResourceFile() []Resource {
	jsonFile, err := os.Open("resources.json")
	if err != nil {
		_, err = os.Create("resources.json")
		checkErr(err)
		return []Resource{}
	}
	fmt.Println("Successfully opened resources.json")
	defer jsonFile.Close()

	var resources []Resource

	bytes, err := ioutil.ReadAll(jsonFile)
	checkErr(err)
	err = json.Unmarshal(bytes, &resources)
	checkErr(err)
	return resources
}

func saveResourceFile() {
	bytes, err := getPrettyPrinted(&AvailableResources)
	checkErr(err)
	err = ioutil.WriteFile("resources.json", bytes, os.ModeAppend)
	checkErr(err)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	splitMessage := strings.Split(m.Message.Content, " ")
	command := findCommand(splitMessage[0])

	if command != nil {
		channel, err := s.Channel(m.ChannelID)
		checkErr(err)

		var arguments []string
		if len(splitMessage) >= 2 {
			arguments = splitMessage[1:]
		} else {
			arguments = []string{}
		}

		command.execute(CommandData{
			Label:     splitMessage[0],
			GuildID:   m.GuildID,
			Message:   m.Message,
			User:      m.Author,
			Arguments: arguments,
			Channel:   channel,
			Session:   s,
		})
	}
}
