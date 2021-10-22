package db

import (
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	"log"
)

//SendLogInformation creates a new lg information in the database
func GetTempAnnouncement() []map[string]string {
	messages, err := db.QueryString("select * from `temp_announcement`")
	if err != nil {
		log.Println(err)
	}
	return messages
}
