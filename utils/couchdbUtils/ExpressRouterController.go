package couchdbUtils

import (
	"net/http"

	"encoding/json"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
)

type ControllerGeneric interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetBySearch(w http.ResponseWriter, r *http.Request)
	GetSingle(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
}

type controllerGeneric[Entities any, Dto any] struct {
	repository     Repository[Entities]
	converter      DtoConverter[Entities, Dto]
	hydrateEntites func(r *http.Request, id string, entities *Entities, isUpdate bool) error
}

func NewControllerGeneric[Entities any, Dto any](repository Repository[Entities], hydrateEntites func(r *http.Request, id string, entities *Entities, isUpdate bool) error) ControllerGeneric {
	return &controllerGeneric[Entities, Dto]{
		repository:     repository,
		converter:      NewDtoConverter[Entities, Dto](),
		hydrateEntites: hydrateEntites,
	}
}

func (c *controllerGeneric[Entities, Dto]) GetAll(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > GetAll"
	pageSize, currentPage := ExtractPageParamsFromRequest(r)
	where, orderBy := ExtractWhereOrderByParamsFromRequest(r)
	result, err := c.repository.FindAllWithPagination(pageSize, currentPage, orderBy, where...)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	resultDto := c.converter.ConvertListToDtoWithPagination(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultDto)
}

func (c *controllerGeneric[Entities, Dto]) GetBySearch(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > GetAll"
	pageSize, currentPage := ExtractPageParamsFromRequest(r)
	search := r.URL.Query().Get("query")
	result, err := c.repository.Search(search, pageSize, currentPage)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	resultDto := c.converter.ConvertListToDtoWithPagination(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultDto)
}

func (c *controllerGeneric[Entities, Dto]) GetSingle(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > GetSingle"
	vars := mux.Vars(r)
	id := vars["id"]
	result, err := c.repository.FindOneByID(id)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	resultDto := c.converter.ToDto(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultDto)
}

func (c *controllerGeneric[Entities, Dto]) Create(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > Create"
	var entityDto Dto
	err := json.NewDecoder(r.Body).Decode(&entityDto)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}

	id, _ := CreateID()

	entity := c.converter.ToEntity(entityDto)

	err = c.hydrateEntites(r, id, &entity, false)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}

	result, err := c.repository.Save(id, entity)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	resultDto := c.converter.ToDto(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultDto)
}

func (c *controllerGeneric[Entities, Dto]) Update(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > Update"
	vars := mux.Vars(r)
	id := vars["id"]
	var entityDto Dto
	err := json.NewDecoder(r.Body).Decode(&entityDto)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}

	entity := c.converter.ToEntity(entityDto)
	err = c.hydrateEntites(r, id, &entity, true)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	result, err := c.repository.UpdateById(id, entity)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}
	resultDto := c.converter.ToDto(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultDto)
}

func (c *controllerGeneric[Entities, Dto]) Delete(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > Delete"
	vars := mux.Vars(r)
	id := vars["id"]
	_, err := c.repository.DeleteById(id)
	if err != nil {
		HandleAndSendError(err, w, methodName)
		return
	}

	w.WriteHeader(204)

}

// EpressRouterController is a function that creates a router for a given entity
// cluster: the couchbase cluster
// router: the router to add the routes
// baseURL: the base url for the entity
// idKey: the key for the id document
// collection: the collection for the entity
// authMiddleware: the middleware for the authorization
// withMiddleware: if the routes should use the middleware
// hydrateEntites: a function to hydrate the entities when the entity is created or updated
// the function will create the following routes: GET /baseURL, POST /baseURL, GET /baseURL/search, GET /baseURL/{id}, PUT /baseURL/{id}, DELETE /baseURL/{id}
func ExpressRouterController[Entities any, Dto any](cluster *gocb.Cluster, router *mux.Router, baseURL string, idKey string, collection *gocb.Collection, authMiddleware func(next http.HandlerFunc) http.HandlerFunc, withMiddleware bool, hydrateEntites func(r *http.Request, id string, entities *Entities, isUpdate bool) (err error)) {
	var controller = NewControllerGeneric[Entities, Dto](NewRepository[Entities](cluster, collection, idKey), hydrateEntites)

	if withMiddleware {
		router.HandleFunc(baseURL, authMiddleware(controller.GetAll)).Methods("GET")
		router.HandleFunc(baseURL, authMiddleware(controller.Create)).Methods("POST")
		router.HandleFunc(baseURL+"/search", authMiddleware(controller.GetBySearch)).Methods("GET")
		router.HandleFunc(baseURL+"/{id}", authMiddleware(controller.GetSingle)).Methods("GET")
		router.HandleFunc(baseURL+"/{id}", authMiddleware(controller.Update)).Methods("PUT")
		router.HandleFunc(baseURL+"/{id}", authMiddleware(controller.Delete)).Methods("DELETE")
	} else {
		router.HandleFunc(baseURL, controller.GetAll).Methods("GET")
		router.HandleFunc(baseURL, controller.Create).Methods("POST")
		router.HandleFunc(baseURL+"/search", controller.GetBySearch).Methods("GET")
		router.HandleFunc(baseURL+"/{id}", controller.GetSingle).Methods("GET")
		router.HandleFunc(baseURL+"/{id}", controller.Update).Methods("PUT")
		router.HandleFunc(baseURL+"/{id}", controller.Delete).Methods("DELETE")
	}

}
