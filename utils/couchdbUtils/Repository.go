package couchdbUtils

import (
	"math"
	"time"

	"github.com/couchbase/gocb/v2"
)

// Repository
// Manage the CRUD operations for the database
// For one collection

type Repository[T any] interface {
	FindByInnerJoin(collectionToJoin *gocb.Collection, joinOnLeft string, joinOnRight string, where ...string) (result []T, err error)
	FindByInnerJoinWithPagination(collectionToJoin *gocb.Collection, joinOnLeft string, joinOnRight string, pageSize int, currentPage int, where ...string) (result ResponseListWithPagination[[]T], err error)
	FindOneByID(id string) (result T, err error)
	FindOne(where ...string) (result T, err error)
	FindAllWithPagination(pageSize int, currentPage int, oderBy string, where ...string) (result ResponseListWithPagination[[]T], err error)
	FindAll(where ...string) (result []T, err error)
	Delete(where ...string) (result bool, err error)
	DeleteById(id string) (result bool, err error)
	Save(uuid string, data T) (result T, err error)
	SaveBatch(data []T) (result bool, err error)
	UpdateById(teamId string, dataUpdate T, fieldToIgnore ...string) (result T, err error)
	GetTableName() (tableName string)
	UpdateTableFields(fieldsNewValue []string, where ...string) (result bool, err error)
	Search(querySearch string, pageSize int, currentPage int, where ...string) (result ResponseListWithPagination[[]T], err error)
}

type repository[T any] struct {
	cluster      *gocb.Cluster
	collection   *gocb.Collection
	indexIdField string
	tableName    string
}

// NewRepository
// Create a new repository
func NewRepository[T any](clusterInit *gocb.Cluster, collectionInit *gocb.Collection, indexIdField string) Repository[T] {
	return &repository[T]{
		cluster:      clusterInit,
		collection:   collectionInit,
		indexIdField: indexIdField,
		tableName:    collectionInit.Bucket().Name() + "." + collectionInit.ScopeName() + "." + collectionInit.Name(),
	}
}

// FindByInnerJoin
// Find documents in the collection by a inner join to an onther collection with a where condition
func (r *repository[T]) FindByInnerJoin(collectionToJoin *gocb.Collection, joinOnLeft string, joinOnRight string, where ...string) (result []T, err error) {

	methodName := "FindByInnerJoin"
	printStart(methodName)

	selectOpt := "Item"

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()

	tableNameTojoin := collectionToJoin.Bucket().Name() + "." + collectionToJoin.ScopeName() + "." + collectionToJoin.Name()
	query := "SELECT " + selectOpt + " FROM " + tableName + " AS Item "

	query += "INNER JOIN " + tableNameTojoin + " AS Itemjoinneds "
	query += "ON Item." + joinOnLeft + " = Itemjoinneds." + joinOnRight + " "

	query += GetWhereFormater("Itemjoinneds", true, false, where...)

	printInfo(methodName, query)
	result, err = ExecuteQuery[T](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return nil, err
	}

	printSuccess(methodName)
	return result, nil
}

// FindByInnerJoinWithPagination
// Find documents in the collection by a inner join to a onther collection with a where condition and pagination
func (r *repository[T]) FindByInnerJoinWithPagination(collectionToJoin *gocb.Collection, joinOnLeft string, joinOnRight string, pageSize int, currentPage int, where ...string) (result ResponseListWithPagination[[]T], err error) {

	data := []T{}
	methodName := "GetJoin"
	printStart(methodName)

	totalCount := 0
	totalPage := 0

	selectOpt := "Item"

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()

	tableNameTojoin := collectionToJoin.Bucket().Name() + "." + collectionToJoin.ScopeName() + "." + collectionToJoin.Name()
	query := "SELECT " + selectOpt + " FROM " + tableName + " AS Item "

	query += "INNER JOIN " + tableNameTojoin + " AS Itemjoinneds "
	query += "ON Item." + joinOnLeft + " = Itemjoinneds." + joinOnRight + " "

	query += GetWhereFormater("Itemjoinneds", true, false, where...)

	printInfo(methodName, query)

	totalCount, err = GetTotalCountFromQuery(r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return result, err
	}

	query += GetQueryPaginationFormated(pageSize, currentPage)

	if pageSize > 0 {
		totalPage = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	data, err = ExecuteQuery[T](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return result, err
	}

	printSuccess(methodName)

	result.Data = data
	result.Pagination.CurrentPage = currentPage
	result.Pagination.PageSize = pageSize
	result.Pagination.TotalPages = totalPage
	result.Pagination.TotalCount = totalCount

	return result, nil

}

// FindOneByID
// Find one document in the collection by id
func (r *repository[T]) FindOneByID(id string) (result T, err error) {
	methodName := "FindOneByID"
	printStart(methodName)

	value, err := r.collection.Get(id, &gocb.GetOptions{})

	if err != nil {
		printError(methodName, err)
		return result, err
	}
	value.Content(&result)

	printSuccess(methodName)
	return result, nil
}

// FindOne
// Find one document in the collection with a where condition
func (r *repository[T]) FindOne(where ...string) (result T, err error) {
	methodName := "FindOne"
	printStart(methodName)

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()
	query := "SELECT * FROM " + tableName + " AS Item "

	query += GetWhereFormater("Item", true, false, where...)

	type item struct {
		Item T `json:"item"`
	}
	resultItem, err := ExecuteQueryOneResult[item](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return result, err
	}

	result = resultItem.Item

	printSuccess(methodName, result)
	return result, nil
}

// FindAllWithPagination
// Find all documents in the collection with a where condition and pagination
func (r *repository[T]) FindAllWithPagination(pageSize int, currentPage int, orderBy string, where ...string) (result ResponseListWithPagination[[]T], err error) {

	methodName := "FindAllWithPagination"
	printStart(methodName)

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()

	query := "SELECT * FROM " + tableName + " AS Item "

	query += GetWhereFormater("Item", true, false, where...)
	if orderBy != "" {
		query += " ORDER BY " + orderBy + " "
	}

	queryCount := "SELECT COUNT(*) as total FROM " + tableName + " AS Item "
	queryCount += GetWhereFormater("Item", true, false, where...)
	type ResultQuery struct {
		Item T `json:"item"`
	}

	resultDataList := []T{}
	resultQuery, err := ExecuteQueryWithPaginationAsync[ResultQuery](r.cluster, queryCount, query, pageSize, currentPage)
	if err != nil {
		printError(methodName, err)
		return result, err
	}

	for _, resultQueryItem := range resultQuery.Data {
		resultDataList = append(resultDataList, resultQueryItem.Item)
	}
	printSuccess(methodName)

	result.Data = resultDataList
	result.Pagination = resultQuery.Pagination

	return result, nil
}

// FindAll
// Find all documents in the collection with a where condition
func (r *repository[T]) FindAll(where ...string) (result []T, err error) {
	methodName := "FindAll"
	printStart(methodName)
	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()
	query := "SELECT * FROM " + tableName + " AS Item "

	query += GetWhereFormater("Item", true, false, where...)

	result, err = ExecuteQuery[T](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return nil, err
	}

	printSuccess(methodName)
	return result, nil
}

// Detete
// Delete a document by a where condition
func (r *repository[T]) Delete(where ...string) (result bool, err error) {
	methodName := "Delete"
	printStart(methodName)

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()

	query := "DELETE FROM " + tableName + " AS Item "
	query += GetWhereFormater("Item", true, false, where...)

	_, err = ExecuteQueryOneResult[any](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return false, nil
	}

	printSuccess(methodName)
	return true, nil

}

// DeleteById
// Delete a document by id
func (r *repository[T]) DeleteById(id string) (result bool, err error) {
	methodName := "Delete"
	printStart(methodName)

	_, err = r.collection.Remove(id, &gocb.RemoveOptions{})

	if err != nil {
		printError(methodName, err)
		return false, err
	}

	printSuccess(methodName)
	return true, nil

}

// UpdateById
// Update a document by id
func (r *repository[T]) UpdateById(id string, dataUpdate T, fieldToIgnore ...string) (result T, err error) {
	methodName := "UpdateById"
	printStart(methodName)

	fieldToIgnore = append(fieldToIgnore, r.indexIdField)

	oldData, err := r.FindOneByID(id)
	if err != nil {
		printError(methodName, err)
		return result, err
	}

	_, err = UpdateFeilds(r.collection, id, oldData, dataUpdate, fieldToIgnore...)

	if err != nil {
		printError(methodName, err)
		return result, err
	}

	result, err = r.FindOneByID(id)

	if err != nil {
		printError(methodName, err)
		return result, err
	}

	printSuccess(methodName)
	return result, nil
}

// Save
// Save a new document in the collection
func (r *repository[T]) Save(uuid string, data T) (result T, err error) {

	_, err = r.collection.Insert(uuid, data, &gocb.InsertOptions{
		Timeout: 1 * time.Second,
	})

	if err != nil {
		printError("Save", "Error insert", err)
		return result, err
	}

	result, err = r.FindOneByID(uuid)

	if err != nil {
		printError("Save", err)
		return result, err
	}

	return result, nil

}

func (r *repository[T]) SaveBatch(data []T) (result bool, err error) {

	methodName := "SaveGroup"
	printStart(methodName)

	for i := 0; i < len(data); i++ {
		id := getValueFromField[T](data[i], r.indexIdField)
		_, err = r.collection.Insert(id, data[i], &gocb.InsertOptions{})

		if err != nil {
			printError(methodName, "Error insert", err)
			return false, err
		}
	}
	if err != nil {
		printError(methodName, err)
		return false, err
	}

	printSuccess(methodName)
	result = true

	return result, nil

}

func (r *repository[T]) GetTableName() (tableName string) {
	return r.tableName
}

func (r *repository[T]) UpdateTableFields(fieldsNewValue []string, where ...string) (result bool, err error) {
	methodName := "UpdateTableFields"
	printStart(methodName)

	tableName := r.GetTableName()

	query := "UPDATE " + tableName + " AS Item "
	query += "SET "
	for i, field := range fieldsNewValue {
		if i != 0 {
			query += ", "
		}
		query += field
	}
	query += " " + GetWhereFormater("Item", true, false, where...)

	_, err = ExecuteQueryOneResult[any](r.cluster, query)

	if err != nil {
		printError(methodName, err)
		return false, err
	}

	printSuccess(methodName)
	return true, nil

}

// search
func (r *repository[T]) Search(querySearch string, pageSize int, currentPage int, where ...string) (result ResponseListWithPagination[[]T], err error) {

	methodName := "Search"
	printStart(methodName)

	tableName := r.collection.Bucket().Name() + "." + r.collection.ScopeName() + "." + r.collection.Name()

	//Query condition

	query := "SELECT * FROM " + tableName + " AS Item "
	queryBase := "WHERE SEARCH(Item, '" + querySearch + "') "
	queryBase += GetWhereFormater("Item", true, true, where...)
	query += queryBase

	queryCount := "SELECT COUNT(*) as total FROM " + tableName + " AS Item "
	queryCount += queryBase

	type ResultQuery struct {
		Item T `json:"item"`
	}

	resultDataList := []T{}
	resultQuery, err := ExecuteQueryWithPaginationAsync[ResultQuery](r.cluster, queryCount, query, pageSize, currentPage)
	if err != nil {
		printError(methodName, err)
		return result, err
	}

	for _, resultQueryItem := range resultQuery.Data {
		resultDataList = append(resultDataList, resultQueryItem.Item)
	}

	result.Data = resultDataList
	result.Pagination = resultQuery.Pagination
	printSuccess(methodName, result.Data)

	return result, nil
}
