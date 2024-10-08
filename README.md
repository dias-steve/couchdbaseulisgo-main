# couchbaseUtilsGo
Package to manage couchbaseDB


# Install the package go

```
go get https://github.com/dias-steve/couchdbaseulisgo-main.git
```

# EpressRouterController

 EpressRouterController is a function that creates a router for a given entity

 The function will create the following routes:
   - GET /baseURL with params: currentPage, pageSize, whereQuery, orderByQuery
   - GET /baseURL/search with params: query
   - GET /baseURL/{id}
   - PUT /baseURL/{id}
   - DELETE /baseURL/{id}
   - POST /baseURL

 ## GET /baseURL
 Return all entities with pagination
### Exemple of use:
```
/baseURL?currentPage=1&pageSize=10&whereQuery=field1==value1|field2<in>value2,value2&orderByQuery=-field1
```
   - Separator for parameter whereQuery: 
```
    |
```
   - Operator available for string comparaison: 
```
   =, !=, <, >, <=, >=, <in>, !<in>
```
   - Operator available for number comparaison:
```
    ==, !==, <<, <<=, >>=, <<in>>, !<<in>>
```
   - Separator for IN operator:
```
,
```

 ## GET /baseURL/search

 Return the entities that match the search
   - Exemple of use: /baseURL/search?query=example

 ## GET /baseURL/{id}

 Return the entity with the given id

 ## POST /baseURL

 Create a new entity

 ## PUT /baseURL/{id}

 Update the entity with the given id

 ## DELETE /baseURL/{id}

 Delete the entity with the given id

## Exemple of Use ExpressRouterController



### Creating of User Entity & DTO

```go

    type UserEntity{
        UserId string `json:"user_id"`
        SessionId string `json:"user_session_id"`
        Name string `json:"user_name"`
        LastName string `json:"user_lastname"`
        UserName string `json:"user_username"`
        Password string `json:"user_password"`
    }

```

We choose to not show the password attribut throw the API
```go
    type UserDTO{
        UserId string `json:"user_id"`
        SessionId string `json:"user_session_id"`
        Name string `json:"user_name"`
        LastName string `json:"user_lastname"`
        UserName string `json:"user_username"`
    }

```


### Init the variable

```go 
    //The cluster couchBase object
    var Cluster *gocb.Cluster 

    	cluster, err = gocb.Connect(
		clusterIP,
		gocb.ClusterOptions{
			Username: clusterAdmin,
			Password: pwdCluster,
		},
	)

    // the user couchBase object
    var userCollection *gocb.Collection 
    
    userCollection= Bucket.Scope(scope).Collection("userList"),

```

### Create the function hydrateEnties
This function will 
This function is called when a user entity is create of update
The goal of this function is to set the right information in the entite object

```go

    hydateEntites := func(r *http.Request, id string, newUser *entities.UserEntity, isUpdate bool) error {

		// Check if the writing is an update or a new document
		if !isUpdate {
			// if is not update > then is a creation, we set the new id generated to the entity
			neUser.UserId = id
		}
		// ====================== Check Validity of the entity - BEGIN ======================
		// Check if the body is empty
		if newUser == nil {
			// if the body is empty, we return an error > then the entity be not saved
			return couchdbUtils.NewError(400, "Body is required")
		}

		// ====================== Check Validity of the entity - END  ======================

		// If all is ok, we return nil > then the entity will be saved
		return nil
	}

```

### Create a GetPreWhereQueryGet (Optional)
This function is called when a Get is requested
This function add where condition to the requested
```go 
    getPreWhereQuery :=func(r *http.Request) []string {

        whereConditionList := []string{}
        
        userId := context.Get(r, "user")
        userIdFormated := userId.(string)

        //Here we want just the user of the session
        whereConditionList = append( whereConditionList , "user_session_id = '"+userIdFormated+"'")
		return whereConditionList
	}
```

### Initialise the expressRouterConfig object
```go
	expressRouterConfig := couchdbUtils.RouterConfig[entities.UserEntity, entities.UserDto]{
		Cluster:         Cluster,
		Router:          router,
		BaseURL:         "/users",
        // the name of the id  attribute of the entity
		IdKey:           "user_id",
		Collection:      UserCollection,

		AuthMiddleware:  nil,

		WithMiddleware:  false,

		HydrateEntities: hydateEntites,
        // the method to not expose
		BlackListMethod: []string{"PUT", "POST", "DELETE"},
        
        GetPreWhereQueryGet: getPreWhereQuery,

	}
```


### Run the ExpressRouterController

```go
    couchdbUtils.ExpressRouterController[entities.UserEntity, entities.UserDto](expressRouterConfig)

```



## Repository


