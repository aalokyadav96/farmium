package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Ingredient struct {
	Name         string  `json:"name"`
	ItemID       string  `json:"itemId"`
	Type         string  `json:"type"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Alternatives []struct {
		Name   string `json:"name"`
		ItemID string `json:"itemId"`
		Type   string `json:"type"`
	} `json:"alternatives"`
}

type Recipe struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `json:"userId"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	PrepTime    string             `json:"prepTime"`
	Tags        []string           `json:"tags"`
	ImageURLs   []string           `json:"imageUrls"`
	Ingredients []Ingredient       `json:"ingredients"`
	Steps       []string           `json:"steps"`
	CreatedAt   int64              `json:"createdAt"`
	Views       int                `json:"views"`
}
