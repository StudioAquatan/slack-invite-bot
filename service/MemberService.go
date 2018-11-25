package service

import (
	"github.com/StudioAquatan/slack-invite-bot/model"
	"github.com/jinzhu/gorm"
)

func CreateMember(db *gorm.DB,member model.Member)  {
	db.Create(&member)
}

func UpdateMember()  {}
