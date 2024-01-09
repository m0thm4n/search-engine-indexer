package elasticsearch

// Elastic search client
import (
	"context"
	"fmt"
	"log"
	"reflect"
	"search-engine-indexer/src/structs"
	"time"

	elastic "github.com/olivere/elastic/v7"
)

const (
	IndexName    = "recipes"
	IndexMapping = `{
		"settings":{
			"number_of_shards":1,
			"number_of_replicas":0,
			"analysis": {
	      "analyzer": {
	        "clean_html": {
						"type": "standard",
	          "char_filter": ["html_strip"]
	        }
	      }
	    }
		},
		"mappings":{
			"properties":{
				"title": {
					"type": "text"
				},
				"description": {
					"type": "text"
				},
				"body": {
					"type": "text"
				},
				"url": {
					"type": "text"
				}
			}
		}
	}`
)

var client *elastic.Client

// NewElasticSearchClient returns an elastic seach client
func NewElasticSearchClient() *elastic.Client {
	var err error
	connected := false
	retries := 0

	// Custom retry strategy for docker-compose initialization
	for connected == false {
		// Create a new elastic client
		client, err = elastic.NewClient(
			elastic.SetURL("http://localhost:9200"), elastic.SetSniff(false))
		if err != nil {
			// log.Fatal(err)
			if retries == 5 {
				log.Fatal(err)
			}
			fmt.Println("Elasticsearch isn't ready for connection", 5-retries, "less")
			retries++
			time.Sleep(3 * time.Second)
		} else {
			connected = true
		}
	}

	// Getting the ES version number is quite common, so there's a shortcut
	esversion, err := client.ElasticsearchVersion("http://localhost:9200")
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)

	return client
}

// ExistsIndex checks if the given index exists or not
func ExistsIndex(i string) bool {
	// Check if index exists
	exists, err := client.IndexExists(i).Do(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	return exists
}

// CreateIndex creates a new index
func CreateIndex(i string) {
	createIndex, err := client.CreateIndex(IndexName).
		Body(IndexMapping).
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if !createIndex.Acknowledged {
		log.Println("CreateIndex was not acknowledged. Check that timeout value is correct.")
	}
}

// DeleteIndex in the indexName constant
func DeleteIndex() {
	ctx := context.Background()
	deleteIndex, err := client.DeleteIndex(IndexName).Do(ctx)
	if err != nil {
		// Handle error
		log.Fatal(err)
	}
	if !deleteIndex.Acknowledged {
		log.Println("DeleteIndex was not acknowledged. Check that timeout value is correct.")
	}
	fmt.Println("Index", IndexName, "deleted")
}

// ExistingPage return a boolean and a page if the link is already
// stored in the database
func ExistingPage(link string) (bool, structs.Page) {
	var exists bool
	var p structs.Page

	ctx := context.Background()
	// Search for a page in the database using Term Query
	// q := elastic.NewTermQuery("url", link)
	q := elastic.NewMultiMatchQuery(link, "url")
	result, err := client.Search().
		Index(IndexName).
		Query(q).
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var ttyp structs.Page
	for _, result := range result.Each(reflect.TypeOf(ttyp)) {
		page := result.(structs.Page)
		if page.URL == link {
			exists, p = true, page
			return exists, p
		}
	}

	return exists, p
}

// CreatePage adds a new page to the database
func CreatePage(p structs.Page) bool {
	ctx := context.Background()

	_, err := client.Index().
		Index("recipes").
		Type("_doc").
		Id(p.ID).
		BodyJson(p).
		Do(ctx)

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// UpdatePage adds a new page to the database
func UpdatePage(id string, params map[string]interface{}) bool {
	ctx := context.Background()

	_, err := client.Update().Index(IndexName).Type("page").Id(id).Doc(params).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
