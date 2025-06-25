package aescbc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"math/big"
	"net"
	"os/exec"
)

const (
	magic   = "87496704531762839457520598147281"
	licName = "lic.dat"
)

func getCiperKey() []byte {
	ret, _ := hex.DecodeString(magic)
	return ret
}

func encrypt(data, key []byte) (rs []byte, err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	size := block.BlockSize()
	data = padding(data, size)
	blockMode := cipher.NewCBCEncrypter(block, key[:size])
	rs = make([]byte, len(data))
	blockMode.CryptBlocks(rs, data)
	return
}

func decrypt(data, key []byte) (rs []byte, err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	size := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:size])
	rs = make([]byte, len(data))
	blockMode.CryptBlocks(rs, data)
	rs = unPadding(rs)
	return
}

func padding(text []byte, size int) []byte {
	padding := size - len(text)%size
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(text, padtext...)
}

func unPadding(origin []byte) []byte {
	l := len(origin)
	unpadding := int(origin[l-1])
	return origin[:l-unpadding]
}

func execShell(command string) (s string, err error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return
	}
	s = out.String()
	return
}

func getMacAndIp() (mac string, ip uint64, err error) {
	var inter []net.Interface
	if inter, err = net.Interfaces(); err != nil {
		return
	}

	for _, v := range inter {
		var addrs []net.Addr
		if addrs, err = v.Addrs(); err != nil {
			continue
		}

		var pip net.IP
		for _, a := range addrs {
			if ip, ok := a.(*net.IPNet); ok {
				if isPublicIP(ip.IP) {
					pip = ip.IP
					break
				}
			}
		}

		if pip == nil {
			continue
		}

		hw := v.HardwareAddr
		if hw == nil || hw.String() == "" {
			continue
		}

		mac = hw.String()
		ip = ipToInt(pip)
		return
	}

	for _, v := range inter {
		var addrs []net.Addr
		if addrs, err = v.Addrs(); err != nil {
			continue
		}

		var okip net.IP
		for _, a := range addrs {
			if ip, ok := a.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				okip = ip.IP
				break
			}
		}

		if okip == nil {
			continue
		}

		hw := v.HardwareAddr
		if hw == nil || hw.String() == "" {
			continue
		}

		mac = hw.String()
		ip = ipToInt(okip)
		return
	}

	return "", 0, errors.New("mac empty")
}

func ipToInt(ip net.IP) uint64 {
	ret := big.NewInt(0)
	ret.SetBytes(ip.To4())
	return ret.Uint64()
}

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}

	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return false
		} else if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return false
		} else if ip4[0] == 192 && ip4[1] == 168 {
			return false
		} else {
			return true
		}
	}
	return false
}
