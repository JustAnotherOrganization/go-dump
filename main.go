// Copyright 2017 Just Another Organization
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

var (
	fileChan = make(chan fileRequest, 200)
)

func main() {
	go handleFiles()
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			rw.WriteHeader(200)
		}()
		fmt.Printf("Request came in from: %v\n", r.Host)
		defer r.Body.Close()

		var (
			fR  fileRequest
			err error
		)
		fR.body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
			return
		}

		fR.header, err = httputil.DumpRequest(r, false)
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
			return
		}

		fR.time, fR.host = time.Now(), r.Host

		fileChan <- fR
	})

	if err := http.ListenAndServe(":9090", nil); err != nil {
		panic(err)
	}

}

func handleFiles() {
	for file := range fileChan {
		if file.host == "" {
			file.host = "unknown"
		}
		err := os.MkdirAll(file.host+"/"+file.time.String(), os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directories: %v\n", err)
			continue
		}
		if err = writeFile(file.host+"/"+file.time.String()+"/body.json",
			file.body); err != nil {
			fmt.Printf("Error writing body: %v\n", err)
			continue
		}
		if err = writeFile(file.host+"/"+file.time.String()+"/headers",
			file.header); err != nil {
			fmt.Printf("Error writing body: %v\n", err)
			continue
		}
	}
}

func writeFile(path string, bts []byte) error {
	if strings.HasPrefix(path, "/") {
		return errors.New("Can not write to root")
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if strings.HasSuffix(path, ".json") {
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, bts, "", "  ")
		if error != nil {
			fmt.Printf("Error Indenting: %v\n", err)
		} else {
			bts = prettyJSON.Bytes()
		}
	}

	_, err = io.Copy(file, bytes.NewReader(bts))
	return err
}

type fileRequest struct {
	body   []byte
	header []byte
	time   time.Time
	host   string
}
