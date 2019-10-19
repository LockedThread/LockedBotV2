package main

type User struct {
	SQLID       int
	Token       string
	DiscordID   string
	Resources   []string
	IPAddresses []string
}
