package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"newstars/framework/glog"
	"time"
)

// PostJSON post json object by http
func PostJSON(url string, req interface{}, rsp interface{}) error {
	c := http.Client{
		Timeout: 30 * time.Second,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(req)
	glog.SInfof("PostJSON url %v Body %v", url, b.String())

	resp, err := c.Post(url, "application/json", b)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.SInfof("PostJSON url %v Rsp Body %v RetCode:%v", url, string(buf), resp.Status)
	err = json.Unmarshal(buf, &rsp)
	if err != nil {
		return err
	}
	return nil
}

// PostJSONNotRsp ignore rsp
func PostJSONNotRsp(url string, req interface{}) error {
	c := http.Client{
		Timeout: 30 * time.Second,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(req)

	resp, err := c.Post(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.SInfof("PostJSON url %v  Rsp Body %v", url, string(buf))
	if resp.StatusCode != http.StatusOK {
		glog.SWarnf("request url:%s,http status:%d", url, resp.StatusCode)
	}
	return err
}

// PostJSONCheckReturn check return string
func PostJSONCheckReturn(url string, req interface{}, check string) error {
	c := http.Client{
		Timeout: 30 * time.Second,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(req)
	glog.SInfof("PostJSON url %v Body %v", url, b.String())

	resp, err := c.Post(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.SInfof("PostJSON url %v Rsp Body %v", url, string(buf))
	if resp.StatusCode != http.StatusOK {
		glog.SWarnf("request url:%s,http status:%d", url, resp.StatusCode)
		return errors.New("StatusCode failed")
	}
	if string(buf) != check {
		return errors.New("rsp check failed")
	}
	return err
}

// GetJSON get json object by http
func GetJSON(url string, rsp interface{}) error {
	c := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.SInfof("GetJSON url %v. Rsp Body %v", url, string(buf))
	err = json.Unmarshal(buf, &rsp)
	if err != nil {
		return err
	}
	return nil
}

// PostForm key value
func PostForm(url string, data url.Values, rsp interface{}) error {
	c := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := c.PostForm(url, data)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	glog.SInfof("PostJSON url %v.Rsp Body %v", url, string(buf))
	err = json.Unmarshal(buf, &rsp)
	if err != nil {
		return err
	}
	return nil
}
