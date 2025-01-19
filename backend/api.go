package main

import "github.com/Junior-Green/heartbeats/uds"

func handlePing() uds.UDSHandler {
	return func(u uds.UDSRequest) uds.UDSResponse {
		return uds.UDSResponse{Status: uds.Success}
	}
}
