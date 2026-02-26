package core

import "github.com/gofiber/fiber/v2"

// Guard is the contract for authentication/authorization middleware providers
// (e.g. ss-keel-jwt, ss-keel-oauth).
//
// Usage:
//
//	route.Use(jwtGuard.Middleware()).WithSecured("bearerAuth")
type Guard interface {
	Middleware() fiber.Handler
}
