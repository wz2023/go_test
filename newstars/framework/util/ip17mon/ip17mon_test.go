package ip17mon

import (
	"math/rand"
	"net"
	"testing"
	"time"
)

const data = "17monipdb.dat"

func TestFind(t *testing.T) {
	if err := Init(data); err != nil {
		t.Fatal("Init failed:", err)
	}
	info, err := Find("58.82.239.132")

	if err != nil {
		t.Fatal("Find failed:", err)
	}

	if info.Country != "中国" {
		t.Fatal("country expect = 中国, but actual =", info.Country)
	}

	if info.Region != "广东" {
		t.Fatal("region expect = 广东, but actual =", info.Region)
	}

	if info.City != "深圳" {
		t.Fatal("city expect = 深圳, but actual =", info.City)
	}

	if info.Isp != Null {
		t.Fatal("isp expect = Null, but actual =", info.Isp)
	}
}

func TestIP(t *testing.T) {
	test, err := net.ResolveUDPAddr("tcp", "192.168.0.24:1234")
	if err != nil {
		t.Fatal(test)
	}
	t.Fatal(test.IP.String())
}

//-----------------------------------------------------------------------------
func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	if err := Init(data); err != nil {
		b.Fatal("Init failed:", err)
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		if FindByUint(rand.Uint32()) == nil {
			b.Fatal("FindByUint found nil val")
		}
	}
}
