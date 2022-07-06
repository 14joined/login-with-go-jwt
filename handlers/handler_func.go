package handlers

import (
	"log"
	"fmt"
	"time"
	"strings"
	"net/http"
	"html/template"
	"encoding/json"
	"path/filepath"
	jwt "github.com/golang-jwt/jwt/v4"
)

var SecretJWTKey = []byte("s3cr3tjwtk3y")
var Users = map[string]string {
	"brownfox": "s3cr3t",
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims 
}

type Error struct {
	Message string `json:"message"`
	Status int `json:"status"`	 
}

type AuthenReponse struct {
	Token string `json:"token"`
	Status int `json:"status"`
}

func GenToken(username string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(120 * time.Second)

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "http://my-domain.superserv.org",
			Subject: "ics|12345",
			Audience: jwt.ClaimStrings{
				"gojwtapp",
			},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt: jwt.NewNumericDate(now),
			ID: username,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return accessToken.SignedString(SecretJWTKey)
}

func MainRoute(files string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[1:]
		if name != "" {
			tokenHeader := r.Header.Get("Authorization")

			if tokenHeader == "" {
				ResponseErr(w, http.StatusForbidden)
				return
			}

			splitted := strings.Split(tokenHeader, " ") // Bearer jwt_token
			if len(splitted) != 2 {
				ResponseErr(w, http.StatusForbidden)
				return
			}

			tokenPart := splitted[1]
			tk := &Claims{}

			token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
				return SecretJWTKey, nil
			})

			if err != nil {
				fmt.Println(err)
				ResponseErr(w, http.StatusInternalServerError)
				return
			}

			if token.Valid {
				log.Println(tokenPart)
				log.Println(token.Claims)
				t, err := template.ParseFiles("public/profile.gohtml")
				
				if err != nil {
					w.Write([]byte("can not parse template"))
				}
				
				data := struct {
					Username string
				}{
					Username: name,
				}

				err = t.Execute(w, data)

				if err != nil {
					w.Write([]byte("error execute from template"))
				}
			}
			return
		}
		http.ServeFile(w, r, filepath.Join(files, "index.html"))
	}
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
				// http.SetCookie(w, &http.Cookie{
				// 	Name:    "token",
				// 	Value:   token,
				// 	Expires: time.Now().Add(120 * time.Second),
				// })
				ResponseOk(w, AuthenReponse{
					Token: token,
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

