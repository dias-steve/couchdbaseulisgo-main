package couchdbUtils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func ExtractPageParamsFromRequest(r *http.Request) (pageSize int, currentPage int) {
	currentPageParams := r.URL.Query().Get("currentPage")
	pageSizeParams := r.URL.Query().Get("pageSize")

	currentPage = 0
	pageSize = 0

	if currentPageParams != "" && pageSizeParams != "" {
		currentPageConverted, errConvertCurrentPage := strconv.Atoi(currentPageParams)

		pageSizeConverted, errConvertPageSize := strconv.Atoi(pageSizeParams)

		if errConvertCurrentPage == nil || errConvertPageSize == nil {
			currentPage = currentPageConverted
			pageSize = pageSizeConverted
		}

	}
	return pageSize, currentPage
}

func ExtractWhereOrderByParamsFromRequest(r *http.Request) (where []string, orderBy string) {
	whereParams := r.URL.Query().Get("whereQuery")
	orderByParams := r.URL.Query().Get("orderByQuery")

	orderBy = ""

	if orderByParams != "" {
		orderBy = ExtractOrderByQuery(orderByParams)
	}

	if whereParams != "" {
		where = ConvertWhereParamToWhereList(whereParams)
	}

	return where, orderBy
}

func ConvertWhereParamToWhereList(input string) (wherelist []string) {

	listInputWhere := strings.Split(input, "|")

	for _, whereItem := range listInputWhere {
		field, operator, value, err := ExtractOperatorWhere(whereItem)
		fmt.Println(field, operator, value, err)
		if err != nil {
			continue
		}

		switch operator {
		case "<in>":
			value = strings.Replace(value, "[", "", -1)
			value = strings.Replace(value, "]", "", -1)
			tableValue := strings.Split(value, ",")

			valueResult := ""
			for i, val := range tableValue {
				valueResult += "'" + val + "'"

				if i < len(tableValue)-1 {
					valueResult += ","
				}
			}
			wherelist = append(wherelist, field+" "+"IN ["+valueResult+"]")

		case "!<in>":
			value = strings.Replace(value, "[", "", -1)
			value = strings.Replace(value, "]", "", -1)
			tableValue := strings.Split(value, ",")

			valueResult := ""
			for i, val := range tableValue {
				valueResult += "'" + val + "'"

				if i < len(tableValue)-1 {
					valueResult += ","
				}
			}
			wherelist = append(wherelist, field+" "+"NOT IN ["+valueResult+"]")

		case "<<in>>":
			wherelist = append(wherelist, field+" IN "+value)

		case "==":
			wherelist = append(wherelist, field+" = "+value)

		case "<<=":
			wherelist = append(wherelist, field+" <= "+value)

		case ">>=":
			wherelist = append(wherelist, field+" >= "+value)

		case "<<":
			wherelist = append(wherelist, field+" < "+value)

		case ">>":
			wherelist = append(wherelist, field+" > "+value)

		case "!<<in>>":
			wherelist = append(wherelist, field+" NOT IN "+value)

		case "!==":
			wherelist = append(wherelist, field+" NOT = "+value)

		case "!<<=":
			wherelist = append(wherelist, field+" NOT <= "+value)

		case "!>>=":
			wherelist = append(wherelist, field+" NOT >= "+value)

		case "!<<":
			wherelist = append(wherelist, field+" NOT < "+value)

		case "!>>":
			wherelist = append(wherelist, field+" NOT > "+value)

		default:
			wherelist = append(wherelist, field+" "+operator+" '"+value+"'")
		}

	}

	return wherelist
}

func ExtractOperatorWhere(input string) (field string, operartor string, value string, err error) {
	// double sign operator is for number comparison
	// single sign operator is for string comparison
	operators := []string{"!<<in>>", "<<in>>", "!<<=", "!>>=", "<<=", ">>=", "!<<", "!>>", "<<", ">>", "!==", "==", "!<=", "!>=", "<=", ">=", "!=", "=", "!<in>", "<in>", "!>", "!<", ">", "<"}

	for _, operatorFound := range operators {
		if strings.Contains(input, operatorFound) {
			operartor = operatorFound
			break
		}
	}

	tableString := strings.Split(input, operartor)

	if len(tableString) < 2 {
		return field, operartor, value, errors.New("No format no valid")
	}

	return tableString[0], operartor, tableString[1], nil
}
