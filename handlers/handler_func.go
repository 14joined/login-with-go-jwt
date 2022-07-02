package handlers

import (
	"log"
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	jwt "github.com/golang-jwt/jwt/v4"
)

var SecretJWTKey = []byte("")
var Users = map[string]string {
	"brownfox": "s3cr3t",
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims 
}

type Error struct {
	Message string `json:"message"`
	Status int `json:"status"`	 
}

type AuthenReponse struct {
	AccessToken string `json:"accessToken"`
	Status int `json:"status"`
}

func GenToken(username string) (string, error) {
	expirationTime := time.Now().Add(120 * time.Second)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return accessToken.SignedString(SecretJWTKey)
}

func RequestAccess() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return 
		}

		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")
		log.Printf("user %s, passwd %s\n", username, password);

		secretPwd, ok := Users[username]
		log.Println(ok)
		if ok {
			if password == secretPwd {
				var token string
				token, err := GenToken(username)
				
				if err != nil {
					log.Println(1)
					ResponseErr(w, http.StatusInternalServerError)
					return
				}
				// Set the new token as the users `token` cookie
				http.SetCookie(w, &http.Cookie{
					Name:    "token",
					Value:   token,
					Expires: time.Now().Add(120 * time.Second),
				})
				ResponseOk(w, AuthenReponse{
					AccessToken: token,
					Status: http.StatusOK,
				})
				return 
			} else {
				ResponseErr(w, http.StatusUnauthorized)
			}
		} else {
			ResponseErr(w, http.StatusUnauthorized)
		}
	}
}

func ResponseErr(w http.ResponseWriter, statusCode int) {
	j_data, err := json.Marshal(Error{
		Status: statusCode,
		Message: http.StatusText(statusCode),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	log.Println("Sent response err")
	w.Write(j_data)
}

func ResponseOk(w http.ResponseWriter, data interface{}) {
	if data == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	j_data, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	log.Println("Sent response ok")
	w.Write(j_data)
}

