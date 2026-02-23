package core

import "github.com/slicesoft/ss-keel-core/openapi"

func toOpenAPIRoutes(routes []Route) []openapi.RouteInput {
	var out []openapi.RouteInput
	for _, r := range routes {
		ri := openapi.RouteInput{
			Method:      r.method,
			Path:        r.path,
			Summary:     r.summary,
			Description: r.description,
			Tags:        r.tags,
			Secured:     r.secured,
		}
		if r.body != nil {
			ri.Body = r.body.Type
		}
		if r.response != nil {
			ri.Response = r.response.Type
			ri.StatusCode = r.response.StatusCode
		}
		out = append(out, ri)
	}
	return out
}
