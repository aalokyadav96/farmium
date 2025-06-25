// // buy.go
// package farms

// import (
// 	"context"
// 	"net/http"
// 	"strconv"
// 	"time"

// 	"naevis/db"
// 	"naevis/models"
// 	"naevis/utils"

// 	"github.com/julienschmidt/httprouter"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // BuyCrop handles purchasing one or more units of a crop:
// //   - decrements the quantity (but not below zero),
// //   - sets outOfStock when it hits zero,
// //   - records an order in OrdersCollection,
// //   - returns the new quantity and outOfStock status.
// func BuyCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	ctx := context.Background()

// 	// 1) Authn: expect userID in context
// 	userIDval := r.Context().Value("userId")
// 	userID, ok := userIDval.(primitive.ObjectID)
// 	if !ok {
// 		utils.RespondWithJSON(w, http.StatusUnauthorized, utils.M{"success": false, "message": "Not authenticated"})
// 		return
// 	}

// 	// 2) Parse farm & crop IDs
// 	farmID, err := primitive.ObjectIDFromHex(ps.ByName("id"))
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid farm ID"})
// 		return
// 	}
// 	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
// 	if err != nil {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
// 		return
// 	}

// 	// 3) Parse optional qty query param (default = 1)
// 	qty := 1
// 	if qstr := r.URL.Query().Get("qty"); qstr != "" {
// 		if q, err := strconv.Atoi(qstr); err == nil && q > 0 {
// 			qty = q
// 		}
// 	}

// 	// 4) Atomic update via pipeline: decrement, floor at 0, set outOfStock, update timestamp
// 	pipeline := mongo.Pipeline{
// 		bson.D{{Key: "$set", Value: bson.D{
// 			{Key: "crops", Value: bson.D{{
// 				Key: "$map", Value: bson.D{
// 					{Key: "input", Value: "$crops"},
// 					{Key: "as", Value: "c"},
// 					{Key: "in", Value: bson.D{{
// 						Key: "$cond", Value: bson.A{
// 							bson.D{{Key: "$eq", Value: bson.A{"$$c._id", cropID}}},
// 							bson.D{{Key: "$mergeObjects", Value: bson.A{
// 								"$$c",
// 								bson.D{
// 									{"quantity", bson.D{{"$max", bson.A{0, bson.D{{"$subtract", bson.A{"$$c.quantity", qty}}}}}}},
// 									{"outOfStock", bson.D{{"$lte", bson.A{bson.D{{"$subtract", bson.A{"$$c.quantity", qty}}}, 0}}}},
// 									{"updatedAt", time.Now()},
// 								},
// 							}}},
// 							"$$c",
// 						},
// 					}}},
// 				},
// 			}}},
// 		}}},
// 	}

// 	filter := bson.M{
// 		"_id":       farmID,
// 		"crops._id": cropID,
// 	}

// 	updateRes, err := db.FarmsCollection.UpdateOne(ctx, filter, pipeline)
// 	if err != nil || updateRes.ModifiedCount == 0 {
// 		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{
// 			"success": false,
// 			"message": "Crop not available or already out of stock",
// 		})
// 		return
// 	}

// 	// 5) Fetch the updated sub-document
// 	var tmp struct{ Crops []models.Crop }
// 	findOpts := options.FindOne().SetProjection(bson.D{
// 		{"crops", bson.D{{"$elemMatch", bson.D{{"_id", cropID}}}}},
// 	})
// 	if err := db.FarmsCollection.FindOne(ctx, filter, findOpts).Decode(&tmp); err != nil || len(tmp.Crops) == 0 {
// 		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.M{"success": false, "message": "Failed to retrieve updated crop"})
// 		return
// 	}
// 	updatedCrop := tmp.Crops[0]

// 	// 6) Record an order in OrdersCollection
// 	order := models.FarmOrder{
// 		ID:              primitive.NewObjectID(),
// 		UserID:          userID,
// 		FarmID:          farmID,
// 		CropID:          cropID,
// 		Quantity:        qty,
// 		PriceAtPurchase: updatedCrop.Price,
// 		BoughtAt:        time.Now(),
// 	}
// 	if _, err := db.FarmOrdersCollection.InsertOne(ctx, order); err != nil {
// 		// non‐fatal: we still return success on the stock update
// 	}

// 	// 7) Return new state
// 	utils.RespondWithJSON(w, http.StatusOK, utils.M{
// 		"success":     true,
// 		"newQuantity": updatedCrop.Quantity,
// 		"outOfStock":  updatedCrop.OutOfStock,
// 	})
// }

package farms

import (
	"context"
	"net/http"
	"time"

	"naevis/db"
	"naevis/globals"
	"naevis/utils"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BuyCrop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	farmID, err := primitive.ObjectIDFromHex(ps.ByName("id"))
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid farm ID"})
		return
	}

	cropID, err := primitive.ObjectIDFromHex(ps.ByName("cropid"))
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Invalid crop ID"})
		return
	}

	// Retrieve user ID
	requestingUserID, ok := r.Context().Value(globals.UserIDKey).(string)
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	_ = requestingUserID

	// Decrement quantity
	filter := bson.M{
		"_id":              farmID,
		"crops._id":        cropID,
		"crops.quantity":   bson.M{"$gt": 0},
		"crops.outOfStock": false,
	}

	update := bson.M{
		"$inc": bson.M{"crops.$.quantity": -1},
		"$set": bson.M{"crops.$.updatedAt": time.Now()},
	}

	result, err := db.FarmsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil || result.ModifiedCount == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.M{"success": false, "message": "Crop not available or already out of stock"})
		return
	}

	// Set outOfStock if quantity hits 0
	filterZero := bson.M{
		"_id":            farmID,
		"crops._id":      cropID,
		"crops.quantity": 0,
	}
	db.FarmsCollection.UpdateOne(context.Background(), filterZero, bson.M{
		"$set": bson.M{"crops.$.outOfStock": true},
	})

	utils.RespondWithJSON(w, http.StatusOK, utils.M{"success": true})
}
