package suggestions

import (
	"context"
	"encoding/json"
	"fmt"
	"naevis/db"
	"naevis/globals"
	"naevis/rdx"
	"naevis/structs"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SuggestFollowers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	currentUserID := r.URL.Query().Get("userid")
	if currentUserID == "" {
		http.Error(w, "Missing userid", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(globals.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Pagination parameters
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	// Fetch user's follow data
	var followData structs.UserFollow
	err = db.FollowingsCollection.FindOne(context.TODO(), bson.M{"userid": currentUserID}).Decode(&followData)
	if err != nil && err != mongo.ErrNoDocuments {
		http.Error(w, "Failed to fetch follow data", http.StatusInternalServerError)
		return
	}

	// Exclude already followed users + current user
	excludedUserIDs := append(followData.Follows, currentUserID, userID)

	// Query for suggested users
	filter := bson.M{"userid": bson.M{"$nin": excludedUserIDs}}
	options := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetProjection(bson.M{
			"userid":   1,
			"username": 1,
			"bio":      1,
		})

	cursor, err := db.UserCollection.Find(context.TODO(), filter, options)
	if err != nil {
		http.Error(w, "Failed to fetch suggestions", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	// Collect suggested users
	var suggestedUsers []structs.UserSuggest
	for cursor.Next(context.TODO()) {
		var suggestedUser structs.UserSuggest
		if err := cursor.Decode(&suggestedUser); err == nil {
			// Explicitly set is_following: false
			suggestedUser.IsFollowing = false
			suggestedUsers = append(suggestedUsers, suggestedUser)
		}
	}

	// Handle empty response case
	if len(suggestedUsers) == 0 {
		suggestedUsers = []structs.UserSuggest{}
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestedUsers); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

/***************************************************/
func GetPlaceSuggestions(ctx context.Context, query string) ([]structs.Suggestion, error) {
	var suggestions []structs.Suggestion

	// Use Redis KEYS command to find matching place suggestions by name
	// (this is a simple approach, you may want a more efficient search strategy)
	keys, err := rdx.Conn.Keys(ctx, fmt.Sprintf("suggestions:place:%s*", query)).Result()
	if err != nil {
		return nil, err
	}

	// Retrieve the corresponding place data
	for _, key := range keys {
		var suggestion structs.Suggestion
		err := rdx.Conn.Get(ctx, key).Scan(&suggestion)
		if err != nil {
			return nil, err
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}
