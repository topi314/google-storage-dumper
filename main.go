package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	storageURL string
	bucketName string
	dataFolder string
)

type ListBucketResult struct {
	XMLName     xml.Name  `xml:"ListBucketResult"`
	Name        string    `xml:"Name"`
	Prefix      string    `xml:"Prefix"`
	Marker      string    `xml:"Marker"`
	NextMarker  string    `xml:"NextMarker"`
	IsTruncated bool      `xml:"IsTruncated"`
	Contents    []Content `xml:"Contents"`
}

type Content struct {
	Key            string `xml:"Key"`
	Generation     int    `xml:"Generation"`
	Metageneration int    `xml:"Metageneration"`
	LastModified   string `xml:"LastModified"`
	ETag           string `xml:"ETag"`
	Size           int64  `xml:"Size"`
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func main() {
	flag.StringVar(&storageURL, "storage-url", "https://storage.googleapis.com/", "Google Cloud Storage URL")
	flag.StringVar(&bucketName, "bucket-name", "pokemongolive", "Bucket name to download from")
	dataFolder = flag.Arg(0)
	if dataFolder == "" {
		dataFolder = "./data"
	}
	flag.Parse()

	log.Printf("Storage URL: %s", storageURL)
	log.Printf("Bucket Name: %s", bucketName)
	log.Printf("Data Folder: %s", dataFolder)

	if err := os.MkdirAll(dataFolder, 0777); err != nil {
		log.Fatalf("failed to create data folder: %v", err)
		return
	}

	contents, err := getAllBucketList()
	if err != nil {
		log.Fatalf("Error getting bucket list: %v", err)
		return
	}

	var er errgroup.Group
	er.SetLimit(10)

	for _, content := range contents.Contents {
		er.Go(func() error {
			if content.Size == 0 {
				// Skip empty files
				return nil
			}
			log.Printf("Key: %s, Size: %d, LastModified: %s\n", content.Key, content.Size, content.LastModified)
			return getContent(content.Key)
		})
	}
	if err = er.Wait(); err != nil {
		log.Fatalf("Error downloading files: %v", err)
		return
	}

	log.Println("All files downloaded successfully.")
}

func getAllBucketList() (*ListBucketResult, error) {
	var allContents []Content
	var lastResult ListBucketResult

	for {
		url := storageURL + bucketName
		if lastResult.NextMarker != "" {
			url += "?marker=" + lastResult.NextMarker
		}

		result, err := getBucketList(url)
		if err != nil {
			return nil, fmt.Errorf("failed to get bucket list: %w", err)
		}

		log.Printf("Fetched %d items from bucket %q, next marker: %q\n", len(result.Contents), result.Name, result.NextMarker)
		lastResult = *result
		allContents = append(allContents, result.Contents...)
		if !result.IsTruncated {
			break
		}
	}

	lastResult.Contents = allContents

	return &lastResult, nil
}

func getBucketList(url string) (*ListBucketResult, error) {
	rq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rs, err := client.Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to access bucket: %s", rs.Status)
	}

	var result ListBucketResult
	err = xml.NewDecoder(rs.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func getContent(key string) error {
	rq, err := http.NewRequest(http.MethodGet, storageURL+bucketName+"/"+key, nil)
	if err != nil {
		return err
	}

	rs, err := client.Do(rq)
	if err != nil {
		return err
	}

	defer rs.Body.Close()
	if rs.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to access content: %s", rs.Status)
	}

	filePath := filepath.Join(dataFolder, key)
	if err = os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %v", filePath, err)
	}
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}

	defer outFile.Close()
	if _, err = outFile.ReadFrom(rs.Body); err != nil {
		return fmt.Errorf("failed to write content to file %s: %v", filePath, err)
	}

	return nil
}
