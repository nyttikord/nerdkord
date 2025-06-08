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

	err = gokord.DB.AutoMigrate(&Nerd{})
	if err != nil {
		panic(err)
	}

	innovations, err := gokord.LoadInnovationFromJson(updatesData)
	if err != nil {
		panic(err)
	}

	latexCmd := gokord.NewCommand("latex", "Compiles latex source").
		SetHandler(commands.Latex)

	latexifyCmd := gokord.NewCommand("latexify", "Converts a math expression to latex").
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The math expression to convert").IsRequired()).
		SetHandler(commands.Latexify)

	calculateCmd := gokord.NewCommand("calculate", "Parses and evaluates a math expression").
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The expression you want to evaluate").IsRequired()).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionInteger,
			"precision",
			"The number of digits you want. Default : 6")).
		SetHandler(commands.Calculate)

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
			latexifyCmd,
			calculateCmd,
		},
		AfterInit:   afterInit,
		Innovations: innovations,
		Version:     Version,
		Intents:     discordgo.IntentsAllWithoutPrivileged,
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	dg.AddHandler(commands.OnLatexModalSubmit)
}
