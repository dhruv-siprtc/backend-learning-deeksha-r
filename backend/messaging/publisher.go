package messaging

import (
	"encoding/json"
	"log"
	"time"
)

type UserEvent struct {
	Event     string    `json:"event"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      UserData  `json:"data"`
}

type UserData struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func PublishUserEvent(
	paota *Paota,
	routingKey string,
	eventName string,
	userID int,
	name string,
	email string,
) error {

	event := UserEvent{
		Event:     eventName,
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		Data: UserData{
			UserID: userID,
			Name:   name,
			Email:  email,
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Use Paota's generic Publish method
	err = paota.Publish("user.events", routingKey, body)
	if err != nil {
		return err
	}

	log.Printf("📤 Event published: %s (%s)", eventName, routingKey)
	return nil
}
