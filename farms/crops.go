package farms

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"naevis/db"
	"naevis/globals"
	"naevis/models"
	"naevis/rdx"
	"naevis/utils"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(globals.UserIDKey).(string)
	return userID, ok
}

func handleImageUpload(r *http.Request, fieldName, dir string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	fullDir := "./static/uploads/" + dir
	path := fullDir + "/" + filename
	os.MkdirAll(fullDir, os.ModePerm)

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	io.Copy(out, file)
	return "/uploads/" + dir + "/" + filename, nil
}

func parseCropForm(r *http.Request) models.Crop {
	crop := models.Crop{
		ID:         primitive.NewObjectID(),
		Name:       r.FormValue("name"),
		Price:      utils.ParseFloat(r.FormValue("price")),
		Quantity:   utils.ParseInt(r.FormValue("quantity")),
		Unit:       r.FormValue("unit"),
		Notes:      r.FormValue("notes"),
		Category:   r.FormValue("category"),
		Featured:   r.FormValue("featured") == "true",
		OutOfStock: r.FormValue("outOfStock") == "true",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if d := utils.ParseDate(r.FormValue("harvestDate")); d != nil {
		crop.HarvestDate = d
	}
	if d := utils.ParseDate(r.FormValue("expiryDate")); d != nil {
		crop.ExpiryDate = d
	}
	return crop
}

func AddCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	farmID, err := primitive.ObjectIDFromHex(ps.ByName("id"))
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid farm ID"})
		return
	}

	if _, ok := getUserIDFromContext(r); !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid form"})
		return
	}

	name := r.FormValue("name")
	if name == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Name is required"})
		return
	}

	crop := parseCropForm(r)
	crop.FarmID = farmID

	if imageURL, err := handleImageUpload(r, "image", "crops"); err == nil {
		crop.ImageURL = imageURL
	}

	_, err = db.CropsCollection.InsertOne(context.Background(), crop)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Insert failed"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "cropId": crop.ID.Hex()})
}

func EditCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
		return
	}

	if _, ok := getUserIDFromContext(r); !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	r.ParseMultipartForm(10 << 20)
	update := bson.M{
		"name":       r.FormValue("name"),
		"unit":       r.FormValue("unit"),
		"price":      utils.ParseFloat(r.FormValue("price")),
		"quantity":   utils.ParseInt(r.FormValue("quantity")),
		"notes":      r.FormValue("notes"),
		"category":   r.FormValue("category"),
		"featured":   r.FormValue("featured") == "true",
		"outOfStock": r.FormValue("outOfStock") == "true",
		"updatedAt":  time.Now(),
	}

	if d := utils.ParseDate(r.FormValue("harvestDate")); d != nil {
		update["harvestDate"] = d
	}
	if d := utils.ParseDate(r.FormValue("expiryDate")); d != nil {
		update["expiryDate"] = d
	}

	if imageURL, err := handleImageUpload(r, "image", "crops"); err == nil {
		update["imageUrl"] = imageURL
	}

	_, err = db.CropsCollection.UpdateOne(context.Background(), bson.M{"_id": cropID}, bson.M{"$set": update})
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true})
}

func DeleteCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
		return
	}

	if _, ok := getUserIDFromContext(r); !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	res, err := db.CropsCollection.DeleteOne(context.Background(), bson.M{"_id": cropID})
	if err != nil || res.DeletedCount == 0 {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to delete crop"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true})
}

func GetFilteredCrops(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := bson.M{}
	params := r.URL.Query()

	if category := params.Get("category"); category != "" {
		query["category"] = category
	}
	if region := params.Get("region"); region != "" {
		query["farmLocation"] = region
	}
	if params.Get("inStock") == "true" {
		query["quantity"] = bson.M{"$gt": 0}
	}

	price := bson.M{}
	if min := utils.ParseFloat(params.Get("minPrice")); min > 0 {
		price["$gte"] = min
	}
	if max := utils.ParseFloat(params.Get("maxPrice")); max > 0 {
		price["$lte"] = max
	}
	if len(price) > 0 {
		query["price"] = price
	}

	cursor, err := db.CropsCollection.Find(context.Background(), query)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
		return
	}
	var crops []models.Crop
	if err = cursor.All(context.Background(), &crops); err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
}

func GetPreCropCatalogue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const redisKey = "crop_catalogue"
	var crops []models.CropCatalogueItem

	if val, err := rdx.Conn.Get(ctx, redisKey).Result(); err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &crops); err == nil {
			utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
			return
		}
	}

	cursor, err := db.CatalogueCollection.Find(ctx, bson.M{})
	if err == nil {
		if err := cursor.All(ctx, &crops); err == nil && len(crops) > 0 {
			if jsonBytes, err := json.Marshal(crops); err == nil {
				_ = rdx.Conn.Set(ctx, redisKey, jsonBytes, 2*time.Hour).Err()
			}
			utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
			return
		}
	}

	file, err := os.Open("data/pre_crop_catalogue.csv")
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to retrieve catalogue"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Invalid CSV"})
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) != len(headers) {
			continue
		}

		entry := models.CropCatalogueItem{}
		for i, field := range headers {
			switch strings.ToLower(field) {
			case "name":
				entry.Name = record[i]
			case "category":
				entry.Category = record[i]
			case "imageurl":
				entry.ImageURL = record[i]
			case "stock":
				entry.Stock, _ = strconv.Atoi(record[i])
			case "unit":
				entry.Unit = record[i]
			case "featured":
				entry.Featured = strings.ToLower(record[i]) == "true"
			case "pricerange":
				parts := strings.Split(record[i], "-")
				if len(parts) == 2 {
					min, _ := strconv.Atoi(parts[0])
					max, _ := strconv.Atoi(parts[1])
					entry.PriceRange = []int{min, max}
				}
			}
		}
		crops = append(crops, entry)
	}

	if jsonBytes, err := json.Marshal(crops); err == nil {
		_ = rdx.Conn.Set(ctx, redisKey, jsonBytes, 2*time.Hour).Err()
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
}

func GetCropCatalogue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.CropsCollection.Find(ctx, bson.M{})
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to fetch crop catalogue"})
		return
	}
	defer cursor.Close(ctx)

	var allCrops []models.Crop
	if err := cursor.All(ctx, &allCrops); err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to decode crops"})
		return
	}

	seen := make(map[string]bool)
	uniqueCrops := []models.Crop{}
	for _, crop := range allCrops {
		key := strings.ToLower(crop.Name + crop.CatalogueId)
		if !seen[key] {
			seen[key] = true
			uniqueCrops = append(uniqueCrops, crop)
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": uniqueCrops})
}

// func GetCropTypes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	cropTypes, err := db.CropsCollection.Distinct(context.Background(), "name", bson.M{})
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{
// 			"success": false,
// 			"error":   "Failed to retrieve crop types",
// 		})
// 		return
// 	}

//		utils.RespondWithJSON(w, http.StatusOK, utils.M{
//			"success": true,
//			"types":   cropTypes,
//		})
//	}
func GetCropTypes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{}}}, // No filters, return all types

		{{
			Key: "$group", Value: bson.M{
				"_id":      "$name",
				"minPrice": bson.M{"$min": "$price"},
				"maxPrice": bson.M{"$max": "$price"},
				"availableCount": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$gt": []interface{}{"$quantity", 0}}, 1, 0,
						},
					},
				},
				"imageUrl": bson.M{"$first": "$imageUrl"},
				"unit":     bson.M{"$first": "$unit"},
			},
		}},
		{{
			Key: "$project", Value: bson.M{
				"name":           "$_id",
				"minPrice":       1,
				"maxPrice":       1,
				"availableCount": 1,
				"imageUrl":       1,
				"unit":           1,
				"_id":            0,
			},
		}},
		{{Key: "$sort", Value: bson.M{"name": 1}}},
	}

	cursor, err := db.CropsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
		return
	}
	var cropTypes []bson.M
	if err := cursor.All(context.Background(), &cropTypes); err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.M{
		"success":   true,
		"cropTypes": cropTypes,
	})
}

// func AddCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	farmID, err := primitive.ObjectIDFromHex(ps.ByName("id"))
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid farm ID"})
// 		return
// 	}

// 	// Retrieve user ID
// 	requestingUserID, ok := r.Context().Value(globals.UserIDKey).(string)
// 	if !ok {
// 		http.Error(w, "Invalid user", http.StatusBadRequest)
// 		return
// 	}
// 	_ = requestingUserID

// 	if err := r.ParseMultipartForm(10 << 20); err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid form"})
// 		return
// 	}

// 	name := r.FormValue("name")
// 	if name == "" {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Name is required"})
// 		return
// 	}

// 	crop := models.Crop{
// 		ID:       primitive.NewObjectID(),
// 		FarmID:   farmID,
// 		Name:     name,
// 		Price:    utils.ParseFloat(r.FormValue("price")),
// 		Quantity: utils.ParseInt(r.FormValue("quantity")),
// 		Unit:     r.FormValue("unit"),
// 		Notes:    r.FormValue("notes"),
// 		// FieldPlot:  r.FormValue("fieldPlot"),
// 		Category:   r.FormValue("category"),
// 		Featured:   r.FormValue("featured") == "true",
// 		OutOfStock: r.FormValue("outOfStock") == "true",
// 		CreatedAt:  time.Now(),
// 		UpdatedAt:  time.Now(),
// 	}

// 	if d := utils.ParseDate(r.FormValue("harvestDate")); d != nil {
// 		crop.HarvestDate = d
// 	}
// 	if d := utils.ParseDate(r.FormValue("expiryDate")); d != nil {
// 		crop.ExpiryDate = d
// 	}

// 	if file, header, err := r.FormFile("image"); err == nil {
// 		defer file.Close()
// 		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
// 		path := "./static/uploads/crops/" + filename
// 		os.MkdirAll("./static/uploads/crops", os.ModePerm)
// 		out, err := os.Create(path)
// 		if err == nil {
// 			defer out.Close()
// 			io.Copy(out, file)
// 			crop.ImageURL = "/uploads/crops/" + filename
// 		}
// 	}

// 	_, err = db.CropsCollection.InsertOne(context.Background(), crop)
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Insert failed"})
// 		return
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "cropId": crop.ID.Hex()})
// }

// func EditCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
// 		return
// 	}

// 	// Retrieve user ID
// 	requestingUserID, ok := r.Context().Value(globals.UserIDKey).(string)
// 	if !ok {
// 		http.Error(w, "Invalid user", http.StatusBadRequest)
// 		return
// 	}
// 	_ = requestingUserID

// 	r.ParseMultipartForm(10 << 20)

// 	updateFields := bson.M{
// 		"name":     r.FormValue("name"),
// 		"unit":     r.FormValue("unit"),
// 		"price":    utils.ParseFloat(r.FormValue("price")),
// 		"quantity": utils.ParseInt(r.FormValue("quantity")),
// 		"notes":    r.FormValue("notes"),
// 		// "fieldPlot":  r.FormValue("fieldPlot"),
// 		"featured":   r.FormValue("featured") == "true",
// 		"outOfStock": r.FormValue("outOfStock") == "true",
// 		"updatedAt":  time.Now(),
// 	}

// 	if d := utils.ParseDate(r.FormValue("harvestDate")); d != nil {
// 		updateFields["harvestDate"] = d
// 	}
// 	if d := utils.ParseDate(r.FormValue("expiryDate")); d != nil {
// 		updateFields["expiryDate"] = d
// 	}

// 	if file, header, err := r.FormFile("image"); err == nil {
// 		defer file.Close()
// 		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
// 		path := "./static/uploads/crops/" + filename
// 		os.MkdirAll("./static/uploads/crops", os.ModePerm)
// 		out, err := os.Create(path)
// 		if err == nil {
// 			defer out.Close()
// 			io.Copy(out, file)
// 			updateFields["imageUrl"] = "/static/uploads/crops/" + filename
// 		}
// 	}

// 	filter := bson.M{"_id": cropID}
// 	_, err = db.CropsCollection.UpdateOne(context.Background(), filter, bson.M{"$set": updateFields})
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
// 		return
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true})
// }
// func DeleteCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
// 		return
// 	}

// 	// Retrieve user ID
// 	requestingUserID, ok := r.Context().Value(globals.UserIDKey).(string)
// 	if !ok {
// 		http.Error(w, "Invalid user", http.StatusBadRequest)
// 		return
// 	}
// 	_ = requestingUserID

// 	res, err := db.CropsCollection.DeleteOne(context.Background(), bson.M{"_id": cropID})
// 	if err != nil || res.DeletedCount == 0 {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to delete crop"})
// 		return
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true})
// }
// func GetFilteredCrops(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	query := bson.M{}
// 	category := r.URL.Query().Get("category")
// 	if category != "" {
// 		query["category"] = category
// 	}
// 	if region := r.URL.Query().Get("region"); region != "" {
// 		query["farmLocation"] = region
// 	}
// 	if r.URL.Query().Get("inStock") == "true" {
// 		query["quantity"] = bson.M{"$gt": 0}
// 	}
// 	if min := utils.ParseFloat(r.URL.Query().Get("minPrice")); min > 0 {
// 		query["price"] = bson.M{"$gte": min}
// 	}
// 	if max := utils.ParseFloat(r.URL.Query().Get("maxPrice")); max > 0 {
// 		if p, ok := query["price"].(bson.M); ok {
// 			p["$lte"] = max
// 		} else {
// 			query["price"] = bson.M{"$lte": max}
// 		}
// 	}
// 	// TODO: implement geolocation filter if farm coords are stored

// 	cursor, err := db.CropsCollection.Find(context.Background(), query)
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
// 		return
// 	}
// 	var crops []models.Crop
// 	if err = cursor.All(context.Background(), &crops); err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false})
// 		return
// 	}
// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
// }
// func GetPreCropCatalogue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	const redisKey = "crop_catalogue"

// 	// Try Redis first
// 	val, err := rdx.Conn.Get(ctx, redisKey).Result()
// 	if err == nil && val != "" {
// 		// Cache hit
// 		var cached []models.CropCatalogueItem
// 		if err := json.Unmarshal([]byte(val), &cached); err == nil {
// 			utils.RespondWithJSON(w, http.StatusOK, utils.M{
// 				"success": true,
// 				"crops":   cached,
// 			})
// 			return
// 		}
// 	}

// 	// Fall back to MongoDB
// 	cursor, err := db.CatalogueCollection.Find(ctx, bson.M{})
// 	if err == nil {
// 		var fromMongo []models.CropCatalogueItem
// 		if err := cursor.All(ctx, &fromMongo); err == nil && len(fromMongo) > 0 {
// 			// Cache it back to Redis
// 			if jsonBytes, err := json.Marshal(fromMongo); err == nil {
// 				_ = rdx.Conn.Set(ctx, redisKey, string(jsonBytes), time.Hour*2).Err()
// 			}
// 			utils.RespondWithJSON(w, http.StatusOK, utils.M{
// 				"success": true,
// 				"crops":   fromMongo,
// 			})
// 			return
// 		}
// 	}

// 	// Final fallback to CSV
// 	file, err := os.Open("data/pre_crop_catalogue.csv")
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{
// 			"success": false,
// 			"message": "Failed to retrieve catalogue from all sources",
// 		})
// 		return
// 	}
// 	defer file.Close()

// 	var crops []models.CropCatalogueItem
// 	rdr := csv.NewReader(file)
// 	headers, err := rdr.Read()
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{
// 			"success": false,
// 			"message": "Invalid CSV format",
// 		})
// 		return
// 	}

// 	for {
// 		record, err := rdr.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil || len(record) != len(headers) {
// 			continue
// 		}

// 		entry := models.CropCatalogueItem{}
// 		for i, field := range headers {
// 			switch strings.ToLower(field) {
// 			case "name":
// 				entry.Name = record[i]
// 			case "category":
// 				entry.Category = record[i]
// 			case "imageurl":
// 				entry.ImageURL = record[i]
// 			case "stock":
// 				entry.Stock, _ = strconv.Atoi(record[i])
// 			case "unit":
// 				entry.Unit = record[i]
// 			case "featured":
// 				entry.Featured = strings.ToLower(record[i]) == "true"
// 			case "pricerange":
// 				// Expects something like "45-65"
// 				rangeParts := strings.Split(record[i], "-")
// 				if len(rangeParts) == 2 {
// 					min, _ := strconv.Atoi(rangeParts[0])
// 					max, _ := strconv.Atoi(rangeParts[1])
// 					entry.PriceRange = []int{min, max}
// 				}
// 			}
// 		}
// 		crops = append(crops, entry)
// 	}

// 	// Cache the parsed result to Redis
// 	if jsonBytes, err := json.Marshal(crops); err == nil {
// 		_ = rdx.Conn.Set(ctx, redisKey, string(jsonBytes), time.Hour*2).Err()
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{
// 		"success": true,
// 		"crops":   crops,
// 	})
// }

// func GetCropCatalogue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cursor, err := db.CropsCollection.Find(ctx, bson.M{})
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{
// 			"success": false,
// 			"message": "Failed to fetch crop catalogue",
// 		})
// 		return
// 	}
// 	defer cursor.Close(ctx)

// 	var allCrops []models.Crop
// 	if err := cursor.All(ctx, &allCrops); err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{
// 			"success": false,
// 			"message": "Failed to decode crops",
// 		})
// 		return
// 	}

// 	type UniqueKey struct {
// 		Name        string
// 		CatalogueId string
// 	}

// 	seen := make(map[UniqueKey]bool)
// 	uniqueCrops := []models.Crop{}

// 	for _, crop := range allCrops {
// 		key := UniqueKey{
// 			Name:        strings.ToLower(crop.Name),
// 			CatalogueId: strings.ToLower(crop.CatalogueId),
// 		}

// 		if !seen[key] {
// 			seen[key] = true
// 			uniqueCrops = append(uniqueCrops, crop)
// 		}
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{
// 		"success": true,
// 		"crops":   uniqueCrops,
// 	})
// }

// func GetCropCatalogue(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cursor, err := db.CropsCollection.Find(ctx, bson.M{})
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to fetch crops"})
// 		return
// 	}
// 	defer cursor.Close(ctx)

// 	var crops []models.Crop
// 	if err = cursor.All(ctx, &crops); err != nil {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to decode crops"})
// 		return
// 	}

// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true, "crops": crops})
// }
