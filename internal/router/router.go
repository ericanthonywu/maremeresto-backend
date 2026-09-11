package router

import (
	"net/http"
	"os"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/controller"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, ctrl *controller.Controller) chi.Router {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(cfg))

	// WebSocket endpoint
	r.Get("/api/v1/ws", ctrl.HandleWebSocket)

	// Static uploaded files
	_ = os.MkdirAll(cfg.UploadDir, 0755)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Public API routes
	r.Group(func(public chi.Router) {
		public.Post("/api/v1/auth/customer-login", ctrl.CustomerPhoneLogin)
		public.Post("/api/v1/auth/admin-login", ctrl.AdminLogin)

		public.Get("/api/v1/branches", ctrl.ListBranches)
		public.Get("/api/v1/branches/{slug}", ctrl.GetBranch)
		public.Get("/api/v1/categories", ctrl.ListCategories)
		public.Get("/api/v1/branches/{branch}/menu", ctrl.ListMenuByBranch)

		public.Post("/api/v1/promos/validate", ctrl.ValidatePromo)

		// Public order access for tracking by ID (even before login/guest tracking)
		public.Get("/api/v1/orders/{id}", ctrl.GetOrder)

		// Midtrans Webhook Notification
		public.Post("/api/v1/payments/notification", ctrl.HandleMidtransNotification)
	})

	// Authenticated customer/guest order creation
	r.Group(func(auth chi.Router) {
		auth.Post("/api/v1/orders", ctrl.CreateOrder)
		auth.Post("/api/v1/orders/{id}/cancel", ctrl.CancelOrder)
		auth.Post("/api/v1/payments", ctrl.CreatePayment)
		auth.Get("/api/v1/auth/me", ctrl.GetCurrentUser)
	})

	// Branch Admin / Owner Routes
	r.Group(func(admin chi.Router) {
		admin.Use(middleware.AuthRequired(cfg, "branch_admin", "owner"))

		admin.Get("/api/v1/admin/dashboard", ctrl.BranchDashboard)
		admin.Get("/api/v1/admin/orders", ctrl.ListOrders)
		admin.Put("/api/v1/admin/orders/{id}/status", ctrl.UpdateOrderStatus)

		admin.Post("/api/v1/admin/menu", ctrl.CreateMenuItem)
		admin.Put("/api/v1/admin/menu/{id}", ctrl.UpdateMenuItem)
		admin.Put("/api/v1/admin/menu/{id}/availability", ctrl.ToggleMenuItemAvailability)
		admin.Delete("/api/v1/admin/menu/{id}", ctrl.DeleteMenuItem)

		admin.Get("/api/v1/admin/settings", ctrl.GetBranchSettings)
		admin.Put("/api/v1/admin/settings", ctrl.UpdateBranchSettings)
		admin.Put("/api/v1/admin/branches/{id}/status", ctrl.UpdateBranchStatus)

		admin.Post("/api/v1/admin/upload", ctrl.UploadImage)
	})

	// Owner HQ Routes
	r.Group(func(owner chi.Router) {
		owner.Use(middleware.AuthRequired(cfg, "owner"))

		owner.Get("/api/v1/owner/dashboard", ctrl.OwnerDashboard)
		owner.Get("/api/v1/owner/orders", ctrl.ListOrders)
	})

	return r
}
