/*
Copyright 2021, 2022 NotAProton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//TODO: add env var to check if signup is allowed
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v4"
)

var (
	RSAprivateKey *rsa.PrivateKey
	RSApublicKey  *rsa.PublicKey
)

func loadRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Get private key

	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, nil, errors.New("failed to parse env var containing the private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// Get public key

	block, _ = pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, nil, errors.New("failed to parse env var containing the Public key")
	}

	parsedPub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	pub := parsedPub.(*rsa.PublicKey)
	return priv, pub, nil
}

//Returns Refresh JWT
func genRefToken(username string) (string, error) {
	uuid := uuid.New().String()

	//exp is 60 days, TODO make it configureable
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "ref_token",
		"aud": username,
		"jti": uuid,
		"exp": time.Now().Add(time.Hour * 1440).Unix(),
	})

	err := dbAdminRefTokenInsert(username, uuid)
	if err != nil {
		return "", err
	}
	tokenString, err := token.SignedString(RSAprivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Takes refresh token, verifies and returns username and access token
func genAccessToken(refToken string) (username string, accToken string, err error) {
	token, err := jwt.Parse(refToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("Unexpected signing method:" + token.Header["alg"].(string))
		}
		return RSApublicKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["sub"].(string) != "ref_token" {
			return claims["aud"].(string), "", errors.New("not refresh token")
		}

		username := claims["aud"].(string)
		uuid := claims["jti"].(string)

		ok, tokenInDB, err := dbAdminRefTokenQuery(username)

		if err != nil {
			return "", "", err
		}

		if !ok {
			return claims["aud"].(string), "", errors.New("user not found")
		}

		if uuid != tokenInDB {
			return claims["aud"].(string), "", errors.New("token mismatch")
		}

		accToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "acc_token",
			"aud": username,
			"exp": time.Now().Add(time.Minute * 15).Unix(),
		})

		accTokenString, err := accToken.SignedString(RSAprivateKey)
		if err != nil {
			return claims["aud"].(string), "", err
		}

		return claims["aud"].(string), accTokenString, nil
	} else {
		return claims["aud"].(string), "", err
	}
}

// Takes Access token, verifies and parses it, returns username and error
func parseAccessToken(accToken string) (username string, err error) {
	token, err := jwt.Parse(accToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("Unexpected signing method:" + token.Header["alg"].(string))
		}
		return RSApublicKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["sub"].(string) != "acc_token" {
			return "", errors.New("not access token")
		}
		username := claims["aud"].(string)
		return username, nil
	} else {
		return "", err
	}

}
