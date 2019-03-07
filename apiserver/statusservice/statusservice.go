package statusservice

/*
Copyright 2017-2019 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"encoding/json"
	"github.com/crunchydata/postgres-operator/apiserver"
	msgs "github.com/crunchydata/postgres-operator/apiservermsgs"
	log "github.com/sirupsen/logrus"
	//"github.com/gorilla/mux"
	"net/http"
)

// StatusHandler ...
// pgo status mycluster
// pgo status --selector=env=research
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	var username, ns string

	//vars := mux.Vars(r)
	clientVersion := r.URL.Query().Get("version")

	namespace := r.URL.Query().Get("namespace")
	log.Debugf("StatusHandler parameters version [%s] namespace [%s]", clientVersion, namespace)

	username, err := apiserver.Authn(apiserver.STATUS_PERM, w, r)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	var resp msgs.StatusResponse
	if clientVersion != msgs.PGO_VERSION {
		resp = msgs.StatusResponse{}
		resp.Status = msgs.Status{Code: msgs.Error, Msg: apiserver.VERSION_MISMATCH_ERROR}
		json.NewEncoder(w).Encode(resp)
		return
	}

	ns, err = apiserver.GetNamespace(apiserver.Clientset, username, namespace)
	if err != nil {
		resp = msgs.StatusResponse{}
		resp.Status = msgs.Status{Code: msgs.Error, Msg: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp = Status(ns)

	json.NewEncoder(w).Encode(resp)
}
