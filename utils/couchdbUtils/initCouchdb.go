package couchdbUtils

import (
	"fmt"
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
)

type NotificationServerUrl struct {
	Expo string
}
type CouchbaseContext[T any] struct {
	Cluster               *gocb.Cluster
	Bucket                *gocb.Bucket
	BucketName            string
	Scope                 string
	clusterIp             string
	Collections           T
	NotificationServerURL NotificationServerUrl
}

func InitCouchDB[T any](couchBaseContext *CouchbaseContext[T], clusterIP string, bucketName string, scope string, clusterAdmin string, pwdCluster string, createIndex [][]string, createSecondaryIndex [][]string) {
	methodName := "initCouchDB new version"
	log.Println("[INIT] ", methodName, " Init DB")
	var err error

	cluster, err := gocb.Connect(
		clusterIP,
		gocb.ClusterOptions{
			Username: clusterAdmin,
			Password: pwdCluster,
		},
	)

	couchBaseContext.Cluster = cluster
	if err != nil {
		panic(err)
	}

	bucket := cluster.Bucket(bucketName)
	couchBaseContext.Bucket = bucket

	// We wait until the bucket is definitely connected and setup.
	err = bucket.WaitUntilReady(10*time.Second, nil)
	if err != nil {
		panic(err)
	}
	//collectionbucketName := bucket.DefaultCollection()

	// Create If not exist
	err = bucket.Collections().CreateScope(scope, &gocb.CreateScopeOptions{})
	if err != nil {
		log.Println("[WARNING]  ", methodName, ", error create scope", scope, err)
	}

	for _, infos := range createIndex {
		name := infos[0]
		err = bucket.Collections().CreateCollection(gocb.CollectionSpec{Name: name, ScopeName: scope}, &gocb.CreateCollectionOptions{})
		if err != nil {
			log.Println("[WARNING] ", methodName, ", error create collection ", name, err)
		}
	}

	for i, infos := range createIndex {
		PrintInfo(methodName, "creating index")
		PrintStatus(i, len(createIndex))
		name := infos[0]
		idxName := infos[1]
		id := infos[2]
		query := fmt.Sprintf("CREATE PRIMARY INDEX ON `default`:`%v`.`%v`.`%v`", bucketName, scope, name)
		rows, err := cluster.Query(query, &gocb.QueryOptions{})
		if err == nil {
			var resp interface{} // this could also be a specific type like Hotel
			rows.One(&resp)
		} else {
			PrintError(methodName, " ==> Creating index ", query, " - ", err)
		}
		query2 := fmt.Sprintf("CREATE INDEX `%v` ON %v.%v.%v(`%v`)", idxName, bucketName, scope, name, id)
		rows, err = cluster.Query(query2, &gocb.QueryOptions{})
		if err == nil {
			var resp interface{} // this could also be a specific type like Hotel
			rows.One(&resp)
		} else {
			PrintError(methodName, " ==> Creating index ", query2, " - ", err)
		}
	}
	for i, infos := range createSecondaryIndex {
		PrintInfo(methodName, "creating secondary index")
		PrintStatus(i, len(createSecondaryIndex))
		name := infos[0]
		idxName := infos[1]
		field := infos[2]

		query := fmt.Sprintf("CREATE INDEX `%v` ON %v.%v.%v(`%v`)", idxName, bucketName, scope, name, field)
		rows, err := cluster.Query(query, &gocb.QueryOptions{})
		if err == nil {
			var resp interface{} // this could also be a specific type like Hotel
			rows.One(&resp)
		} else {
			PrintError(methodName, " ==> Creating secondary index ", query, " - ", err)
		}
	}

	PrintSuccess(methodName, "INIT complete")
}
func CreateCouchbaseContext[T any](bucketName string, scope string, clusterIp string, clusterAdmin string, pwdCluster string, indexList [][]string, secondaryIndex [][]string, initListCollection func(*CouchbaseContext[T])) *CouchbaseContext[T] {
	log.Println("[INIT-COUCHBASE] ", " START")
	couchbaseContext := &CouchbaseContext[T]{
		BucketName: bucketName,
		Scope:      scope,
		clusterIp:  clusterIp,
	}

	InitCouchDB(couchbaseContext, clusterIp, bucketName, scope, clusterAdmin, pwdCluster, indexList, secondaryIndex)

	initListCollection(couchbaseContext)

	log.Println("[INIT-COUCHBASE] ", " SUCCESS - DB Ready")
	return couchbaseContext
}
