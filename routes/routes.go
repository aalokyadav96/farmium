package routes

import (
	"naevis/auth"
	"naevis/cart"
	"naevis/farms"
	"naevis/home"
	"naevis/middleware"
	"naevis/profile"
	"naevis/ratelim"
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

func AddHomeRoutes(router *httprouter.Router) {
	router.GET("/api/home/:apiRoute", middleware.Authenticate(home.GetHomeContent))
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

}

func RegisterCropRoutes(router *httprouter.Router) {
	// router.GET("/api/crops", middleware.Authenticate(crops.GetCrops))
	// router.GET("/api/crops/:id", middleware.Authenticate(crops.GetCropByID))
	// router.GET("/api/crops/:id/sellers", middleware.Authenticate(crops.GetCropSellers))
	// router.GET("/api/farms/:id", middleware.Authenticate(crops.GetFarmByID))
	// router.GET("/api/farms/:id/crops", middleware.Authenticate(crops.GetCropsByFarm))
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
