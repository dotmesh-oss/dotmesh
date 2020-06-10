package auth

import (
	"context"
	"net/http"

	"github.com/dotmesh-oss/dotmesh/pkg/user"
)

const authenticationUserIDContextKey = "authenticated-user-id"
const authenticationUserObjectContextKey = "authenticated-user-object"
const authenticationPasswordAuthContextKey = "authentication-type"

// GetUserID - gets current user ID from this request
func GetUserID(r *http.Request) (id string) {
	if accountID := r.Context().Value(authenticationUserIDContextKey); accountID != nil {
		return accountID.(string)
	}
	return ""
}

// GetUserIDFromCtx - get user ID from context
func GetUserIDFromCtx(ctx context.Context) string {
	if u := ctx.Value(authenticationUserIDContextKey); u != nil {
		return u.(string)
	}
	return ""
}

// GetUserFromCtx- get user from given context
func GetUserFromCtx(ctx context.Context) *user.User {
	if u := ctx.Value(authenticationUserObjectContextKey); u != nil {
		return u.(*user.User)
	}
	return nil
}

// SetUserIDCtx - set user ID to given context
func SetUserIDCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, authenticationUserIDContextKey, id)
}

// GetUser - gets user object from request context
func GetUser(r *http.Request) *user.User {
	return GetUserFromCtx(r.Context())
}

// GetAuthenticationType - get authentication type
func GetAuthenticationType(r *http.Request) user.AuthenticationType {
	if at := r.Context().Value(authenticationPasswordAuthContextKey); at != nil {
		return at.(user.AuthenticationType)
	}
	return user.AuthenticationTypeNone
}

// SetAuthenticationDetails sets user details for this request
func SetAuthenticationDetails(r *http.Request, user *user.User, authenticationType user.AuthenticationType) *http.Request {
	return r.WithContext(SetAuthenticationDetailsCtx(r.Context(), user, authenticationType))
}

// SetAuthenticationDetailsCtx set user details to the given context and return a new ctx
func SetAuthenticationDetailsCtx(ctx context.Context, user *user.User, authenticationType user.AuthenticationType) context.Context {
	ctx = context.WithValue(ctx, authenticationUserIDContextKey, user.Id)
	ctx = context.WithValue(ctx, authenticationUserObjectContextKey, user)
	return context.WithValue(ctx, authenticationPasswordAuthContextKey, authenticationType)
}
