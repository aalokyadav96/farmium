package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactInfo struct {
	Phone   string `bson:"phone,omitempty" json:"phone,omitempty"`
	Email   string `bson:"email,omitempty" json:"email,omitempty"`
	Website string `bson:"website,omitempty" json:"website,omitempty"`
}

type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"     json:"id"`
	UserID    primitive.ObjectID `bson:"userId"            json:"userId"`
	Rating    int                `bson:"rating"            json:"rating"`
	Comment   string             `bson:"comment,omitempty" json:"comment,omitempty"`
	CreatedAt time.Time          `bson:"createdAt"         json:"createdAt"`
}

type Farm struct {
	FarmID primitive.ObjectID `bson:"_id,omitempty"         json:"id"`
	// ID                 primitive.ObjectID `bson:"_id,omitempty"         json:"id"`
	Name               string      `bson:"name"                  json:"name"`
	Location           string      `bson:"location"              json:"location"`
	Latitude           float64     `bson:"latitude,omitempty"    json:"latitude,omitempty"`
	Longitude          float64     `bson:"longitude,omitempty"   json:"longitude,omitempty"`
	Description        string      `bson:"description,omitempty" json:"description,omitempty"`
	Owner              string      `bson:"owner"                 json:"owner"`
	ContactInfo        ContactInfo `bson:"contactInfo,omitempty" json:"contactInfo,omitempty"`
	AvailabilityTiming string      `bson:"availabilityTiming,omitempty" json:"availabilityTiming,omitempty"`
	Tags               []string    `bson:"tags,omitempty"        json:"tags,omitempty"`
	Photo              string      `bson:"photo,omitempty"       json:"photo,omitempty"`
	Crops              []Crop      `bson:"crops" json:"crops,omitempty"` // loaded via lookup or separate query
	Media              []string    `bson:"media,omitempty"       json:"media,omitempty"`
	AvgRating          float64     `bson:"avgRating,omitempty"   json:"avgRating,omitempty"`
	ReviewCount        int         `bson:"reviewCount,omitempty" json:"reviewCount,omitempty"`
	FavoritesCount     int64       `bson:"favoritesCount,omitempty" json:"favoritesCount,omitempty"`
	CreatedBy          string      `bson:"createdBy"             json:"createdBy"`
	CreatedAt          time.Time   `bson:"createdAt"             json:"createdAt"`
	UpdatedAt          time.Time   `bson:"updatedAt"             json:"updatedAt"`
	Contact            string      `json:"contact"`
}

// type Farm struct {
// 	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	Name               string             `json:"name"`
// 	Location           string             `json:"location"`
// 	Description        string             `json:"description"`
// 	Owner              string             `json:"owner"`
// 	Contact            string             `json:"contact"`
// 	AvailabilityTiming string             `json:"availabilityTiming"`
// 	Crops              []Crop             `json:"crops"`
// 	CreatedAt          time.Time          `json:"createdAt"`
// 	UpdatedAt          time.Time          `json:"updatedAt"`
// 	Photo              string             `json:"photo,omitempty" bson:"photo,omitempty"`
// 	Tags               []string           `bson:"tags,omitempty" json:"tags,omitempty"`
// }

type PricePoint struct {
	Date  time.Time `json:"date" bson:"date"`
	Price float64   `json:"price" bson:"price"`
}

type Crop struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `json:"name"`
	Price        float64            `json:"price"`
	Quantity     int                `json:"quantity"`
	Unit         string             `json:"unit"`
	ImageURL     string             `json:"imageUrl,omitempty"`
	Notes        string             `json:"notes,omitempty"`
	Category     string             `json:"category,omitempty"`
	CatalogueId  string             `json:"catalogueid,omitempty"`
	Featured     bool               `json:"featured,omitempty"`
	OutOfStock   bool               `json:"outOfStock,omitempty"`
	HarvestDate  *time.Time         `json:"harvestDate,omitempty"`
	ExpiryDate   *time.Time         `json:"expiryDate,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	PriceHistory []PricePoint       `json:"priceHistory,omitempty"`
	FieldPlot    string             `json:"fieldPlot,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
	FarmID       primitive.ObjectID `bson:"farmId,omitempty" json:"farmId,omitempty"`
}

type FarmOrder struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	UserID          primitive.ObjectID `bson:"userId"         json:"userId"`
	FarmID          primitive.ObjectID `bson:"farmId"         json:"farmId"`
	CropID          primitive.ObjectID `bson:"cropId"         json:"cropId"`
	Quantity        int                `bson:"quantity"       json:"quantity"`
	PriceAtPurchase float64            `bson:"priceAtPurchase" json:"priceAtPurchase"`
	BoughtAt        time.Time          `bson:"boughtAt"       json:"boughtAt"`
}

type CropCatalogueItem struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	ImageURL   string `json:"imageUrl"`
	Stock      int    `json:"stock"`
	Unit       string `json:"unit"`
	Featured   bool   `json:"featured"`
	PriceRange []int  `json:"priceRange,omitempty"`
}
