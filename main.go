package main

import (
	"github.com/StudioAquatan/slack-invite-bot/model"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
)


func main() {
	// Initialize DB
	InitDB()

	e := echo.New()

	// Routing
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/new/member/", registerMember)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}

func InitDB() {
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Printf("[ERROR] Invalid action was submitted: %s", err)
		panic("failed to connect database")

	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&model.Member{})
}

func registerMember(c echo.Context) (err error) {
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Printf("[ERROR] Invalid action was submitted: %s", err)
		panic("failed to connect database")

	}
	defer db.Close()

	m := new(model.Member)
	if err = c.Bind(m); err != nil {
		return
	}

	//service.CreateMember(db, *m)
	db.Create(&m)
	return c.JSON(http.StatusOK, m)
}
