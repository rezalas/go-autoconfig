package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go-autoconfig/internal/loader"
	"go-autoconfig/internal/render"
	"go-autoconfig/internal/validate"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	templates loader.TemplateRegistry
	checker   validate.DomainChecker
}

// New creates a Handler with the given template registry and domain checker.
func New(templates loader.TemplateRegistry, checker validate.DomainChecker) *Handler {
	return &Handler{templates: templates, checker: checker}
}

// RegisterRoutes iterates over all loaded ClientConfigs and registers a Gin
// route for each supported endpoint, choosing the correct HTTP method based on
// the endpoint path convention.
func RegisterRoutes(r *gin.Engine, configs []*loader.ClientConfig, registry loader.TemplateRegistry, checker validate.DomainChecker) {
	h := New(registry, checker)

	for _, cfg := range configs {
		for _, endpoint := range cfg.SupportedEndpoints {

			// Autodiscover uses POST; everything else uses GET.
			method := http.MethodGet
			if strings.Contains(strings.ToLower(endpoint), "autodiscover") {
				method = http.MethodPost
			}

			r.Handle(method, endpoint, func(c *gin.Context) {
				h.serve(c, cfg)
			})
		}
	}
}

// serve is the shared dispatch path for all registered vendor endpoints.
func (h *Handler) serve(c *gin.Context, cfg *loader.ClientConfig) {
	email := resolveEmail(c, cfg)
	if email == "" {
		c.String(http.StatusBadRequest, "email address is required")
		return
	}

	domain := domainFromEmail(email)
	if domain == "" {
		c.String(http.StatusBadRequest, "invalid email address")
		return
	}

	// Domain validation
	ok, err := h.checker.DomainExists(c.Request.Context(), domain)
	if err != nil {
		c.String(http.StatusInternalServerError, "domain lookup error")
		return
	}
	if !ok {
		c.String(http.StatusNotFound, "domain not supported")
		return
	}

	tmpl, err := h.templates.Get(cfg.TemplateFile)
	if err != nil {
		c.String(http.StatusInternalServerError, "template not available")
		return
	}

	contentType := contentTypeFor(cfg.Vendor)
	c.Status(http.StatusOK)
	c.Header("Content-Type", contentType)

	if err := render.Render(c.Writer, tmpl, cfg, email); err != nil {
		// Header already sent; log only.
		_ = c.Error(err)
	}
}

// resolveEmail extracts the email address from the request using the
// convention appropriate for each vendor.
func resolveEmail(c *gin.Context, cfg *loader.ClientConfig) string {
	vendor := strings.ToUpper(cfg.Vendor)

	switch vendor {
	case "MICROSOFT":
		// Autodiscover POX XML: <EMailAddress>user@domain.com</EMailAddress>
		var body struct {
			Request struct {
				EMailAddress string `xml:"EMailAddress"`
			} `xml:"Request"`
		}
		if err := c.ShouldBindXML(&body); err == nil && body.Request.EMailAddress != "" {
			return body.Request.EMailAddress
		}
		// Fall through to query param as a convenience for testing
		return c.Query("emailaddress")

	default:
		// Mozilla autoconfig and Apple mobileconfig both use a query param.
		return c.Query("emailaddress")
	}
}

func domainFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return ""
	}
	return strings.ToLower(parts[1])
}

func contentTypeFor(vendor string) string {
	switch strings.ToUpper(vendor) {
	case "APPLE":
		return "application/x-apple-aspen-config"
	default:
		return "text/xml; charset=utf-8"
	}
}
