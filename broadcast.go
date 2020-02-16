package main

func BroadcastRequest(request *Request) {
	// todo: mutex for local.RemoteInstances0
	for i, _ := range local.RemoteInstances {
		if local.RemoteInstances[i].connected {
			local.RemoteInstances[i].nextRequests <- request
		}
	}
}
