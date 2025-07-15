package routes

import (
	"naevis/admin"
	"naevis/auth"
	"naevis/cart"
	"naevis/chats"
	"naevis/comments"
	"naevis/discord"
	"naevis/farms"
	"naevis/home"
	"naevis/middleware"
	"naevis/newchat"
	"naevis/profile"
	"naevis/ratelim"
	"naevis/recipes"
	"naevis/reports"
	"naevis/reviews"
	"naevis/search"
	"naevis/settings"
	"naevis/suggestions"
	"naevis/userdata"
	"naevis/utils"
	"net/http"
	_ "net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

func AddStaticRoutes(router *httprouter.Router) {
	router.ServeFiles("/static/postpic/*filepath", http.Dir("static/postpic"))
	router.ServeFiles("/static/merchpic/*filepath", http.Dir("static/merchpic"))
	router.ServeFiles("/static/menupic/*filepath", http.Dir("static/menupic"))
	router.ServeFiles("/static/uploads/*filepath", http.Dir("static/uploads"))
	router.ServeFiles("/static/placepic/*filepath", http.Dir("static/placepic"))
	router.ServeFiles("/static/businesspic/*filepath", http.Dir("static/eventpic"))
	router.ServeFiles("/static/userpic/*filepath", http.Dir("static/userpic"))
	router.ServeFiles("/static/eventpic/*filepath", http.Dir("static/eventpic"))
	router.ServeFiles("/static/artistpic/*filepath", http.Dir("static/artistpic"))
	router.ServeFiles("/static/cartoonpic/*filepath", http.Dir("static/cartoonpic"))
	router.ServeFiles("/static/chatpic/*filepath", http.Dir("static/chatpic"))
	router.ServeFiles("/static/newchatpic/*filepath", http.Dir("static/newchatpic"))
	router.ServeFiles("/static/threadpic/*filepath", http.Dir("static/threadpic"))
}

func AddAdminRoutes(router *httprouter.Router) {
	router.GET("/api/v1/admin/reports", middleware.Authenticate(admin.GetReports))
}

func AddRecipeRoutes(router *httprouter.Router) {
	router.GET("/api/v1/recipes/tags", ratelim.RateLimit(recipes.GetRecipeTags))
	router.GET("/api/v1/recipes", middleware.OptionalAuth(recipes.GetRecipes))
	router.GET("/api/v1/recipes/recipe/:id", middleware.OptionalAuth(recipes.GetRecipe))
	router.POST("/api/v1/recipes", middleware.Authenticate(recipes.CreateRecipe))
	router.PUT("/api/v1/recipes/recipe/:id", middleware.Authenticate(recipes.UpdateRecipe))
	router.DELETE("/api/v1/recipes/recipe/:id", middleware.Authenticate(recipes.DeleteRecipe))
}

func AddDiscordRoutes(router *httprouter.Router) {
	router.GET("/api/v1/merechats/all", middleware.Authenticate(discord.GetUserChats))
	router.POST("/api/v1/merechats/start", middleware.Authenticate(discord.StartNewChat))
	router.GET("/api/v1/merechats/chat/:chatId", middleware.Authenticate(discord.GetChatByID))
	router.GET("/api/v1/merechats/chat/:chatId/messages", middleware.Authenticate(discord.GetChatMessages))
	router.POST("/api/v1/merechats/chat/:chatId/message", middleware.Authenticate(discord.SendMessageREST))
	router.PATCH("/api/v1/meremessages/:messageId", middleware.Authenticate(discord.EditMessage))
	router.DELETE("/api/v1/meremessages/:messageId", middleware.Authenticate(discord.DeleteMessage))
	router.GET("/ws/merechat", middleware.Authenticate(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		discord.HandleWebSocket(w, r, httprouter.Params{}) // or just nil
	}))
	router.POST("/api/v1/merechats/chat/:chatId/upload", middleware.Authenticate(discord.UploadAttachment))
	router.GET("/api/v1/merechats/chat/:chatId/search", middleware.Authenticate(discord.SearchMessages))
	router.GET("/api/v1/meremessages/unread-count", middleware.Authenticate(discord.GetUnreadCount))
	router.POST("/api/v1/meremessages/:messageId/read", middleware.Authenticate(discord.MarkAsRead))

}

func AddNewChatRoutes(router *httprouter.Router, hub *newchat.Hub) {
	router.GET("/api/v1/newchats/all", middleware.Authenticate(chats.GetUserChats))
	router.GET("/ws/newchat/:room", newchat.WebSocketHandler(hub))
	router.POST("/newchat/upload", middleware.Authenticate(newchat.UploadHandler(hub)))
	router.POST("/newchat/edit", newchat.EditMessageHandler(hub))
	router.POST("/newchat/delete", newchat.DeleteMessageHandler(hub))

}

func AddChatRoutes(router *httprouter.Router) {
	router.GET("/api/v1/chats/all", middleware.Authenticate(chats.GetUserChats))
	router.POST("/api/v1/chats/init", middleware.Authenticate(chats.InitChat))
	router.GET("/api/v1/chat/:chatid", middleware.Authenticate(chats.GetChat))
	router.POST("/api/v1/chat/:chatid/message", middleware.Authenticate(chats.CreateMessage))
	router.PUT("/api/v1/chat/:chatid/message/:msgid", middleware.Authenticate(chats.UpdateMessage))
	// router.GET("/api/v1/chat/:chatid", middleware.Authenticate(chats.GetMessage))
	router.DELETE("/api/v1/chat/:chatid/message/:msgid", middleware.Authenticate(chats.DeleteMessage))
	router.GET("/ws/chat", chats.ChatWebSocket)
	router.GET("/api/v1/chat/:chatid/search", middleware.Authenticate(chats.SearchChat))
}

func AddHomeRoutes(router *httprouter.Router) {
	router.GET("/api/v1/home/:apiRoute", middleware.OptionalAuth(home.GetHomeContent))
}

func AddReportRoutes(router *httprouter.Router) {
	router.POST("/api/v1/report", ratelim.RateLimit(middleware.Authenticate(reports.ReportContent)))
	router.GET("/api/v1/reports", ratelim.RateLimit(middleware.Authenticate(reports.GetReports)))
	router.PUT("/api/v1/report/:id", ratelim.RateLimit(middleware.Authenticate(reports.UpdateReport)))
}

func AddCommentsRoutes(router *httprouter.Router) {
	router.POST("/api/v1/comments/:entitytype/:entityid", comments.CreateComment)
	router.GET("/api/v1/comments/:entitytype/:entityid", comments.GetComments)
	router.GET("/api/v1/comments/:entitytype", comments.GetComment)
	router.PUT("/api/v1/comments/:entitytype/:entityid/:commentid", comments.UpdateComment)
	router.DELETE("/api/v1/comments/:entitytype/:entityid/:commentid", comments.DeleteComment)
}

func AddAuthRoutes(router *httprouter.Router) {
	router.POST("/api/v1/auth/register", ratelim.RateLimit(auth.Register))
	router.POST("/api/v1/auth/login", ratelim.RateLimit(auth.Login))
	router.POST("/api/v1/auth/logout", middleware.Authenticate(auth.LogoutUser))
	router.POST("/api/v1/auth/token/refresh", ratelim.RateLimit(middleware.Authenticate(auth.RefreshToken)))

	router.POST("/api/v1/auth/verify-otp", ratelim.RateLimit(auth.VerifyOTPHandler))
	router.POST("/api/v1/auth/request-otp", ratelim.RateLimit(auth.VerifyOTPHandler))
}

func AddCartRoutes(router *httprouter.Router) {
	// Cart operations
	router.POST("/api/v1/cart", middleware.Authenticate(cart.AddToCart))
	router.GET("/api/v1/cart", middleware.Authenticate(cart.GetCart))
	router.POST("/api/v1/cart/update", middleware.Authenticate(cart.UpdateCart))
	router.POST("/api/v1/cart/checkout", middleware.Authenticate(cart.InitiateCheckout))

	// Checkout session creation
	router.POST("/api/v1/checkout/session", middleware.Authenticate(cart.CreateCheckoutSession))

	// Order placement
	router.POST("/api/v1/order", middleware.Authenticate(cart.PlaceOrder))
}

// // RegisterFarmRoutes wires up endpoints to the given router
// func RegisterFarmRoutes(router *httprouter.Router) {
// 	router.POST("/api/v1/farms", middleware.Authenticate(farms.CreateFarm))
// 	router.GET("/api/v1/farms", farms.GetPaginatedFarms)
// 	router.GET("/api/v1/farms/:id", middleware.Authenticate(farms.GetFarm))
// 	router.PUT("/api/v1/farms/:id", middleware.Authenticate(farms.EditFarm))
// 	router.DELETE("/api/v1/farms/:id", middleware.Authenticate(farms.DeleteFarm))
// 	// Crop routes
// 	router.POST("/api/v1/farms/:id/crops", middleware.Authenticate(farms.AddCrop))
// 	router.PUT("/api/v1/farms/:id/crops/:cropid", middleware.Authenticate(farms.EditCrop))
// 	router.DELETE("/api/v1/farms/:id/crops/:cropid", middleware.Authenticate(farms.DeleteCrop))
// 	router.PUT("/api/v1/farms/:id/crops/:cropid/buy", middleware.Authenticate(farms.BuyCrop))

// 	router.GET("/api/v1/dash/farms", middleware.Authenticate(farms.GetFarmDash))
// 	router.GET("/api/v1/farmorders/mine", middleware.Authenticate(farms.GetMyFarmOrders))
// 	router.GET("/api/v1/farmorders/incoming", middleware.Authenticate(farms.GetIncomingFarmOrders))

// 	router.GET("/api/v1/crops", farms.GetFilteredCrops)
// 	router.GET("/api/v1/crops/catalogue", farms.GetCropCatalogue)
// 	router.GET("/api/v1/crops/precatalogue", farms.GetPreCropCatalogue)
// 	router.GET("/api/v1/crops/types", farms.GetCropTypes)
// 	// router.GET("/api/v1/crops/crop/:cropid", middleware.Authenticate(farms.GetCropFarms))
// 	router.GET("/api/v1/crops/crop/:cropname", middleware.Authenticate(farms.GetCropTypeFarms))

// 	router.GET("/api/v1/farm/items", farms.GetItems)
// 	router.GET("/api/v1/farm/products", farms.GetProducts)
// 	router.GET("/api/v1/farm/tools", farms.GetTools)
// 	router.GET("/api/v1/farm/items/categories", farms.GetItemCategories)

// 	router.POST("/api/v1/farm/product", farms.CreateProduct)
// 	router.PUT("/api/v1/farm/product/:id", farms.UpdateProduct)
// 	router.DELETE("/api/v1/farm/product/:id", farms.DeleteProduct)

// 	router.POST("/api/v1/farm/tool", farms.CreateTool)
// 	router.PUT("/api/v1/farm/tool/:id", farms.UpdateTool)
// 	router.DELETE("/api/v1/farm/tool/:id", farms.DeleteTool)

//		router.POST("/api/v1/upload/images", utils.UploadImages)
//	}
func RegisterFarmRoutes(router *httprouter.Router) {
	// 🌾 Farm CRUD
	router.POST("/api/v1/farms", middleware.Authenticate(farms.CreateFarm))
	router.GET("/api/v1/farms", farms.GetPaginatedFarms)
	router.GET("/api/v1/farms/:id", middleware.OptionalAuth(farms.GetFarm))
	router.PUT("/api/v1/farms/:id", middleware.Authenticate(farms.EditFarm))
	router.DELETE("/api/v1/farms/:id", middleware.Authenticate(farms.DeleteFarm))

	// 🌱 Crops (within farm)
	router.POST("/api/v1/farms/:id/crops", middleware.Authenticate(farms.AddCrop))
	router.PUT("/api/v1/farms/:id/crops/:cropid", middleware.Authenticate(farms.EditCrop))
	router.DELETE("/api/v1/farms/:id/crops/:cropid", middleware.Authenticate(farms.DeleteCrop))
	router.PUT("/api/v1/farms/:id/crops/:cropid/buy", middleware.Authenticate(farms.BuyCrop))

	// 📊 Dashboard
	router.GET("/api/v1/dash/farms", middleware.Authenticate(farms.GetFarmDash))

	// 📦 Farm Orders
	router.GET("/api/v1/orders/mine", middleware.Authenticate(farms.GetMyFarmOrders))           // my own farm orders
	router.GET("/api/v1/orders/incoming", middleware.Authenticate(farms.GetIncomingFarmOrders)) // orders from buyers to me
	router.POST("/api/v1/farmorders/:id/accept", middleware.Authenticate(farms.AcceptOrder))
	router.POST("/api/v1/farmorders/:id/reject", middleware.Authenticate(farms.RejectOrder))
	router.POST("/api/v1/farmorders/:id/deliver", middleware.Authenticate(farms.MarkOrderDelivered))
	router.POST("/api/v1/farmorders/:id/markpaid", middleware.Authenticate(farms.MarkOrderPaid))
	router.GET("/api/v1/farmorders/:id/receipt", middleware.Authenticate(farms.DownloadReceipt))

	// 🌾 Crop catalogue & type browsing
	router.GET("/api/v1/crops", farms.GetFilteredCrops)                                         // for search/filter
	router.GET("/api/v1/crops/catalogue", farms.GetCropCatalogue)                               // full list
	router.GET("/api/v1/crops/precatalogue", farms.GetPreCropCatalogue)                         // pre-published
	router.GET("/api/v1/crops/types", farms.GetCropTypes)                                       // types list
	router.GET("/api/v1/crops/crop/:cropname", middleware.OptionalAuth(farms.GetCropTypeFarms)) // farms by crop name

	// 🛒 Items, Products, Tools
	// -- GET
	// router.GET("/api/v1/farm/products", farms.GetProducts)
	// router.GET("/api/v1/farm/tools", farms.GetTools)

	router.GET("/api/v1/farm/items", farms.GetItems)
	router.GET("/api/v1/farm/items/categories", farms.GetItemCategories)

	// -- Products (CRUD)
	router.POST("/api/v1/farm/product", farms.CreateProduct)
	router.PUT("/api/v1/farm/product/:id", farms.UpdateProduct)
	router.DELETE("/api/v1/farm/product/:id", farms.DeleteProduct)

	// -- Tools (CRUD)
	router.POST("/api/v1/farm/tool", farms.CreateTool)
	router.PUT("/api/v1/farm/tool/:id", farms.UpdateTool)
	router.DELETE("/api/v1/farm/tool/:id", farms.DeleteTool)

	// 🖼 Upload
	router.POST("/api/v1/upload/images", utils.UploadImages)
}

func AddSuggestionsRoutes(router *httprouter.Router) {
	router.GET("/api/v1/suggestions/follow", ratelim.RateLimit(middleware.Authenticate(suggestions.SuggestFollowers)))
}

func AddReviewsRoutes(router *httprouter.Router) {
	router.GET("/api/v1/reviews/:entityType/:entityId", ratelim.RateLimit(middleware.Authenticate(reviews.GetReviews)))
	router.GET("/api/v1/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.GetReview)))
	router.POST("/api/v1/reviews/:entityType/:entityId", ratelim.RateLimit(middleware.Authenticate(reviews.AddReview)))
	router.PUT("/api/v1/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.EditReview)))
	router.DELETE("/api/v1/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.DeleteReview)))
}

func AddProfileRoutes(router *httprouter.Router) {
	router.GET("/api/v1/profile/profile", middleware.Authenticate(profile.GetProfile))
	router.PUT("/api/v1/profile/edit", middleware.Authenticate(profile.EditProfile))
	router.PUT("/api/v1/profile/avatar", middleware.Authenticate(profile.EditProfilePic))
	router.PUT("/api/v1/profile/banner", middleware.Authenticate(profile.EditProfileBanner))
	router.DELETE("/api/v1/profile/delete", middleware.Authenticate(profile.DeleteProfile))

	router.GET("/api/v1/user/:username", ratelim.RateLimit(profile.GetUserProfile))

	router.GET("/api/v1/user/:username/data", ratelim.RateLimit(middleware.Authenticate(userdata.GetUserProfileData)))
	router.GET("/api/v1/user/:username/udata", ratelim.RateLimit(middleware.Authenticate(userdata.GetOtherUserProfileData)))

	router.PUT("/api/v1/follows/:id", ratelim.RateLimit(middleware.Authenticate(profile.ToggleFollow)))
	router.DELETE("/api/v1/follows/:id", ratelim.RateLimit(middleware.Authenticate(profile.ToggleUnFollow)))
	router.GET("/api/v1/follows/:id/status", ratelim.RateLimit(middleware.Authenticate(profile.DoesFollow)))
	router.GET("/api/v1/followers/:id", ratelim.RateLimit(middleware.Authenticate(profile.GetFollowers)))
	router.GET("/api/v1/following/:id", ratelim.RateLimit(middleware.Authenticate(profile.GetFollowing)))

}

func AddUtilityRoutes(router *httprouter.Router, rateLimiter *ratelim.RateLimiter) {
	router.GET("/api/v1/csrf", rateLimiter.Limit(middleware.Authenticate(utils.CSRF)))
}

func AddSettingsRoutes(router *httprouter.Router) {
	router.GET("/api/v1/settings/init/:userid", middleware.Authenticate(settings.InitUserSettings))
	// router.GET("/api/v1/settings/setting/:type", getUserSettings)
	router.GET("/api/v1/settings/all", ratelim.RateLimit(middleware.Authenticate(settings.GetUserSettings)))
	router.PUT("/api/v1/settings/setting/:type", ratelim.RateLimit(middleware.Authenticate(settings.UpdateUserSetting)))
}

func AddSearchRoutes(router *httprouter.Router) {
	router.GET("/api/v1/ac", search.Autocompleter)
	router.GET("/api/v1/search/:entityType", ratelim.RateLimit(search.SearchHandler))
	router.POST("/emitted", search.EventHandler)
}
