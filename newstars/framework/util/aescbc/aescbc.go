package aescbc

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
)

// Decrypt lic info
func Decrypt(data []byte) (info AuthInfo, err error) {
	var by []byte
	key := getCiperKey()
	if by, err = decrypt(data, key); err != nil {
		return
	}

	var buf bytes.Buffer
	if _, err = buf.Write(by); err != nil {
		return
	}

	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&info)
	return
}

// Encrypt lic info
func Encrypt(info AuthInfo) (rs []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(info); err != nil {
		return
	}
	key := getCiperKey()
	return encrypt(buf.Bytes(), key)
}

// WriteFile for lic
func WriteFile(data []byte) error {
	f, err := os.Create(licName)
	if err != nil {
		return err
	}
	defer f.Close()

	bufWriter := bufio.NewWriter(f)
	bufWriter.Write(data)
	bufWriter.Flush()
	return nil
}

// ReadFile for lic
func ReadFile() ([]byte, error) {
	return os.ReadFile(licName)
}

func GetAuthInfo() (auth AuthInfo, err error) {
	auth = AuthInfo{}
	if auth.UUID, err = execShell("dmidecode -s system-uuid"); err != nil {
		return
	}

	if auth.UUID == "" {
		err = errors.New("uuid empty")
		return
	}

	if auth.SerialNumber, err = execShell("dmidecode -s system-serial-number"); err != nil {
		return
	}

	if auth.Mac, auth.IP, err = getMacAndIp(); err != nil {
		return
	}
	return
}

// CheckAuth 检查认证信息
func CheckAuth() {
	// orign, err := GetAuthInfo()
	// if err != nil {
	// 	glog.SFatalf("get auth info %v", err)
	// }

	// by, err := ReadFile()
	// if err != nil {
	// 	glog.SFatalf("read file %v", err)
	// }

	// info, err := Decrypt(by)
	// if err != nil {
	// 	glog.SFatalf("decrypt info %v", err)
	// }

	// if info.EndTime < time.Now().Unix() {
	// 	glog.Fatal("auth time out")
	// }

	// if info.UUID != orign.UUID {
	// 	glog.Fatal("auth info uuid no equals")
	// }

	// if info.SerialNumber != orign.SerialNumber {
	// 	glog.Fatal("auth info serial number no equals")
	// }

	// if info.Mac != orign.Mac {
	// 	glog.Fatal("auth info mac no equals")
	// }

	// if info.IP != orign.IP {
	// 	glog.Fatal("auth info ip no equals")
	// }

	// glog.Infof("check auth finish, end time:%s", time.Unix(info.EndTime, 0).Format(time.RFC3339))
}
