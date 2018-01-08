package token

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	authclient "github.com/fabric8-services/fabric8-auth/token"
	"github.com/fabric8-services/fabric8-toggles-service/auth/authservice"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-auth/log"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// tokenManagerConfiguration represents configuration needed to construct a token manager
type tokenManagerConfiguration interface {
	GetAuthServiceURL() string
	GetKeycloakDevModeURL() string
}

// Manager generate and find auth token information
type Manager interface {
	Locate(ctx context.Context) (uuid.UUID, error)
	ParseToken(ctx context.Context, tokenString string) (*TokenClaims, error)
	PublicKey(kid string) *rsa.PublicKey
	PublicKeys() []*rsa.PublicKey
	//IsServiceAccount(ctx context.Context) bool
}

// AuthorizationPayload represents an authz payload in the rpt token
type AuthorizationPayload struct {
	Permissions []Permissions `json:"permissions"`
}

// Permissions represents a "permissions" in the AuthorizationPayload
type Permissions struct {
	ResourceSetName *string `json:"resource_set_name"`
	ResourceSetID   *string `json:"resource_set_id"`
}
type PublicKey struct {
	KeyID string
	Key   *rsa.PublicKey
}
type tokenManager struct {
	publicKeysMap map[string]*rsa.PublicKey
	publicKeys    []*PublicKey
}

// TokenClaims represents access token claims
type TokenClaims struct {
	Name          string                `json:"name"`
	Username      string                `json:"preferred_username"`
	GivenName     string                `json:"given_name"`
	FamilyName    string                `json:"family_name"`
	Email         string                `json:"email"`
	Company       string                `json:"company"`
	SessionState  string                `json:"session_state"`
	Authorization *AuthorizationPayload `json:"authorization"`
	jwt.StandardClaims
}

// ParseToken parses token claims
func (mgm *tokenManager) ParseToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"]
		if !ok {
			log.Error(ctx, map[string]interface{}{}, "There is no 'kid' header in the token")
			return nil, errors.New("there is no 'kid' header in the token")
		}
		key := mgm.PublicKey(fmt.Sprintf("%s", kid))
		if key == nil {
			log.Error(ctx, map[string]interface{}{
				"kid": kid,
			}, "There is no public key with such ID")
			return nil, errors.New(fmt.Sprintf("there is no public key with such ID: %s", kid))
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(*TokenClaims)
	if token.Valid {
		return claims, nil
	}
	return nil, errors.WithStack(errors.New("token is not valid"))
}

func (mgm *tokenManager) Locate(ctx context.Context) (uuid.UUID, error) {
	token := goajwt.ContextJWT(ctx)
	if token == nil {
		return uuid.UUID{}, errors.New("Missing token") // TODO, make specific tokenErrors
	}
	id := token.Claims.(jwt.MapClaims)["sub"]
	if id == nil {
		return uuid.UUID{}, errors.New("Missing sub")
	}
	idTyped, err := uuid.FromString(id.(string))
	if err != nil {
		return uuid.UUID{}, errors.New("uuid not of type string")
	}
	return idTyped, nil
}

// PublicKey returns the public key by the ID
func (mgm *tokenManager) PublicKey(kid string) *rsa.PublicKey {
	return mgm.publicKeysMap[kid]
}

// PublicKeys returns all the public keys
func (mgm *tokenManager) PublicKeys() []*rsa.PublicKey {
	keys := make([]*rsa.PublicKey, 0, len(mgm.publicKeysMap))
	for _, key := range mgm.publicKeys {
		keys = append(keys, key.Key)
	}
	return keys
}

// NewManager returns a new token Manager for handling tokens
func NewManager(config tokenManagerConfiguration) (Manager, error) {
	// Load public keys from Auth service and add them to the manager
	tm := &tokenManager{
		publicKeysMap: map[string]*rsa.PublicKey{},
	}

	keysEndpoint := fmt.Sprintf("%s%s", config.GetAuthServiceURL(), authservice.KeysTokenPath())
	remoteKeys, err := authclient.FetchKeys(keysEndpoint)
	if err != nil {
		log.Error(nil, map[string]interface{}{
			"err":      err,
			"keys_url": keysEndpoint,
		}, "unable to load public keys from remote service")
		return nil, errors.New("unable to load public keys from remote service")
	}
	for _, remoteKey := range remoteKeys {
		tm.publicKeysMap[remoteKey.KeyID] = remoteKey.Key
		tm.publicKeys = append(tm.publicKeys, &PublicKey{KeyID: remoteKey.KeyID, Key: remoteKey.Key})
		log.Info(nil, map[string]interface{}{
			"kid": remoteKey.KeyID,
		}, "Public key added")
	}

	devModeURL := config.GetKeycloakDevModeURL()
	if devModeURL != "" {
		remoteKeys, err = authclient.FetchKeys(fmt.Sprintf("%s/protocol/openid-connect/certs", devModeURL))
		if err != nil {
			log.Error(nil, map[string]interface{}{
				"err":      err,
				"keys_url": devModeURL,
			}, "unable to load public keys from remote service in Dev Mode")
			return nil, errors.New("unable to load public keys from remote service in Dev Mode")
		}
		for _, remoteKey := range remoteKeys {
			tm.publicKeysMap[remoteKey.KeyID] = remoteKey.Key
			tm.publicKeys = append(tm.publicKeys, &PublicKey{KeyID: remoteKey.KeyID, Key: remoteKey.Key})
			log.Info(nil, map[string]interface{}{
				"kid": remoteKey.KeyID,
			}, "Public key added")
		}
		// Add the public key which will be used to verify tokens generated in dev mode
		rsaKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(configuration.DevModeRsaPrivateKey))
		if err != nil {
			return nil, err
		}
		tm.publicKeysMap["test-key"] = &rsaKey.PublicKey
		tm.publicKeys = append(tm.publicKeys, &PublicKey{KeyID: "test-key", Key: &rsaKey.PublicKey})
		log.Info(nil, map[string]interface{}{
			"kid": "test-key",
		}, "Public key added")
	}

	return tm, nil
}
