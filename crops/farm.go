package crops

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Farm struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Rating      float64 `json:"rating"`
	Description string  `json:"description"`
}

// type Crop struct {
// 	ID       int     `json:"id"`
// 	Name     string  `json:"name"`
// 	Category string  `json:"category"`
// 	Price    float64 `json:"price"`
// 	Stock    int     `json:"stock"`
// 	FarmID   int     `json:"farmId"`
// }

// Sample data
var farms = []Farm{
	{101, "Green Valley Farms", "Haryana", 4.5, "Family-run organic farm"},
	{102, "Sunrise Orchards", "Punjab", 4.2, "Fresh produce straight from the orchards"},
}

var farmCrops = []Crop{
	{1, "Apple", "Fruits", 120.0, 100, 101},
	{2, "Carrot", "Vegetables", 35.0, 50, 101},
	{3, "Banana", "Fruits", 40.0, 80, 102},
}

// GET /api/farms/:id
func GetFarmByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid farm ID", http.StatusBadRequest)
		return
	}

	for _, farm := range farms {
		if farm.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(farm)
			return
		}
	}

	http.NotFound(w, r)
}

// GET /api/farms/:id/crops
func GetCropsByFarm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid farm ID", http.StatusBadRequest)
		return
	}

	var crops []Crop
	for _, crop := range farmCrops {
		if crop.FarmID == id {
			crops = append(crops, crop)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crops)
}
