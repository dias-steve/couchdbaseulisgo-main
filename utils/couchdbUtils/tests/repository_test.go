package tests

import (
	"fmt"
	"log"
	"testing"

	"github.com/dias-steve/couchdbaseulisgo-main/utils/couchdbUtils"

	"github.com/couchbase/gocb/v2"
)

type collectionListAuthor struct {
	AuthorList *gocb.Collection
}

type Author struct {
	Id        string `json:"author_id"`
	LastName  string `json:"author_lastname"`
	FirstName string `json:"author_firstname"`
}

func initCollectionListAuthor(couchbaseContext *couchdbUtils.CouchbaseContext[collectionListAuthor]) {

	methodName := " InitListCollection"
	log.Println("[INIT-Collection] ", methodName, " Hydrates Collection")
	Bucket := couchbaseContext.Bucket
	scope := couchbaseContext.Scope

	log.Println("[INIT-Bucket] ", Bucket.Name())

	log.Println("[INIT-Collection] ", methodName, " Hydrates Collection-start: ", scope)

	couchbaseContext.Collections = collectionListAuthor{
		AuthorList: Bucket.Scope(scope).Collection("testAuthorList"),
	}
	log.Println("[INIT-Collection] ", methodName, " SUCCESS")
}

func CreateCouchbaseContextTestAuthor() *couchdbUtils.CouchbaseContext[collectionListAuthor] {
	var indexList = [][]string{
		{"testAuthorList", "idx_author_id", "author_id"},
	}

	var secondaryIndexList = [][]string{}

	context := *couchdbUtils.CreateCouchbaseContext[collectionListAuthor]("Comptatest", "compta", "127.0.0.1:8091", "Administrator", "Steve2023", indexList, secondaryIndexList, initCollectionListAuthor)
	context.Cluster.Buckets().FlushBucket("Comptatest", nil)
	return &context

}

func TestRepositorySave(t *testing.T) {

	fmt.Println("hi")
	CouchBaseContext := CreateCouchbaseContextTestAuthor()

	fmt.Println(CouchBaseContext.BucketName)

	authorRepository := couchdbUtils.NewRepository[Author](CouchBaseContext.Cluster, CouchBaseContext.Collections.AuthorList, "author_id")

	authorSave, err := authorRepository.Save("1", Author{
		Id:        "1",
		LastName:  "Doe",
		FirstName: "John",
	})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if authorSave.Id != "1" || authorSave.LastName != "Doe" || authorSave.FirstName != "John" {
		t.Error("Save failed")
		t.Fail()
	}

}

func TestRepositoryFindById(t *testing.T) {

	CouchBaseContext := CreateCouchbaseContextTestAuthor()

	fmt.Println(CouchBaseContext.BucketName)

	authorRepository := couchdbUtils.NewRepository[Author](CouchBaseContext.Cluster, CouchBaseContext.Collections.AuthorList, "author_id")

	_, err := authorRepository.Save("1", Author{
		Id:        "1",
		LastName:  "Doe",
		FirstName: "John",
	})

	_, err = authorRepository.Save("2", Author{
		Id:        "2",
		LastName:  "Doe 2",
		FirstName: "John 2",
	})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	authorFound, err := authorRepository.FindOneByID("2")
	if authorFound.Id != "2" || authorFound.LastName != "Doe 2" || authorFound.FirstName != "John 2" {
		t.Error("Save failed")
		t.Fail()
	}

}

func TestRepositoryFindOne(t *testing.T) {

	CouchBaseContext := CreateCouchbaseContextTestAuthor()

	fmt.Println(CouchBaseContext.BucketName)

	authorRepository := couchdbUtils.NewRepository[Author](CouchBaseContext.Cluster, CouchBaseContext.Collections.AuthorList, "author_id")

	_, err := authorRepository.Save("1", Author{
		Id:        "1",
		LastName:  "Doe",
		FirstName: "John",
	})

	_, err = authorRepository.Save("2", Author{
		Id:        "2",
		LastName:  "Doe 2",
		FirstName: "John 2",
	})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	authorFound, err := authorRepository.FindOne("author_lastname = 'Doe 2'")

	if authorFound.Id != "2" || authorFound.LastName != "Doe 2" || authorFound.FirstName != "John 2" {
		t.Error("not found")
		t.Error(authorFound)
		t.Fail()
	}

}

func TestRepositorySearch(t *testing.T) {

	CouchBaseContext := CreateCouchbaseContextTestAuthor()

	fmt.Println(CouchBaseContext.BucketName)

	authorRepository := couchdbUtils.NewRepository[Author](CouchBaseContext.Cluster, CouchBaseContext.Collections.AuthorList, "author_id")

	_, err := authorRepository.Save("1", Author{
		Id:        "1",
		LastName:  "Doe",
		FirstName: "John",
	})
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	_, err = authorRepository.Save("2", Author{
		Id:        "2",
		LastName:  "Marta",
		FirstName: "Laura",
	})

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	authorFoundList, err := authorRepository.Search("Marta", 0, 0)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if len(authorFoundList.Data) == 0 {
		t.Error("No reults found")
		t.Fail()
	}
	authorFound := authorFoundList.Data[0]

	if authorFound.Id != "2" || authorFound.LastName != "Marta" || authorFound.FirstName != "Laura" {

		t.Error(authorFound)
		t.Fail()
	}

	t.Log(authorFound)

}
