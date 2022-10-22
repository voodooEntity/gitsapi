package auth

import (
	"errors"
	"fmt"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func Login(username string, password string) (string, error) {
	entities, err := gits.GetEntitiesByTypeAndValue("User", username, "match", "")
	if nil != err {
		return "", err
	}

	if 0 == len(entities) {
		return "", errors.New("Unknown user given")
	}

	if !CheckPasswordHash(password, entities[0].Properties["Password"]) {
		return "", errors.New("Wrong password given")
	}

	token := randomString(20)
	query.Execute(query.New().Update("User").Match("Value", "==", username).Set("Properties.Token", token).Set("Properties.TokenTime", strconv.FormatInt(time.Now().UTC().UnixNano(), 10)))

	return token, nil
}

func ValidateUserAuthToken(username string, token string) bool {
	ret := query.Execute(query.New().Read("User").Match("Value", "==", username).Match("Properties.Token", "==", token))
	if 0 == len(ret.Entities) {
		return false
	}
	return true
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateApiKey(username string, apikey string) bool {
	if "" == apikey {
		return false
	}

	ret := query.Execute(query.New().Read("User").Match("Value", "==", username).To(query.New().Read("ApiKey").Match("Value", "==", apikey)))
	if 0 == len(ret.Entities) {
		return false
	}

	return true
}

func Setup() {
	ret := query.Execute(query.New().Read("User"))
	if 0 == len(ret.Entities) {
		defaultPwd, err := HashPassword("logmein")
		if nil != err {
			archivist.Info("Problem using bcrypt - exiting server due to possible vulnerabilities if proceeding")
			os.Exit(0)
		}

		gits.MapTransportData(transport.TransportEntity{
			ID:         -1,
			Type:       "User",
			Value:      "default",
			Context:    "autoinserted",
			Properties: map[string]string{"Password": defaultPwd},
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
