package gitsapi

import (
	"encoding/json"
	"errors"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"github.com/voodooEntity/gitsapi/src/auth"
	"github.com/voodooEntity/gitsapi/src/config"
	"github.com/voodooEntity/gitsapi/src/user"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
)

var ServeMux = http.NewServeMux()

var gi *gits.Gits

func Start() {
	archivist.Info("> Bootin HTTP API")
	auth.Setup()
	gi = gits.GetDefault()

	// Route: /v1/ping
	ServeMux.HandleFunc("/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		respond("pong", 200, w)
	})

	// Route: /v1/mapJson
	ServeMux.HandleFunc("/v1/mapJson", func(w http.ResponseWriter, r *http.Request) {
		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// unpack the json
		var transportData transport.TransportEntity
		if err := json.Unmarshal(body, &transportData); err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object "+err.Error(), 422)
			return
		}

		// lets pass the body to our mapper
		// that will recursive map the entities
		responseData := gi.MapData(transportData)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(transport.Transport{
			Entities: []transport.TransportEntity{responseData},
		}, w)
	})

	// Route: /v1/query
	ServeMux.HandleFunc("/v1/query", func(w http.ResponseWriter, r *http.Request) {
		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// unpack the json
		var qry query.Query
		err = json.Unmarshal(body, &qry)
		if err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object ", 422)
			return
		}

		// lets pass the body to our mapper
		// that will recursive map the entities
		responseData := gi.Query().Execute(&qry)

		respondOk(responseData, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// Direct storage functions
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -

	// Route: /v1/getEntityByTypeAndId
	ServeMux.HandleFunc("/v1/getEntityByTypeAndId", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// get type id for given string
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			http.Error(w, string(err.Error()), 404)
			return
		}

		// read the data
		data, err := gi.Storage().GetEntityByPath(typeID, id, "")

		// if error respond
		if nil != err {
			http.Error(w, string(err.Error()), 404)
			return
		}

		// retrieve the type string
		typeStr, err := gi.Storage().GetTypeStringById(data.Type)
		if err != nil {
			http.Error(w, string(err.Error()), 404)
			return
		}

		// all seems fine lets return the data
		respondOk(transport.Transport{
			Entities: []transport.TransportEntity{
				{
					ID:         data.ID,
					Type:       typeStr,
					Context:    data.Context,
					Value:      data.Value,
					Properties: data.Properties,
					Version:    data.Version,
				},
			},
		}, w)
	})

	// Route: /v1/createEntity
	ServeMux.HandleFunc("/v1/createEntity", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// lets see if the body json is valid tho
		var newEntity transport.TransportEntity
		err = json.Unmarshal(body, &newEntity)
		if nil != err {
			http.Error(w, "Malformed json body.", 422)
			return
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(newEntity.Type)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we create the entity
		newID, err := gi.Storage().CreateEntity(types.StorageEntity{
			Type:       typeID,
			ID:         -1,
			Properties: newEntity.Properties,
			Context:    newEntity.Context,
		})
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(transport.Transport{
			Entities: []transport.TransportEntity{
				{
					ID:         newID,
					Type:       newEntity.Type,
					Value:      newEntity.Value,
					Context:    newEntity.Context,
					Properties: newEntity.Properties,
					Version:    1,
				},
			},
		}, w)
	})

	// Route: /v1/getEntitiesByType
	ServeMux.HandleFunc("/v1/getEntitiesByType", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 403)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default for mode and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// ok we seem to be fine on types, lets call the actual getter method
		entities, err := gi.Storage().GetEntitiesByType(urlParams["type"], context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// translate the storage format to our transport format
		responseData := transport.Transport{
			Entities: []transport.TransportEntity{},
		}

		for _, val := range entities {
			responseData.Entities = append(responseData.Entities, transport.TransportEntity{
				ID:         val.ID,
				Type:       requiredUrlParams["type"],
				Context:    val.Context,
				Value:      val.Value,
				Properties: val.Properties,
				Version:    val.Version,
			})
		}

		// all seems fine lets return the data
		respondOk(responseData, w)
	})

	// Route: /v1/getEntitiesByTypeAndValue
	ServeMux.HandleFunc("/v1/getEntitiesByTypeAndValue", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["value"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["mode"] = ""
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default for mode and
		// overwrite if given
		mode := "match"
		if _, ok := urlParams["mode"]; urlParams["mode"] != "" && ok {
			mode = urlParams["mode"]
		}

		// lets make a default for mode and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// ok we seem to be fine on types, lets call the actual getter method
		entities, err := gi.Storage().GetEntitiesByTypeAndValue(urlParams["type"], urlParams["value"], mode, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// translate the storage format to our transport format
		responseData := transport.Transport{
			Entities: []transport.TransportEntity{},
		}

		for _, val := range entities {
			responseData.Entities = append(responseData.Entities, transport.TransportEntity{
				ID:         val.ID,
				Type:       requiredUrlParams["type"],
				Context:    val.Context,
				Value:      val.Value,
				Properties: val.Properties,
				Version:    val.Version,
			})
		}

		// all seems fine lets return the data
		respondOk(responseData, w)
	})

	// Route: /v1/deleteEntity
	ServeMux.HandleFunc("/v1/deleteEntity", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "DELETE" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we delete the entity
		gi.Storage().DeleteEntity(typeID, id)

		respond("", 200, w)
	})

	// Route: /v1/updateEntity
	ServeMux.HandleFunc("/v1/updateEntity", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "PUT" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// lets see if the body json is valid tho
		var newEntity transport.TransportEntity
		err = json.Unmarshal(body, &newEntity)
		if nil != err {
			http.Error(w, "Malformed json body.", 422)
			return
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(newEntity.Type)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we update the entity
		err = gi.Storage().UpdateEntity(types.StorageEntity{
			Type:       typeID,
			ID:         newEntity.ID,
			Value:      newEntity.Value,
			Context:    newEntity.Context,
			Properties: newEntity.Properties,
			Version:    newEntity.Version,
		})

		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respond("", 200, w)
	})

	// Route: /v1/getChildEntities
	ServeMux.HandleFunc("/v1/getChildEntities", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default context and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// retrieve the child entities if given
		childRelations, err := gi.Storage().GetChildRelationsBySourceTypeAndSourceId(typeID, id, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// build response data by getting entities based on childRelation data into transport format
		returnData := transport.Transport{
			Entities: []transport.TransportEntity{},
		}

		for _, val := range childRelations {
			entity, err := gi.Storage().GetEntityByPath(val.TargetType, val.TargetID, "")
			if nil != err {
				returnData.Entities = append(returnData.Entities, transport.TransportEntity{
					ID:         entity.ID,
					Type:       urlParams["type"],
					Value:      entity.Value,
					Context:    entity.Context,
					Properties: entity.Properties,
					Version:    entity.Version,
				})
			}
		}

		respondOk(returnData, w)
	})

	// Route: /v1/getParentEntities
	ServeMux.HandleFunc("/v1/getParentEntities", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default context and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// retrieve the child entities if given
		parentRelations, err := gi.Storage().GetParentRelationsByTargetTypeAndTargetId(typeID, id, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// build response data by getting entities based on childRelation data into transport format
		returnData := transport.Transport{
			Entities: []transport.TransportEntity{},
		}

		for _, val := range parentRelations {
			entity, err := gi.Storage().GetEntityByPath(val.SourceType, val.SourceID, "")
			if nil != err {
				returnData.Entities = append(returnData.Entities, transport.TransportEntity{
					ID:         entity.ID,
					Type:       urlParams["type"],
					Value:      entity.Value,
					Context:    entity.Context,
					Properties: entity.Properties,
					Version:    entity.Version,
				})
			}
		}

		respondOk(returnData, w)
	})

	// Route: /v1/getRelationsTo
	ServeMux.HandleFunc("/v1/getRelationsTo", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default context and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// retrieve the child entities if given
		relations, err := gi.Storage().GetParentRelationsByTargetTypeAndTargetId(typeID, id, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// since we could have every possible type in our results we gonne go the easy way and retrieve all entity types for easier result translation
		entityTypes := gi.Storage().GetEntityTypes()

		// prepare return data and write retrieved relations into the fitting format
		returnData := transport.Transport{
			Relations: []transport.TransportRelation{},
		}
		for _, val := range relations {
			returnData.Relations = append(returnData.Relations, transport.TransportRelation{
				SourceID:   val.SourceID,
				SourceType: entityTypes[val.SourceType],
				TargetID:   val.TargetID,
				TargetType: urlParams["type"],
				Context:    val.Context,
				Properties: val.Properties,
				Version:    val.Version,
			})
		}

		respondOk(returnData, w)
	})

	// Route: /v1/getRelationsFrom
	ServeMux.HandleFunc("/v1/getRelationsFrom", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		requiredUrlParams["id"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		id, err := strconv.Atoi(urlParams["id"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default context and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// translate the type from string to id
		typeID, err := gi.Storage().GetTypeIdByString(urlParams["type"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// retrieve the child entities if given
		relations, err := gi.Storage().GetChildRelationsBySourceTypeAndSourceId(typeID, id, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// since we could have every possible type in our results we gonne go the easy way and retrieve all entity types for easier result translation
		entityTypes := gi.Storage().GetEntityTypes()

		// prepare return data and write retrieved relations into the fitting format
		returnData := transport.Transport{
			Relations: []transport.TransportRelation{},
		}
		for _, val := range relations {
			returnData.Relations = append(returnData.Relations, transport.TransportRelation{
				SourceID:   val.SourceID,
				SourceType: urlParams["type"],
				TargetID:   val.TargetID,
				TargetType: entityTypes[val.SourceType],
				Context:    val.Context,
				Properties: val.Properties,
				Version:    val.Version,
			})
		}

		respondOk(returnData, w)
	})

	// Route: /v1/getRelation
	ServeMux.HandleFunc("/v1/getRelation", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["srcType"] = ""
		requiredUrlParams["srcID"] = ""
		requiredUrlParams["targetType"] = ""
		requiredUrlParams["targetID"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id
		srcID, err := strconv.Atoi(urlParams["srcID"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// int conv id
		targetID, err := strconv.Atoi(urlParams["targetID"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// translate the type from string to id
		srcTypeID, err := gi.Storage().GetTypeIdByString(urlParams["srcType"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}
		targetTypeID, err := gi.Storage().GetTypeIdByString(urlParams["targetType"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		relation, err := gi.Storage().GetRelation(srcTypeID, srcID, targetTypeID, targetID)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(transport.Transport{
			Relations: []transport.TransportRelation{
				{
					SourceType: urlParams["srcType"],
					SourceID:   relation.SourceID,
					TargetType: urlParams["targetType"],
					TargetID:   relation.TargetID,
					Context:    relation.Context,
					Properties: relation.Properties,
					Version:    relation.Version,
				},
			},
		}, w)
	})

	// Route: /v1/getEntitiesByValue
	ServeMux.HandleFunc("/v1/getEntitiesByValue", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["value"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// now we get optional params
		optionalUrlParams := make(map[string]string)
		optionalUrlParams["mode"] = ""
		optionalUrlParams["context"] = ""
		urlParams = getOptionalUrlParams(optionalUrlParams, urlParams, r)

		// lets make a default for mode and
		// overwrite if given
		mode := "match"
		if _, ok := urlParams["mode"]; urlParams["mode"] != "" && ok {
			mode = urlParams["mode"]
		}

		// lets make a default context and
		// overwrite if given
		context := ""
		if _, ok := urlParams["context"]; ok {
			context = urlParams["context"]
		}

		// retrieve the entities
		entities, err := gi.Storage().GetEntitiesByValue(urlParams["value"], mode, context)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// since we could have every possible type in our results we gonne go the easy way and retrieve all entity types for easier result translation
		entityTypes := gi.Storage().GetEntityTypes()

		// write return data
		returnData := transport.Transport{
			Entities: []transport.TransportEntity{},
		}

		for _, val := range entities {
			returnData.Entities = append(returnData.Entities, transport.TransportEntity{
				ID:         val.ID,
				Type:       entityTypes[val.Type],
				Context:    val.Context,
				Value:      val.Value,
				Version:    val.Version,
				Properties: val.Properties,
			})
		}

		respondOk(returnData, w)
	})

	// Route: /v1/getEntityTypes
	ServeMux.HandleFunc("/v1/getEntityTypes", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve all entity types
		types := gi.Storage().GetEntityTypes()

		// than we gonne json encode it
		// build the json
		responseData, err := json.Marshal(types)
		if nil != err {
			http.Error(w, "Error building response data json", 500)
			return
		}
		// finally we gonne send our response
		respond(string(responseData), 200, w)
	})

	// Route: /v1/updateRelation
	ServeMux.HandleFunc("/v1/updateRelation", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "PUT" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// lets see if the body json is valid tho
		var newRelation transport.TransportRelation
		err = json.Unmarshal(body, &newRelation)
		if nil != err {
			http.Error(w, "Malformed json body.", 422)
			return
		}

		// translate the type from string to id
		srcTypeID, err := gi.Storage().GetTypeIdByString(newRelation.SourceType)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}
		targetTypeID, err := gi.Storage().GetTypeIdByString(newRelation.TargetType)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we update the entity
		_, err = gi.Storage().UpdateRelation(srcTypeID, newRelation.SourceID, targetTypeID, newRelation.TargetID, types.StorageRelation{
			SourceID:   newRelation.SourceID,
			SourceType: srcTypeID,
			TargetID:   newRelation.TargetID,
			TargetType: targetTypeID,
			Context:    newRelation.Context,
			Properties: newRelation.Properties,
			Version:    newRelation.Version,
		})
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respond("", 200, w)
	})

	// Route: /v1/createRelation
	ServeMux.HandleFunc("/v1/createRelation", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		// lets see if the body json is valid tho
		var newRelation transport.TransportRelation
		err = json.Unmarshal(body, &newRelation)
		if nil != err {
			http.Error(w, "Malformed json body.", 422)
			return
		}

		// translate the type from string to id
		srcTypeID, err := gi.Storage().GetTypeIdByString(newRelation.SourceType)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}
		targetTypeID, err := gi.Storage().GetTypeIdByString(newRelation.TargetType)
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we update the entity
		_, err = gi.Storage().CreateRelation(srcTypeID, newRelation.SourceID, targetTypeID, newRelation.TargetID, types.StorageRelation{
			SourceID:   newRelation.SourceID,
			SourceType: srcTypeID,
			TargetID:   newRelation.TargetID,
			TargetType: targetTypeID,
			Context:    newRelation.Context,
			Properties: newRelation.Properties,
		})

		respond("", 200, w)
	})

	// Route: /v1/createRelation
	ServeMux.HandleFunc("/v1/deleteRelation", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "DELETE" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["srcType"] = ""
		requiredUrlParams["srcID"] = ""
		requiredUrlParams["targetType"] = ""
		requiredUrlParams["targetID"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		// int conv id's
		srcID, err := strconv.Atoi(urlParams["srcID"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// int conv id's
		targetID, err := strconv.Atoi(urlParams["targetID"])
		if nil != err {
			// handle error
			http.Error(w, "Invalid param id given", 422)
			return
		}

		// translate the type from string to id
		srcTypeID, err := gi.Storage().GetTypeIdByString(requiredUrlParams["srcType"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}
		targetTypeID, err := gi.Storage().GetTypeIdByString(requiredUrlParams["targetType"])
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// finally we delete the entity
		gi.Storage().DeleteRelation(srcTypeID, srcID, targetTypeID, targetID)

		respond("", 200, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// Auth / User
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -

	// Route: /v1/user/auth
	ServeMux.HandleFunc("/v1/user/auth", func(w http.ResponseWriter, r *http.Request) {
		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		var usr user.User
		if err := json.Unmarshal(body, &usr); err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object "+err.Error(), 422)
			return
		}

		token, err := auth.Login(usr.Name, usr.Password)
		if nil != err {
			archivist.Error("Could not login", err.Error())
			http.Error(w, "Could not login. ", 422)
			return
		}

		// finally we gonne send our response
		respond(token, 200, w)
	})

	ServeMux.HandleFunc("/v1/user/create", func(w http.ResponseWriter, r *http.Request) {
		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "POST" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		var usr user.User
		if err := json.Unmarshal(body, &usr); err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object "+err.Error(), 422)
			return
		}

		usr2, err := user.Create(usr.Name, usr.Password, usr.PasswordControle, usr.ApiKey)
		if nil != err {
			archivist.Error("Could not create user", err.Error())
			http.Error(w, "could not create user. ", 422)
			return
		}

		// finally we gonne send our response
		respond(strconv.Itoa(usr2), 200, w)
	})

	ServeMux.HandleFunc("/v1/user/update", func(w http.ResponseWriter, r *http.Request) {
		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "PUT" != r.Method {
			http.Error(w, "Invalid http method for this path", 422)
			return
		}

		// retrieve data from request
		body, err := getRequestBody(r)
		if nil != err {
			archivist.Error("Could not read http request body", err.Error())
			http.Error(w, "Malformed or no body. ", 422)
			return
		}

		var usr user.User
		if err := json.Unmarshal(body, &usr); err != nil {
			archivist.Error("Invalid json query object", err.Error())
			http.Error(w, "Invalid json query object "+err.Error(), 422)
			return
		}

		if err = user.Update(usr.Name, usr.Password, usr.PasswordControle, usr.ApiKey); nil != err {
			archivist.Error("Could not update user", err.Error())
			http.Error(w, "could not update user.", 422)
			return
		}

		// finally we gonne send our response
		respond("", 200, w)
	})

	// Route: /v1/user/search
	ServeMux.HandleFunc("/v1/user/search", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// check http method
		if "GET" != r.Method {
			http.Error(w, "Invalid http method for this path", 403)
			return
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["search"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)

		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 422)
			return
		}

		// ok we seem to be fine on types, lets call the actual getter method
		responseData := user.GetUserListBySearch(urlParams["search"])

		// all seems fine lets return the data
		respondOk(responseData, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// Stats
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// Route: /v1/statistics/getEntityAmount
	ServeMux.HandleFunc("/v1/statistics/getEntityAmount", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		amount := gi.Storage().GetEntityAmount()
		respond(strconv.Itoa(amount), 200, w)
	})

	// Route: /v1/statistics/getEntityAmountByType
	ServeMux.HandleFunc("/v1/statistics/getEntityAmountByType", func(w http.ResponseWriter, r *http.Request) {
		if !handleAuth(r) {
			respond("", 401, w)
			return
		}

		if "" != config.GetValue("CORS_ORIGIN") || "" != config.GetValue("CORS_HEADER") {
			if "OPTIONS" == r.Method {
				respond("", 200, w)
				return
			}
		}

		// first we get the params
		requiredUrlParams := make(map[string]string)
		requiredUrlParams["type"] = ""
		urlParams, err := getRequiredUrlParams(requiredUrlParams, r)
		// required params check
		if nil != err {
			// handle error
			http.Error(w, err.Error(), 404)
			return
		}
		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		entityTypes := gi.Storage().GetEntityRTypes()
		// we should have a way to compare instead of checking an index, this could have
		// overflow/escap/bug chances
		if _, ok := entityTypes[urlParams["type"]]; !ok {
			respond("Unknown entity type given", 404, w)
		}

		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		amount, _ := gi.Storage().GetEntityAmountByType(entityTypes[urlParams["type"]])
		respond(strconv.Itoa(amount), 200, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// NOT IMPLEMENTED YET (seperator)
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -

	// Route: /v1/template
	//ServeMux.HandleFunc("/v1/template", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "Hello, you hit foo!")
	//})

	// building server listen string by
	// config values and print it - than listen
	connectString := buildListenConfigString()
	archivist.Info("> Server listening settings by config (" + connectString + ")")
	if "https" == config.GetValue("PROTOCOL") {
		http.ListenAndServeTLS(connectString, config.GetValue("SSL_CERT_FILE"), config.GetValue("SSL_KEY_FILE"), ServeMux)
	} else if "http" == config.GetValue("PROTOCOL") {
		http.ListenAndServe(connectString, ServeMux)
	} else {
		archivist.Error("Unknown server protocol given (valid http / https) - given: " + config.GetValue("PROTOCOL"))
		os.Exit(0)
	}

}

func getOptionalUrlParams(optionalUrlParams map[string]string, urlParams map[string]string, r *http.Request) map[string]string {
	tmpParams := r.URL.Query()
	for paramName := range optionalUrlParams {
		val, ok := tmpParams[paramName]
		if ok {
			urlParams[paramName] = val[0]
		}
	}
	return urlParams
}

func getRequiredUrlParams(requiredUrlParams map[string]string, r *http.Request) (map[string]string, error) {
	urlParams := r.URL.Query()
	for paramName := range requiredUrlParams {
		val, ok := urlParams[paramName]
		if !ok {
			return nil, errors.New("Missing required url param")
		}
		requiredUrlParams[paramName] = val[0]
	}
	return requiredUrlParams, nil
}

func respond(message string, responseCode int, w http.ResponseWriter) {

	corsAllowHeaders := config.GetValue("CORS_HEADER")
	if "" != corsAllowHeaders {
		w.Header().Add("Access-Control-Allow-Headers", corsAllowHeaders)
	}
	corsAllowOrigin := config.GetValue("CORS_ORIGIN")
	if "" != corsAllowOrigin {
		w.Header().Add("Access-Control-Allow-Origin", corsAllowOrigin)
	}

	w.WriteHeader(responseCode)
	messageBytes := []byte(message)

	_, err := w.Write(messageBytes)
	if nil != err {
		archivist.Error("Could not write http response body ", err, message)
	}
}

func respondOk(data transport.Transport, w http.ResponseWriter) {
	// than we gonne json encode it
	// build the json
	responseData, err := json.Marshal(data)
	if nil != err {
		http.Error(w, "Error building response data json", 500)
		return
	}

	// finally we gonne send our response
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(200)
	_, err = w.Write(responseData)
	if nil != err {
		archivist.Error("Could not write http response body ", err, data)
	}
}

func getRequestBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func buildListenConfigString() string {
	var connectString string
	connectString += config.GetValue("HOST")
	connectString += ":"
	connectString += config.GetValue("PORT")
	return connectString
}

func handleAuth(r *http.Request) bool {
	// first we check if auth is active
	if "no" == config.GetValue("AUTH_ACTIVE") {
		return true
	}

	username := r.Header.Get("GITSAPI-AUTH-USER")
	if "" == username {
		archivist.Info("trying to access API without username given.")
		return false
	}

	token := r.Header.Get("GITSAPI-AUTH-TOKEN")
	if "" != token {
		return auth.ValidateUserAuthToken(username, token)
	}

	apikey := r.Header.Get("GITSAPI-AUTH-APIKEY")
	if "" != apikey {
		return auth.ValidateApiKey(username, apikey)
	}

	return false
}
