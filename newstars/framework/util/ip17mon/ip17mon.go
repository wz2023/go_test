package ip17mon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
)

// Null for n/a
const Null = "N/A"

// ErrInvalidIp er
var (
	ErrInvalidIP = errors.New("invalid ip format")
	std          *Locator
)

// Init defaut locator with dataFile
func Init(dataFile string) (err error) {
	if std != nil {
		return
	}
	std, err = NewLocator(dataFile)
	return
}

// InitWithData Init defaut locator with data
func InitWithData(data []byte) {
	if std != nil {
		return
	}
	std = NewLocatorWithData(data)
	return
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func Find(ipstr string) (*LocationInfo, error) {
	return std.Find(ipstr)
}

// FindByUint Find locationInfo by uint32
func FindByUint(ip uint32) *LocationInfo {
	return std.FindByUint(ip)
}

//-----------------------------------------------------------------------------

// NewLocator New locator with dataFile
func NewLocator(dataFile string) (loc *Locator, err error) {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return
	}
	loc = NewLocatorWithData(data)
	return
}

// NewLocatorWithData New locator with data
func NewLocatorWithData(data []byte) (loc *Locator) {
	loc = new(Locator)
	loc.init(data)
	return
}

// Locator Locator
type Locator struct {
	textData   []byte
	indexData1 []uint32
	indexData2 []int
	indexData3 []int
	index      []int
}

// LocationInfo info
type LocationInfo struct {
	Country string
	Region  string
	City    string
	Isp     string
}

// var exceptName = []string{
// 	"广东广州",
// 	"北京",
// 	"上海",
// 	"四川成都",
// 	"广东深圳",
// }

func (locinfo *LocationInfo) String() string {
	if locinfo.Country == "中国" {
		if locinfo.City == Null {
			return fmt.Sprintf("%v", locinfo.Region)
		}
		return fmt.Sprintf("%v", locinfo.City)
	}
	return "未知"
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func (loc *Locator) Find(ipstr string) (info *LocationInfo, err error) {
	ip := net.ParseIP(ipstr).To4()
	if ip == nil || ip.To4() == nil {
		err = ErrInvalidIP
		return
	}
	info = loc.FindByUint(binary.BigEndian.Uint32([]byte(ip)))
	return
}

// FindByUint Find locationInfo by uint32
func (loc *Locator) FindByUint(ip uint32) (info *LocationInfo) {
	end := len(loc.indexData1) - 1
	if ip>>24 != 0xff {
		end = loc.index[(ip>>24)+1]
	}
	idx := loc.findIndexOffset(ip, loc.index[ip>>24], end)
	off := loc.indexData2[idx]
	return newLocationInfo(loc.textData[off : off+loc.indexData3[idx]])
}

// binary search
func (loc *Locator) findIndexOffset(ip uint32, start, end int) int {
	for start < end {
		mid := (start + end) / 2
		if ip > loc.indexData1[mid] {
			start = mid + 1
		} else {
			end = mid
		}
	}

	if loc.indexData1[end] >= ip {
		return end
	}

	return start
}

func (loc *Locator) init(data []byte) {
	textoff := int(binary.BigEndian.Uint32(data[:4]))

	loc.textData = data[textoff-1024:]

	loc.index = make([]int, 256)
	for i := 0; i < 256; i++ {
		off := 4 + i*4
		loc.index[i] = int(binary.LittleEndian.Uint32(data[off : off+4]))
	}

	nidx := (textoff - 4 - 1024 - 1024) / 8

	loc.indexData1 = make([]uint32, nidx)
	loc.indexData2 = make([]int, nidx)
	loc.indexData3 = make([]int, nidx)

	for i := 0; i < nidx; i++ {
		off := 4 + 1024 + i*8
		loc.indexData1[i] = binary.BigEndian.Uint32(data[off : off+4])
		loc.indexData2[i] = int(uint32(data[off+4]) | uint32(data[off+5])<<8 | uint32(data[off+6])<<16)
		loc.indexData3[i] = int(data[off+7])
	}
	return
}

func newLocationInfo(str []byte) *LocationInfo {

	var info *LocationInfo

	fields := bytes.Split(str, []byte("\t"))
	switch len(fields) {
	case 4:
		// free version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
		}
	case 5:
		// pay version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
			Isp:     string(fields[4]),
		}
	default:
		panic("unexpected ip info:" + string(str))
	}

	if len(info.Country) == 0 {
		info.Country = Null
	}
	if len(info.Region) == 0 {
		info.Region = Null
	}
	if len(info.City) == 0 {
		info.City = Null
	}
	if len(info.Isp) == 0 {
		info.Isp = Null
	}
	return info
}
