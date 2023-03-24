package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
	TB
)

// path represents the path at which the downloaded file will be written
var path string

// interval represent the amount of seconds between onProgress invocations
var interval int

// WriteCounter keeps track of the bytes being written by an io.Writer.
// A callback function can be provided and called at a set interval
type WriteCounter struct {
	downloaded int64                               // # bytes read
	total      int64                               // total # of bytes to read
	interval   time.Duration                       // seconds between onProgress calls
	onProgress func(downloaded int64, total int64) // callback
	done       chan bool                           // channel on which to send when the download is complete
}

// New instantiates a WriteCounter expecting to read the given size bytes
// and call the provided function at regular intervals
func New(size int64, interval int, onProgress func(downloaded, total int64)) *WriteCounter {
	return &WriteCounter{
		total:      size,
		interval:   time.Second * time.Duration(interval),
		onProgress: onProgress,
		done:       make(chan bool),
	}
}

// Write implements the io.Writer interface.
// Always completes and never returns an error.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.downloaded += int64(n)
	return n, nil
}

// NotifyProgress invokes the onProgress callback at regular intervals
// and stops on a send to the WriteCounter done channel
func (wc *WriteCounter) NotifyProgress() {
	ticker := time.NewTicker(wc.interval)
	start := time.Now()
	for {
		select {
		case <-wc.done:
			close(wc.done)
			ticker.Stop()
			return
		case t := <-ticker.C:
			elapsed := int(t.Sub(start).Seconds())
			fmt.Printf("Running for %d seconds ...\n", elapsed)
			wc.onProgress(wc.downloaded, wc.total)
		}
	}
}

// parseFlags parses command line args to setup flag variables
func parseFlags() error {
	flag.StringVar(&path, "p", "", "The path to which the file will be downloaded")
	flag.IntVar(&interval, "i", 1, "The interval at which the logs will be produced, in seconds (must be >= 1)")
	flag.Parse()

	if interval < 1 {
		return errors.New("interval must be >= 1")
	}
	return nil
}

// isCompleteUrl returns a boolean representing if the given string
// can be parsed to a complete, valid, absolute URL or not
func isCompleteUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// getFileSize parses the content-length header in a HTTP Response
// and returns the downloaded file size
func getFileSize(resp *http.Response) (int, error) {
	contentLength := resp.Header.Get("Content-Length")
	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}
	return length, nil
}

// download setups a WriteCounter and pipes the content of an HTTP response's body through it,
// then copies it to the desired (or default) file path
func download(url string, onProgress func(downloaded, total int64)) error {
	client := http.DefaultClient
	resp, err := client.Head(url)
	if err != nil {
		return err
	}

	// parse header to get file size
	size, err := getFileSize(resp)
	if err != nil {
		return err
	}
	// if path was not specifiedc
	// parse URL to retrieve filename
	if path == "" {
		path = string(url[strings.LastIndex(url, "/")+1:])
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept-Encoding", "identity")
	// Make GET request
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	wc := New(int64(size), interval, onProgress)
	go wc.NotifyProgress()
	// pipe stream
	body := io.TeeReader(resp.Body, wc)
	_, err = io.Copy(file, body)

	wc.done <- true
	return err
}

func main() {
	err := parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// read URL from CLI args and validate it
	url := os.Args[len(os.Args)-1]
	if !isCompleteUrl(url) {
		fmt.Printf("Invalid or missing URL: %s\n", url)
		os.Exit(1)
	}

	fmt.Printf("Starting download of: %s\n", url)
	// custom callback provided here with inline func
	err = download(url, func(downloaded, total int64) {
		fmt.Printf("Downloaded %d MB out of %d total MB\n", downloaded/MB, total/MB)
	})
	if err != nil {
		fmt.Printf("Failed with error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done!\nYou can find the downloaded file at %s\n", path)
	os.Exit(0)
}
