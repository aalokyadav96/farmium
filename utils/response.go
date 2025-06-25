package utils

import (
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

type M map[string]interface{}

// func MergeCarts(a, b models.CartData) models.CartData {
// 	return models.CartData{
// 		Crops:       mergeItems(a.Crops, b.Crops, true),
// 		Merchandise: mergeItems(a.Merchandise, b.Merchandise, false),
// 		Tickets:     mergeItems(a.Tickets, b.Tickets, false),
// 		Menu:        mergeItems(a.Menu, b.Menu, false),
// 	}
// }

// func mergeItems(a, b []models.CartItem, includeFarm bool) []models.CartItem {
// 	merged := map[string]models.CartItem{}
// 	for _, item := range a {
// 		key := item.Item
// 		if includeFarm {
// 			key += "__" + item.Farm
// 		}
// 		merged[key] = item
// 	}
// 	for _, item := range b {
// 		key := item.Item
// 		if includeFarm {
// 			key += "__" + item.Farm
// 		}
// 		if existing, ok := merged[key]; ok {
// 			existing.Quantity += item.Quantity
// 			merged[key] = existing
// 		} else {
// 			merged[key] = item
// 		}
// 	}
// 	var result []models.CartItem
// 	for _, v := range merged {
// 		result = append(result, v)
// 	}
// 	return result
// }
