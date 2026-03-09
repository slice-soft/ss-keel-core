package core

import (
	"strings"

	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/openapi"
)

// toBuildInput maps App configuration and routes to the OpenAPI BuildInput structure.
func toBuildInput(cfg KConfig, routes []httpx.Route) openapi.BuildInput {
	bi := openapi.BuildInput{
		Title:       cfg.Docs.Title,
		Version:     cfg.Docs.Version,
		Description: cfg.Docs.Description,
		Routes:      toOpenAPIRoutes(routes),
	}
	if cfg.Docs.Contact != nil {
		bi.Contact = &openapi.Contact{
			Name:  cfg.Docs.Contact.Name,
			URL:   cfg.Docs.Contact.URL,
			Email: cfg.Docs.Contact.Email,
		}
	}
	if cfg.Docs.License != nil {
		bi.License = &openapi.License{
			Name: cfg.Docs.License.Name,
			URL:  cfg.Docs.License.URL,
		}
	}
	for _, s := range cfg.Docs.Servers {
		parts := strings.SplitN(s, " - ", 2)
		si := openapi.ServerInfo{URL: parts[0]}
		if len(parts) == 2 {
			si.Description = parts[1]
		}
		bi.Servers = append(bi.Servers, si)
	}
	for _, tag := range cfg.Docs.Tags {
		bi.Tags = append(bi.Tags, openapi.TagInfo{Name: tag.Name, Description: tag.Description})
	}
	return bi
}

// toOpenAPIRoutes converts internal Route objects to OpenAPI RouteInput format.
func toOpenAPIRoutes(routes []httpx.Route) []openapi.RouteInput {
	var out []openapi.RouteInput
	for _, r := range routes {
		ri := openapi.RouteInput{
			Method:      r.Method(),
			Path:        r.Path(),
			Summary:     r.Summary(),
			Description: r.Description(),
			Tags:        r.Tags(),
			Secured:     r.Secured(),
			Deprecated:  r.Deprecated(),
		}
		if r.Body() != nil {
			ri.Body = r.Body().Type
		}
		if r.Response() != nil {
			ri.Response = r.Response().Type
			ri.StatusCode = r.Response().StatusCode
		}
		for _, qp := range r.QueryParams() {
			ri.QueryParams = append(ri.QueryParams, openapi.QueryParamInput{
				Name:        qp.Name,
				Type:        qp.Type,
				Description: qp.Description,
				Required:    qp.Required,
			})
		}
		out = append(out, ri)
	}
	return out
}
