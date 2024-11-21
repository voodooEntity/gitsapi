package auth

import (
	"errors"
	"fmt"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gitsapi/src/config"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func Login(username string, password string) (string, error) {
	user, err := gits.GetDefault().Storage().GetEntitiesByTypeAndValue("User", username, "match", "")
	if nil != err {
		return "", err
	}

	if 0 == len(user) {
		return "", errors.New("Unknown user given.")
	}

	if !CheckPasswordHash(password+user[0].Properties["Salt"], user[0].Properties["Password"]) {
		return "", errors.New("Wrong password given")
	}

	token := randomString(20)
	gits.GetDefault().MapData(transport.TransportEntity{
		ID:    0,
		Type:  "User",
		Value: username,
		ChildRelations: []transport.TransportRelation{
			{
				Context: "",
				Target: transport.TransportEntity{
					ID:         -1,
					Type:       "Token",
					Value:      token,
					Properties: map[string]string{"time": strconv.FormatInt(time.Now().UTC().Unix(), 10)},
				},
			},
		},
	})

	return token, nil
}

func ValidateUserAuthToken(username string, token string) bool {
	tokenLifetime, err := strconv.ParseInt(config.GetValue("TOKEN_LIFETIME"), 10, 64)
	if nil != err {
		archivist.Error("invalid token lifetime given. exiting .", tokenLifetime)
		os.Exit(0)
		return false
	}
	gits.GetDefault().Query().Execute(query.New().Read("User").Match("Value", "==", username))

	checkTime := strconv.FormatInt(time.Now().UTC().Unix()-tokenLifetime, 10)
	ret := gits.GetDefault().Query().Execute(
		query.New().Read("User").Match("Value", "==", username).To(
			query.New().Read("Token").Match("Value", "==", token).Match("Properties.time", ">", checkTime),
		),
	)

	if 0 == len(ret.Entities) {
		return false
	}
	return true
}

func HashPassword(password string) (string, string, error) {
	salt := randomString(20)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password+salt), 14)
	return string(bytes), salt, err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateApiKey(username string, apikey string) bool {
	if "" == apikey {
		return false
	}

	ret := gits.GetDefault().Query().Execute(query.New().Read("User").Match("Value", "==", username).Match("Properties.ApiKey", "==", apikey))
	if 0 == len(ret.Entities) {
		return false
	}

	return true
}

func Setup() {
	ret := gits.GetDefault().Query().Execute(query.New().Read("User"))
	if 0 == len(ret.Entities) {
		defaultPwd, salt, err := HashPassword("logmein")
		if nil != err {
			archivist.Info("Problem using bcrypt - exiting server due to possible vulnerabilities if proceeding")
			os.Exit(0)
		}

		gits.GetDefault().MapData(transport.TransportEntity{
			ID:         -1,
			Type:       "User",
			Value:      "default",
			Context:    "autoinserted",
			Properties: map[string]string{"Password": defaultPwd, "Salt": salt},
		})

		archivist.Info("Created default user")
	}
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
