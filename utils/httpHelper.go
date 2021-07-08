package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func GetWithTime(url string, time time.Duration) ([]byte, error) {
	client := http.Client{
		Timeout: time,
	}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	//res.Body.Read()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
