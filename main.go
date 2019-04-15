package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mpetavy/common"
)

const version = "1.0.0"

var (
	filename *string
	url      *string
	username *string
	password *string
	method   *string
	header   *string
)

func init() {
	url = flag.String("u", "", "URL to JNLP file")
	filename = flag.String("f", "", "filename")
	username = flag.String("username", "", "username")
	password = flag.String("password", "", "password")
	method = flag.String("method", "GET", "http method")
	header = flag.String("header", "", "header")
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// download loads a remote resource via http(s) and stores it to the given filename
func download(href string) (*bytes.Buffer, error) {
	var b bytes.Buffer

	client := &http.Client{}

	request, err := http.NewRequest(*method, *url, nil)

	if *username != "" {
		request.Header.Add("Authorization", "Basic "+basicAuth(*username, *password))
	}

	if *header != "" {
		headers := strings.Split(*header, ";")

		for _, h := range headers {
			kv := strings.Split(h, "=")

			if len(kv) < 2 {
				return nil, fmt.Errorf("invalid header format: %s", h)
			}

			request.Header.Add(kv[0], kv[1])
		}
	}

	// get a response from the remote source
	response, err := client.Do(request)
	// response, err := http.Get(href)
	if err != nil {
		return &b, err
	}

	// care about final cleanup of reponse body
	defer response.Body.Close()

	// download the remote resource to the file
	_, err = io.Copy(&b, response.Body)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func run() error {
	now := time.Now()

	b, err := download(*url)

	if err != nil {
		return err
	}

	if err == nil {
		elapsed := int64(time.Since(now).Nanoseconds()) / 1000 / 1000
		fmt.Printf("time needed: %d msec\n", elapsed)
		fmt.Printf("bytes written: %d\n", b.Len())

		if *filename != "" {
			err = ioutil.WriteFile(*filename, b.Bytes(), os.ModePerm)
			if err != nil {
				if err != nil {
					return err
				}
			}

			fmt.Printf("bytes written to file: %s\n", *filename)
		} else {
			fmt.Printf("%s\n", string(b.Bytes()))
		}

	}

	return nil
}

func main() {
	defer common.Cleanup()

	common.New(&common.App{"httpc", "1.0.4", "2017", "simple http tool", "mpetavy", common.APACHE, "https://github.com/mpetavy/httpc", false, nil,nil, nil, run, time.Duration(0)}, []string{"u"})
	common.Run()
}
