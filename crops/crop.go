package crops

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Crop struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	FarmID   int     `json:"farmId"`
}

type Seller struct {
	FarmID   int     `json:"farmId"`
	FarmName string  `json:"farmName"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	Location string  `json:"location"`
	Rating   float64 `json:"rating"`
}

var (
	crops           []Crop
	sellersByCropID map[int][]Seller
)

func init() {
	crops = []Crop{
		{ID: 1, Name: "Apple", Category: "Fruits", Price: 120.0, Stock: 100, FarmID: 101},
		{ID: 2, Name: "Banana", Category: "Fruits", Price: 40.0, Stock: 80, FarmID: 102},
		{ID: 3, Name: "Carrot", Category: "Vegetables", Price: 30.0, Stock: 50, FarmID: 103},
		{ID: 4, Name: "Potato", Category: "Vegetables", Price: 25.0, Stock: 200, FarmID: 104},
		{ID: 5, Name: "Milk", Category: "Dairy", Price: 50.0, Stock: 150, FarmID: 105},
		{ID: 6, Name: "Paneer", Category: "Dairy", Price: 250.0, Stock: 60, FarmID: 106},
		{ID: 7, Name: "Wheat", Category: "Grains", Price: 250.0, Stock: 60, FarmID: 107},
		{ID: 8, Name: "Urad Dal", Category: "Millets/Pulses", Price: 90.0, Stock: 90, FarmID: 108},
		{ID: 9, Name: "Bajra", Category: "Millets/Pulses", Price: 40.0, Stock: 70, FarmID: 109},
	}

	sellersByCropID = map[int][]Seller{
		1: {
			{FarmID: 101, FarmName: "Green Valley Farms", Price: 120.0, Stock: 100, Location: "Haryana", Rating: 4.5},
			{FarmID: 102, FarmName: "Sunrise Orchards", Price: 115.0, Stock: 80, Location: "Punjab", Rating: 4.2},
		},
		2: {
			{FarmID: 103, FarmName: "Tropical Farm Co", Price: 38.0, Stock: 150, Location: "Kerala", Rating: 4.7},
		},
	}
}

// Handlers

func GetCrops(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	category := r.URL.Query().Get("category")
	var filtered []Crop

	if category == "" {
		filtered = crops
	} else {
		for _, crop := range crops {
			if crop.Category == category {
				filtered = append(filtered, crop)
			}
		}
	}

	respondWithJSON(w, filtered)
}

func GetCropByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid crop ID", http.StatusBadRequest)
		return
	}

	for _, crop := range crops {
		if crop.ID == id {
			respondWithJSON(w, crop)
			return
		}
	}

	http.NotFound(w, r)
}

func GetCropSellers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid crop ID", http.StatusBadRequest)
		return
	}

	sellers, exists := sellersByCropID[id]
	if !exists {
		http.NotFound(w, r)
		return
	}

	respondWithJSON(w, sellers)
}

// Utility

func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
