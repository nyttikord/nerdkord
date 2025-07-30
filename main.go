package main

import (
	_ "embed"
	"errors"
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/nyttikord/nerdkord/commands"
	"github.com/nyttikord/nerdkord/db"
	"os"
)

var (
	token   string
	Version = &gokord.Version{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}
	//go:embed updates.json
	updatesData []byte
)

func init() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		utils.SendWarn("Error while loading .env file", "error", err.Error())
	}

	flag.StringVar(&token, "token", os.Getenv("TOKEN"), "token of the bot")
}

func main() {
	flag.Parse()
	gokord.UseRedis = false
	err := gokord.SetupConfigs(&Config{}, nil)
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&db.Nerd{})
	if err != nil {
		panic(err)
	}

	innovations, err := gokord.LoadInnovationFromJson(updatesData)
	if err != nil {
		panic(err)
	}

	latexCmd := gokord.NewCommand("latex", "Compiles latex source").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"source",
			"LaTeX source code",
		)).
		SetHandler(commands.Latex)

	latexifyCmd := gokord.NewCommand("latexify", "Converts a math expression to latex").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The math expression to convert").IsRequired()).
		SetHandler(commands.Latexify)

	calculateCmd := gokord.NewCommand("calculate", "Parses and evaluates a math expression").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The expression you want to evaluate").IsRequired()).
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionInteger,
			"precision",
			"The number of digits you want. Default : 6")).
		SetHandler(commands.Calculate)

	preamble := gokord.NewCommand("preamble", "Show and edit your preamble").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		SetHandler(commands.Preamble)

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
			latexCmd, latexifyCmd, calculateCmd, preamble,
		},
		AfterInit:   afterInit,
		Innovations: innovations,
		Version:     Version,
		Intents:     discordgo.IntentsAllWithoutPrivileged,
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	//commands: latex
	dg.AddHandler(commands.OnLatexModalSubmit)
	dg.AddHandler(commands.OnSourceButton)
	//commands: preamble
	dg.AddHandler(commands.OnEditPreambleButton)
	dg.AddHandler(commands.OnResetPromptPreambleButton)
	dg.AddHandler(commands.OnPreambleModalSubmit)
}
