package utils

import (
	"DBImageCache/config"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func GetWithTime(url string, time time.Duration) ([]byte, error) {
	client := http.Client{
		Timeout: time,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.UserAgent)
	res, err := client.Do(req)
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
