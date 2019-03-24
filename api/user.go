package api

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var hmacSecret = []byte(getSecret("JWT_SECRET", "not secure!"))
var users = make(map[uuid.UUID]*User)

type User struct {
	UUID uuid.UUID
	Room string
}

func createUser() *User {
	u := &User{
		UUID: uuid.New(),
	}
	if _, exists := users[u.UUID]; exists == true {
		log.Error("A new UUID was created but it already exists!")
	}
	users[u.UUID] = u
	return u
}

func (user *User) updateCookie(w http.ResponseWriter, r *http.Request) {
	existing := ""
	if c, err := r.Cookie("identity"); err != nil && c != nil {
		existing = c.Value
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"UUID": user.UUID,
		"Room": user.Room,
	})
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		log.Error(err)
	} else if tokenString != existing {
		// don't re-send the cookie header if the value hasn't changed
		http.SetCookie(w, &http.Cookie{
			Name:   "identity",
			Value:  tokenString,
			MaxAge: math.MaxInt32,
		})
	}
}

func UserMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		valid := false
		c, err := r.Cookie("identity")
		if err == nil {
			user, err := getValidUser(c.Value)
			if err == nil {
				ctx = context.WithValue(r.Context(), "user", user)
				log.Info("Matched existing user:", user.UUID)
				valid = true
			} else {
				log.Error(err)
			}
		}

		if valid == false {
			user := createUser()
			user.updateCookie(w, r)
			ctx = context.WithValue(r.Context(), "user", user)
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getSecret(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
		log.Warnf("No %s supplied, falling back to _insecure_ default value!", key)
	}
	return value
}

func getValidUser(tokenString string) (*User, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if idString, ok := claims["UUID"].(string); ok {
			if id, err := uuid.Parse(idString); err != nil {
				return nil, fmt.Errorf("User submitted valid jwt, but could not parse UUID")
			} else {
				if u, exists := users[id]; exists == true {
					return u, nil
				}
			}
		} else {
			log.Error("Could not get UUID type")
		}

		return nil, fmt.Errorf("User submitted valid jwt, but did did not exist in users register")
	}
	return nil, fmt.Errorf("User submitted invalid jwt")
}
