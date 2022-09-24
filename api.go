package runner

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type APIError string

const endpoint = "https://gitlab.com/api/v4"

var Client *http.Client

func Register(data url.Values) (string, error) {

	var res, err = Client.PostForm(endpoint+"/runners", data)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 201 {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return "", APIError(res.Status)
	}

	var register struct {
		Token string
	}

	err = json.NewDecoder(res.Body).Decode(&register)
	res.Body.Close()

	if err != nil {
		return "", err
	}

	return register.Token, nil
}

func Request(input url.Values, output any) (bool, error) {

	var res, err = Client.PostForm(endpoint+"/jobs/request", input)
	if err != nil {
		return false, err
	}

	var is204 = res.StatusCode == 204
	if is204 || res.StatusCode != 201 {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()

		if is204 {
			return false, nil
		}

		return false, APIError(res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(output)
	res.Body.Close()

	if err != nil {
		return false, err
	}

	return true, nil
}

func Update(id string, data any) error {

	var enc, err = json.Marshal(data)
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPut, endpoint+"/jobs/"+id, bytes.NewReader(enc))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	var res *http.Response
	res, err = Client.Do(req)
	if err != nil {
		return err
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	return nil
}

func SendTrace(id, token string, trace io.Reader) (bool, error) {

	var req, err = http.NewRequest(http.MethodPatch, endpoint+"/jobs/"+id+"/trace", trace)
	if err != nil {
		return false, err
	}

	if req.ContentLength <= 0 {
		return false, nil
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("JOB-TOKEN", token)
	req.Header.Set("Content-Range", "0-"+strconv.FormatInt(req.ContentLength, 10))

	var res *http.Response
	res, err = Client.Do(req)
	if err != nil {
		return false, err
	}

	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	return true, nil
}

func (api APIError) Error() string {
	return string(api)
}
