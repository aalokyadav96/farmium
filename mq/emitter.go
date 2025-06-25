package mq

import (
	"fmt"
)

type Index struct {
	EntityType string `json:"entity_type"`
	Method     string `json:"method"`
	EntityId   string `json:"entity_id"`
	ItemId     string `json:"item_id"`
	ItemType   string `json:"item_type"`
}

// Emit event by sending JSON data to QUIC server
func Emit(eventName string, content Index) error {
	fmt.Println(eventName, "emitted", content)
	return nil
}

// Notify event (placeholder function)
func Notify(eventName string, content Index) error {
	fmt.Println(eventName, "Notified")
	return nil
}
