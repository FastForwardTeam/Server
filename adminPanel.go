//TODO: add env var to check if signup is allowed
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	RSAprivateKey *rsa.PrivateKey
	RSApublicKey  *rsa.PublicKey
)

func adminPanelRouters(h *http.ServeMux) {
	fs := http.FileServer(http.Dir(panelDir))
	h.Handle("/admin/", http.StripPrefix("/admin/", fs))

	h.HandleFunc("/admin/api/signup", signUp)
	h.HandleFunc("/admin/api/login", authToken)
}

// Inserts creds into db
func signUp(w http.ResponseWriter, r *http.Request) {

	type Input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input Input

	// decode input
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbAdminCredsInsert(input.Username, string(hashedPassword))
}

func loadRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Get private key

	privPEM, err := os.ReadFile("./key.rsa")
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, nil, errors.New("failed to parse PEM file containing the private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// Get public key

	pubPEM, err := os.ReadFile("./key.rsa.pub")
	if err != nil {
		return nil, nil, err
	}

	block, _ = pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, nil, errors.New("failed to parse PEM block containing the Public key")
	}

	parsedPub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	pub := parsedPub.(*rsa.PublicKey)
	return priv, pub, nil
}

// returns JWT
func authToken(w http.ResponseWriter, r *http.Request) {

	type Input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input Input

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, hashedpassword := dbAdminCredsQuery(input.Username)

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(input.Password))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": input.Username,
	})

	// The claims object allows you to store information in the actual token.

	tokenString, _ := token.SignedString(RSAprivateKey)

	// tokenString Contains the actual token you should share with your client.

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, tokenString)

}
