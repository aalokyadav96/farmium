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

var crops = []Crop{
	{1, "Apple", "Fruits", 120.0, 100, 101},
	{2, "Banana", "Fruits", 40.0, 80, 102},
	{3, "Carrot", "Vegetables", 30.0, 50, 103},
	{4, "Potato", "Vegetables", 25.0, 200, 104},
	{5, "Milk", "Dairy", 50.0, 150, 105},
	{6, "Paneer", "Dairy", 250.0, 60, 106},
	{7, "Urad Dal", "Millets/Pulses", 90.0, 90, 107},
	{8, "Bajra", "Millets/Pulses", 40.0, 70, 108},
}

func GetCrops(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// type Crop struct {
// 	ID       int     `json:"id"`
// 	Name     string  `json:"name"`
// 	Category string  `json:"category"`
// 	Price    float64 `json:"price"`
// 	Stock    int     `json:"stock"`
// 	FarmID   int     `json:"farmId"`
// }

type Seller struct {
	FarmID   int     `json:"farmId"`
	FarmName string  `json:"farmName"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	Location string  `json:"location"`
	Rating   float64 `json:"rating"`
}

// var crops = []Crop{
// 	{1, "Apple", "Fruits", 120.0, 100, 101},
// 	{2, "Banana", "Fruits", 40.0, 80, 102},
// 	{3, "Carrot", "Vegetables", 30.0, 50, 103},
// 	// Add more as needed
// }

var sellersByCropID = map[int][]Seller{
	1: {
		{101, "Green Valley Farms", 120.0, 100, "Haryana", 4.5},
		{102, "Sunrise Orchards", 115.0, 80, "Punjab", 4.2},
	},
	2: {
		{103, "Tropical Farm Co", 38.0, 150, "Kerala", 4.7},
	},
	// Add sellers for other crops
}

func GetCropByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crop ID", http.StatusBadRequest)
		return
	}

	for _, crop := range crops {
		if crop.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(crop)
			return
		}
	}

	http.NotFound(w, r)
}

func GetCropSellers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid crop ID", http.StatusBadRequest)
		return
	}

	sellers := sellersByCropID[id]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sellers)
}
