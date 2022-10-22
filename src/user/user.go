package user

import (
	"errors"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gitsapi/src/auth"
)

type User struct {
	Name             string
	Password         string
	PasswordControle string
	ApiKey           string
}

func Create(username string, password string, passwordControle string, apiKey string) error {
	if usernameExists(username) {
		return errors.New("Username already exists. Please choose a different one")
	}

	if 3 > len(username) {
		return errors.New("Username is to short. Please use at least a username length of 3 characters or longer")
	}

	if 8 > len(password) {
		return errors.New("Password is to short. Please use at least a username length of 8 characters or longer")
	}

	if password != passwordControle {
		return errors.New("Password and controle password dont match. Please correct and retry")
	}

	passwordHash, _ := auth.HashPassword(password)

	userEntity := transport.TransportEntity{
		ID:         -1,
		Value:      username,
		Properties: map[string]string{"Password": passwordHash, "Token": "", "TokenTime": ""},
	}

	if "" != apiKey {
		if 18 < len(apiKey) {
			return errors.New("Password and controle password dont match. Please correct and retry")
		}

		apiKeyEntity := transport.TransportEntity{
			ID:      -1,
			Type:    "ApiKey",
			Value:   apiKey,
			Context: "",
		}

		userEntity.ChildRelations = []transport.TransportRelation{
			{
				Context: "",
				Target:  apiKeyEntity,
			},
		}
	}

	gits.MapTransportData(userEntity)
	return nil
}

func Update(username string, password string, passwordControle string, apiKey string) error {
	if !usernameExists(username) {
		return errors.New("Username does not exist. Unable to update")
	}

	if "" != password {
		if 8 > len(password) {
			return errors.New("Password is to short. Please use at least a username length of 8 characters or longer")
		}

		if password != passwordControle {
			return errors.New("Password and controle password dont match. Please correct and retry")
		}

		passwordHash, _ := auth.HashPassword(password)

		qry := query.New().Update("User").Match("Value", "==", username).Set("Properties.Password", passwordHash)
		query.Execute(qry)
	}

	if "" != apiKey {
		if 18 < len(apiKey) {
			return errors.New("Password and controle password dont match. Please correct and retry")
		}

		qry := query.New().Update("ApiKey").Reduce("User").Match("Value", "==", username).Set("Value", apiKey)
		query.Execute(qry)
	}

	return nil
}

func usernameExists(username string) bool {
	ret, _ := gits.GetEntitiesByTypeAndValue("User", username, "match", "")
	if 0 < len(ret) {
		return true
	}
	return false
}
