package main

import (
	_ "embed"
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/bwmarrin/discordgo"
)

var (
	token   string
	Version = &gokord.Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
	//go:embed updates.json
	updatesData []byte
)

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
}

func main() {
	flag.Parse()
	gokord.UseRedis = false
	err := gokord.SetupConfigs(&Config{}, nil)
	if err != nil {
		panic(err)
	}

	innovations, err := gokord.LoadInnovationFromJson(updatesData)
	if err != nil {
		panic(err)
	}

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.WatchStatus,
				Content: "Les Copaings",
			},
			{
				Type:    gokord.GameStatus,
				Content: "être dev par @anhgelus",
			},
			{
				Type:    gokord.ListeningStatus,
				Content: "http 418, I'm a tea pot",
			},
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot " + Version.String(),
			},
		},
		Commands:    []gokord.CommandBuilder{},
		AfterInit:   afterInit,
		Innovations: innovations,
		Version:     Version,
		Intents: discordgo.IntentsAllWithoutPrivileged |
			discordgo.IntentsMessageContent |
			discordgo.IntentGuildMembers,
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	// handles here
}
