package tests

import (
	"fmt"
	"testing"
	"time"

	"gitlab.com/myjoule/couchbaseutilsgo/utils/couchdbUtils"
)

func TestGetInsertFeildsFormatedQuery(t *testing.T) {
	type ObjectT struct {
		Name string    `json:"object_name"`
		Age  int       `json:"object_age"`
		Date time.Time `json:"date"`
	}

	objectTInst1 := ObjectT{
		Name: "olà",
		Age:  29,
		Date: time.Now(),
	}

	objectTInst2 := ObjectT{
		Name: "olà2",
		Age:  30,
		Date: time.Now(),
	}

	list := []ObjectT{objectTInst1, objectTInst2}
	dateString := objectTInst1.Date.Format(time.RFC3339Nano)
	dateString2 := objectTInst2.Date.Format(time.RFC3339Nano)
	result := couchdbUtils.GetInsertFieldsFormatedQuery[ObjectT](list, "tableName")

	fmt.Println(result)

	t.Log(result)

	resultExpected := "INSERT INTO tableName (object_name, object_age, date) VALUES ('olà', 29, '" + dateString + "'), ('olà2', 30, '" + dateString2 + "')"
	if result != resultExpected {
		t.Error("GenerateSQLInsertFieldList failed")
		t.Fail()
	}

}

func TestGetValueFromField(t *testing.T) {
	type ObjectT struct {
		Name string    `json:"object_name"`
		Age  int       `json:"object_age"`
		Date time.Time `json:"date"`
	}

	objectTInst := ObjectT{
		Name: "olà",
		Age:  29,
		Date: time.Now(),
	}

	result := couchdbUtils.GetValueFromField[ObjectT](objectTInst, "object_name")

	fmt.Println(result)

	t.Log(result)

	if result != "olà" {
		t.Error("GetValueFromField failed")
		t.Fail()
	}

}
