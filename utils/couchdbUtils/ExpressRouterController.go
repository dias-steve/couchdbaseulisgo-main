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

// NewControllerGeneric will create a new controller for the given entity
func NewControllerGeneric[Entities any, Dto any](repository Repository[Entities], hydrateEntites func(r *http.Request, id string, entities *Entities, isUpdate bool) error) ControllerGeneric {
	return &controllerGeneric[Entities, Dto]{
		repository:     repository,
		converter:      NewDtoConverter[Entities, Dto](),
		hydrateEntites: hydrateEntites,
	}
}

// GetAll will return all the entities
//
// Exemple of use: /baseURL?currentPage=1&pageSize=10&whereQuery=field1==ok|field2<in>value2,value2&orderByQuery=-field1
//   - Separator for parameter whereQuery: |
//   - Operator available for string comparaison: =, !=, <, >, <=, >=, <in>, !<in>
//   - Operator available for number comparaison: ==, !==, <<, <>, <<=, >>=, <<in>>, !<<in>>
//   - Separator for in comparaison: ,
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

// GetBySearch will return the entities that match the search
//
// Exemple of use: /baseURL/search?query=example
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

// GetSingle will return the entity with the given id
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

// Create will create a new entity
func (c *controllerGeneric[Entities, Dto]) Create(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > Create"

	if r.Body == nil {
		err := NewError(400, "Body is required")
		HandleAndSendError(err, w, methodName)
		return
	}
	var entityDto Dto
	err := json.NewDecoder(r.Body).Decode(&entityDto)
	if err != nil {
		err = NewError(400, err.Error())
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

// Update will update the entity with the given id
func (c *controllerGeneric[Entities, Dto]) Update(w http.ResponseWriter, r *http.Request) {
	methodName := "Conroller Generic > Update"

	if r.Body == nil {
		err := NewError(400, "Body is required")
		HandleAndSendError(err, w, methodName)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	var entityDto Dto
	err := json.NewDecoder(r.Body).Decode(&entityDto)
	if err != nil {
		err = NewError(400, err.Error())
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

// Delete will delete the entity with the given id
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

// RouterConfig is the configuration for the router
type RouterConfig[Entities any, Dto any] struct {
	// cluster: the couchbase cluster
	Cluster *gocb.Cluster
	// router: the router to add the routes
	Router *mux.Router
	// baseURL: the base url for the entity
	BaseURL string
	// idKey: the key for the id document
	IdKey string
	// collection: the collection couchbase
	Collection *gocb.Collection
	// authMiddleware: the middleware for the authorization
	AuthMiddleware func(next http.HandlerFunc) http.HandlerFunc
	// withMiddleware: if the routes should use the middleware : set true to use the middleware
	WithMiddleware bool
	// hydrateEntites: a function to hydrate the entities when the entity is created or updated
	HydrateEntities func(r *http.Request, id string, entities *Entities, isUpdate bool) error
	// blackListMethod: the methods to not expose
	BlackListMethod []string
}

// # EpressRouterController is a function that creates a router for a given entity
//
// The function will create the following routes:
//   - GET /baseURL with params: currentPage, pageSize, whereQuery, orderByQuery
//   - GET /baseURL/search with params: query
//   - GET /baseURL/{id}
//   - PUT /baseURL/{id}
//   - DELETE /baseURL/{id}
//   - POST /baseURL
//
// # GET /baseURL
// Return all the entities
//   - Exemple of use: /baseURL?currentPage=1&pageSize=10&whereQuery=field1==ok|field2<in>value2,value2&orderByQuery=-field1
//   - Separator for parameter whereQuery: |
//   - Operator available for string comparaison: =, !=, <, >, <=, >=, <in>, !<in>
//   - Operator available for number comparaison: ==, !==, <<, <>, <<=, >>=, <<in>>, !<<in>>
//   - Separator for in comparaison: ,
//
// # GET /baseURL/search
//
// Return the entities that match the search
//   - Exemple of use: /baseURL/search?query=example
//
// # GET /baseURL/{id}
//
// # Return the entity with the given id
//
// # POST /baseURL
//
// # Create a new entity
//
// # PUT /baseURL/{id}
//
// # Update the entity with the given id
//
// # DELETE /baseURL/{id}
//
// Delete the entity with the given id
func ExpressRouterController[Entities any, Dto any](config RouterConfig[Entities, Dto]) {
	var controller = NewControllerGeneric[Entities, Dto](NewRepository[Entities](config.Cluster, config.Collection, config.IdKey), config.HydrateEntities)

	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		if config.AuthMiddleware != nil && config.WithMiddleware {
			return config.AuthMiddleware(next)
		}
		return next
	}

	if !includes(config.BlackListMethod, "GET") {
		config.Router.HandleFunc(config.BaseURL, middleware(controller.GetAll)).Methods("GET")
		config.Router.HandleFunc(config.BaseURL+"/search", middleware(controller.GetBySearch)).Methods("GET")
		config.Router.HandleFunc(config.BaseURL+"/{id}", middleware(controller.GetSingle)).Methods("GET")
	}

	if !includes(config.BlackListMethod, "POST") {
		config.Router.HandleFunc(config.BaseURL, middleware(controller.Create)).Methods("POST")
	}

	if !includes(config.BlackListMethod, "PUT") {
		config.Router.HandleFunc(config.BaseURL+"/{id}", middleware(controller.Update)).Methods("PUT")
	}

	if !includes(config.BlackListMethod, "DELETE") {
		config.Router.HandleFunc(config.BaseURL+"/{id}", middleware(controller.Delete)).Methods("DELETE")
	}

}
