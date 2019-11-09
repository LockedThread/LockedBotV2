package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

import _ "github.com/go-sql-driver/mysql"

var (
	mySQL      *sql.DB = nil
	token      string
	commandMap map[string]*Command
	config     *Configuration

	stmtInsertResourceRow  *sql.Stmt
	stmtFindResourceRow    *sql.Stmt
	stmtFindResourceColumn *sql.Stmt
	stmtUpdateResourceName *sql.Stmt
	stmtGetAllResources    *sql.Stmt
	stmtDeleteResource     *sql.Stmt

	stmtUpdateUserResourceColumn *sql.Stmt
	stmtInsertUserRow            *sql.Stmt
	stmtFindUserRow              *sql.Stmt
	stmtUpdateUserIPColumn       *sql.Stmt
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	config = config.SetupConfig()

	db, err := sql.Open("mysql", config.User+":"+config.Password+"@tcp("+config.Host+")/"+config.DbName)
	CheckErr(err)

	mySQL = db
	InitPreparedStatements()

	commandMap = make(map[string]*Command)
	discord, err := discordgo.New("Bot " + token)
	CheckErr(err)
	err = discord.Open()
	CheckErr(err)

	err = discord.UpdateListeningStatus("LockedThread's Front Door")
	CheckErr(err)

	discord.AddHandler(messageCreate)

	RegisterCommand(&Command{
		Aliases: []string{"-importclients"},
		Execute: func(data CommandData) {
			if IsOwner(data.User) {
				guild := data.GetGuild()
				for e := range guild.Members {
					member := guild.Members[e]

					resourceArray := GetResourcesFromRoles(data.Session, member)
					if len(resourceArray) == 0 {
						continue
					}

					bytes, err := json.Marshal(resourceArray)
					CheckErr(err)

					rows, err := stmtFindUserRow.Query(member.User.ID)
					CheckErr(err)

					if !rows.Next() {
						licenseKey := RandomString(10)
						_, err = stmtInsertUserRow.Exec(licenseKey, member.User.ID, string(bytes), "[]")
						CheckErr(err)
						channel, err := data.Session.UserChannelCreate(member.User.ID)
						if err != nil {
							data.SendMessage("Unable to message %s their new information.", member.User.Mention())
						} else {
							_, err = data.Session.ChannelMessageSendEmbed(channel.ID, NewEmbed().
								SetTitle("RESOURCE IMPORT").
								SetDescription("New License Key: %s\nLicensed Resources: %s\nIf you are missing any resource please message LockedThread ASAP. Thank you for being an amazing customer :heart:", licenseKey, JoinArray(resourceArray)).
								SetColor(Green).MessageEmbed)
							if err != nil {
								data.SendMessage("Unable to message %s their new information.", member.User.Mention())
							}
						}
					} else {
						_, err = stmtUpdateUserResourceColumn.Exec(string(bytes), member.User.ID)
						/*channel, err := data.Session.UserChannelCreate(member.User.ID)
						if err != nil {
							data.SendMessage("Unable to message %s their new information.", member.User.Mention())
						} else {
							_, err = data.Session.ChannelMessageSendEmbed(channel.ID, NewEmbed().
								SetTitle("RESOURCE IMPORT").
								SetDescription("Licensed Resources: %s\nIf you are missing any resource please message LockedThread ASAP. Thank you for being an amazing customer :heart:", JoinArray(resourceArray)).
								SetColor(Green).MessageEmbed)
							if err != nil {
								data.SendMessage("Unable to message %s their new information.", member.User.Mention())
							}
						}*/
					}
				}
			} else {
				data.SendNoPermission()
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-help"},
		func(data CommandData) {

			description :=
			//"-prices | dms you prices on all of our products & services\n" +
				"*-addresource {@mention} {resource} | Adds resource to client for the auth system\n" +
					"*-createresource {name} | Creates resource for the auth system\n" +
					"*-createclient {@mention} {token} | Creates client in database\n" +
					"*-update {resource} {changelog} | Updates a resource with a changelog message\n" +
					"*-rename {resource} {new name} | Changes the name of a resource\n" +
					"*-setupdatechannel {resource} {channel} | Sets the update channel for a resource\n" +
					"*-deleteresource {resource} | Deletes that resource\n" +
					"-listresources | Shows you all of our resources\n" +
					"-addip {ip-address} | Adds ip to your whitelisted ip addresses\n" +
					"-setupclient {token} | Sets up your client data in the database\n" +
					"-clientinfo {@mention} | Displays all information about a client\n\n" +
					"All commands annotated by * are admin only!"

			data.SendEmbed(NewEmbed().
				SetTitle("Help for LockedBot V2").
				SetDescription(description).
				SetColor(Green))
		},
	})

	RegisterCommand(&Command{
		Aliases: []string{"-deleteresource"},
		Execute: func(data CommandData) {
			if IsOwner(data.User) {
				if len(data.Arguments) == 1 {
					result, err := stmtDeleteResource.Exec(data.Arguments[0])
					if err != nil {
						data.SendEmbed(NewEmbed().
							SetTitle("ERROR").
							SetDescription("There was some type of SQL error. (%s)", err.Error()).
							SetColor(Red))
					} else {
						rowsAffected, _ := result.RowsAffected()
						if rowsAffected == 0 {
							data.SendEmbed(NewEmbed().
								SetTitle("ERROR").
								SetDescription("There was no resources found with %s to delete", data.Arguments[0]).
								SetColor(Red))
						} else {
							data.SendEmbed(NewEmbed().
								SetTitle("SUCCESS").
								SetDescription("Successfully removed %s resource", data.Arguments[0]).
								SetColor(Green))
						}
					}
				} else {
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -deleteresource {resource}").
						SetColor(Red))
				}
			} else {
				data.SendNoPermission()
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-setupclient"},
		func(data CommandData) {

			switch len(data.Arguments) {
			case 0:
				data.SendEmbed(NewEmbed().
					SetTitle("Incorrect Syntax").
					SetDescription("Incorrect Syntax. Please do -setupclient {token}").
					SetColor(Red))
				break
			case 1:
				rows, err := stmtFindUserRow.Query(data.User.ID)
				CheckErr(err)

				next := rows.Next()
				if next {
					data.SendEmbed(NewEmbed().
						SetTitle("ERROR").
						SetDescription("You have already setup yourself in the database. Make sure to add your ips using -addip <3").
						SetColor(Red))
				} else {
					_, err := stmtInsertUserRow.Exec(data.Arguments[0], data.User.ID, "[]", "[]")
					CheckErr(err)

					data.SendEmbed(NewEmbed().
						SetTitle("SUCCESS").
						SetDescription("Thank you for setting up yourself as a client, make sure to add your ips using -addip <3").
						SetColor(Green))
				}
				err = rows.Close()
				CheckErr(err)
				break
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-addip"},
		func(data CommandData) {
			switch len(data.Arguments) {
			case 0:
				data.SendEmbed(NewEmbed().
					SetTitle("Incorrect Syntax").
					SetDescription("Incorrect Syntax. Please do -addip [ip]").
					SetColor(Red))
				break
			case 1:
				if !IsValidIP4(data.Arguments[0]) {
					data.SendEmbed(NewEmbed().
						SetTitle("ERROR").
						SetDescription("%s is not a valid IPv4", data.Arguments[0]).
						SetColor(Red))
					break
				}
				user, err := GetUser(data.User)
				if err != nil {
					data.SendEmbed(NewEmbed().
						SetTitle("ERROR").
						SetDescription(err.Error()).
						SetColor(Red))
					break
				} else {
					for e := range user.IPAddresses {
						if user.IPAddresses[e] == data.Arguments[0] {
							embed := data.SendEmbed(NewEmbed().
								SetTitle("ERROR").
								SetDescription("You already have that IP in your whitelisted ip address list").
								SetColor(Red))
							timer := time.NewTimer(3 * time.Second)
							go func() {
								<-timer.C
								err = data.Session.ChannelMessageDelete(data.Channel.ID, embed.ID)
								CheckErr(err)
								err = data.Session.ChannelMessageDelete(data.Channel.ID, data.Message.ID)
								CheckErr(err)
							}()

							return
						}
					}

					user.IPAddresses = append(user.IPAddresses, data.Arguments[0])

					bytes, err := json.Marshal(user.IPAddresses)
					CheckErr(err)
					_, err = stmtUpdateUserIPColumn.Exec(string(bytes), user.DiscordID)
					CheckErr(err)

					embed := data.SendEmbed(NewEmbed().
						SetTitle("SUCCESS").
						SetDescription("You have added %s to your whitelisted ip address list", data.Arguments[0]).
						SetColor(Green))

					timer := time.NewTimer(3 * time.Second)
					go func() {
						<-timer.C
						err = data.Session.ChannelMessageDelete(data.Channel.ID, embed.ID)
						CheckErr(err)
						err = data.Session.ChannelMessageDelete(data.Channel.ID, data.Message.ID)
						CheckErr(err)
					}()
					break
				}
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-listresources"},
		func(data CommandData) {
			resources := GetAllResources()

			var description string

			if len(resources) == 0 {
				description = "**None**"
			} else {
				resourceNameArray := make([]string, len(resources))
				for e := range resources {
					resourceNameArray[e] = resources[e].Name
				}
				description = JoinArray(resourceNameArray)
			}

			data.SendEmbed(NewEmbed().
				SetTitle("Resources").
				SetDescription(description).
				SetColor(Yellow))
		},
	})

	RegisterCommand(&Command{
		[]string{"-rename"},
		func(data CommandData) {
			if IsOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -rename {resource} {new name}").
						SetColor(Red))
					break
				default:
					resourceString := data.Arguments[0]
					resource, err := GetResource(resourceString)
					if err != nil {
						data.SendEmbed(NewEmbed().
							SetTitle("ERROR").
							SetDescription("Unable to find resource with name %s", resourceString).
							SetColor(Red))
					} else {
						_, err := stmtUpdateResourceName.Exec(data.Arguments[1], resource.Name)
						if err != nil {
							data.SendEmbed(NewEmbed().
								SetTitle("ERROR").
								SetDescription("There was a SQL error, this shouldn't be happening message LockedThread").
								SetColor(Red))
						} else {
							data.SendEmbed(NewEmbed().
								SetTitle("SUCCESS").
								SetDescription("Successfully changed resource %s to %s ", resource.Name, data.Arguments[1]).
								SetColor(Green))
						}
					}
					break
				}
			} else {

				data.SendNoPermission()

			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-update"},
		func(data CommandData) {
			if IsOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -update {resource} {changelog}").
						SetColor(Red))
					break
				default:
					resourceString := data.Arguments[0]
					resource, err := GetResource(resourceString)
					if err != nil {
						data.SendEmbed(NewEmbed().
							SetTitle("ERROR").
							SetDescription("Unable to find resource with name %s", resourceString).
							SetColor(Red))
					} else if len(data.Message.Attachments) > 0 {
						files := make([]*discordgo.File, len(data.Message.Attachments))
						var resp *http.Response
						for e := range data.Message.Attachments {
							attachment := data.Message.Attachments[e]
							resp, err = http.Get(attachment.URL)
							CheckErr(err)
							files[e] = &discordgo.File{
								Name:   attachment.Filename,
								Reader: resp.Body,
							}
						}
						defer resp.Body.Close()
						_, err := data.Session.ChannelMessageSendComplex(resource.DiscordChannelID, &discordgo.MessageSend{
							Embed: NewEmbed().
								SetTitle("UPDATE").
								SetDescription(strings.Join(data.Arguments[1:len(data.Arguments)], " ")).
								SetColor(Green).MessageEmbed,
							Files: files,
						})
						CheckErr(err)
					} else {
						_, err := data.Session.ChannelMessageSendEmbed(resource.DiscordChannelID, NewEmbed().
							SetTitle("UPDATE").
							SetDescription(strings.Join(data.Arguments[1:len(data.Arguments)], " ")).
							SetColor(Green).MessageEmbed)
						CheckErr(err)
					}
					break
				}
			} else {

				data.SendNoPermission()

			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-clientinfo"},
		func(data CommandData) {
			switch len(data.Arguments) {
			case 0:
				data.SendEmbed(NewEmbed().
					SetTitle("Incorrect Syntax").
					SetDescription("Incorrect Syntax. Please do -clientinfo [@mention]").
					SetColor(Red))
				break
			case 1:
				mentions := data.Message.Mentions

				if len(mentions) == 0 {
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -clientinfo [@mention]").
						SetColor(Red))
				} else {
					guildMember, err := data.Session.GuildMember(data.GuildID, mentions[0].ID)
					CheckErr(err)

					resources := GetResources(guildMember.User)

					var description string
					if len(resources) == 0 {
						description = "Resources: **None**"
					} else {
						description = "Resources:  " + JoinArray(resources)
					}

					data.SendEmbed(NewEmbed().
						SetTitle("INFO").
						SetDescription(description).
						SetColor(Yellow))
				}
				break
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-addresource"},
		func(data CommandData) {
			if IsOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -addresource [@mention] [resource/role]").
						SetColor(Red))
					break
				case 2:
					mentions := data.Message.Mentions
					if len(mentions) == 0 {
						data.SendEmbed(NewEmbed().
							SetTitle("Incorrect Syntax").
							SetDescription("Incorrect Syntax. Please do -addresource [@mention] [token]").
							SetColor(Red))
					} else {
						guildMember, err := data.Session.GuildMember(data.GuildID, mentions[0].ID)
						CheckErr(err)
						if data.Arguments[1] != "*" {
							guild := data.GetGuild()
							role := GetRole(guild, data.Arguments[1])
							if role == nil {
								data.SendEmbed(NewEmbed().
									SetTitle("ERROR").
									SetDescription("Unable to add role %[1]s to %[2]s because that role doesn't exist!", data.Arguments[1], guildMember.Mention()).
									SetColor(Red))
								return
							}
							hasRole := HasRole(guildMember, role.ID)
							if !hasRole {
								err := data.Session.GuildMemberRoleAdd(guild.ID, guildMember.User.ID, role.ID)
								if err != nil {
									data.SendEmbed(NewEmbed().
										SetTitle("ERROR").
										SetDescription("An error occured report this to LockedThread now!").
										SetColor(Red))
								}
							} else {
								data.SendEmbed(NewEmbed().
									SetTitle("SUCCESS").
									SetDescription("%s already has that role but we will update their resource list in the database.", guildMember.Mention()).
									SetColor(Green))
							}
						}
						var resources []string
						if data.Arguments[1] == "*" {
							resources = []string{"*"}
						} else {
							resources = GetResources(guildMember.User)
							for e := range resources {
								if resources[e] == "*" {
									data.SendEmbed(NewEmbed().
										SetTitle("ERROR").
										SetDescription("That client has a resource wildcard, no point in adding a resource!").
										SetColor(Red))
									return
								} else if strings.ToLower(resources[e]) == strings.ToLower(data.Arguments[1]) {
									data.SendEmbed(NewEmbed().
										SetTitle("ERROR").
										SetDescription("That resource is already found for %s in the database", guildMember.Mention()).
										SetColor(Red))
									return
								}
							}
							resources = append(resources, data.Arguments[1])
						}

						bytes, err := json.Marshal(resources)
						CheckErr(err)
						_, err = stmtUpdateUserResourceColumn.Exec(string(bytes), guildMember.User.ID)
						CheckErr(err)

						data.SendEmbed(NewEmbed().
							SetTitle("Success").
							SetDescription("You have added the resource %[1]s to %[2]s.", data.Arguments[1], guildMember.Mention()).
							SetColor(Green))
						break
					}
				}
			} else {
				data.SendNoPermission()
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-createresource"},
		func(data CommandData) {
			if IsOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -createresource {resource/rolename} [channel-mention]").
						SetColor(Red))
					break
				case 1:
				case 2:
					guild := data.GetGuild()
					role := GetRole(guild, data.Arguments[0])
					var roleFound, createdRole bool
					if role == nil {
						role, err = data.Session.GuildRoleCreate(guild.ID)
						CheckErr(err)
						role, err = data.Session.GuildRoleEdit(guild.ID, role.ID, data.Arguments[0], 0xdb7c23, role.Hoist, 3263553, false)
						CheckErr(err)
						createdRole = true
					} else {
						roleFound = true
					}
					rows, err := stmtFindResourceRow.Query(role.Name)
					CheckErr(err)

					next := rows.Next()
					if next {
						if roleFound {
							data.SendEmbed(NewEmbed().
								SetTitle("ERROR").
								SetDescription("Resource already found with name %s", role.Name).
								SetColor(Red))
						}
						if createdRole {
							data.SendEmbed(NewEmbed().
								SetTitle("SUCCESS").
								SetDescription("Create role with name %s, found resource in database", role.Name).
								SetColor(Green))
						}
					} else {
						if len(data.Arguments) == 2 {
							mentions := GetChannelMentions(data.Message, 1)
							CheckErr(err)
							_, err = stmtInsertResourceRow.Exec(role.Name, "", mentions[0])
						} else {
							_, err = stmtInsertResourceRow.Exec(role.Name, "", "")
						}
						CheckErr(err)
						var message string
						if createdRole {
							message = "Create role & resource with name %s"
						} else {
							message = "Created resource with name %s"
						}

						data.SendEmbed(NewEmbed().
							SetTitle("SUCCESS").
							SetDescription(message, role.Name).
							SetColor(Green))

						err = rows.Close()
						CheckErr(err)
					}
					break
				}
			} else {
				data.SendNoPermission()
			}
		},
	})

	RegisterCommand(&Command{
		[]string{"-createclient"},
		func(data CommandData) {
			if IsOwner(data.User) {
				switch len(data.Arguments) {
				case 0:
				case 1:
					data.SendEmbed(NewEmbed().
						SetTitle("Incorrect Syntax").
						SetDescription("Incorrect Syntax. Please do -createclient [@mention] [token]").
						SetColor(Red))
					break
				case 2:
					mentions := data.Message.Mentions
					if len(mentions) == 1 {
						mentionedUser := mentions[0]

						rows, err := stmtFindUserRow.Query(mentionedUser.ID)
						CheckErr(err)

						next := rows.Next()
						if next {
							data.SendEmbed(NewEmbed().
								SetTitle("ERROR").
								SetDescription("Unable to create client for %s because that client already exists in the database!", mentionedUser.Mention()).
								SetColor(Red))
						} else {
							_, err := stmtInsertUserRow.Exec(data.Arguments[1], mentionedUser.ID, "[]", "[]")
							CheckErr(err)

							data.SendEmbed(NewEmbed().
								SetTitle("SUCCESS").
								SetDescription("Created client for %s.", mentionedUser.Mention()).
								SetColor(Green))
						}
						err = rows.Close()
						CheckErr(err)
					} else {
						data.SendEmbed(NewEmbed().
							SetTitle("Incorrect Syntax").
							SetDescription("Incorrect Syntax. Please do -createclient [@mention] [token]").
							SetColor(Red))
					}
					break
				}
			} else {
				data.SendNoPermission()
			}
		},
	})

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = mySQL.Close()
	_ = discord.Close()
}

func InitPreparedStatements() {
	var err error
	stmtInsertResourceRow, err = mySQL.Prepare("INSERT INTO " + config.Tables.ResourcesTable + " (resource_name, response_data, channel_id) VALUES(?,?,?)")
	CheckErr(err)

	stmtFindResourceRow, err = mySQL.Prepare("SELECT * FROM " + config.Tables.ResourcesTable + " WHERE resource_name = ?")
	CheckErr(err)

	stmtFindResourceColumn, err = mySQL.Prepare("SELECT resources FROM " + config.Tables.UserTable + " WHERE discord_id = ?")
	CheckErr(err)

	stmtInsertUserRow, err = mySQL.Prepare("INSERT INTO " + config.Tables.UserTable + " (token, discord_id, resources, ip_addresses) VALUES(?,?,?,?)")
	CheckErr(err)

	stmtFindUserRow, err = mySQL.Prepare("SELECT * FROM " + config.Tables.UserTable + " WHERE discord_id = ?")
	CheckErr(err)

	stmtUpdateUserIPColumn, err = mySQL.Prepare("UPDATE " + config.Tables.UserTable + " SET ip_addresses = ? WHERE discord_id = ?")
	CheckErr(err)

	stmtUpdateUserResourceColumn, err = mySQL.Prepare("UPDATE " + config.Tables.UserTable + " SET resources = ? WHERE discord_id = ?")
	CheckErr(err)

	stmtUpdateResourceName, err = mySQL.Prepare("UPDATE " + config.Tables.ResourcesTable + " SET resource_name = ? WHERE resource_name = ?")
	CheckErr(err)

	stmtGetAllResources, err = mySQL.Prepare("SELECT * FROM " + config.Tables.ResourcesTable)
	CheckErr(err)

	stmtDeleteResource, err = mySQL.Prepare("DELETE FROM " + config.Tables.ResourcesTable + " where resource_name = ?")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	splitMessage := strings.Split(m.Message.Content, " ")
	command := FindCommand(splitMessage[0])

	if command != nil {
		channel, err := s.Channel(m.ChannelID)
		CheckErr(err)

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
