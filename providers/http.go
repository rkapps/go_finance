package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func runHTTPGet(url string, in interface{}) error {

	// log.Printf("url: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		// log.Print(err
		return err
	}

	// log.Printf("%v", resp.Header.Get("X-MBX-USED-WEIGHT"))

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err1 := json.Unmarshal(body, &in)
	if err1 != nil {
		return err1
	}

	return nil

}

func runHTTPPost(url string, p interface{}, in interface{}) error {

	body, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "applicatoin/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	err1 := json.Unmarshal(result, &in)
	// fmt.Printf("Result: %v \nError: %v\n", in, err1)
	if err1 != nil {
		log.Printf("Unmarshal error: %v", err1)
		return err1
	}
	return nil
}
