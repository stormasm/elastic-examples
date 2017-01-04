package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"errors"
	"flag"
	"log"
	"math/rand"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"github.com/stormasm/elastic"
)

func main() {
	done := make(chan bool)
	url := "http://127.0.0.1:3000/omdb.json"
	json := getJson(url)
	fmt.Println(len(json))
	doc_chan := getChannel(json)
	//done <- true
	fmt.Println("Channel length ",len(doc_chan))

//func churn(newIndex <-chan float64, newData chan<- datum) {
go func() {

	//<-done // Wait for doc_chan to have data
	fmt.Println("processChannel length ",len(doc_chan))

	var (
		index    = flag.String("index", "", "Elasticsearch index name")
		typ      = flag.String("type", "", "Elasticsearch type name")
		n        = flag.Int("n", 0, "Number of documents to bulk insert")
		bulkSize = flag.Int("bulk-size", 0, "Number of documents to collect before committing")
	)
	flag.Parse()
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())

	if *index == "" {
		log.Fatal("missing index parameter")
	}
	if *typ == "" {
		log.Fatal("missing type parameter")
	}
	if *n <= 0 {
		log.Fatal("n must be a positive number")
	}
	if *bulkSize <= 0 {
		log.Fatal("bulk-size must be a positive number")
	}


	// Do a trace log
	tracelog := log.New(os.Stdout, "", 0)
	client, err := elastic.NewClient(elastic.SetTraceLog(tracelog))
	// Or with nothing...
	// client, err := elastic.NewClient()

	if err != nil {
		// Handle error
		log.Fatal(err)
	}



	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())

	g.Go(func() error {

		bulk := client.Bulk().Index(*index).Type(*typ)
		count := 0
		for d := range doc_chan {
			fmt.Println(count, " ", d)
			// Enqueue the document
			countstr := strconv.Itoa(count)
			bulk.Add(elastic.NewBulkStringRequest().Id(countstr).SetSource(d))
			count = count + 1
			if bulk.NumberOfActions() >= *bulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					// Look up the failed documents with res.Failed(), and e.g. recommit
					return errors.New("bulk commit failed")
				}
				// "bulk" is reset after Do, so you can reuse it
			}
		}

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			_, err := bulk.Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
	done <- true
}()
<-done
}

func getChannel(json []byte) <-chan string {
	doc_chan := make(chan string)
	go func() {
		reader := bytes.NewReader(json)
		scanner := bufio.NewScanner(reader)
		count := 0
		var doc string
		for scanner.Scan() {
			evenodd := count % 2
			if evenodd == 0 {
				scanner.Text()
			} else {
				doc = scanner.Text()
				doc_chan <- doc
			}
			count = count + 1
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		close(doc_chan)
	}()
	return doc_chan
}

func getJson(url string) (buf []byte) {
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}
	response, err := netClient.Get(url)
	if err != nil {
		fmt.Println("get: ", err)
	}

	buf, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ioutil: ", err)
	}
	return buf
}
