package main

import "strings"

type User struct {
	SQLID       int
	Token       string
	DiscordID   string
	Resources   []string
	IPAddresses []string
}

type Resource struct {
	RoleID   string `json:"roleId"`
	RoleName string `json:"roleName"`
}

func findResource(name string) *Resource {
	for resourceIndex := range AvailableResources {
		resource := AvailableResources[resourceIndex]
		if strings.ToLower(resource.RoleName) == name {
			return &resource
		}
	}
	return nil
}
