package gitsapi

import (
	"encoding/json"
	"errors"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
)

func Start() {
	archivist.Info("> Bootin HTTP API")
	h := http.NewServeMux()

	// Route: /v1/ping
	h.HandleFunc("/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		respond("pong", 200, w)
	})

	// Route: /v1/mapJson
	h.HandleFunc("/v1/mapJson", func(w http.ResponseWriter, r *http.Request) {
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
			archivist.Error("Invalid json query object", errors.New("Invalid Json"))
			http.Error(w, "Invalid json query object ", 422)
			return
		}

		// lets pass the body to our mapper
		// that will recursive map the entities
		responseData := gits.MapTransportData(transportData)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(transport.Transport{
			Data: []transport.TransportEntity{responseData},
		}, w)
	})

	// Route: /v1/query
	h.HandleFunc("/v1/query", func(w http.ResponseWriter, r *http.Request) {
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
		if err := json.Unmarshal(body, &qry); err != nil {
			archivist.Error("Invalid json query object", errors.New("Invalid Json"))
			http.Error(w, "Invalid json query object ", 422)
			return
		}

		// lets pass the body to our mapper
		// that will recursive map the entities
		responseData := query.Execute(&qry)
		if nil != err {
			http.Error(w, err.Error(), 422)
			return
		}

		respondOk(responseData, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// CUSTOMS (seperator)
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// Route: /v1/statistics/getEntityAmount
	h.HandleFunc("/v1/statistics/getEntityAmount", func(w http.ResponseWriter, r *http.Request) {
		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		amount := gits.GetEntityAmount()
		respond(strconv.Itoa(amount), 200, w)
	})

	// Route: /v1/statistics/getEntityAmountByType
	h.HandleFunc("/v1/statistics/getEntityAmountByType", func(w http.ResponseWriter, r *http.Request) {
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
		entityTypes := gits.GetEntityRTypes()
		// we should have a way to compare instead of checking an index, this could have
		// overflow/escap/bug chances
		if _, ok := entityTypes[urlParams["type"]]; !ok {
			respond("Unknown entity type given", 404, w)
		}

		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		amount, _ := gits.GetEntityAmountByType(entityTypes[urlParams["type"]])
		respond(strconv.Itoa(amount), 200, w)
	})

	// Route: /v1/statistics/getAmountPersistencePayloadsPending
	h.HandleFunc("/v1/statistics/getAmountPersistencePayloadsPending", func(w http.ResponseWriter, r *http.Request) {
		// calling storage directly from API is very bad ### bad bad entity change this and move to mapper
		amount := gits.GetAmountPersistencePayloadsPending()
		respond(strconv.Itoa(amount), 200, w)
	})

	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -
	// NOT IMPLEMENTED YET (seperator)
	// -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -

	// Route: /v1/template
	//h.HandleFunc("/v1/template", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "Hello, you hit foo!")
	//})

	// building server listen string by
	// config values and print it - than listen
	connectString := buildHttpListenConfigString()
	archivist.Info("> Server listening settings by config (" + connectString + ")")
	http.ListenAndServe(connectString, h)
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

func buildHttpListenConfigString() string {
	var connectString string
	//connectString += config.GetValue("HOST")
	connectString += ""
	connectString += ":"
	connectString += "8765"
	//connectString += config.GetValue("PORT")
	return connectString
}
