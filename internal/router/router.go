package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/controller"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, ctrl *controller.Controller) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(cfg))
	r.Use(securityHeaders)

	// Buckets sized to the cost of each endpoint: signing in and paying are
	// expensive and abusable, browsing the menu is neither.
	authLimiter := middleware.NewRateLimiter(10, 20) // 10 burst, 20/min
	orderLimiter := middleware.NewRateLimiter(8, 30)
	geocodeLimiter := middleware.NewRateLimiter(15, 40)

	r.Get("/api/v1/health", ctrl.Health)

	// The WebSocket route is registered before the request-timeout middleware
	// is applied: a 30s timeout would sever every live order feed, and the
	// timeout handler's ResponseWriter cannot be hijacked for the upgrade.
	r.Get("/api/v1/ws", ctrl.HandleWebSocket)

	// Uploaded menu images.
	_ = os.MkdirAll(cfg.UploadDir, 0o755)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/",
		uploadFileServer(http.Dir(cfg.UploadDir))))

	// Every route below is plain request/response, so a hard timeout is safe.
	r.Group(func(rest chi.Router) {
		rest.Use(chimw.Timeout(30 * time.Second))
		registerRESTRoutes(rest, cfg, ctrl, authLimiter, orderLimiter, geocodeLimiter)
	})

	return r
}

func registerRESTRoutes(
	r chi.Router,
	cfg *config.Config,
	ctrl *controller.Controller,
	authLimiter, orderLimiter, geocodeLimiter *middleware.RateLimiter,
) {
	// ---- Public: storefront reads and the payment gateway callback ----
	r.Group(func(public chi.Router) {
		public.Get("/api/v1/branches", ctrl.ListBranches)
		public.Get("/api/v1/branches/{slug}", ctrl.GetBranch)
		public.Get("/api/v1/categories", ctrl.ListCategories)
		public.Get("/api/v1/branches/{branch}/menu", ctrl.ListMenuByBranch)

		// Midtrans posts here; the SHA-512 signature is the authentication.
		public.Post("/api/v1/payments/notification", ctrl.HandleMidtransNotification)
	})

	r.Group(func(limited chi.Router) {
		limited.Use(authLimiter.Middleware)
		limited.Post("/api/v1/auth/customer-login", ctrl.CustomerPhoneLogin)
		limited.Post("/api/v1/auth/admin-login", ctrl.AdminLogin)
	})

	// Quoting and geocoding are public (the customer has not signed in while
	// browsing) but rate limited, because each geocode call hits a third party.
	r.Group(func(quote chi.Router) {
		quote.Use(geocodeLimiter.Middleware)
		quote.Post("/api/v1/delivery/quote", ctrl.QuoteDelivery)
		quote.Get("/api/v1/geocode/search", ctrl.SearchAddress)
		quote.Get("/api/v1/geocode/reverse", ctrl.ReverseGeocode)
		quote.Post("/api/v1/promos/validate", ctrl.ValidatePromo)
	})

	// ---- Checkout: the order itself allows guests, everything afterwards
	//      requires the session created during checkout. These routes
	//      previously had no auth middleware at all, so any caller could
	//      read or cancel any order by id.
	r.Group(func(checkout chi.Router) {
		checkout.Use(orderLimiter.Middleware)
		checkout.Use(middleware.OptionalAuth(cfg))
		checkout.Post("/api/v1/orders", ctrl.CreateOrder)
	})

	r.Group(func(customer chi.Router) {
		customer.Use(middleware.AuthRequired(cfg))
		customer.Get("/api/v1/auth/me", ctrl.GetCurrentUser)
		customer.Get("/api/v1/orders", ctrl.ListMyOrders)
		customer.Get("/api/v1/orders/{id}", ctrl.GetOrder)
		customer.Post("/api/v1/orders/{id}/cancel", ctrl.CancelOrder)
		customer.Get("/api/v1/orders/{id}/payment", ctrl.GetPaymentStatus)
	})

	r.Group(func(pay chi.Router) {
		pay.Use(orderLimiter.Middleware)
		pay.Use(middleware.AuthRequired(cfg))
		pay.Use(middleware.Idempotency)
		pay.Post("/api/v1/payments", ctrl.CreatePayment)
	})

	// ---- Branch admin & owner ----
	r.Group(func(admin chi.Router) {
		admin.Use(middleware.AuthRequired(cfg, "branch_admin", "owner"))

		admin.Get("/api/v1/admin/dashboard", ctrl.BranchDashboard)

		admin.Get("/api/v1/admin/orders", ctrl.ListOrders)
		admin.Get("/api/v1/admin/orders/{id}", ctrl.GetOrder)
		admin.Put("/api/v1/admin/orders/{id}/status", ctrl.UpdateOrderStatus)
		admin.Put("/api/v1/admin/orders/{id}/driver", ctrl.AssignDriver)
		admin.Post("/api/v1/admin/orders/acknowledge", ctrl.AcknowledgeOrders)

		admin.Post("/api/v1/admin/menu", ctrl.CreateMenuItem)
		admin.Put("/api/v1/admin/menu/{id}", ctrl.UpdateMenuItem)
		admin.Put("/api/v1/admin/menu/{id}/availability", ctrl.ToggleMenuItemAvailability)
		admin.Delete("/api/v1/admin/menu/{id}", ctrl.DeleteMenuItem)

		admin.Get("/api/v1/admin/settings", ctrl.GetBranchSettings)
		admin.Put("/api/v1/admin/settings", ctrl.UpdateBranchSettings)
		admin.Put("/api/v1/admin/branches/{id}/status", ctrl.UpdateBranchStatus)

		admin.Post("/api/v1/admin/upload", ctrl.UploadImage)
	})

	// ---- Owner HQ ----
	r.Group(func(owner chi.Router) {
		owner.Use(middleware.AuthRequired(cfg, "owner"))
		owner.Get("/api/v1/owner/dashboard", ctrl.OwnerDashboard)
		owner.Get("/api/v1/owner/orders", ctrl.ListOrders)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// uploadFileServer serves menu images. Directory listings are disabled and
// only image extensions are served, so an uploaded file can never be handed
// back as HTML or a script.
func uploadFileServer(root http.FileSystem) http.Handler {
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".svg": false,
	}
	fs := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := filepath.Clean("/" + r.URL.Path)
		if strings.HasSuffix(clean, "/") || !allowed[strings.ToLower(filepath.Ext(clean))] {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fs.ServeHTTP(w, r)
	})
}
