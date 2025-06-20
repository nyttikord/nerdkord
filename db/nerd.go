package db

import "github.com/anhgelus/gokord"

type Nerd struct {
	ID        uint `gorm:"primarykey"`
	DiscordID string
	Preamble  string
}

func GetNerd(id string) (*Nerd, error) {
	n := Nerd{}
	n.DiscordID = id
	if err := n.Load(); err != nil {
		return nil, err
	}
	return &n, nil
}

func (n *Nerd) Load() error {
	return gokord.DB.Where("discord_id = ?", n.DiscordID).FirstOrCreate(n).Error
}

func (n *Nerd) Save() error {
	return gokord.DB.Save(n).Error
}
