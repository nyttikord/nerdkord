package commands

import (
	"bytes"
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/data"
	"github.com/nyttikord/nerdkord/libs/latex2png"
	"strings"
	"text/template"
)

const (
	EditPreambleID        = "edit_preamble"
	ResetPreambleID       = "reset_preamble"
	ReallyResetPreambleID = "really_reset_preamble"
)

var (
	defaultPreamble = ""
)

func OnEditPreambleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	componentData := i.MessageComponentData()
	if componentData.CustomID != EditPreambleID {
		utils.SendDebug("commands/preamble.go - not a profile button ID")
		return
	}
	var u *discordgo.User
	if i.User != nil {
		u = i.User
	} else {
		u = i.Member.User
	}
	var val string
	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendWarn("Getting nerd profile", "err", err.Error(), "discord_id", u.ID)

		val, err = getDefaultPreamble()
		if err == nil {
			utils.SendDebug("Using default preamble as placeholder")
		} else {
			utils.SendAlert("commands/preamble.go - Getting default preamble", err.Error())
			utils.SendDebug("Using empty preamble")
			val = ""
		}
	} else {
		val = nerd.Preamble
		if len(nerd.Preamble) == 0 {
			val, err = getDefaultPreamble()
			if err != nil {
				utils.SendWarn("Getting default preamble", "err", err.Error())
			}
		}
	}
	err = utils.NewResponseBuilder(s, i).
		IsModal().
		SetTitle("Edit preamble").
		SetCustomID(EditPreambleID).
		AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.TextInput{
				CustomID:    "source",
				Label:       "Preamble",
				Style:       discordgo.TextInputParagraph,
				Placeholder: `\usepackage[french]{babel}`,
				Value:       val,
				Required:    true,
				MinLength:   0,
				MaxLength:   4000,
			},
		}}).Send()
	if err != nil {
		var u *discordgo.User
		if i.User == nil {
			u = i.Member.User
		} else {
			u = i.User
		}
		utils.SendAlert("commands/preamble.go - Sending modal to edit preamble", err.Error(), "discord_id", u.ID)
	}
}

func OnResetPromptPreambleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	componentData := i.MessageComponentData()
	if componentData.CustomID == ResetPreambleID {
		err := utils.NewResponseBuilder(s, i).
			IsEphemeral().
			SetMessage("Are you sure you want to reset your preamble ?").
			AddComponent(&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Yes, reset my preamble",
						Style:    discordgo.DangerButton,
						Disabled: false,
						CustomID: ReallyResetPreambleID,
					},
				},
			}).Send()

		if err != nil {
			utils.SendAlert("commands/preamble.go - Sending reset confirmation", err.Error())
		}
		return
	}
	if componentData.CustomID != ReallyResetPreambleID {
		utils.SendDebug("commands/preamble.go - not a reset button ID")
		return
	}

	resp := utils.NewResponseBuilder(s, i).IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}

	nerd.Preamble = ""

	err = nerd.Save()
	if err != nil {
		utils.SendAlert("commands/preamble.go - Resetting preamble", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while resetting your preamble. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending error resetting preamble", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble reset to default").Send(); err != nil {
		utils.SendAlert("commands/preamble.go - Sending reset success", err.Error())
	}
}

func OnPreambleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	submitData := i.ModalSubmitData()
	if submitData.CustomID != EditPreambleID {
		utils.SendDebug("commands/preamble.go - not a profile modal ID")
		return
	}
	resp := utils.NewResponseBuilder(s, i).IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}

	val := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	if strings.Contains(val, `\documentclass`) {
		if err = resp.SetMessage("You can't use `\\documentclass`").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending error \\documentclass is present", err.Error())
		}
		return
	}
	if err = resp.IsDeferred().Send(); err != nil {
		utils.SendAlert("commands/preamble.go - Sending deferred", err.Error(), "discord_id", u.ID)
		return
	}
	utils.SendDebug("Checking preamble's validity", "discord_id", u.ID)
	file := new(bytes.Buffer)
	err = latex2png.Compile(file, "hey, this code was written by a furry", &latex2png.Options{
		LatexBinary:          "latex",
		DvipngBinary:         "dvipng",
		OutputFormat:         latex2png.PNG,
		BackgroundColor:      bgColor,
		ForegroundColor:      fgColor,
		ImageDPI:             10, // reduce DPI for faster results
		PreprocessingOptions: defaultPreprocessingOptions,
	})
	if err != nil {
		if err = resp.SetMessage("Your preamble is invalid.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending invalid preamble", err.Error())
		}
		return
	}
	utils.SendDebug("Updating nerd's preamble", "discord_id", u.ID)
	nerd.Preamble = val
	err = nerd.Save()
	if err != nil {
		utils.SendAlert("commands/preamble.go - Saving preamble", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while saving your preamble. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Sending error getting nerd", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble saved").Send(); err != nil {
		utils.SendAlert("commands/preamble.go - Sending success", err.Error())
	}
}

func Preamble(_ *discordgo.Session, i *discordgo.InteractionCreate, _ utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/preamble.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your preamble. Please report.").Send(); err != nil {
			utils.SendAlert("commands/preamble.go - Getting nerd error", err.Error(), "discord_id", u.ID)
		}
		return
	}
	if len(nerd.Preamble) == 0 {
		nerd.Preamble, err = getDefaultPreamble()
		if err != nil {
			if err = resp.SetMessage("An error occurred. Please report the bug.").Send(); err != nil {
				utils.SendAlert("commands/preamble.go - Sending error occurred while parsing template", err.Error())
			}
			return
		}
	}
	err = resp.AddEmbed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's preamble", u.Username),
		Description: fmt.Sprintf("Your preamble:\n```tex\n%s\n```", nerd.Preamble),
		Color:       0,
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Edit",
			Style:    discordgo.PrimaryButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "✏️"},
			CustomID: EditPreambleID,
		},
		discordgo.Button{
			Label:    "Reset",
			Style:    discordgo.DangerButton,
			Disabled: false,
			Emoji:    &discordgo.ComponentEmoji{Name: "🔄"},
			CustomID: ResetPreambleID,
		},
	}}).Send()
	if err != nil {
		utils.SendAlert("commands/preamble.go - Sending preamble", err.Error(), "discord_id", u.ID)
	}
}

func getDefaultPreamble() (string, error) {
	if len(defaultPreamble) == 0 {
		t, err := template.ParseFiles(defaultPreprocessingOptions.TemplateFile)
		if err != nil {
			utils.SendAlert(
				"commands/preamble.go - Parsing template file", err.Error(),
				"path", defaultPreprocessingOptions.TemplateFile,
			)
		} else {
			wr := new(bytes.Buffer)
			err = t.ExecuteTemplate(wr, "defaultPreamble", nil)
			if err != nil {
				return "", err
			}
			defaultPreamble = wr.String()
		}
	}
	if len(defaultPreamble) == 0 {
		return "Default one", nil
	}
	return defaultPreamble, nil
}
