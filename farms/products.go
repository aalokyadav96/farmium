package farms

import (
	"context"
	"encoding/json"
	"naevis/db"
	"naevis/models"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetItems(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query parameters
	itemType := r.URL.Query().Get("type")     // "product" or "tool"
	search := r.URL.Query().Get("search")     // search text
	category := r.URL.Query().Get("category") // filter by category
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int64(10)
	offset := int64(0)

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = int64(l)
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = int64(o)
	}

	// Build filter
	filter := bson.M{}
	if itemType != "" {
		filter["type"] = itemType
	}
	if category != "" {
		filter["category"] = category
	}
	if search != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(offset).
		SetSort(bson.M{"name": 1})

	cursor, err := db.ProductCollection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var items []models.Product
	if err := cursor.All(ctx, &items); err != nil {
		http.Error(w, "Failed to decode items", http.StatusInternalServerError)
		return
	}

	// Total count (optional)
	count, err := db.ProductCollection.CountDocuments(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to count items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
		"total": count,
	})
}

func GetProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.ProductCollection.Find(ctx, struct{}{})
	if err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		http.Error(w, "Failed to parse products", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

func GetTools(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.ProductCollection.Find(ctx, struct{}{})
	if err != nil {
		http.Error(w, "Failed to fetch tools", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var tools []models.Tool
	if err := cursor.All(ctx, &tools); err != nil {
		http.Error(w, "Failed to parse tools", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tools)
}

func CreateProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	createItem(w, r, "product")
}

func CreateTool(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	createItem(w, r, "tool")
}

func createItem(w http.ResponseWriter, r *http.Request, itemType string) {
	var item models.Product
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item.Type = itemType

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.ProductCollection.InsertOne(ctx, item)
	if err != nil {
		http.Error(w, "Failed to insert item", http.StatusInternalServerError)
		return
	}

	item.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(item)
}
func UpdateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	updateItem(w, r, ps, "product")
}

func UpdateTool(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	updateItem(w, r, ps, "tool")
}

func updateItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params, itemType string) {
	idParam := ps.ByName("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var item models.Product
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	item.Type = itemType
	update := bson.M{"$set": item}

	_, err = db.ProductCollection.UpdateByID(ctx, objID, update)
	if err != nil {
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
func DeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deleteItem(w, r, ps)
}

func DeleteTool(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	deleteItem(w, r, ps)
}

func deleteItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idParam := ps.ByName("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ProductCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
