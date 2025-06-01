<h1 align="center">
  <br>
  <a href="https://github.com/voodooEntity/gitsapi/"><img src="https://raw.githubusercontent.com/voodooEntity/gitsapi/main/DOCS/IMAGES/gitsapi_logo.png" alt="GITSAPI" width="300"></a>
  <br>
  GITSAPI
  <br>
</h1>

<h4 align="center">A <span style="color:#35b9e9">RESTful HTTP API</span> for seamless interaction with GITS in-memory graph storage.</h4>

<p align="center">
  <a href="#key-features">Key Features</a> •
  <a href="#about">About</a> •
  <a href="#use-cases">Use Cases</a> •
  <a href="#how-to-use">How To Use</a> •
  <a href="#api-reference">API Reference</a> •
  <a href="#changelog">Changelog</a> •
  <a href="#license">License</a>
</p>

---

## Key Features

* **HTTP/S Access:** Interact with GITS instances over standard HTTP or secure HTTPS.
* **JSON-Native:** All requests and responses use JSON for easy integration with any programming language.
* **Unified Data Interface:** Leverages `transport.TransportEntity` and `transport.Transport` for consistent data mapping and querying.
* **Direct Storage Operations:** Perform CRUD (Create, Read, Update, Delete) operations on individual entities and relations.
* **Query Language Support:** Execute complex GITS query builder statements via the API.
* **Graph Traversal:** Navigate relationships by fetching child and parent entities/relations.
* **Multi-Storage Support:** Select a specific GITS instance using the `Storage` HTTP header, or use the default.
* **CORS Enabled:** Configurable Cross-Origin Resource Sharing for flexible web application integration.
* **Real-time Insights:** Access entity type lists and overall/type-specific entity counts.

---

## About

GITSAPI serves as the HTTP interface layer for the **GITS (Graph In-memory Thread-safe Storage)** library. It exposes GITS's powerful in-memory graph capabilities through a simple and intuitive RESTful API, enabling the usage of gits without the need of including it into your project as package and also enable other applications to interact with it.

This API is designed for applications that need to:

* **Integrate GITS with non-Go services:** Any language capable of making HTTP requests can now leverage GITS.
* **Decouple GITS from the main application:** Run GITSAPI as a standalone microservice.
* **Provide a centralized graph data store:** Offer a single point of access for various clients to interact with GITS data.

By providing a network-accessible interface, GITSAPI makes GITS's high-performance, concurrency-safe graph storage accessible across diverse architectural landscapes.

You choose to either bootstrap it inside your GITS-using application, or run it standalone with the provided command.

---

## Use Cases

GITSAPI broadens the applicability of GITS to a wider range of scenarios, including:

* **Frontend Dashboards:** Powering web-based dashboards with real-time graph data from GITS, but now accessible via API for a JavaScript frontend).
* **Distributed Web Crawlers:** Microservices written in different languages can inject discovered links and pages into a central GITS instance via GITSAPI, enabling distributed crawl management.
* **API Gateways:** Acting as a backend data store for API gateways that need to manage complex, interconnected data from various upstream services.
* **Automation Workflows:** Integrating GITS-based data processing into automation platforms that prefer HTTP interactions over direct library calls.
* **Multi-language Environments:** Teams using different programming languages can all interact with the same GITS data through a standardized HTTP interface.

---

## How To Use
As mentioned before, you can either use GITSAPI as package in your go project, or as standalone server.

### Setup

#### GITSAPI used in your existing go project:

First we require the library into your existing project
```bash
go get github.com/voodooEntity/gitsapi
```

afterwards we initialize the server from code
```go
    // init the base config package 
	config.Init(make(map[string]string))

    // init the archivist internal logger
    archivist.Init(config.GetValue("LOG_LEVEL"), config.GetValue("LOG_TARGET"), config.GetValue("LOG_PATH"))

    // if not already existing, you need to create at least one storage before starting the API. 
	// if not told different via HTTP Header, GITSAPI will default to the default set storage
    gits.NewInstance("api")

    // finally you start the server
    gitsapi.Start()
```



#### GITSAPI runs as a standalone server. You'll typically build and run it as an executable.

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/voodooEntity/gitsapi.git](https://github.com/voodooEntity/gitsapi.git)
    cd gitsapi
    ```
2.  **Configure:** GITSAPI uses environment variables or a configuration file (managed by the `config` package) for its settings. Key settings include:
    * `HOST`: The IP address to listen on (e.g., `0.0.0.0`).
    * `PORT`: The port number to listen on (e.g., `8080`).
    * `PROTOCOL`: `http` or `https`.
    * `SSL_CERT_FILE` / `SSL_KEY_FILE`: Paths to SSL certificate and key files (if using HTTPS).
    * `CORS_ORIGIN`: Allowed CORS origin (e.g., `*` or `http://localhost:3000`).
    * `CORS_HEADER`: Allowed CORS headers (e.g., `*` or `Content-Type, Authorization`).
*It's also required to have a gits.api.config.json existing with all required values. This requirement is set in order to make sure that GITSAPI might never be bootstrapped with missing values and no proper fallback.*


3.  **Run:**
    ```bash
    cd cmd/httpserver/ && go run . # or compile and run the executable from httpserver.go
    ```
    Once running, GITSAPI will start listening on the configured address and port.

### Interacting with GITSAPI

You can interact with GITSAPI using any HTTP client (e.g., `curl`, Postman, browser `fetch` API, or programming language HTTP libraries).

*It is important to mention, that at this point GITSAPI does not implement any way of authentication/authorization. Since GITS in considered an in memory graph storage, it doesnt include a specific role/permissions system. Access to the API, if protection necessary, should be done previous access.*

#### Selecting a GITS Instance

By default, GITSAPI interacts with the **default** GITS instance (`gits.GetDefault()`). However, if you're running multiple named GITS instances within the same application (e.g., `myGitsInstance := gits.NewInstance("my_specific_storage")`), you can specify which instance to use for a request by including the `Storage` HTTP header:


Storage: my_specific_storage


**Example `curl` with Storage Header:**

```bash
curl -X POST http://localhost:8080/v1/mapJson \
     -H "Content-Type: application/json" \
     -H "Storage: my_specific_storage" \
     -d '{
           "Type": "User",
           "Value": "john.doe",
           "Properties": {"email": "john.doe@example.com"}
         }'
```

#### Extending the API
While the GITSAPI is designed to enable a wide variety on possibilities to interact with the storge, that for the case the application is used in context of existing systems or as package, there might be some endpoints you want to expose without adding a second API/Interface. Therefor GITSAPI exposes its **gitsapi.ServeMux** as public member. This enables you to programmatically add more endpoints to the server and extend the abilities of your application.

To achieve this just add a HandleFunc to the exposed ServeMux like:
```go
// Route: /my/custom/gitsapi/extended/endpoint
gitsapi.ServeMux.HandleFunc("/my/custom/gitsapi/extended/endpoint", func(w http.ResponseWriter, r *http.Request) {
    // Your endpoint action goes here
})
```

-----

## API Reference

GITSAPI exposes a comprehensive set of endpoints for managing and querying your GITS data. All request and response bodies are JSON.

For detailed information on the `transport.TransportEntity`, `transport.Transport`, and `query.Query` JSON structures, please refer to the main [GITS Documentation](https://www.google.com/search?q=https://github.com/voodooEntity/gits/DOCS/README.md).

### Core Operations

-----

### `/v1/ping`

  * **Method:** `GET`
  * **Purpose:** Simple health check to verify the API is running.
  * **Response:** `text/plain` body with "pong".
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/ping
    # Response: pong
    ```

-----

### `/v1/mapJson`

  * **Method:** `POST`
  * **Purpose:** Maps a `transport.TransportEntity` (including nested `ChildRelations`) into the GITS storage. This is the primary endpoint for bulk data injection and creating complex graph structures at once.
  * **Request Body:** A JSON object representing a `transport.TransportEntity`.
    ```json
    {
      "Type": "string",
      "ID": -1,              // Use -1 or storage.MAP_FORCE_CREATE for new entities
      "Value": "string",
      "Context": "string",
      "Properties": {        // Optional
        "key1": "value1",
        "key2": "value2"
      },
      "ChildRelations": [    // Optional: for mapping nested structures
        {
          "Context": "string",
          "Properties": {},
          "Target": {
            "Type": "string",
            "ID": -1,
            "Value": "string"
          }
        }
      ]
    }
    ```
  * **Response Body (200 OK):** A `transport.Transport` object containing the root `transport.TransportEntity` that was mapped (with its newly assigned GITS ID).
    ```json
    {
      "Entities": [
        {
          "Type": "string",
          "ID": 123,           // GITS assigned ID
          "Value": "string",
          "Context": "string",
          "Version": 1,
          "Properties": {
            "key1": "value1"
          },
          "ChildRelations": []
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON.
  * **Example:**
    ```bash
    curl -X POST http://localhost:8080/v1/mapJson \
         -H "Content-Type: application/json" \
         -d '{
               "Type": "Article",
               "Value": "GITSAPI Release",
               "Context": "Blog",
               "Properties": {"author": "VoodooEntity"},
               "ChildRelations": [
                 {
                   "Context": "LinksTo",
                   "Target": {
                     "Type": "GitHubRepo",
                     "Value": "gitsapi",
                     "Properties": {"url": "[https://github.com/voodooEntity/gitsapi](https://github.com/voodooEntity/gitsapi)"}
                   }
                 }
               ]
             }'
    ```

-----

### `/v1/query`

  * **Method:** `POST`
  * **Purpose:** Executes a GITS query builder statement to retrieve complex graph data.
  * **Request Body:** A JSON object representing a `query.Query` struct. Refer to the [GITS Query Language Reference](https://www.google.com/search?q=https://github.com/voodooEntity/gits/DOCS/QUERY.md) for detailed query syntax.

  * **Response Body (200 OK):** A `transport.Transport` object containing `Entities` and `Relations` that match the query.
    ```json
    {
      "Entities": [
        {
          "Type": "string",
          "ID": 1,
          "Value": "string",
          "Context": "string",
          "Properties": {}
        }
      ],
      "Relations": [
        {
          "SourceType": "string",
          "SourceID": 1,
          "TargetType": "string",
          "TargetID": 2,
          "Context": "string",
          "Properties": {}
        }
      ]
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON query object.
  * **Example:**
    ```bash
    curl -X POST http://localhost:8080/v1/query \
         -H "Content-Type: application/json" \
         -d 'YOURGITSQUERY###'
    ```

-----

### `/v1/getEntityByTypeAndId`

  * **Method:** `GET`
  * **Purpose:** Retrieves a single GITS entity by its `Type` and `ID`.
  * **URL Parameters:**
      * `type` (required, string): The entity type (e.g., `User`, `Domain`).
      * `id` (required, integer): The unique ID of the entity within its type.
  * **Response Body (200 OK):** A `transport.Transport` object containing the requested `transport.TransportEntity`.
    ```json
    {
      "Entities": [
        {
          "ID": 123,
          "Type": "User",
          "Value": "john.doe",
          "Context": "ApplicationData",
          "Properties": { "email": "john.doe@example.com" },
          "Version": 1
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters (e.g., `id` not an integer).
      * `404 Not Found`: Entity type or entity with the given ID not found.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getEntityByTypeAndId?type=User&id=123
    ```

-----

### `/v1/createEntity`

  * **Method:** `POST`
  * **Purpose:** Creates a new single entity in GITS. This differs from `mapJson` by only accepting a single entity and not processing nested relations.
  * **Request Body:** A JSON object representing a `transport.TransportEntity`. Only `Type`, `Value`, `Context`, and `Properties` are typically used for creation. `ID` should be set to `-1` or `storage.MAP_FORCE_CREATE`.
    ```json
    {
      "Type": "string",
      "Value": "string",
      "Context": "string",
      "Properties": {}
    }
    ```
  * **Response Body (200 OK):** A `transport.Transport` object containing the newly created `transport.TransportEntity` with its assigned GITS ID.
    ```json
    {
      "Entities": [
        {
          "ID": 456,           // GITS assigned ID
          "Type": "Product",
          "Value": "New Gadget",
          "Context": "Inventory",
          "Properties": {},
          "Version": 1
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON.
  * **Example:**
    ```bash
    curl -X POST http://localhost:8080/v1/createEntity \
         -H "Content-Type: application/json" \
         -d '{
               "Type": "Product",
               "Value": "New Gadget",
               "Context": "Inventory",
               "Properties": {"price": "99.99"}
             }'
    ```

-----

### `/v1/getEntitiesByType`

  * **Method:** `GET`
  * **Purpose:** Retrieves all entities of a specified type.
  * **URL Parameters:**
      * `type` (required, string): The entity type (e.g., `Device`, `Service`).
      * `context` (optional, string): Filter entities by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of `transport.TransportEntity` objects matching the criteria.
    ```json
    {
      "Entities": [
        {
          "ID": 101,
          "Type": "Device",
          "Value": "Server-01",
          "Context": "Network",
          "Properties": {}
        },
        {
          "ID": 102,
          "Type": "Device",
          "Value": "Router-05",
          "Context": "Network",
          "Properties": {}
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing required URL parameter `type`.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getEntitiesByType?type=Device
    curl http://localhost:8080/v1/getEntitiesByType?type=Device&context=Network
    ```

-----

### `/v1/getEntitiesByTypeAndValue`

  * **Method:** `GET`
  * **Purpose:** Retrieves entities of a specific type that also match a given `Value`.
  * **URL Parameters:**
      * `type` (required, string): The entity type.
      * `value` (required, string): The value to match against the `Value` field of entities.
      * `mode` (optional, string): Match mode. Defaults to `match`. Can be `match`, `contains`, `startsWith`, `endsWith`, `fuzzy`.
      * `context` (optional, string): Filter entities by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of matching `transport.TransportEntity` objects.
    ```json
    {
      "Entities": [
        {
          "ID": 201,
          "Type": "File",
          "Value": "index.html",
          "Context": "Web",
          "Properties": {}
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing required URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getEntitiesByTypeAndValue?type=File&value=index.html
    curl "http://localhost:8080/v1/getEntitiesByTypeAndValue?type=File&value=html&mode=contains"
    ```

-----

### `/v1/deleteEntity`

  * **Method:** `DELETE`
  * **Purpose:** Deletes a specific entity from GITS by its `Type` and `ID`.
  * **URL Parameters:**
      * `type` (required, string): The entity type.
      * `id` (required, integer): The unique ID of the entity.
  * **Response (200 OK):** Empty body.
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl -X DELETE http://localhost:8080/v1/deleteEntity?type=User&id=123
    ```

-----

### `/v1/updateEntity`

  * **Method:** `PUT`
  * **Purpose:** Updates an existing entity's `Value`, `Context`, `Properties`, or `Version`.
  * **Request Body:** A JSON object representing a `transport.TransportEntity`. The `ID` and `Type` fields are crucial for identifying the entity to update.
    ```json
    {
      "Type": "string",
      "ID": 123,
      "Value": "string",
      "Context": "string",
      "Properties": {}
    }
    ```
  * **Response (200 OK):** Empty body.
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON.
  * **Example:**
    ```bash
    curl -X PUT http://localhost:8080/v1/updateEntity \
         -H "Content-Type: application/json" \
         -d '{
               "Type": "Task",
               "ID": 789,
               "Value": "Updated Task Description",
               "Properties": {"status": "completed"}
             }'
    ```

-----

### `/v1/getChildEntities`

  * **Method:** `GET`
  * **Purpose:** Retrieves all direct child entities linked from a specified source entity.
  * **URL Parameters:**
      * `type` (required, string): The type of the source entity.
      * `id` (required, integer): The ID of the source entity.
      * `context` (optional, string): Filter child relations by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of `transport.TransportEntity` objects that are children of the source.
    ```json
    {
      "Entities": [
        {
          "ID": 1001,
          "Type": "Subtask",
          "Value": "Subtask A",
          "Context": "Project",
          "Properties": {}
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getChildEntities?type=Project&id=500
    curl http://localhost:8080/v1/getChildEntities?type=Project&id=500&context=Contains
    ```

-----

### `/v1/getParentEntities`

  * **Method:** `GET`
  * **Purpose:** Retrieves all direct parent entities that link to a specified target entity.
  * **URL Parameters:**
      * `type` (required, string): The type of the target entity.
      * `id` (required, integer): The ID of the target entity.
      * `context` (optional, string): Filter parent relations by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of `transport.TransportEntity` objects that are parents of the target.
    ```json
    {
      "Entities": [
        {
          "ID": 2001,
          "Type": "Team",
          "Value": "Development Team",
          "Context": "Organization",
          "Properties": {}
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getParentEntities?type=Employee&id=600
    ```

-----

### `/v1/getRelationsTo`

  * **Method:** `GET`
  * **Purpose:** Retrieves all relations that point *to* a specified target entity (`TargetID`, `TargetType`).
  * **URL Parameters:**
      * `type` (required, string): The type of the target entity.
      * `id` (required, integer): The ID of the target entity.
      * `context` (optional, string): Filter relations by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of `transport.TransportRelation` objects.
    ```json
    {
      "Entities": [],
      "Relations": [
        {
          "SourceID": 101,
          "SourceType": "User",
          "TargetID": 202,
          "TargetType": "Role",
          "Context": "HasRole",
          "Properties": {"assigned_by": "admin"},
          "Version": 1
        }
      ]
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getRelationsTo?type=Role&id=202
    ```

-----

### `/v1/getRelationsFrom`

  * **Method:** `GET`
  * **Purpose:** Retrieves all relations that originate *from* a specified source entity (`SourceID`, `SourceType`).
  * **URL Parameters:**
      * `type` (required, string): The type of the source entity.
      * `id` (required, integer): The ID of the source entity.
      * `context` (optional, string): Filter relations by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of `transport.TransportRelation` objects.
    ```json
    {
      "Entities": [],
      "Relations": [
        {
          "SourceID": 101,
          "SourceType": "User",
          "TargetID": 303,
          "TargetType": "Permission",
          "Context": "Grants",
          "Properties": {"level": "read"},
          "Version": 1
        }
      ]
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getRelationsFrom?type=User&id=101
    ```

-----

### `/v1/getRelation`

  * **Method:** `GET`
  * **Purpose:** Retrieves a specific relation between a source entity and a target entity.
  * **URL Parameters:**
      * `srcType` (required, string): The type of the source entity.
      * `srcID` (required, integer): The ID of the source entity.
      * `targetType` (required, string): The type of the target entity.
      * `targetID` (required, integer): The ID of the target entity.
  * **Response Body (200 OK):** A `transport.Transport` object containing the requested `transport.TransportRelation`.
    ```json
    {
      "Entities": [],
      "Relations": [
        {
          "SourceType": "User",
          "SourceID": 101,
          "TargetType": "Role",
          "TargetID": 202,
          "Context": "HasRole",
          "Properties": {},
          "Version": 1
        }
      ]
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getRelation?srcType=User&srcID=101&targetType=Role&targetID=202
    ```

-----

### `/v1/getEntitiesByValue`

  * **Method:** `GET`
  * **Purpose:** Retrieves all entities that match a given `Value` across all types.
  * **URL Parameters:**
      * `value` (required, string): The value to match against the `Value` field of entities.
      * `mode` (optional, string): Match mode. Defaults to `match`. Can be `match`, `contains`, `startsWith`, `endsWith`, `fuzzy`.
      * `context` (optional, string): Filter entities by their `Context` field.
  * **Response Body (200 OK):** A `transport.Transport` object containing an array of matching `transport.TransportEntity` objects.
    ```json
    {
      "Entities": [
        {
          "ID": 101,
          "Type": "Username",
          "Value": "admin",
          "Context": "Authentication",
          "Properties": {}
        },
        {
          "ID": 201,
          "Type": "Hostname",
          "Value": "admin-server",
          "Context": "Network",
          "Properties": {}
        }
      ],
      "Relations": []
    }
    ```
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing required URL parameter `value`.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getEntitiesByValue?value=admin
    curl "http://localhost:8080/v1/getEntitiesByValue?value=server&mode=contains"
    ```

-----

### `/v1/getEntityTypes`

  * **Method:** `GET`
  * **Purpose:** Retrieves a list of all entity types currently known and stored in GITS.
  * **Response Body (200 OK):** A JSON object where keys are type names and values are their internal IDs (e.g., `{"User":1, "Domain":2}`).
    ```json
    {
      "User": 1,
      "Domain": 2,
      "IP": 3
    }
    ```
  * **Error Responses:**
      * `500 Internal Server Error`: If there's an issue marshalling the response data.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/getEntityTypes
    ```

-----

### `/v1/updateRelation`

  * **Method:** `PUT`
  * **Purpose:** Updates the `Context`, `Properties`, or `Version` of an existing relation between two entities.
  * **Request Body:** A JSON object representing a `transport.TransportRelation`. `SourceType`, `SourceID`, `TargetType`, and `TargetID` are required to identify the relation.
    ```json
    {
      "SourceType": "string",
      "SourceID": 123,
      "TargetType": "string",
      "TargetID": 456,
      "Context": "string",      // New context
      "Properties": {}          // New properties
    }
    ```
  * **Response (200 OK):** Empty body.
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON.
  * **Example:**
    ```bash
    curl -X PUT http://localhost:8080/v1/updateRelation \
         -H "Content-Type: application/json" \
         -d '{
               "SourceType": "Project",
               "SourceID": 100,
               "TargetType": "Task",
               "TargetID": 200,
               "Context": "HasCriticalTask",
               "Properties": {"priority": "high"}
             }'
    ```

-----

### `/v1/createRelation`

  * **Method:** `POST`
  * **Purpose:** Creates a new relation between two existing entities.
  * **Request Body:** A JSON object representing a `transport.TransportRelation`.
    ```json
    {
      "SourceType": "string",
      "SourceID": 123,
      "TargetType": "string",
      "TargetID": 456,
      "Context": "string",      // Optional
      "Properties": {}          // Optional
    }
    ```
  * **Response (200 OK):** Empty body.
  * **Error Responses:**
      * `422 Unprocessable Entity`: Invalid HTTP method, malformed body, or invalid JSON.
  * **Example:**
    ```bash
    curl -X POST http://localhost:8080/v1/createRelation \
         -H "Content-Type: application/json" \
         -d '{
               "SourceType": "User",
               "SourceID": 101,
               "TargetType": "Group",
               "TargetID": 501,
               "Context": "MemberOf"
             }'
    ```

-----

### `/v1/deleteRelation`

  * **Method:** `DELETE`
  * **Purpose:** Deletes a specific relation between a source and target entity.
  * **URL Parameters:**
      * `srcType` (required, string): The type of the source entity.
      * `srcID` (required, integer): The ID of the source entity.
      * `targetType` (required, string): The type of the target entity.
      * `targetID` (required, integer): The ID of the target entity.
  * **Response (200 OK):** Empty body.
  * **Error Responses:**
      * `422 Unprocessable Entity`: Missing or invalid URL parameters.
  * **Example:**
    ```bash
    curl -X DELETE http://localhost:8080/v1/deleteRelation?srcType=User&srcID=101&targetType=Group&targetID=501
    ```

-----

### Statistics

-----

### `/v1/statistics/getEntityAmount`

  * **Method:** `GET`
  * **Purpose:** Retrieves the total number of entities stored in the GITS instance.
  * **Response (200 OK):** `text/plain` body with the integer count.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/statistics/getEntityAmount
    # Response: 12345
    ```

-----

### `/v1/statistics/getEntityAmountByType`

  * **Method:** `GET`
  * **Purpose:** Retrieves the number of entities for a specific type.
  * **URL Parameters:**
      * `type` (required, string): The entity type (e.g., `Domain`, `IP`).
  * **Response (200 OK):** `text/plain` body with the integer count.
  * **Error Responses:**
      * `404 Not Found`: Missing required URL parameter `type` or unknown entity type.
  * **Example:**
    ```bash
    curl http://localhost:8080/v1/statistics/getEntityAmountByType?type=Domain
    # Response: 500
    ```

-----

## Changelog

[Full Changelog](CHANGELOG.md) - [Latest Release](https://www.google.com/search?q=https://github.com/voodooEntity/gitsapi/releases)

-----

## License

[GNU General Public License v3.0](https://www.google.com/search?q=./LICENSE)

-----

> [laughingman.dev](https://blog.laughingman.dev)  · 
> GitHub [@voodooEntity](https://github.com/voodooEntity)



```
```