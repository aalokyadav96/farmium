package home

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// GetFarmHomeContent handles all of the dashboard endpoints under /home/:apiRoute
// func GetFarmHomeContent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
func GetHomeContent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	apiRoute := strings.ToLower(ps.ByName("apiRoute"))

	var (
		data interface{}
		err  error
	)

	switch apiRoute {
	case "farms":
		data, err = getTopFarms()
	case "categories":
		data, err = getFarmCategories()
	case "offers":
		data, err = getSpecialOffers()
	case "blogs":
		data, err = getBlogPosts()
	case "seasonal-tips":
		data, err = getSeasonalTips()
	case "locations":
		data, err = getFarmLocations()
	default:
		http.Error(w, "Invalid API route", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getTopFarms returns a list of top farms
func getTopFarms() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"id": 201, "name": "Green Valley Farms", "rating": 4.8, "region": "Haryana"},
		{"id": 202, "name": "Sunrise Orchards", "rating": 4.6, "region": "Punjab"},
		{"id": 203, "name": "Riverbank Ranch", "rating": 4.7, "region": "Uttar Pradesh"},
	}, nil
}

// getFarmCategories returns the categories of farms/products
func getFarmCategories() ([]string, error) {
	return []string{
		"Fruits",
		"Vegetables",
		"Dairy",
		"Grains",
		"Millets/Pulses",
	}, nil
}

// getSpecialOffers returns current offers running on the platform
func getSpecialOffers() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"id":     301,
			"title":  "Buy 5 kg Mangoes, get 1 kg free",
			"endsAt": "2025-07-15",
		},
		{
			"id":     302,
			"title":  "20% off on all Dairy products",
			"endsAt": "2025-06-30",
		},
	}, nil
}

// getBlogPosts returns the latest blog posts
func getBlogPosts() ([]map[string]string, error) {
	return []map[string]string{
		{"title": "Sustainable Farming Practices", "link": "/blogs/sustainable-farming"},
		{"title": "How to Store Grains Safely", "link": "/blogs/grain-storage"},
	}, nil
}

// getSeasonalTips returns a list of seasonal farming tips
func getSeasonalTips() ([]string, error) {
	return []string{
		"🌾 Time to sow wheat in North India",
		"🍅 Tomatoes thrive in warm afternoons",
		"🥬 Use shade nets for spinach during peak sun",
	}, nil
}

// getFarmLocations returns basic location info for mapping
func getFarmLocations() ([]map[string]string, error) {
	return []map[string]string{
		{"name": "Green Valley Farms", "region": "Haryana"},
		{"name": "Sunrise Orchards", "region": "Punjab"},
		{"name": "Riverbank Ranch", "region": "Uttar Pradesh"},
	}, nil
}
