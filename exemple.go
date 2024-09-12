package exmeple

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dias-steve/couchdbaseulisgo-main/entities"
	"github.com/dias-steve/couchdbaseulisgo-main/utils/couchdbUtils"

	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func mainExemple() {
	r := mux.NewRouter()

	apiPort := ":8080"
	//CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	})

	//Exemple of use of the couchdbUtils.ExpressRouterController

	GrafanaCredentialRouterExemple(r)

	handler := c.Handler(r)

	fmt.Println(time.Now().Format("2006/01/02 15:04:05"), ": INFO ::  Server is running on port", apiPort)

	log.Fatal(http.ListenAndServe(apiPort, handler))
}

// Exemple of use of the couchdbUtils.ExpressRouterController
func GrafanaCredentialRouterExemple(router *mux.Router) {

	var CoGrafanaCredentials *gocb.Collection // is to set

	var Cluster *gocb.Cluster // is to set

	hydateEntites := func(r *http.Request, id string, newCredential *entities.GrafanaCredentialsEntity, isUpdate bool) error {
		// Check if the writing is an update or a new document
		if !isUpdate {
			// if is not update > then is a creation, we set the new id generated to the entity
			newCredential.CredentialId = id
		}

		// ====================== Check Validity of the entity - BEGIN ======================
		// Check if the body is empty
		if newCredential == nil {
			// if the body is empty, we return an error > then the entity be not saved
			return couchdbUtils.NewError(400, "Body is required")
		}

		// Check if the required fields are filled
		if newCredential.PasswordGrafana == "" {
			// if the password is empty, we return an error > then the entity be not saved
			return couchdbUtils.NewError(400, "Password is required")
		}
		if newCredential.UsernameGrafana == "" {
			// if the password is empty, we return an error > then the entity be not saved
			return couchdbUtils.NewError(400, "Username is required")
		}

		// ====================== Check Validity of the entity - END  ======================

		// If all is ok, we return nil > then the entity will be saved
		return nil
	}
	expressRouterConfig := couchdbUtils.RouterConfig[entities.GrafanaCredentialsEntity, entities.GrafanaCredentialsDto]{
		Cluster:         Cluster,
		Router:          router,
		BaseURL:         "/credentials",
		IdKey:           "credential_id",
		Collection:      CoGrafanaCredentials,
		AuthMiddleware:  nil,
		WithMiddleware:  false,
		HydrateEntities: hydateEntites,
		BlackListMethod: []string{"PUT", "POST", "DELETE"},
	}

	// Exemple of use of the couchdbUtils.CreateCouchbaseContext
	//
	// The function will create the GET /credentials, POST /credentials, GET /credentials/search, GET /credentials/{id}, PUT /credentials/{id}, DELETE /credentials/{id}
	//
	//couchdbUtils.CreateCouchbaseContext[<entity in DB>, <entity shown out of the router>](<Cluster Object>, <RouterObject>, <AtttributeIDName>, <CollectionObject>, <MidlewareFunc> ,<isActivate Middleware> ,<Create or Update function>)
	couchdbUtils.ExpressRouterController[entities.GrafanaCredentialsEntity, entities.GrafanaCredentialsDto](expressRouterConfig)

}
