package couchdbUtils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func ExtractStatusCode(input string, defaultStatusCode ...int) (statusCode int, err error) {

	defaultStatusCodeFound := func() int {

		if len(defaultStatusCode) > 0 {
			return defaultStatusCode[0]
		}
		return 404
	}()
	// Utilisation d'une expression régulière pour extraire le code de statut
	re := regexp.MustCompile(`\[STATUSCODE:(\d+)\](.+)`)

	// Trouver les correspondances dans la chaîne d'entrée
	matches := re.FindStringSubmatch(input)

	// Vérifier s'il y a des correspondances
	if len(matches) < 2 {
		return defaultStatusCodeFound, fmt.Errorf("no status code found in input string")
	}

	// La sous-chaîne correspondant au code de statut est à l'indice 1
	statusCodeString := matches[1]

	statusCode, err = strconv.Atoi(statusCodeString)
	if err != nil {
		return defaultStatusCodeFound, fmt.Errorf("impossible to convert status code")
	}

	return statusCode, nil
}

func HandleAndSendError(err error, w http.ResponseWriter, methodName string, defaultStatusCode ...int) {
	errorString := fmt.Sprintf("%s", err)

	statusCode, _ := ExtractStatusCode(errorString, defaultStatusCode...)

	result := ErrorResponse{
		StatusCode: statusCode,
		Message:    errorString,
	}
	log.Println("xxxx[ERROR]xxxx - func : ", methodName, " :: ", errorString)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	return
}

func NewError(statusCode int, message string) error {
	return errors.New("[STATUSCODE:" + strconv.Itoa(statusCode) + "]" + message)
}
