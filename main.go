package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mpetavy/common"
)

var (
	filename  *string
	url       *string
	username  *string
	password  *string
	method    *string
	header    *string
	ignoreTls *bool
)

func init() {
	common.Init(false, "1.0.4", "", "2017", "simple http tool", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, run, 0)

	url = flag.String("c", "", "connection URL")
	filename = flag.String("f", "", "filename")
	username = flag.String("u", "", "username")
	password = flag.String("p", "", "password")
	method = flag.String("m", "GET", "http method")
	header = flag.String("h", "", "header")
	ignoreTls = flag.Bool("i", true, "header")
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// download loads a remote resource via http(s) and stores it to the given filename
func download() (*http.Response, *bytes.Buffer, error) {
	var b bytes.Buffer

	if *ignoreTls {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{}

	request, err := http.NewRequest(*method, *url, nil)
	if common.Error(err) {
		return nil, nil, err
	}

	if *username != "" {
		request.Header.Add("Authorization", "Basic "+basicAuth(*username, *password))
	}

	if *header != "" {
		headers := strings.Split(*header, ";")

		for _, h := range headers {
			kv := strings.Split(h, "=")

			if len(kv) < 2 {
				return nil, nil, fmt.Errorf("invalid header format: %s", h)
			}

			request.Header.Add(kv[0], kv[1])
		}
	}

	// get a response from the remote source
	response, err := client.Do(request)
	// response, err := http.Get(href)
	if common.Error(err) {
		return nil, &b, err
	}

	// care about final cleanup of reponse body
	defer func() {
		common.Error(response.Body.Close())
	}()

	// download the remote resource to the file
	_, err = io.Copy(&b, response.Body)
	if common.Error(err) {
		return nil, nil, err
	}

	return response, &b, nil
}

func run() error {
	now := time.Now()

	r, b, err := download()

	if common.Error(err) {
		return err
	}

	elapsed := int64(time.Since(now).Nanoseconds()) / 1000 / 1000
	fmt.Printf("time needed:   %d msec\n", elapsed)
	fmt.Printf("status:        %s\n", r.Status)
	fmt.Printf("bytes written: %d\n", b.Len())
	fmt.Printf("body:\n")

	if *filename != "" {
		err = os.WriteFile(*filename, b.Bytes(), common.DefaultFileMode)
		if common.Error(err) {
			return err
		}

		fmt.Printf("bytes written to file: %s\n", *filename)
	} else {
		fmt.Printf("%s\n", string(b.Bytes()))
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run([]string{"c"})
}
