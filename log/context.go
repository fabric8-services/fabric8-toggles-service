package log

import (
	"context"
)

// extractIdentityID obtains the identity ID out of the authentication context
func extractIdentityID(ctx context.Context) (string, error) {
	//token := goajwt.ContextJWT(ctx)
	//if token == nil {
	//	return "", errors.New("Missing token")
	//}
	//id := token.Claims.(jwt.MapClaims)["sub"]
	//if id == nil {
	//	return "", errors.New("Missing sub")
	//}
	//return id.(string), nil
	return "", nil
}

// ExtractRequestID obtains the request ID either from a goa client or middleware
func ExtractRequestID(ctx context.Context) string {
	reqID := ""
	//reqID := middleware.ContextRequestID(ctx)
	//if reqID == "" {
	//	return client.ContextRequestID(ctx)
	//}

	return reqID
}
