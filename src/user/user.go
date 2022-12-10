package user

import (
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gitsapi/src/auth"
)

type User struct {
	ID               int
	Name             string
	Password         string
	PasswordControle string
	ApiKey           string
}

func Create(username string, password string, passwordControle string, apiKey string) (int, error) {
	if usernameExists(username) {
		return -1, errors.New("Username already exists. Please choose a different one")
	}

	if 3 > len(username) {
		return -1, errors.New("Username is to short. Please use at least a username length of 3 characters or longer")
	}

	if 8 > len(password) {
		return -1, errors.New("Password is to short. Please use at least a username length of 8 characters or longer")
	}
	archivist.Debug("Password provided for new user", password)
	archivist.Debug("PasswordControle provided for new user", passwordControle)
	if password != passwordControle {
		return -1, errors.New("Password and controle password dont match. Please correct and retry")
	}

	passwordHash, salt, err := auth.HashPassword(password)

	if nil != err {
		return -1, errors.New("failure in password hash generation progress.")
	}

	userEntity := transport.TransportEntity{
		ID:         -1,
		Type:       "User",
		Value:      username,
		Properties: map[string]string{"Password": passwordHash, "Salt": salt, "Token": "", "TokenTime": ""},
	}

	if "" != apiKey {
		if 18 < len(apiKey) {
			return -1, errors.New("Password and controle password dont match. Please correct and retry")
		}
		userEntity.Properties["ApiKey"] = apiKey
	}

	usr := gits.MapTransportData(userEntity)
	return usr.ID, nil
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

		passwordHash, salt, err := auth.HashPassword(password)

		if nil != err {
			return errors.New("failure in password hash generation progress.")
		}

		qry := query.New().Update("User").Match("Value", "==", username).Set("Properties.Password", passwordHash).Set("Properties.Salt", salt)
		query.Execute(qry)
	}

	if "" != apiKey {
		if 18 < len(apiKey) {
			return errors.New("API Key to short - it should at least be of length 18. Please correct and retry")
		}
		qry := query.New().Update("User").Match("Value", "==", username).Set("Properties.ApiKey", apiKey)
		query.Execute(qry)
	}

	return nil
}

func GetUserListBySearch(search string) transport.Transport {
	ret := transport.Transport{
		Entities: []transport.TransportEntity{},
		Amount:   0,
	}

	users := query.Execute(query.New().Read("User").Match("Value", "contain", search))
	if 0 < len(users.Entities) {
		for _, user := range users.Entities {
			ret.Entities = append(ret.Entities, transport.TransportEntity{
				ID:             user.ID,
				Value:          user.Value,
				Context:        user.Context,
				ChildRelations: user.ChildRelations,
			})
		}
		ret.Amount = users.Amount
	}
	return ret
}

func usernameExists(username string) bool {
	ret, _ := gits.GetEntitiesByTypeAndValue("User", username, "match", "")
	if 0 < len(ret) {
		return true
	}
	return false
}
