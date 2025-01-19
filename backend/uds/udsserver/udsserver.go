package udsserver

import "github.com/Junior-Green/heartbeats/uds"

type UDSServer struct {
	getHandlers    map[string]uds.UDSHandler
	postHandlers   map[string]uds.UDSHandler
	putHandlers    map[string]uds.UDSHandler
	deleteHandlers map[string]uds.UDSHandler
}

func (s *UDSServer) AddGetHandler(resource string, handler uds.UDSHandler) {
	s.getHandlers[resource] = handler
}
func (s *UDSServer) AddPostHandler(resource string, handler uds.UDSHandler) {
	s.postHandlers[resource] = handler
}
func (s *UDSServer) AddPutHandler(resource string, handler uds.UDSHandler) {
	s.putHandlers[resource] = handler
}
func (s *UDSServer) AddDeleteHandler(resource string, handler uds.UDSHandler) {
	s.deleteHandlers[resource] = handler
}

func (s *UDSServer) UDSRequestHandler() uds.UDSHandler {
	return func(req uds.UDSRequest) uds.UDSResponse {
		switch req.Action {
		case uds.GET:
			if _, ok := s.getHandlers[req.Resource]; !ok {
				return uds.UDSResponse{Status: uds.BadRequest}
			}
			return s.getHandlers[req.Resource](req)
		case uds.PUT:
			if _, ok := s.putHandlers[req.Resource]; !ok {
				return uds.UDSResponse{Status: uds.BadRequest}
			}
			return s.putHandlers[req.Resource](req)
		case uds.POST:
			if _, ok := s.postHandlers[req.Resource]; !ok {
				return uds.UDSResponse{Status: uds.BadRequest}
			}
			return s.postHandlers[req.Resource](req)
		case uds.DELETE:
			if _, ok := s.deleteHandlers[req.Resource]; !ok {
				return uds.UDSResponse{Status: uds.BadRequest}
			}
			return s.deleteHandlers[req.Resource](req)
		default:
			return uds.UDSResponse{Status: uds.BadRequest}
		}
	}
}
