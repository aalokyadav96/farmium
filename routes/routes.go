package routes

import (
	"naevis/auth"
	"naevis/cart"
	"naevis/chats"
	"naevis/comments"
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
	"naevis/suggestions"
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

func AddNewChatRoutes(router *httprouter.Router, hub *newchat.Hub) {
	router.GET("/api/newchats/all", middleware.Authenticate(chats.GetUserChats))
	// router.POST("/api/newchats/init", middleware.Authenticate(newchat.InitNewChat))
	router.GET("/ws/newchat/:room", newchat.WebSocketHandler(hub))
	router.POST("/newchat/upload", middleware.Authenticate(newchat.UploadHandler(hub)))
	router.POST("/newchat/edit", newchat.EditMessageHandler(hub))
	router.POST("/newchat/delete", newchat.DeleteMessageHandler(hub))

}

func AddChatRoutes(router *httprouter.Router) {
	router.GET("/api/chats/all", middleware.Authenticate(chats.GetUserChats))
	router.POST("/api/chats/init", middleware.Authenticate(chats.InitChat))
	router.GET("/api/chat/:chatid", middleware.Authenticate(chats.GetChat))
	router.POST("/api/chat/:chatid/message", middleware.Authenticate(chats.CreateMessage))
	router.PUT("/api/chat/:chatid/message/:msgid", middleware.Authenticate(chats.UpdateMessage))
	// router.GET("/api/chat/:chatid", middleware.Authenticate(chats.GetMessage))
	router.DELETE("/api/chat/:chatid/message/:msgid", middleware.Authenticate(chats.DeleteMessage))
	router.GET("/ws/chat", chats.ChatWebSocket)
	router.GET("/api/chat/:chatid/search", middleware.Authenticate(chats.SearchChat))
}

func AddHomeRoutes(router *httprouter.Router) {
	router.GET("/api/home/:apiRoute", middleware.Authenticate(home.GetHomeContent))
}

func AddReportRoutes(router *httprouter.Router) {
	router.POST("/api/report", ratelim.RateLimit(middleware.Authenticate(reports.ReportContent)))
	router.GET("/api/reports", ratelim.RateLimit(middleware.Authenticate(reports.GetReports)))
	router.PUT("/api/report/:id", ratelim.RateLimit(middleware.Authenticate(reports.UpdateReport)))
}

func AddCommentsRoutes(router *httprouter.Router) {
	router.POST("/api/comments/:entitytype/:entityid", comments.CreateComment)
	router.GET("/api/comments/:entitytype/:entityid", comments.GetComments)
	router.GET("/api/comments/:entitytype", comments.GetComment)
	router.PUT("/api/comments/:entitytype/:entityid/:commentid", comments.UpdateComment)
	router.DELETE("/api/comments/:entitytype/:entityid/:commentid", comments.DeleteComment)
}

func AddAuthRoutes(router *httprouter.Router) {
	router.POST("/api/auth/register", ratelim.RateLimit(auth.Register))
	router.POST("/api/auth/login", ratelim.RateLimit(auth.Login))
	router.POST("/api/auth/logout", middleware.Authenticate(auth.LogoutUser))
	router.POST("/api/auth/token/refresh", ratelim.RateLimit(middleware.Authenticate(auth.RefreshToken)))

	router.POST("/api/auth/verify-otp", ratelim.RateLimit(auth.VerifyOTPHandler))
	router.POST("/api/auth/request-otp", ratelim.RateLimit(auth.VerifyOTPHandler))
}

func AddCartRoutes(router *httprouter.Router) {
	// Cart operations
	router.POST("/api/cart", middleware.Authenticate(cart.AddToCart))
	router.GET("/api/cart", middleware.Authenticate(cart.GetCart))
	router.POST("/api/cart/update", middleware.Authenticate(cart.UpdateCart))
	router.POST("/api/cart/checkout", middleware.Authenticate(cart.InitiateCheckout))

	// Checkout session creation
	router.POST("/api/checkout/session", middleware.Authenticate(cart.CreateCheckoutSession))

	// Order placement
	router.POST("/api/order", middleware.Authenticate(cart.PlaceOrder))
}

// RegisterFarmRoutes wires up endpoints to the given router
func RegisterFarmRoutes(router *httprouter.Router) {
	router.POST("/api/farms", middleware.Authenticate(farms.CreateFarm))
	router.GET("/api/farms", farms.GetPaginatedFarms)
	router.GET("/api/farms/:id", middleware.Authenticate(farms.GetFarm))
	router.PUT("/api/farms/:id", middleware.Authenticate(farms.EditFarm))
	router.DELETE("/api/farms/:id", middleware.Authenticate(farms.DeleteFarm))
	// Crop routes
	router.POST("/api/farms/:id/crops", middleware.Authenticate(farms.AddCrop))
	router.PUT("/api/farms/:id/crops/:cropid", middleware.Authenticate(farms.EditCrop))
	router.DELETE("/api/farms/:id/crops/:cropid", middleware.Authenticate(farms.DeleteCrop))
	router.PUT("/api/farms/:id/crops/:cropid/buy", middleware.Authenticate(farms.BuyCrop))
	router.GET("/api/crops", farms.GetFilteredCrops)

	router.GET("/api/crops/catalogue", farms.GetCropCatalogue)
	router.GET("/api/crops/precatalogue", farms.GetPreCropCatalogue)
	router.GET("/api/crops/types", farms.GetCropTypes)
	// router.GET("/api/crops/crop/:cropid", middleware.Authenticate(farms.GetCropFarms))
	router.GET("/api/crops/crop/:cropname", middleware.Authenticate(farms.GetCropTypeFarms))

	router.GET("/api/farm/items", farms.GetItems)
	router.GET("/api/farm/products", farms.GetProducts)
	router.GET("/api/farm/tools", farms.GetTools)
	router.GET("/api/farm/items/categories", farms.GetItemCategories)

	router.POST("/api/farm/product", farms.CreateProduct)
	router.PUT("/api/farm/product/:id", farms.UpdateProduct)
	router.DELETE("/api/farm/product/:id", farms.DeleteProduct)

	router.POST("/api/farm/tool", farms.CreateTool)
	router.PUT("/api/farm/tool/:id", farms.UpdateTool)
	router.DELETE("/api/farm/tool/:id", farms.DeleteTool)

	router.POST("/api/upload/images", utils.UploadImages)
}

func AddRecipeRoutes(router *httprouter.Router) {
	router.GET("/api/recipes/tags", middleware.Authenticate(recipes.GetRecipeTags))
	router.GET("/api/recipes", middleware.Authenticate(recipes.GetRecipes))
	router.GET("/api/recipes/recipe/:id", middleware.Authenticate(recipes.GetRecipe))
	router.POST("/api/recipes", middleware.Authenticate(recipes.CreateRecipe))
	router.PUT("/api/recipes/recipe/:id", middleware.Authenticate(recipes.UpdateRecipe))
	router.DELETE("/api/recipes/recipe/:id", middleware.Authenticate(recipes.DeleteRecipe))
}

func AddSuggestionsRoutes(router *httprouter.Router) {
	router.GET("/api/suggestions/follow", ratelim.RateLimit(middleware.Authenticate(suggestions.SuggestFollowers)))
}

func AddReviewsRoutes(router *httprouter.Router) {
	router.GET("/api/reviews/:entityType/:entityId", ratelim.RateLimit(middleware.Authenticate(reviews.GetReviews)))
	router.GET("/api/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.GetReview)))
	router.POST("/api/reviews/:entityType/:entityId", ratelim.RateLimit(middleware.Authenticate(reviews.AddReview)))
	router.PUT("/api/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.EditReview)))
	router.DELETE("/api/reviews/:entityType/:entityId/:reviewId", ratelim.RateLimit(middleware.Authenticate(reviews.DeleteReview)))
}

func AddProfileRoutes(router *httprouter.Router) {
	router.GET("/api/profile/profile", middleware.Authenticate(profile.GetProfile))
	router.PUT("/api/profile/edit", middleware.Authenticate(profile.EditProfile))
	router.PUT("/api/profile/avatar", middleware.Authenticate(profile.EditProfilePic))
	router.PUT("/api/profile/banner", middleware.Authenticate(profile.EditProfileBanner))
	router.DELETE("/api/profile/delete", middleware.Authenticate(profile.DeleteProfile))

	router.GET("/api/user/:username", ratelim.RateLimit(profile.GetUserProfile))
	router.PUT("/api/follows/:id", ratelim.RateLimit(middleware.Authenticate(profile.ToggleFollow)))
	router.DELETE("/api/follows/:id", ratelim.RateLimit(middleware.Authenticate(profile.ToggleUnFollow)))
	router.GET("/api/follows/:id/status", ratelim.RateLimit(middleware.Authenticate(profile.DoesFollow)))
	router.GET("/api/followers/:id", ratelim.RateLimit(middleware.Authenticate(profile.GetFollowers)))
	router.GET("/api/following/:id", ratelim.RateLimit(middleware.Authenticate(profile.GetFollowing)))

}

func AddUtilityRoutes(router *httprouter.Router, rateLimiter *ratelim.RateLimiter) {
	router.GET("/api/csrf", rateLimiter.Limit(middleware.Authenticate(utils.CSRF)))
}

func AddSearchRoutes(router *httprouter.Router) {
	router.GET("/api/ac", search.Autocompleter)
	router.GET("/api/search/:entityType", ratelim.RateLimit(search.SearchHandler))
	router.POST("/emitted", search.EventHandler)
}

func AddMiscRoutes(router *httprouter.Router) {
	// Example Routes
	// router.GET("/", ratelim.RateLimit(wrapHandler(proxyWithCircuitBreaker("frontend-service"))))

	// router.GET("/api/search/:entityType", ratelim.RateLimit(searchEvents))

	// router.POST("/api/check-file", rateLimiter.Limit(filecheck.CheckFileExists))
	// router.POST("/api/upload", rateLimiter.Limit(filecheck.UploadFile))
	// router.POST("/api/feed/remhash", rateLimiter.Limit(filecheck.RemoveUserFile))

	// router.POST("/agi/home_feed_section", ratelim.RateLimit(middleware.Authenticate(agi.GetHomeFeed)))
	// router.GET("/resize/:folder/*filename", cdn.ServeStatic)

}
