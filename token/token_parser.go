package token

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-auth/log"
	"github.com/fabric8-services/fabric8-auth/token"
	authclient "github.com/fabric8-services/fabric8-toggles-service/auth/client"
	"github.com/fabric8-services/fabric8-wit/rest"
	errs "github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

// NewParser initializes a new token parser
func NewParser(authClient *authclient.Client) (token.Parser, error) {
	p := parserImpl{
		authClient: authClient,
		publicKeys: make(map[string]*rsa.PublicKey),
	}
	err := p.loadKeys(context.Background())
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// parserImpl the actual Parser implementation
type parserImpl struct {
	authClient *authclient.Client
	publicKeys map[string]*rsa.PublicKey
}

func (p *parserImpl) Parse(ctx context.Context, raw string) (*jwt.Token, error) {
	keyFunc := p.keyFunction(ctx)
	jwtToken, err := jwt.Parse(raw, keyFunc)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err,
		}, "unable to parse the token")
		return nil, errs.Wrapf(err, "unable to parse the token")
	}
	return jwtToken, nil
}

func (p *parserImpl) PublicKeys() []*rsa.PublicKey {
	keys := make([]*rsa.PublicKey, 0)
	for _, key := range p.publicKeys {
		keys = append(keys, key)
	}
	return keys
}

func (p *parserImpl) keyFunction(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		kid := token.Header["kid"]
		if kid == nil {
			log.Error(ctx, map[string]interface{}{}, "there is no 'kid' header in the token")
			return nil, errs.New("there is no 'kid' header in the token")
		}
		key, found := p.publicKeys[fmt.Sprintf("%s", kid)]
		if !found {
			log.Error(ctx, map[string]interface{}{
				"kid": fmt.Sprintf("%s", kid),
			}, "there is no public key with such an ID")
			return nil, errs.Errorf("there is no public key with such an ID: %s", kid)
		}
		return key, nil
	}
}

func (p *parserImpl) loadKeys(ctx context.Context) error {
	res, err := p.authClient.KeysToken(ctx, authclient.KeysTokenPath(), nil)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err.Error(),
		}, "unable to get public keys from the auth service")
		return errs.Wrap(err, "unable to get public keys from the auth service")
	}
	defer res.Body.Close()
	bodyString := rest.ReadBody(res.Body)
	if res.StatusCode != http.StatusOK {
		if err != nil {
			log.Error(ctx, map[string]interface{}{
				"err": err.Error(),
			}, "unable to read public keys from the auth service")
			return errs.Wrap(err, "unable to read public keys from the auth service")
		}
	}
	keys, err := unmarshalKeys([]byte(bodyString))
	if err != nil {
		return errs.Wrapf(err, "unable to load keys from auth service")
	}
	log.Info(nil, map[string]interface{}{
		"url":            authclient.KeysTokenPath(),
		"number_of_keys": len(keys),
	}, "Public keys loaded")
	for _, k := range keys {
		p.publicKeys[k.KeyID] = k.Key
	}
	return nil
}

// PublicKey a public key loaded from auth service
type PublicKey struct {
	KeyID string
	Key   *rsa.PublicKey
}

// JSONKeys the JSON structure for unmarshalling the keys
type JSONKeys struct {
	Keys []interface{} `json:"keys"`
}

func unmarshalKeys(jsonData []byte) ([]*PublicKey, error) {
	var keys []*PublicKey
	var raw JSONKeys
	err := json.Unmarshal(jsonData, &raw)
	if err != nil {
		return nil, err
	}
	for _, key := range raw.Keys {
		jsonKeyData, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		publicKey, err := unmarshalKey(jsonKeyData)
		if err != nil {
			return nil, err
		}
		keys = append(keys, publicKey)
	}
	return keys, nil
}

func unmarshalKey(jsonData []byte) (*PublicKey, error) {
	var key *jose.JSONWebKey
	key = &jose.JSONWebKey{}
	err := key.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.Key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Key is not an *rsa.PublicKey")
	}
	log.Info(nil, map[string]interface{}{"key_id": key.KeyID}, "unmarshalled public key")
	return &PublicKey{
			KeyID: key.KeyID,
			Key:   rsaKey},
		nil
}
