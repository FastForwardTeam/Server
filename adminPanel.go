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
	"strings"

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
	h.HandleFunc("/admin/api/login", genToken)
	h.HandleFunc("/admin/api/reported", returnReported)
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
	w.WriteHeader(http.StatusOK)
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
func genToken(w http.ResponseWriter, r *http.Request) {

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
		"Username": input.Username,
	})

	tokenString, _ := token.SignedString(RSAprivateKey)

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, tokenString)

}

// Converts sql rows and coloumns into json string

func checkToken(tokenString string) (username string, err error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return RSApublicKey, nil
	})

	if err != nil {
		return "", err
	}

	if token.Valid {
		return claims["Username"].(string), nil
	}

	return "", errors.New("Token invalid")
}

func parseAuthHeader(r *http.Request) (username string, ok bool) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return "", false
	}
	reqToken = splitToken[1]
	username, err := checkToken(reqToken)
	if err != nil {
		return "", false
	}
	return username, true
}

func returnReported(w http.ResponseWriter, r *http.Request) {
	username, valid := parseAuthHeader(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Println(username + " accessed /admin/api/reported")
	jsonData, err := dbQueryReported()
	if err != nil {
		panic(err)
	}

	w.Write(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

}
