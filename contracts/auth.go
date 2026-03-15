package contracts

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

// TokenSigner signs a JWT for an authenticated user.
// Implemented by ss-keel-jwt; any custom implementation also works.
//
// subject is a unique identifier for the user, typically formatted as
// "<provider>:<user-id>" (e.g. "google:1234567890") or simply a user ID.
// data holds arbitrary claims that will be embedded in the token payload.
//
// Usage in ss-keel-oauth:
//
//	oauth.Config{Signer: jwtProvider}
type TokenSigner interface {
	Sign(subject string, data map[string]any) (string, error)
}
