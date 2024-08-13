package couchdbUtils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sync"

	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	uuid "github.com/satori/go.uuid"
)

func includes(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func GetWhereFormater(tableLable string, isAnd bool, startWithAnd bool, where ...string) (query string) {

	query = ""

	if len(where) > 0 {
		if startWithAnd {
			query += "AND "
		} else {
			query += "WHERE "
		}

		for i := 0; i < len(where); i++ {

			if tableLable != "" {
				query += tableLable + "." + where[i]
			} else {
				query += where[i]
			}
			if i < len(where)-1 {
				if isAnd {
					query += " AND "
				} else {
					query += " OR "
				}

			}
		}
	}

	return query
}

func GetTotalCount(cluster *gocb.Cluster, collection *gocb.Collection, where ...string) (result int, err error) {
	methodName := "GetTotalCount"
	PrintStart(methodName)

	var query = ""
	tableName := collection.Bucket().Name() + "." + collection.ScopeName() + "." + collection.Name()

	query = "SELECT COUNT(*) AS total FROM " + tableName + " AS Item "

	query += GetWhereFormater("Item", true, false, where...)

	type TotalResult struct {
		Total int `json:"total"`
	}

	var totalFound TotalResult
	totalFound, err = ExecuteQueryOneResult[TotalResult](cluster, query)

	if err != nil {
		PrintError(methodName, err)
		return 0, err
	}

	result = totalFound.Total

	PrintSuccess(methodName)
	return result, nil
}

func GetTotalCountFromQuery(cluster *gocb.Cluster, queryBase string) (result int, err error) {
	methodName := "GetTotalCount"
	PrintStart(methodName)

	var query = ""

	queryWordList := strings.Split(queryBase, " ")

	for i := 0; i < len(queryWordList); i++ {
		if queryWordList[i] == "SELECT" {
			queryWordList[i+1] = "COUNT(" + queryWordList[i+1] + ") AS Total"
		}

	}

	query = strings.Join(queryWordList, " ")

	type TotalResult struct {
		Total int `json:"total"`
	}

	var totalFound TotalResult
	totalFound, err = ExecuteQueryOneResult[TotalResult](cluster, query)

	if err != nil {
		PrintError(methodName, err)
		return 0, err
	}

	result = totalFound.Total

	PrintSuccess(methodName)
	return result, nil
}

func ExecuteQueryOneResultAsync[T any](cluster *gocb.Cluster, query string, resultChan chan T, errChan chan error, wg *sync.WaitGroup) bool {
	methodName := "ExecuteQueryOneResultAsync"
	PrintSuccess(methodName)
	defer wg.Done()
	result, err := ExecuteQueryOneResult[T](cluster, query)

	if err != nil {
		errChan <- err
		PrintError(methodName, err)

		return false
	}
	resultChan <- result
	errChan <- err

	PrintSuccess(methodName, err)
	return true
}

/**
* Execute query
*
**/
func ExecuteQuery[T any](cluster *gocb.Cluster, query string) (result []T, err error) {
	methodName := "ExecuteQuery"
	PrintStart(methodName)
	rows, err := cluster.Query(query, &gocb.QueryOptions{})

	if err != nil {
		PrintError(methodName, err)
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var item struct {
			Item T `json:"Item"`
		}

		err := rows.Row(&item)

		if err != nil {
			PrintError(methodName, "Scanning", err)
			return result, err
		}
		result = append(result, item.Item)
	}

	if result == nil {
		result = []T{}
	}

	PrintSuccess(methodName)
	return result, nil
}

func ExecuteQueryRaw[T any](cluster *gocb.Cluster, query string) (result []T, err error) {
	methodName := "ExecuteQueryRaw"
	PrintStart(methodName)
	rows, err := cluster.Query(query, &gocb.QueryOptions{})

	if err != nil {
		PrintError(methodName, err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item T

		err := rows.Row(&item)

		if err != nil {
			PrintError(methodName, "Scanning", err)
			return result, err
		}
		result = append(result, item)
	}

	if result == nil {
		result = []T{}
	}

	PrintSuccess(methodName)
	return result, nil
}

func ExecuteQueryRawAsync[T any](cluster *gocb.Cluster, query string, resultChan chan []T, errChan chan error, wg *sync.WaitGroup) bool {
	methodName := "ExecuteQueryRawAsync"
	PrintStart(methodName)
	defer wg.Done()
	result, err := ExecuteQueryRaw[T](cluster, query)

	if err != nil {
		PrintError(methodName, err)
		errChan <- err

		return false
	}

	resultChan <- result
	errChan <- err

	PrintSuccess(methodName)
	return true
}

func ExecuteQueryWithPagination[T any](cluster *gocb.Cluster, countQuery string, query string, pageSize int, currentPage int) (result ResponseListWithPagination[[]T], err error) {
	methodName := "ExecuteQueryWithPagination"

	totalCount := 0
	totalPage := 0

	PrintStart(methodName)
	PrintInfo(methodName, "Count Query", countQuery)
	PrintInfo(methodName, "Query", query)

	type TotalResult struct {
		Total int `json:"total"`
	}

	var totalFound TotalResult
	totalFound, err = ExecuteQueryOneResult[TotalResult](cluster, countQuery)

	if err != nil {
		PrintError(methodName, err)
		return result, err
	}

	totalCount = totalFound.Total

	if pageSize > 0 {
		totalPage = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	query += GetQueryPaginationFormated(pageSize, currentPage)

	dataFound, err := ExecuteQueryRaw[T](cluster, query)
	if err != nil {
		PrintError(methodName, err)
		return result, err
	}

	if len(dataFound) > 0 {

		p := dataFound[0]
		jsonData, _ := json.Marshal(p)
		fmt.Println(string(jsonData))
	}

	result.Data = dataFound
	result.Pagination.CurrentPage = currentPage
	result.Pagination.PageSize = pageSize
	result.Pagination.TotalPages = totalPage
	result.Pagination.TotalCount = totalCount

	PrintSuccess(methodName)
	return result, err

}

func ExecuteQueryWithPaginationAsync[T any](cluster *gocb.Cluster, countQuery string, query string, pageSize int, currentPage int) (result ResponseListWithPagination[[]T], err error) {
	methodName := "ExecuteQueryWithPagination"
	query += GetQueryPaginationFormated(pageSize, currentPage)
	PrintStart(methodName)
	PrintInfo(methodName, "Query : ", query)
	PrintInfo(methodName, "Count Query :", countQuery)
	type TotalResult struct {
		Total int `json:"total"`
	}
	resultChan := make(chan []T)
	resultTotalChan := make(chan TotalResult)
	errChan := make(chan error)
	var wg *sync.WaitGroup = &sync.WaitGroup{}

	wg.Add(1)
	go ExecuteQueryOneResultAsync[TotalResult](cluster, countQuery, resultTotalChan, errChan, wg)

	wg.Add(1)
	go ExecuteQueryRawAsync[T](cluster, query, resultChan, errChan, wg)

	go func() {

		wg.Wait()
		close(resultChan)
		close(errChan)
		close(resultTotalChan)
	}()

	result.Data = <-resultChan

	totalFound := <-resultTotalChan

	totalPage := 0
	if pageSize > 0 {
		totalCount := totalFound.Total
		totalPage = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	result.Pagination.TotalCount = totalFound.Total
	result.Pagination.CurrentPage = currentPage
	result.Pagination.PageSize = pageSize
	result.Pagination.TotalPages = totalPage

	err = <-errChan

	if err != nil {
		PrintError(methodName, err)
		return result, err
	}
	PrintSuccess(methodName)

	return result, err

}
func ExecuteQueryOneResult[T any](cluster *gocb.Cluster, query string) (result T, err error) {
	methodName := "ExecuteQueryOneResult"

	PrintStart(methodName)
	PrintInfo(methodName, "Query", query)
	rows, err := cluster.Query(query, &gocb.QueryOptions{})
	if err != nil {
		PrintError(methodName, err)
		return result, err
	}

	defer rows.Close()
	var resultTab = []T{}

	for rows.Next() {

		var item T
		err := rows.Row(&item)

		if err != nil {
			PrintError(methodName, "Scanning", err)
			return result, err
		}
		resultTab = append(resultTab, item)
	}

	if isSelectQuery(query) && (len(resultTab) <= 0) {
		err = errors.New("[STATUSCODE:404]No results")
		return result, err
	}

	if isSelectQuery(query) && (len(resultTab) > 0) {
		result = resultTab[0]
	}

	PrintSuccess(methodName, result)

	return result, nil
}

func GetQueryPaginationFormated(pageSize int, currentPage int) (query string) {
	var offset = 0

	if currentPage > 0 {
		offset = (currentPage - 1) * pageSize
	}
	var limit = pageSize

	if pageSize > 0 {
		query += " OFFSET  " + strconv.Itoa(offset)
	}

	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	return query
}

func GetUpdateFormater(dataToUpdate []string) (query string) {

	query = ""
	if len(dataToUpdate) > 0 {
		for i := 0; i < len(dataToUpdate); i++ {
			if i == 0 {
				query += "SET "
			}
			query += dataToUpdate[i]
			if i < len(dataToUpdate)-1 {
				query += ", "
			}
		}
	}

	return query
}

func UpdateFeilds[T any](collection *gocb.Collection, objectId string, oldData T, objectUpdated T, feildToIgnore ...string) (isUpdateDone bool, err error) {

	v := reflect.ValueOf(objectUpdated)
	old := reflect.ValueOf(oldData)

	for i := 0; i < v.NumField(); i++ {

		fieldValue := v.Field(i).Interface()
		oldValue := old.Field(i).Interface()

		currentField := v.Type().Field(i).Tag.Get("json")

		var oldValueString = fmt.Sprintf("%v", oldValue)
		var newValueString = fmt.Sprintf("%v", fieldValue)

		if includes(feildToIgnore, currentField) {
			continue
		} else if oldValueString == newValueString {
			continue
		} else {

			_, err = UpdateREPLACE(Update{
				ElementID:  objectId,
				Field:      currentField,
				NewData:    fieldValue,
				Collection: collection,
			})
			if err != nil {
				fmt.Print("Error Update", err)
				return false, err

			}
		}

	}

	return true, nil

}

func GetInsertFieldsFormatedQuery[T any](objectList []T, tableName string, feildToIgnore ...string) (query string) {

	var fieldList = "INSERT INTO " + tableName + " "
	var valueList = " VALUES"

	if len(objectList) == 0 {
		return ""
	}
	for index, object := range objectList {

		var fields []string
		var values []string
		v := reflect.ValueOf(object)
		for i := 0; i < v.NumField(); i++ {

			//fieldName := typeOfS.Field(i).Name
			fieldValue := v.Field(i).Interface()

			currentField := v.Type().Field(i).Tag.Get("json")
			if includes(feildToIgnore, currentField) {
				continue
			}

			// Si le champ est une chaîne, ajoutez des guillemets autour de sa valeur
			kind := reflect.TypeOf(fieldValue).Kind()
			isNumber := kind >= reflect.Int && kind <= reflect.Float64
			if isNumber {

				values = append(values, fmt.Sprintf("%v", fieldValue))
			} else if kind == reflect.TypeOf(time.Time{}).Kind() {

				values = append(values, fmt.Sprintf("'%v'", fieldValue.(time.Time).Format(time.RFC3339Nano)))

			} else {

				values = append(values, fmt.Sprintf("'%v'", fieldValue))
			}

			fields = append(fields, currentField)

		}

		if index == 0 {
			fieldList += "(" + strings.Join(fields, ", ") + ")"
		}

		valueList += " (" + strings.Join(values, ", ") + ")"
		if index != len(objectList)-1 {
			valueList += ","
		}
	}

	return fieldList + valueList

}

func GetValueFromField[T any](object T, fieldName string) (result string) {

	v := reflect.ValueOf(object)

	for i := 0; i < v.NumField(); i++ {

		currentField := v.Type().Field(i).Tag.Get("json")

		if currentField == fieldName {
			fieldValue := v.Field(i).Interface()
			kind := reflect.TypeOf(fieldValue).Kind()
			isNumber := kind >= reflect.Int && kind <= reflect.Float64
			if isNumber {
				return fmt.Sprintf("%v", fieldValue)
			} else if kind == reflect.TypeOf(time.Time{}).Kind() {
				return fmt.Sprintf("%v", fieldValue.(time.Time).Format(time.RFC3339Nano))
			} else {
				return fmt.Sprintf("%v", fieldValue)
			}
		}
	}

	return result
}

func GetCollectionNameQuery(collection *gocb.Collection) string {
	str := collection.Bucket().Name() + "." + collection.ScopeName() + "." + collection.Name()
	return str
}

func UpdateREPLACE(infos Update) (string, error) {
	methodName := "updateREPLACE"
	fmt.Println(methodName, infos.ElementID, infos)
	err := ReplaceValueToDB(infos.Collection, infos.ElementID, infos.Field, infos.NewData)
	if err != nil {
		fmt.Println(methodName, err)
		return "", err
	}
	return "updated successfully", nil
}

func ReplaceValueToDB(collection *gocb.Collection, id string, path string, infos interface{}) error {
	mops := []gocb.MutateInSpec{
		//gocb.ReplaceSpec(path, infos, &gocb.ReplaceSpecOptions{}),
		gocb.UpsertSpec(path, infos, &gocb.UpsertSpecOptions{
			CreatePath: true,
		}),
	}
	_, err := collection.MutateIn(id, mops, &gocb.MutateInOptions{
		Timeout: 1 * time.Second,
	})

	return err
}

func GetOrderByQueryFormated(tableName string, field string, idDesc bool) (query string) {

	tableLabel := ""
	if tableName != "" {
		tableLabel = tableName + "."
	}
	query = " ORDER BY " + tableLabel + field
	if idDesc {
		query += " DESC "
	}
	return query
}

func FormatValueWithOperator(value string) string {
	if value == "" {
		return ""
	}
	return " = '" + value + "'"
}

// Find the right operator for the value
func TransformOperator(input string, isValueString bool) string {
	// Définir une expression régulière pour capturer l'opérateur et la valeur
	re := regexp.MustCompile(`^([<>]=?)\s*(.*)$`)

	// Trouver des correspondances dans la chaîne d'entrée
	matches := re.FindStringSubmatch(input)

	if len(matches) == 3 {
		operator := matches[1]
		value := strings.TrimSpace(matches[2])

		// Construire la chaîne transformée
		if isValueString {
			return operator + " \"" + value + "\""
		}
		return operator + " " + value
	}

	if isValueString {
		return "= \"" + input + "\""
	}

	// Si aucune correspondance n'est trouvée, renvoyer la chaîne d'origine
	return "= " + input
}

func GenerateID() string {
	result, _ := CreateID()
	return result
}

func getValueFromField[T any](object T, fieldName string) (result string) {

	v := reflect.ValueOf(object)

	for i := 0; i < v.NumField(); i++ {

		currentField := v.Type().Field(i).Tag.Get("json")

		if currentField == fieldName {
			fieldValue := v.Field(i).Interface()
			kind := reflect.TypeOf(fieldValue).Kind()
			isNumber := kind >= reflect.Int && kind <= reflect.Float64
			if isNumber {
				return fmt.Sprintf("%v", fieldValue)
			} else if kind == reflect.TypeOf(time.Time{}).Kind() {
				return fmt.Sprintf("%v", fieldValue.(time.Time).Format(time.RFC3339Nano))
			} else {
				return fmt.Sprintf("%v", fieldValue)
			}
		}
	}

	return result
}

func CreateID() (string, error) {
	newID := uuid.NewV4()
	// if err != nil {
	// 	printError("CREATE ID - newID", err)
	// 	return "", err
	// }
	// printInfo("CREATE id", newID)
	return newID.String(), nil
}
func isSelectQuery(query string) bool {
	match, _ := regexp.MatchString(`(?i)^\s*SELECT\s+`, query)
	return match
}

func ExtractOrderByQuery(rawOrderBy string) string {
	result := rawOrderBy
	if strings.HasPrefix(rawOrderBy, "-") {
		field := strings.TrimPrefix(rawOrderBy, "-")
		result = field + " DESC"

	}

	return result
}

func PrintSuccess(methodName string, msg ...interface{}) {
	log.Println("_________[SUCCESS]_________", methodName, " => ", msg)
}
func PrintStart(methodName string) {
	log.Println("_________[START]_________", methodName, "_______")
}
func PrintInfo(methodName string, msg ...interface{}) {
	log.Println("*****[INFO]***** - func : ", methodName, " :: ", msg)
}
func PrintError(methodName string, msg ...interface{}) {
	log.Println("xxxx[ERROR]xxxx - func : ", methodName, " :: ", msg)
}
func PrintDebug(methodName string, printDebugStatus bool, msg ...interface{}) {
	if printDebugStatus {
		log.Println("-----[DEBUG]----- - func : ", methodName, " :: ", msg)

	}
}
func PrintStatus(index int, length int) {
	log.Println(" --- STATUS --- ", index+1, " / ", length)
}
