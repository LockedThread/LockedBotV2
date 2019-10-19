package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

import _ "github.com/go-sql-driver/mysql"

var (
	MySQL      *sql.DB = nil
	Token      string
	CommandMap map[string]*Command
	UserMap    map[string]*User
	Config     *Configuration

	StmtInsertResourceRow  *sql.Stmt
	StmtFindResourceRow    *sql.Stmt
	StmtFindResourceColumn *sql.Stmt

	StmtUpdateUserResourceColumn *sql.Stmt
	StmtInsertUserRow            *sql.Stmt
	StmtFindUserRow              *sql.Stmt
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
					data.sendMessage("Incorrect Syntax. Please do -addresource [@mention] [resource/role]")
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
					data.sendMessage("Incorrect Syntax. Please do -createresource [resource/rolename]")
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
					rows, err := StmtFindResourceRow.Query(role.Name)
					checkErr(err)

					next := rows.Next()
					if next == false {
						_, err := StmtInsertResourceRow.Exec(role.Name, "")
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

	registerCommand(&Command{
		[]string{"-createclient"},
		func(data CommandData) {
			if isOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.sendMessage("Incorrect Syntax. Please do -createclient [@mention] [token]")
					break
				case 2:
					mentions := data.Message.Mentions
					if len(mentions) == 1 {
						mentionedUser := mentions[0]

						rows, err := StmtFindUserRow.Query(mentionedUser.ID)
						checkErr(err)

						next := rows.Next()
						if next {
							data.sendMessage("Unable to create client for %s because that client already exists in the database!", mentionedUser.Mention())
						} else {
							_, err := StmtInsertUserRow.Exec(data.Arguments[1], mentionedUser.ID, "", "")
							checkErr(err)
							data.sendMessage("Created client for %s.", mentionedUser.Mention())
						}
						err = rows.Close()
						checkErr(err)

					} else {
						data.sendMessage("Incorrect Syntax. Please do -createclient [@mention] [token]")
					}
					break
				}
			} else {
				data.sendNoPermission()
			}
		},
	})

	registerCommand(&Command{
		[]string{"-addresource"},
		func(data CommandData) {
			if isOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.sendMessage("Incorrect Syntax. Please do -addresource [@mention] [resource]")
					break
				case 2:
					mentions := data.Message.Mentions
					if len(mentions) == 1 {
						mentionedUser := mentions[0]
						var resources []string
						if data.Arguments[1] == "*" {
							resources = []string{"*"}
						} else {
							resources = getResources(mentionedUser)
							for e := range resources {
								if resources[e] == "*" {
									data.sendMessage("That client has a resource wildcard, no point in adding a resource!")
									return
								} else if strings.ToLower(resources[e]) == strings.ToLower(data.Arguments[1]) {
									data.sendMessage("That resource is already found for %s", mentionedUser.Mention())
									return
								}
							}
							resources = append(resources, data.Arguments[1])
						}

						bytes, err := json.Marshal(resources)
						checkErr(err)
						_, err = StmtUpdateUserResourceColumn.Exec(string(bytes), mentionedUser.ID)
						checkErr(err)

						data.sendMessage("Added resource to %s", mentionedUser.Mention())
					} else {
						data.sendMessage("Incorrect Syntax. Please do -addresource [@mention] [resource]")
					}
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
}

func initPreparedStatements() {
	stmt, err := MySQL.Prepare("INSERT INTO " + Config.Tables.ResourcesTable + " (resource_name, response_data) VALUES(?,?)")
	checkErr(err)
	StmtInsertResourceRow = stmt

	stmt, err = MySQL.Prepare("SELECT * FROM " + Config.Tables.ResourcesTable + " WHERE resource_name = ?")
	checkErr(err)
	StmtFindResourceRow = stmt

	stmt, err = MySQL.Prepare("SELECT resources FROM " + Config.Tables.UserTable + " WHERE discord_id = ?")
	checkErr(err)
	StmtFindResourceColumn = stmt

	stmt, err = MySQL.Prepare("INSERT INTO " + Config.Tables.UserTable + " (token, discord_id, resources, ip_addresses) VALUES(?,?,?,?)")
	checkErr(err)
	StmtInsertUserRow = stmt

	stmt, err = MySQL.Prepare("SELECT * FROM " + Config.Tables.UserTable + " WHERE discord_id = ?")
	checkErr(err)
	StmtFindUserRow = stmt

	stmt, err = MySQL.Prepare("UPDATE " + Config.Tables.UserTable + " SET resources = ? WHERE discord_id = ?")
	checkErr(err)
	StmtUpdateUserResourceColumn = stmt
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
