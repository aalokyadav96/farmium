package comments

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"naevis/db"
	"naevis/middleware"
	"naevis/models"
	"naevis/utils"
)

func CreateComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("ok")
	entityType := ps.ByName("entitytype")
	entityID := ps.ByName("entityid")

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		log.Printf("JWT validation error: %v", err) // Log the error for debugging
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comment := models.Comment{
		EntityType: entityType,
		EntityID:   entityID,
		CreatedBy:  claims.UserID,
		Content:    body.Content,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	res, err := db.CommentsCollection.InsertOne(context.TODO(), comment)
	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}
	comment.ID = res.InsertedID.(primitive.ObjectID).Hex()

	utils.RespondWithJSON(w, http.StatusOK, comment)
}

func GetComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entityType := ps.ByName("entitytype")
	entityID := ps.ByName("entityid")

	filter := bson.M{"entity_type": entityType, "entity_id": entityID}
	cursor, err := db.CommentsCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "DB find failed", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var comments []models.Comment
	if err := cursor.All(context.TODO(), &comments); err != nil {
		http.Error(w, "Cursor decode failed", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, comments)
}

func GetComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	commentID := ps.ByName("entitytype")

	filter := bson.M{"_id": commentID}
	cursor, err := db.CommentsCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "DB find failed", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var comment models.Comment
	if err := cursor.All(context.TODO(), &comment); err != nil {
		http.Error(w, "Cursor decode failed", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, comment)
}

func UpdateComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	commentID := ps.ByName("commentid")

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if comment exists and belongs to user
	var existing models.Comment
	err = db.CommentsCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if existing.CreatedBy != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"content":    body.Content,
			"updated_at": time.Now(),
		},
	}

	_, err = db.CommentsCollection.UpdateByID(context.TODO(), objID, update)
	if err != nil {
		http.Error(w, "DB update failed", http.StatusInternalServerError)
		return
	}

	var updated models.Comment
	if err := db.CommentsCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&updated); err != nil {
		http.Error(w, "Fetch failed", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updated)
}

// func UpdateComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	commentID := ps.ByName("commentid")

// 	var body struct {
// 		Content string `json:"content"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	objID, err := primitive.ObjectIDFromHex(commentID)
// 	if err != nil {
// 		http.Error(w, "Invalid ID", http.StatusBadRequest)
// 		return
// 	}

// 	update := bson.M{
// 		"$set": bson.M{
// 			"content":    body.Content,
// 			"updated_at": time.Now(),
// 		},
// 	}

// 	_, err = db.CommentsCollection.UpdateByID(context.TODO(), objID, update)
// 	if err != nil {
// 		http.Error(w, "DB update failed", http.StatusInternalServerError)
// 		return
// 	}

// 	var updated models.Comment
// 	if err := db.CommentsCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&updated); err != nil {
// 		http.Error(w, "Fetch failed", http.StatusInternalServerError)
// 		return
// 	}

//		utils.RespondWithJSON(w, http.StatusOK, updated)
//	}

func DeleteComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	commentID := ps.ByName("commentid")

	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	claims, err := middleware.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var existing models.Comment
	err = db.CommentsCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&existing)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if existing.CreatedBy != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = db.CommentsCollection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// func DeleteComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	commentID := ps.ByName("commentid")

// 	objID, err := primitive.ObjectIDFromHex(commentID)
// 	if err != nil {
// 		http.Error(w, "Invalid ID", http.StatusBadRequest)
// 		return
// 	}

// 	_, err = db.CommentsCollection.DeleteOne(context.TODO(), bson.M{"_id": objID})
// 	if err != nil {
// 		http.Error(w, "Delete failed", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }
