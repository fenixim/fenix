package models

import (
	"encoding/json"
	"log"
)

type JSONModel interface {
	ToJSON() *MessageType
}

type MessageType struct {
	MessageType string `json:"type"`
	Data        string `json:"data"`
}

type BadFormat struct {
	Message string `json:"data"`
}

func (b BadFormat) ToJSON() *MessageType {
	data, err := json.Marshal(b)
	if err != nil {
		log.Printf("Error marshalling JSON, %v", err)
	}

	return &MessageType{MessageType: "BadFormat", Data: string(data[:])}
}
