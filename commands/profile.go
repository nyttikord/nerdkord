package commands

import (
	"bytes"
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nyttikord/nerdkord/data"
	"strings"
	"text/template"
)

const (
	EditPreambleID = "edit_preamble"
)

var (
	defaultPreamble = ""
)

func OnProfileButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	componentData := i.MessageComponentData()
	if componentData.CustomID != EditPreambleID {
		utils.SendDebug("commands/profile.go - not a profile button ID")
		return
	}
	err := utils.NewResponseBuilder(s, i).
		IsModal().
		SetTitle("Edit preamble").
		SetCustomID(EditPreambleID).
		AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.TextInput{
				CustomID:    "source",
				Label:       "Preamble",
				Style:       discordgo.TextInputParagraph,
				Placeholder: `\usepackage[french]{babel}`,
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
		utils.SendAlert("commands/profile.go - Sending modal to edit preamble", err.Error(), "discord_id", u.ID)
	}
}

func OnProfileModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	submitData := i.ModalSubmitData()
	if submitData.CustomID != EditPreambleID {
		utils.SendDebug("commands/profile.go - not a profile modal ID")
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
		utils.SendAlert("commands/profile.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/profile.go - Sending error getting nerd", err.Error())
		}
		return
	}

	val := submitData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	if strings.Contains(val, `\documentclass`) {
		if err = resp.SetMessage("You can't use `\\documentclass`").Send(); err != nil {
			utils.SendAlert("commands/profile.go - Sending error getting document class", err.Error())
		}
		return
	}
	utils.SendDebug("Updating nerd's preamble", "discord_id", u.ID)
	nerd.Preamble = val
	err = nerd.Save()
	if err != nil {
		utils.SendAlert("commands/profile.go - Saving preamble", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while saving your profile. Please report the bug.").Send(); err != nil {
			utils.SendAlert("commands/profile.go - Sending error getting nerd", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Preamble saved").Send(); err != nil {
		utils.SendAlert("commands/profile.go - Sending success", err.Error())
	}
}

func Profile(_ *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	var u *discordgo.User
	if i.User == nil {
		u = i.Member.User
	} else {
		u = i.User
	}
	nerd, err := data.GetNerd(u.ID)
	if err != nil {
		utils.SendAlert("commands/profile.go - Getting nerd", err.Error(), "discord_id", u.ID)
		if err = resp.SetMessage("Error while getting your profile. Please report.").Send(); err != nil {
			utils.SendAlert("commands/profile.go - Getting nerd error", err.Error(), "discord_id", u.ID)
		}
		return
	}
	if len(nerd.Preamble) == 0 {
		nerd.Preamble, err = getDefaultPreamble()
		if err != nil {
			if err = resp.SetMessage("An error occurred. Please report the bug.").Send(); err != nil {
				utils.SendAlert("commands/profile.go - Sending error occurred while parsing template", err.Error())
			}
			return
		}
	}
	err = resp.AddEmbed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's nerd profile", u.Username),
		Description: fmt.Sprintf("Your preamble:\n```tex\n%s\n```", nerd.Preamble),
		Color:       0,
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "Edit preamble",
			Style:    discordgo.PrimaryButton,
			Disabled: false,
			Emoji:    nil,
			CustomID: EditPreambleID,
		},
	}}).Send()
	if err != nil {
		utils.SendAlert("commands/profile.go - Sending profile", err.Error(), "discord_id", u.ID)
	}
}

func getDefaultPreamble() (string, error) {
	if len(defaultPreamble) == 0 {
		t, err := template.ParseFiles(defaultPreprocessingOptions.TemplateFile)
		if err != nil {
			utils.SendAlert(
				"commands/profile.go - Parsing template file", err.Error(),
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
