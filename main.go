package main

import (
	_ "embed"
	"errors"
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/nyttikord/nerdkord/commands"
	"github.com/nyttikord/nerdkord/db"
	"github.com/nyttikord/nerdkord/latex"
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
		logger.Warn("Error while loading .env file", "error", err.Error())
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

	latexCmd := cmd.New("latex", "Compiles latex source").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionString,
			"source",
			"LaTeX source code",
		)).
		SetHandler(commands.Latex)

	latexifyCmd := cmd.New("latexify", "Converts a math expression to latex").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The math expression to convert").IsRequired()).
		SetHandler(commands.Latexify)

	calculateCmd := cmd.New("calculate", "Parses and evaluates a math expression").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionString,
			"expression",
			"The expression you want to evaluate").IsRequired()).
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionInteger,
			"precision",
			"The number of digits you want. Default : 6")).
		SetHandler(commands.Calculate)

	preamble := cmd.New("preamble", "Show and edit your preamble").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		SetHandler(commands.Preamble)

	about := cmd.New("about", "About the bot").
		AddIntegrationType(discordgo.ApplicationIntegrationGuildInstall).
		AddIntegrationType(discordgo.ApplicationIntegrationUserInstall).
		AddContext(discordgo.InteractionContextGuild).
		AddContext(discordgo.InteractionContextPrivateChannel).
		AddContext(discordgo.InteractionContextBotDM).
		SetHandler(commands.About)

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
		Commands: []cmd.CommandBuilder{
			latexCmd, latexifyCmd, calculateCmd, preamble, about,
		},
		Innovations: innovations,
		Version:     Version,
		Intents:     discordgo.IntentsAllWithoutPrivileged,
	}

	//commands: latex
	bot.HandleModal(commands.OnLatexModalSubmit, commands.LaTeXModalID)
	bot.HandleMessageComponent(commands.OnSourceButton, latex.GetSourceID)
	//commands: preamble
	bot.HandleMessageComponent(commands.OnEditPreambleButton, commands.EditPreambleID)
	bot.HandleMessageComponent(commands.OnResetPromptPreambleButton, commands.ResetPreambleID)
	bot.HandleMessageComponent(commands.OnReallyResetPromptPreambleButton, commands.ReallyResetPreambleID)
	//event: latex
	bot.AddHandler(latex.HandleLatexSourceCode)

	bot.Start()
}
