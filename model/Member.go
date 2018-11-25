package model

import (
	"github.com/jinzhu/gorm"
)

type Member struct {
	gorm.Model
	Name string		`json:"name" gorm:"not null"`
	Email string	`json:"email" gorm:"not null"`
	GithubId string	`json:"github_id"`
	TrelloId string	`json:"trello_id"`
}