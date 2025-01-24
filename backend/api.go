package main

import (
	"encoding/json"

	"github.com/Junior-Green/heartbeats/database"
	"github.com/Junior-Green/heartbeats/server"
	"github.com/Junior-Green/heartbeats/uds"
)

// handlePing returns a UDSHandler that handles ping requests.
// It sets the response status to uds.Success.
func handlePing() uds.UDSHandler {
	return func(_ uds.UDSRequest, r *uds.UDSResponse) {
		uds.Ok(r, nil)
	}
}

// handleGetAllServers handles the request to retrieve all servers from the database.
// It takes a pointer to a SqliteDatabase and returns a UDSHandler function.
// The handler function processes the UDSRequest and sets the UDSResponse accordingly.
// If there is an error retrieving the servers from the database, it sets the response status to BadRequest.
// If the servers are successfully retrieved, it marshals them into JSON and sets the response status to Success with the JSON payload.
func handleGetAllServers(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		servers, err := db.GetAllServers()
		if err != nil {
			uds.Error(res, "Error getting all servers", uds.Internal)
			return
		}

		bytes, err := json.Marshal(servers)
		if err != nil {
			uds.Error(res, "Error marshalling server array", uds.Internal)
			return
		}
		uds.Ok(res, bytes)
	}
}

// handleGetMetrics handles the retrieval of metrics from the database based on the host data provided in the request payload.
// It returns a UDSHandler function that processes the request and sets the appropriate response.
//
// Parameters:
//   - db: A pointer to a SqliteDatabase instance used to query metrics.
//
// The returned UDSHandler function performs the following steps:
//  1. Defines a local struct type 'body' to parse the JSON payload from the request.
//  2. Unmarshals the request payload into the 'body' struct. If unmarshalling fails, it sets the response status to BadRequest.
//  3. Queries the database for metrics using the host data from the unmarshalled payload. If the query fails, it sets the response status to BadRequest.
//  4. Marshals the retrieved metrics into JSON format. If marshalling fails, it sets the response status to BadRequest.
//  5. Sets the marshalled JSON as the response payload.
func handleGetMetrics(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		type body struct {
			Host string `json:"host"`
		}

		var b body
		if err := json.Unmarshal(req.Payload, &b); err != nil {
			uds.Error(res, "Error unmarshalling JSON payload", uds.Internal)
			return
		}

		metrics, err := db.GetMetricsByHost(b.Host)
		if err != nil {
			uds.Error(res, "Error getting metrics", uds.Internal)
			return
		}

		bytes, err := json.Marshal(metrics)
		if err != nil {
			uds.Error(res, "Error marshalling metrics array", uds.Internal)
			return
		}

		uds.Ok(res, bytes)
	}
}

// handleGetServerByHost handles the request to get a server by its host.
// It takes a SqliteDatabase instance as a parameter and returns a UDSHandler function.
// The handler function unmarshals the request payload to extract the host data,
// retrieves the server information from the database, and marshals the server data
// into the response payload. If any error occurs during these processes, it sets
// the response status to BadRequest.
//
// Parameters:
//   - db: A pointer to a SqliteDatabase instance.
//
// Returns:
//   - A UDSHandler function that processes the request and response.
func handleGetServerByHost(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		type body struct {
			Host string `json:"host"`
		}

		var b body
		if err := json.Unmarshal(req.Payload, &b); err != nil {
			res.Status = uds.BadRequest
			return
		}

		server, err := db.GetServerByHost(b.Host)
		if err != nil {
			res.Status = uds.BadRequest
			return
		}

		bytes, err := json.Marshal(server)
		if err != nil {
			res.Status = uds.BadRequest
			return
		}

		res.Status = uds.Success
		res.Payload = bytes
	}
}

// handleCreateServer handles the creation of a new server.
// It takes a SqliteDatabase instance as a parameter and returns a UDSHandler function.
// The UDSHandler function processes a UDSRequest and populates a UDSResponse.
// It expects the request payload to contain a JSON object with a "data" field representing the server to be created.
// If the payload is invalid or the server cannot be added to the database, it sets the response status to BadRequest.
// On successful creation of the server, it sets the response status to Success.
func handleCreateServer(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		type body struct {
			Data server.Server `json:"data"`
		}

		var s server.Server
		if err := json.Unmarshal(req.Payload, &s); err != nil {
			res.Status = uds.BadRequest
			return
		}

		if err := db.AddServer(s); err != nil {
			res.Status = uds.BadRequest
			return
		}

		res.Status = uds.Success
	}
}

// handleDeleteServerByHost handles the deletion of a server by its host.
// It takes a SqliteDatabase instance and returns a UDSHandler function.
// The handler function expects a UDSRequest with a JSON payload containing
// the host data to be deleted. If the payload is invalid or the deletion
// fails, it sets the response status to BadRequest. Otherwise, it sets the
// response status to Success.
//
// Parameters:
//   - db: A pointer to a SqliteDatabase instance.
//
// Returns:
//   - A UDSHandler function that processes the deletion request.
func handleDeleteServerByHost(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		type body struct {
			Host string `json:"host"`
		}

		var b body
		if err := json.Unmarshal(req.Payload, &b); err != nil {
			res.Status = uds.BadRequest
			return
		}

		if err := db.DeleteServerByHost(b.Host); err != nil {
			res.Status = uds.BadRequest
			return
		}

		res.Status = uds.Success
	}
}

// handleUpdateFavorite handles the update of a favorite status for a given host.
// It expects a JSON payload with the host and favorite status in the request body.
// If the payload is invalid, it responds with a BadRequest status.
// If the database update fails, it responds with an Error status.
// On success, it responds with a Success status.
//
// Parameters:
// - db: A pointer to the SqliteDatabase instance.
//
// Returns:
// - A UDSHandler function that processes the update favorite request.
func handleUpdateFavorite(db *database.SqliteDatabase) uds.UDSHandler {
	return func(req uds.UDSRequest, res *uds.UDSResponse) {
		type body struct {
			Host     string `json:"host"`
			Favorite bool   `json:"favorite"`
		}

		var b body
		if err := json.Unmarshal(req.Payload, &b); err != nil {
			res.Status = uds.BadRequest
			return
		}

		if err := db.UpdateFavoriteByHost(b.Host, b.Favorite); err != nil {
			res.Status = uds.Internal
			return
		}

		res.Status = uds.Success
	}
}
