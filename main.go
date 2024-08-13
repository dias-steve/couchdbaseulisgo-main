package main

import (
	"couchbaseUtilsGo/entities"
	"couchbaseUtilsGo/utils/couchdbUtils"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	apiPort := ":8080"
	//CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	})

	GrafanaCredentialRouterExemple(r)

	handler := c.Handler(r)

	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), ": INFO ::  Server is running on port", apiPort)

	log.Fatal(http.ListenAndServe(apiPort, handler))
}

func GrafanaCredentialRouterExemple(router *mux.Router) {

	var CoGrafanaCredentials *gocb.Collection

	var Cluster *gocb.Cluster

	couchdbUtils.ExpressRouterController[entities.GrafanaCredentialsEntity, entities.GrafanaCredentialsEntity](Cluster, router, "credentials", "credential_id", CoGrafanaCredentials, nil, false, func(r *http.Request, id string, newCredential *entities.GrafanaCredentialsEntity, isUpdate bool) (err error) {
		if newCredential == nil {
			return couchdbUtils.NewError(400, "Body is required")
		}

		if !isUpdate {
			newCredential.CredentialId = id
		}

		if newCredential.PasswordGrafana == "" {
			return couchdbUtils.NewError(400, "Password is required")
		}
		if newCredential.UsernameGrafana == "" {
			return couchdbUtils.NewError(400, "Username is required")
		}

		return nil
	})

}
