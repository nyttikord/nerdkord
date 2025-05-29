package main

import (
	_ "embed"
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/commands"
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

	latexCmd := gokord.NewCommand("latex", "Compiles latex source").
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"source",
			"The latex source you want to compile").IsRequired()).
		SetHandler(commands.Latex)

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.GameStatus,
				Content: "dev by nyttikord",
			},
			{
				Type:    gokord.ListeningStatus,
				Content: "maths",
			},
			{
				Type:    gokord.GameStatus,
				Content: "nerdkord " + Version.String(),
			},
		},
		Commands: []gokord.CommandBuilder{
			latexCmd,
		},
		AfterInit:   afterInit,
		Innovations: innovations,
		Version:     Version,
		Intents:     discordgo.IntentsAllWithoutPrivileged,
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	// handles here
}
