// discord/discord.go
package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"naevis/db"
	"naevis/utils"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// —— Models ——————————————————————————————————————————

type Chat struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"    json:"id"`
	Participants []string           `bson:"participants"     json:"participants"`
	CreatedAt    time.Time          `bson:"createdAt"        json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"        json:"updatedAt"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"     json:"id"`
	ChatID    primitive.ObjectID `bson:"chatId"            json:"chatId"`
	Sender    string             `bson:"sender"            json:"sender"`
	Content   string             `bson:"content"           json:"content"`
	CreatedAt time.Time          `bson:"createdAt"         json:"createdAt"`
	EditedAt  *time.Time         `bson:"editedAt,omitempty" json:"editedAt,omitempty"`
	Deleted   bool               `bson:"deleted"           json:"deleted"`
}

// —— Globals & Initialization ————————————————————————————————————

var (
	clients = struct {
		sync.RWMutex
		m map[string]*websocket.Conn
	}{m: make(map[string]*websocket.Conn)}

	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	ctx      = context.Background()
)

// —— REST Handlers ——————————————————————————————————————————

func GetUserChats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := utils.GetUserIDFromRequest(r)
	cursor, err := db.ChatsCollection.Find(ctx, bson.M{"participants": user})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cursor.Close(ctx)

	var chats []Chat
	if err := cursor.All(ctx, &chats); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// ensure non-nil slice
	if chats == nil {
		chats = make([]Chat, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func StartNewChat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := utils.GetUserIDFromRequest(r)
	var body struct{ Participants []string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	// ensure the requesting user is in the participants list
	found := false
	for _, p := range body.Participants {
		if p == user {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "must include yourself", 400)
		return
	}
	// check existing chat
	filter := bson.M{"participants": bson.M{"$all": body.Participants}}
	var existing Chat
	err := db.ChatsCollection.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
		return
	}
	if err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), 500)
		return
	}
	// create new chat
	now := time.Now()
	chat := Chat{
		Participants: body.Participants,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	res, err := db.ChatsCollection.InsertOne(ctx, chat)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	chat.ID = res.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

func GetChatByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
	if err != nil {
		http.Error(w, "invalid chatId", 400)
		return
	}
	var chat Chat
	if err := db.ChatsCollection.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

func GetChatMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
	if err != nil {
		http.Error(w, "invalid chatId", 400)
		return
	}
	// pagination
	limit := int64(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := parseInt64(l); err == nil {
			limit = v
		}
	}
	skip := int64(0)
	if s := r.URL.Query().Get("skip"); s != "" {
		if v, err := parseInt64(s); err == nil {
			skip = v
		}
	}

	opts := options.Find().SetSort(bson.M{"createdAt": 1}).SetLimit(limit).SetSkip(skip)
	cursor, err := db.MessagesCollection.Find(ctx, bson.M{"chatId": chatID}, opts)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cursor.Close(ctx)

	var msgs []Message
	if err := cursor.All(ctx, &msgs); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// ensure non-nil slice
	if msgs == nil {
		msgs = make([]Message, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func SendMessageREST(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
	if err != nil {
		http.Error(w, "invalid chatId", 400)
		return
	}
	var body struct{ Content string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	msg, err := persistMessage(chatID, utils.GetUserIDFromRequest(r), body.Content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func EditMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msgID, err := primitive.ObjectIDFromHex(ps.ByName("messageId"))
	if err != nil {
		http.Error(w, "invalid messageId", 400)
		return
	}
	var body struct{ Content string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	now := time.Now()
	res, err := db.MessagesCollection.UpdateOne(ctx,
		bson.M{"_id": msgID},
		bson.M{"$set": bson.M{"content": body.Content, "editedAt": now}},
	)
	if err != nil || res.MatchedCount == 0 {
		http.Error(w, "not found or no permission", 404)
		return
	}
	w.WriteHeader(204)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msgID, err := primitive.ObjectIDFromHex(ps.ByName("messageId"))
	if err != nil {
		http.Error(w, "invalid messageId", 400)
		return
	}
	res, err := db.MessagesCollection.UpdateOne(ctx,
		bson.M{"_id": msgID},
		bson.M{"$set": bson.M{"deleted": true}},
	)
	if err != nil || res.MatchedCount == 0 {
		http.Error(w, "not found or no permission", 404)
		return
	}
	w.WriteHeader(204)
}

// —— Utility & Persistence ————————————————————————————————————————

func parseInt64(s string) (int64, error) {
	var v int64
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return 0, err
	}
	return v, nil
}

func persistMessage(chatID primitive.ObjectID, sender, content string) (*Message, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}
	msg := &Message{
		ChatID:    chatID,
		Sender:    sender,
		Content:   content,
		CreatedAt: time.Now(),
	}
	res, err := db.MessagesCollection.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.ID = res.InsertedID.(primitive.ObjectID)
	// update chat's UpdatedAt
	db.ChatsCollection.UpdateOne(ctx,
		bson.M{"_id": chatID},
		bson.M{"$set": bson.M{"updatedAt": time.Now()}},
	)
	return msg, nil
}

// —— WebSocket Handler ————————————————————————————————————————

func HandleWebSocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := utils.GetUserIDFromRequest(r)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade:", err)
		return
	}
	// register client
	clients.Lock()
	clients.m[userID] = conn
	clients.Unlock()
	log.Println("WS connected:", userID)
	defer func() {
		clients.Lock()
		delete(clients.m, userID)
		clients.Unlock()
		conn.Close()
		log.Println("WS disconnected:", userID)
	}()

	for {
		// 1) accept a “type” field in incoming JSON
		var in struct {
			Type    string `json:"type"`    // "message" | "typing" | "presence"
			ChatID  string `json:"chatId"`  // for message & typing
			Content string `json:"content"` // for message
			Online  bool   `json:"online"`  // for presence
		}
		if err := conn.ReadJSON(&in); err != nil {
			break
		}

		switch in.Type {
		case "message":
			// unchanged: persistMessage & broadcast full payload

		case "typing":
			// broadcast to other participants:
			broadcastToChat(in.ChatID, map[string]interface{}{
				"type":   "typing",
				"from":   userID,
				"chatId": in.ChatID,
			})

		case "presence":
			// broadcast presence status to *all* connected peers
			broadcastGlobal(map[string]interface{}{
				"type":   "presence",
				"from":   userID,
				"online": in.Online,
			})
		}
	}
}

// broadcast typing or presence only to relevant clients
func broadcastToChat(chatHex string, payload interface{}) {
	cid, _ := primitive.ObjectIDFromHex(chatHex)
	var chat Chat
	if err := db.ChatsCollection.FindOne(ctx, bson.M{"_id": cid}).Decode(&chat); err != nil {
		return
	}
	for _, p := range chat.Participants {
		clients.RLock()
		if peer, ok := clients.m[p]; ok {
			peer.WriteJSON(payload)
		}
		clients.RUnlock()
	}
}

// send to everyone (for presence)
func broadcastGlobal(payload interface{}) {
	clients.RLock()
	defer clients.RUnlock()
	for _, conn := range clients.m {
		conn.WriteJSON(payload)
	}
}

// // UploadAttachment handles multipart file uploads
// func UploadAttachment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	user := utils.GetUserIDFromRequest(r)
// 	chatIDHex := ps.ByName("chatId")
// 	chatID, err := primitive.ObjectIDFromHex(chatIDHex)
// 	if err != nil {
// 		http.Error(w, "invalid chatId", 400)
// 		return
// 	}

// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "invalid file", 400)
// 		return
// 	}
// 	defer file.Close()

// 	// 1) save the file to disk or object‑store
// 	//    e.g. path := fmt.Sprintf("uploads/%s_%s", uuid.New(), header.Filename)
// 	//    io.Copy(destFile, file)
// 	url := saveFileSomehow(file, header)

// 	// 2) return an attachment descriptor
// 	att := map[string]interface{}{
// 		"id":   primitive.NewObjectID().Hex(),
// 		"url":  url,
// 		"type": header.Header.Get("Content-Type"),
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(att)
// }

// func SearchMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	chatID, _ := primitive.ObjectIDFromHex(ps.ByName("chatId"))
// 	term := r.URL.Query().Get("term")
// 	filter := bson.M{"chatId": chatID}
// 	if term != "" {
// 		filter["content"] = bson.M{"$regex": primitive.Regex{Pattern: term, Options: "i"}}
// 	}
// 	// then exactly as GetChatMessages: apply skip/limit, ensure non‑nil slice, return JSON
// }

// func GetUnreadCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	user := utils.GetUserIDFromRequest(r)
// 	// query messages where `readBy` array does not contain `user`
// 	// return map[chatID]count
// }

// func MarkAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	msgID, _ := primitive.ObjectIDFromHex(ps.ByName("messageId"))
// 	user := utils.GetUserIDFromRequest(r)
// 	db.MessagesCollection.UpdateOne(ctx,
// 		bson.M{"_id": msgID},
// 		bson.M{"$addToSet": bson.M{"readBy": user}},
// 	)
// 	w.WriteHeader(204)
// }

func UploadAttachment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := utils.GetUserIDFromRequest(r)
	chatIDHex := ps.ByName("chatId")
	chatID, err := primitive.ObjectIDFromHex(chatIDHex)
	if err != nil {
		http.Error(w, "invalid chatId", http.StatusBadRequest)
		return
	}
	_ = user
	_ = chatID

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads dir if needed
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "cannot create upload dir", http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	fname := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	destPath := filepath.Join(uploadDir, fname)

	// Save file to disk
	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "error writing file", http.StatusInternalServerError)
		return
	}

	// Build public URL (adjust prefix to your static‑file handler)
	url := fmt.Sprintf("/static/%s", fname)

	resp := map[string]string{
		"id":   uuid.New().String(),
		"url":  url,
		"type": header.Header.Get("Content-Type"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func SearchMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
	if err != nil {
		http.Error(w, "invalid chatId", http.StatusBadRequest)
		return
	}
	term := r.URL.Query().Get("term")

	// pagination
	limit := int64(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := parseInt64(l); err == nil {
			limit = v
		}
	}
	skip := int64(0)
	if s := r.URL.Query().Get("skip"); s != "" {
		if v, err := parseInt64(s); err == nil {
			skip = v
		}
	}

	filter := bson.M{"chatId": chatID}
	if term != "" {
		filter["content"] = bson.M{"$regex": primitive.Regex{Pattern: term, Options: "i"}}
	}

	opts := options.Find().
		SetSort(bson.M{"createdAt": 1}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := db.MessagesCollection.Find(ctx, filter, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var msgs []Message
	if err := cursor.All(ctx, &msgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if msgs == nil {
		msgs = make([]Message, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}
func GetUnreadCount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := utils.GetUserIDFromRequest(r)

	// First, find all chats the user participates in
	cursor, err := db.ChatsCollection.Find(ctx, bson.M{"participants": user})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	type Unread struct {
		ChatID string `json:"chatId"`
		Count  int64  `json:"count"`
	}
	var result []Unread

	for cursor.Next(ctx) {
		var chat Chat
		if err := cursor.Decode(&chat); err != nil {
			continue
		}
		count, err := db.MessagesCollection.CountDocuments(ctx, bson.M{
			"chatId": chat.ID,
			"readBy": bson.M{"$ne": user},
		})
		if err != nil {
			continue
		}
		result = append(result, Unread{
			ChatID: chat.ID.Hex(),
			Count:  count,
		})
	}
	if result == nil {
		result = make([]Unread, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
func MarkAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msgID, err := primitive.ObjectIDFromHex(ps.ByName("messageId"))
	if err != nil {
		http.Error(w, "invalid messageId", http.StatusBadRequest)
		return
	}
	user := utils.GetUserIDFromRequest(r)

	res, err := db.MessagesCollection.UpdateOne(ctx,
		bson.M{"_id": msgID},
		bson.M{"$addToSet": bson.M{"readBy": user}},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.MatchedCount == 0 {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//
// // discord.go
// package discord

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"log"
// 	"naevis/db"
// 	"naevis/utils"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/julienschmidt/httprouter"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// // —— Models ——————————————————————————————————————————

// type Chat struct {
// 	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	Participants []string           `bson:"participants"    json:"participants"`
// 	CreatedAt    time.Time          `bson:"createdAt"       json:"createdAt"`
// 	UpdatedAt    time.Time          `bson:"updatedAt"       json:"updatedAt"`
// }

// type Message struct {
// 	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id"`
// 	ChatID    primitive.ObjectID `bson:"chatId"          json:"chatId"`
// 	Sender    string             `bson:"sender"          json:"sender"`
// 	Content   string             `bson:"content"         json:"content"`
// 	CreatedAt time.Time          `bson:"createdAt"       json:"createdAt"`
// 	EditedAt  *time.Time         `bson:"editedAt,omitempty" json:"editedAt,omitempty"`
// 	Deleted   bool               `bson:"deleted"         json:"deleted"`
// }

// // —— Globals & Initialization ————————————————————————————————————

// var (
// 	clients = struct {
// 		sync.RWMutex
// 		m map[string]*websocket.Conn
// 	}{m: make(map[string]*websocket.Conn)}
// 	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
// 	ctx      = context.Background()
// )

// // —— REST Handlers ——————————————————————————————————————————

// func GetUserChats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	user := utils.GetUserIDFromRequest(r)
// 	cursor, err := db.ChatsCollection.Find(ctx, bson.M{"participants": user})
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	defer cursor.Close(ctx)

// 	var chats []Chat
// 	if err := cursor.All(ctx, &chats); err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(chats)
// }

// func StartNewChat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	user := utils.GetUserIDFromRequest(r)
// 	var body struct{ Participants []string }
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		http.Error(w, "invalid body", 400)
// 		return
// 	}
// 	// ensure the requesting user is in the participants list
// 	found := false
// 	for _, p := range body.Participants {
// 		if p == user {
// 			found = true
// 			break
// 		}
// 	}
// 	if !found {
// 		http.Error(w, "must include yourself", 400)
// 		return
// 	}
// 	// check existing chat
// 	filter := bson.M{"participants": bson.M{"$all": body.Participants}}
// 	var existing Chat
// 	err := db.ChatsCollection.FindOne(ctx, filter).Decode(&existing)
// 	if err == nil {
// 		json.NewEncoder(w).Encode(existing)
// 		return
// 	}
// 	if err != mongo.ErrNoDocuments {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	// create new chat
// 	now := time.Now()
// 	chat := Chat{
// 		Participants: body.Participants,
// 		CreatedAt:    now,
// 		UpdatedAt:    now,
// 	}
// 	res, err := db.ChatsCollection.InsertOne(ctx, chat)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	chat.ID = res.InsertedID.(primitive.ObjectID)
// 	json.NewEncoder(w).Encode(chat)
// }

// func GetChatByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
// 	if err != nil {
// 		http.Error(w, "invalid chatId", 400)
// 		return
// 	}
// 	var chat Chat
// 	if err := db.ChatsCollection.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat); err != nil {
// 		http.Error(w, "not found", 404)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(chat)
// }

// func GetChatMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
// 	if err != nil {
// 		http.Error(w, "invalid chatId", 400)
// 		return
// 	}
// 	// pagination
// 	limit := int64(50)
// 	if l := r.URL.Query().Get("limit"); l != "" {
// 		if v, err := parseInt64(l); err == nil {
// 			limit = v
// 		}
// 	}
// 	skip := int64(0)
// 	if s := r.URL.Query().Get("skip"); s != "" {
// 		if v, err := parseInt64(s); err == nil {
// 			skip = v
// 		}
// 	}

// 	opts := options.Find().SetSort(bson.M{"createdAt": 1}).SetLimit(limit).SetSkip(skip)
// 	cursor, err := db.MessagesCollection.Find(ctx, bson.M{"chatId": chatID}, opts)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	defer cursor.Close(ctx)

// 	var msgs []Message
// 	if err := cursor.All(ctx, &msgs); err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(msgs)
// }

// func SendMessageREST(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	chatID, err := primitive.ObjectIDFromHex(ps.ByName("chatId"))
// 	if err != nil {
// 		http.Error(w, "invalid chatId", 400)
// 		return
// 	}
// 	var body struct{ Content string }
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		http.Error(w, "invalid body", 400)
// 		return
// 	}
// 	msg, err := persistMessage(chatID, utils.GetUserIDFromRequest(r), body.Content)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(msg)
// }

// func EditMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	msgID, err := primitive.ObjectIDFromHex(ps.ByName("messageId"))
// 	if err != nil {
// 		http.Error(w, "invalid messageId", 400)
// 		return
// 	}
// 	var body struct{ Content string }
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		http.Error(w, "invalid body", 400)
// 		return
// 	}
// 	now := time.Now()
// 	res, err := db.MessagesCollection.UpdateOne(ctx,
// 		bson.M{"_id": msgID},
// 		bson.M{"$set": bson.M{"content": body.Content, "editedAt": now}},
// 	)
// 	if err != nil || res.MatchedCount == 0 {
// 		http.Error(w, "not found or no permission", 404)
// 		return
// 	}
// 	w.WriteHeader(204)
// }

// func DeleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	msgID, err := primitive.ObjectIDFromHex(ps.ByName("messageId"))
// 	if err != nil {
// 		http.Error(w, "invalid messageId", 400)
// 		return
// 	}
// 	// soft delete
// 	res, err := db.MessagesCollection.UpdateOne(ctx,
// 		bson.M{"_id": msgID},
// 		bson.M{"$set": bson.M{"deleted": true}},
// 	)
// 	if err != nil || res.MatchedCount == 0 {
// 		http.Error(w, "not found or no permission", 404)
// 		return
// 	}
// 	w.WriteHeader(204)
// }

// // —— Utility & Persistence ————————————————————————————————————————

// func parseInt64(s string) (int64, error) {
// 	var v int64
// 	if err := json.Unmarshal([]byte(s), &v); err != nil {
// 		return 0, err
// 	}
// 	return v, nil
// }

// func persistMessage(chatID primitive.ObjectID, sender, content string) (*Message, error) {
// 	if content == "" {
// 		return nil, errors.New("empty content")
// 	}
// 	msg := &Message{
// 		ChatID:    chatID,
// 		Sender:    sender,
// 		Content:   content,
// 		CreatedAt: time.Now(),
// 	}
// 	res, err := db.MessagesCollection.InsertOne(ctx, msg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	msg.ID = res.InsertedID.(primitive.ObjectID)
// 	// update chat's UpdatedAt
// 	db.ChatsCollection.UpdateOne(ctx, bson.M{"_id": chatID}, bson.M{"$set": bson.M{"updatedAt": time.Now()}})
// 	return msg, nil
// }

// // —— WebSocket Handler ————————————————————————————————————————

// func HandleWebSocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	userID := utils.GetUserIDFromRequest(r)
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("ws upgrade:", err)
// 		return
// 	}
// 	// register client
// 	clients.Lock()
// 	clients.m[userID] = conn
// 	clients.Unlock()
// 	log.Println("WS connected:", userID)
// 	defer func() {
// 		clients.Lock()
// 		delete(clients.m, userID)
// 		clients.Unlock()
// 		conn.Close()
// 		log.Println("WS disconnected:", userID)
// 	}()

// 	for {
// 		// expect: {"chatId":"...", "content":"..."}
// 		var in struct {
// 			ChatID  string `json:"chatId"`
// 			Content string `json:"content"`
// 		}
// 		if err := conn.ReadJSON(&in); err != nil {
// 			break
// 		}
// 		cid, err := primitive.ObjectIDFromHex(in.ChatID)
// 		if err != nil {
// 			continue
// 		}
// 		msg, err := persistMessage(cid, userID, in.Content)
// 		if err != nil {
// 			continue
// 		}
// 		// broadcast to all participants
// 		var chat Chat
// 		if err := db.ChatsCollection.FindOne(ctx, bson.M{"_id": cid}).Decode(&chat); err != nil {
// 			continue
// 		}
// 		out := struct {
// 			ChatID  string   `json:"chatId"`
// 			Message *Message `json:"message"`
// 		}{
// 			ChatID:  cid.Hex(),
// 			Message: msg,
// 		}
// 		for _, p := range chat.Participants {
// 			clients.RLock()
// 			if peer, ok := clients.m[p]; ok {
// 				peer.WriteJSON(out)
// 			}
// 			clients.RUnlock()
// 		}
// 	}
// }
