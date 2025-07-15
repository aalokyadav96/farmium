package recipes

import (
	"context"
	"encoding/json"
	"io"
	"naevis/db"
	"naevis/models"
	"naevis/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get all recipes
func GetRecipes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := context.TODO()
	query := bson.M{}

	// --- Parse query params ---
	search := r.URL.Query().Get("search")
	ingredient := r.URL.Query().Get("ingredient")
	sortParam := r.URL.Query().Get("sort")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	// --- Search by title or description (case-insensitive) ---
	if search != "" {
		query["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// --- Filter by ingredient ---
	if ingredient != "" {
		query["ingredients.name"] = bson.M{"$regex": ingredient, "$options": "i"}
	}

	// --- Pagination ---
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// --- Sorting ---
	sort := bson.D{{Key: "createdAt", Value: -1}} // default: newest
	switch sortParam {
	case "oldest":
		sort = bson.D{{Key: "createdAt", Value: 1}}
	case "popular":
		sort = bson.D{{Key: "views", Value: -1}}
	}

	// --- Execute query ---
	opts := options.Find().
		SetSort(sort).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := db.RecipeCollection.Find(ctx, query, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var recipes []models.Recipe
	if err = cursor.All(ctx, &recipes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(recipes) == 0 {
		recipes = []models.Recipe{}
	}

	json.NewEncoder(w).Encode(recipes)
}

// func GetRecipes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	cursor, err := db.RecipeCollection.Find(context.TODO(), bson.M{})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer cursor.Close(context.TODO())

// 	var recipes []models.Recipe
// 	if err = cursor.All(context.TODO(), &recipes); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(recipes)
// }

// Get one recipe
func GetRecipe(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))
	var recipe models.Recipe
	err := db.RecipeCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&recipe)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(recipe)
}

// Create
func CreateRecipe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	recipe := models.Recipe{
		UserID:      r.FormValue("userId"),
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		PrepTime:    r.FormValue("prepTime"),
		Tags:        splitCSV(r.FormValue("tags")),
		Steps:       splitLines(r.FormValue("steps")),
		CreatedAt:   time.Now().Unix(),
		Views:       0,
	}

	uploadFolder := "./static/uploads"
	files := r.MultipartForm.File["imageUrls"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		savedName, err := utils.SaveFile(file, fileHeader, uploadFolder)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		recipe.ImageURLs = append(recipe.ImageURLs, savedName)
	}

	result, err := db.RecipeCollection.InsertOne(context.TODO(), recipe)
	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// Update
func UpdateRecipe(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	updates := bson.M{
		"title":       r.FormValue("title"),
		"description": r.FormValue("description"),
		"prepTime":    r.FormValue("prepTime"),
		"tags":        splitCSV(r.FormValue("tags")),
		"steps":       splitLines(r.FormValue("steps")),
		// add additional fields as needed
	}

	// Handle new image uploads
	files := r.MultipartForm.File["imageUrls"]
	var imagePaths []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		filename := "uploads/" + fileHeader.Filename
		dst, err := os.Create(filename)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}

		imagePaths = append(imagePaths, filename)
	}
	if len(imagePaths) > 0 {
		updates["imageUrls"] = imagePaths
	}

	_, err := db.RecipeCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"status":"updated"}`))
}

// Delete
func DeleteRecipe(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := primitive.ObjectIDFromHex(ps.ByName("id"))
	_, err := db.RecipeCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"status":"deleted"}`))
}

func GetRecipeTags(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := context.TODO()

	// Use MongoDB aggregation to extract unique tags
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$tags"}},
		{{Key: "$group", Value: bson.M{
			"_id":  nil,
			"tags": bson.M{"$addToSet": "$tags"},
		}}},
	}

	cursor, err := db.RecipeCollection.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var result []struct {
		Tags []string `bson:"tags"`
	}
	if err := cursor.All(ctx, &result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(result) > 0 {
		json.NewEncoder(w).Encode(result[0].Tags)
	} else {
		json.NewEncoder(w).Encode([]string{})
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
