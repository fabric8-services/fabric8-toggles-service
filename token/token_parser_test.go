package token_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-auth/log"
	"github.com/fabric8-services/fabric8-toggles-service/auth"
	testsupport "github.com/fabric8-services/fabric8-toggles-service/test"
	"github.com/fabric8-services/fabric8-toggles-service/test/recorder"
	"github.com/fabric8-services/fabric8-toggles-service/token"
	errs "github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

type ParserConfig struct {
}

func TestParseToken(t *testing.T) {
	if log.IsDebug() {
		jsonkey, err := generateJSONWebKey()
		require.NoError(t, err)
		log.Debug(nil, map[string]interface{}{"json_key": jsonkey}, "ensure that this data is in the recorded server response!")
	}
	//given
	r, err := recorder.New("../test/data/token/auth_get_keys")
	require.NoError(t, err)
	c, err := auth.NewClient(
		context.Background(),
		"http://authservice",
		auth.WithHTTPClient(
			&http.Client{
				Transport: r.Transport,
			}),
	)
	require.NoError(t, err)
	p, err := token.NewParser(c)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		// given a token that will expire in 1hr
		raw, err := generateRawToken("../test/private_key.pem", "foo", time.Now().Add(1*time.Hour))
		require.NoError(t, err)
		// when
		result, err := p.Parse(context.Background(), *raw)
		// then
		require.NoError(t, err)
		claims := result.Claims.(jwt.MapClaims)
		assert.Equal(t, "foo", claims["sub"])
	})

	t.Run("invalid token", func(t *testing.T) {

		t.Run("wrong signing key", func(t *testing.T) {
			// given a token that will expire in 1hr, but signed with another key
			raw, err := generateRawToken("../test/private_key2.pem", "foo", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			// when
			_, err = p.Parse(context.Background(), *raw)
			// then parsing should fail because the private key used to sign the token as no known/loaded public counterpart
			require.Error(t, err)
		})

		t.Run("expired token", func(t *testing.T) {
			// given a token that expired 1 hr ago
			raw, err := generateRawToken("../test/private_key.pem", "foo", time.Now().Add(-1*time.Hour))
			require.NoError(t, err)
			// when
			_, err = p.Parse(context.Background(), *raw)
			// then parsing should fail because the private key used to sign the token as no known/loaded public counterpart
			require.Error(t, err)
		})
	})
}

func TestPublicKeys(t *testing.T) {
	//given
	r, err := recorder.New("../test/data/token/auth_get_keys")
	require.NoError(t, err)
	c, err := auth.NewClient(
		context.Background(),
		"http://authservice",
		auth.WithHTTPClient(
			&http.Client{
				Transport: r.Transport,
			}),
	)
	require.NoError(t, err)
	p, err := token.NewParser(c)
	require.NoError(t, err)
	// when
	result := p.PublicKeys()
	// then
	assert.Len(t, result, 3)
}

func generateRawToken(filename, subject string, exp time.Time) (*string, error) {
	claims := jwt.MapClaims{}
	claims["sub"] = subject
	claims["exp"] = exp.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	// use the test private key to sign the token
	key, err := testsupport.PrivateKey(filename)
	if err != nil {
		return nil, errs.Wrap(err, "failed to generate token")
	}
	token.Header["kid"] = "test_key"
	signed, err := token.SignedString(key)
	if err != nil {
		return nil, errs.Wrap(err, "failed to generate token")
	}
	log.Debug(nil, map[string]interface{}{"raw_token": signed}, "token generated")
	return &signed, nil
}

// utility function to generate the content to put in the `test/data/token/auth_get_keys.yaml` file
func generateJSONWebKey() (interface{}, error) {
	publickey, err := testsupport.PublicKey("../test/public_key.pem")
	if err != nil {
		return nil, err
	}
	key := token.PublicKey{
		KeyID: "foo",
		Key:   publickey,
	}
	jwk := jose.JSONWebKey{Key: key.Key, KeyID: key.KeyID, Algorithm: "RS256", Use: "sig"}
	keyData, err := jwk.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var raw interface{}
	err = json.Unmarshal(keyData, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
