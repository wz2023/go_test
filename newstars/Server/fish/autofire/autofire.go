package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"newstars/Protocol/plr"
	"newstars/Server/fish/conf"
	"newstars/Server/fish/version"
	"newstars/framework/core/connector"
	"newstars/framework/glog"
	"newstars/framework/util/decimal"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

func main() {

	db, err := sql.Open(conf.Conf.Dbtype, conf.Conf.Dsn)
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}

	glog.SInfof("NewFishServer Go Go Go Version:%v", version.Version)
	db.SetConnMaxLifetime(10 * time.Minute)

	rows, err := db.Query(`select w.userid,accountname,password, machineid,ipaddr from account_t t join userwealth_t w on w.userid=t.userid where acctype=1
		and w.userid between 682500 and 683000 and w.wealth>1000 limit 10`)
	if err != nil {
		fmt.Printf("db querey error,err:%v", err)
		return
	}
	for rows.Next() {
		var machineid, ip, account, password string
		var userid int32
		rows.Scan(&userid, &account, &password, &machineid, &ip)
		autofire(userid, machineid, ip, account, password)
		time.Sleep(1 * time.Second)
	}

	rows.Close()
	fmt.Printf("running .....")
	time.Sleep(3600 * time.Second)
}

func autofire(userid int32, machineid, ip, account, password string) {
	tool := newFireTool()
	tool.setUserInfo(userid, machineid, ip, account, password)
	tool.Go()
	tool.userLogin()
}

type FishInfo struct {
	fish      *plr.P3080001_FishInfo
	deadtimes int64
}

type FireTool struct {
	fishes      map[int32]*FishInfo
	con         *connector.Connector
	bid         int32
	chFunction  chan func()
	quit        chan int
	bullets     map[int32]*plr.N3080003
	tid         int32
	uid         string
	sid         int32
	vx          float64
	vy          float64
	paths       map[int32]int32
	machineid   string
	IPAddr      string
	AccountName string
	password    string
	userid      int32
}

func (p *FireTool) setUserInfo(userid int32, machineid, ip, account, password string) {
	p.userid = userid
	p.machineid = machineid
	p.IPAddr = ip
	p.AccountName = account
	p.password = password
}

func newFireTool() *FireTool {
	con := connector.NewConnector()
	err := con.StartWs("ws://localhost:36530/")
	if err != nil {
		//fmt.Printf("connect failed. err:%v", err)
		return nil
	}

	paths := make(map[int32]int32)
	paths[0] = 31
	paths[1] = 39
	paths[2] = 31
	paths[3] = 38
	paths[4] = 27
	paths[5] = 33
	paths[6] = 27
	paths[7] = 41
	paths[8] = 38
	paths[9] = 19
	paths[10] = 31
	paths[11] = 47
	paths[12] = 34
	paths[13] = 33
	paths[14] = 40
	paths[15] = 46
	paths[16] = 39
	paths[17] = 30
	paths[18] = 20
	paths[19] = 26
	paths[20] = 21
	paths[21] = 21
	paths[22] = 20
	paths[23] = 20
	paths[24] = 20
	paths[25] = 52
	paths[26] = 20
	paths[27] = 20
	paths[28] = 23
	paths[29] = 46
	paths[30] = 26
	paths[31] = 25
	paths[32] = 20
	paths[33] = 21
	paths[34] = 43
	paths[35] = 22
	paths[36] = 23
	paths[37] = 39
	paths[38] = 22
	paths[39] = 26
	paths[40] = 30
	paths[41] = 20
	paths[42] = 19
	paths[43] = 20
	paths[44] = 20
	paths[45] = 20
	paths[46] = 12

	tool := &FireTool{
		con:        con,
		fishes:     make(map[int32]*FishInfo),
		bullets:    make(map[int32]*plr.N3080003),
		chFunction: make(chan func(), 1024),
		paths:      paths,
	}
	tool.con.On("P3080001", tool.onP3080001)
	tool.con.On("P3080006", tool.onP3080006)
	tool.con.On("P3080009", tool.onP3080009)
	tool.con.On("P3080010", tool.onP3080010)
	tool.con.On("P3080002", tool.onP3080002)
	tool.con.On("P3080004", tool.onP3080004)
	tool.con.On("P3080011", tool.onP3080011)

	return tool
}

// Go run
func (p *FireTool) Go() {
	go p.run()

}

// Invoke do in goroutine
func (p *FireTool) Invoke(fn func()) {
	p.chFunction <- fn
}

func pinvoke(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			glog.SErrorf("invoke failed: %v stack:%v", err, stack())
		}
	}()
	fn()
}

func stack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)
	return s
}

func (p *FireTool) run() {
	ticker := time.NewTicker(300 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			p.doTicker()
		case fn := <-p.chFunction:
			pinvoke(fn)
		case <-p.quit:
			return
		}
	}
}

func (p *FireTool) doTicker() {
	// now := time.Now().Unix()
	// if now%30 == 0 {
	// 	fmt.Printf("tick:%v,uid:%v\n", p.userid, now)
	// }
	p.Invoke(p.doFire)
	p.Invoke(p.doshootfish)
}

func (p *FireTool) doshootfish() {
	now := time.Now().Unix() + 8
	bids := make([]int32, 0)
	for _, v := range p.bullets {
		var fishid int32
		for i, v := range p.fishes {
			if v.deadtimes > now {
				fishid = i
				break
			} else if v.deadtimes < now-2 {
				delete(p.fishes, fishid)
			}
		}
		if fishid == 0 {
			continue
		}

		v.FishID = fishid
		p.con.Notify("N3080003", v)
		bids = append(bids, v.BulletID)
		// //fmt.Printf("shot fish:%v\n", v)
		break
	}

	for _, v := range bids {
		delete(p.bullets, v)
		// //fmt.Printf("delte bullets:%v\n", v)
	}

}

func (p *FireTool) userLogin() {
	//fmt.Printf("-----doshoot1--------")
	//msg := &plr.C0000001{
	//	MachineID:   p.machineid,
	//	PlatFormID:  1,
	//	IPAddr:      p.IPAddr,
	//	AccountName: p.AccountName,
	//	Password:    p.password,
	//}
	////fmt.Printf("-----doshoot2--------")
	//s01 := C0000001(msg, p.con)
	////fmt.Printf("-----doshoot3--------")
	//msg01 := &plr.C3080001{
	//	GameKindID: 8,
	//}
	//C3080001(msg01, p.con)
	//
	//msg02 := &plr.C3080002{
	//	RoomID: 32,
	//	UserID: s01.GetUserID(),
	//}
	//s02 := C3080002(msg02, p.con)
	//p.tid = s02.GetTableID()
	//p.uid = s01.GetUserID()
	//p.sid = s02.GetSeatNo()

}

func (p *FireTool) doFire() {

	if len(p.fishes) == 0 {
		return
	}
	p.bid = p.bid + 1

	// vx := int64(rand.Intn(10))
	// vy := int64(rand.Intn(10))

	if p.vx == 0 {
		vx := rand.Intn(1920)
		p.vx, _ = decimal.New(int64(vx), 0).Float64()
		// vy := int64(rand.Intn(10))
	}

	if p.vy == 0 {
		vy := rand.Intn(1080)
		p.vy, _ = decimal.New(int64(vy), 0).Float64()
		// vx := int64(rand.Intn(10))
		// vy := int64(rand.Intn(10))
	}
	// vx = p.vx + vx
	// vy = p.vy + vy

	// if p.sid == 1 {
	// 	vx = -vx
	// } else if p.sid == 2 {
	// 	vx = -vx
	// 	vy = -vy
	// } else if p.sid == 3 {
	// 	vy = -vy
	// }

	// x, _ := decimal.New(int64(vx), 0).Float64()
	// y, _ := decimal.New(int64(vy), 0).Float64()
	n02 := &plr.N3080002{
		BulletID: p.bid,
		UserID:   p.uid,
		TableID:  p.tid,
		SeatNo:   p.sid,
		Ratio:    0.1,
		Speed:    1800,
		VectorX:  p.vx,
		VectorY:  p.vy,
	}
	p.con.Notify("N3080002", n02)

	//fmt.Printf("fire:%v\n", n02)
	n03 := &plr.N3080003{
		BulletID: p.bid,
		UserID:   p.uid,
		TableID:  p.tid,
		SeatNo:   p.sid,
	}
	p.bullets[n03.BulletID] = n03

}

func (p *FireTool) onP3080001(data interface{}) {
	m := &plr.P3080001{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		now := time.Now().Unix()
		for _, f := range m.GetFishes() {
			fsh := &plr.P3080001_FishInfo{
				FishID: f.GetFishID(),
				KindID: f.GetKindID(),
				//KindType: f.GetKindType(),
				//RoomID:   f.GetRoomID(),
				//TableID:  f.GetTableID(),
				Path: f.GetPath(),
			}

			pfish := &FishInfo{}
			pfish.fish = fsh
			pfish.deadtimes = now + int64(p.paths[fsh.Path])
			p.fishes[fsh.FishID] = pfish
			//fmt.Printf("add fish deadtime:%v fish:%v \n", pfish.deadtimes, fsh)

		}
	}
}

func (p *FireTool) onP3080006(data interface{}) {
	m := &plr.P3080006{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		fid := m.GetFishID()
		//fmt.Printf("delte fish:%v\n", fid)
		delete(p.fishes, fid)
	}
}

func (p *FireTool) onP3080009(data interface{}) {
	p.fishes = make(map[int32]*FishInfo)
	// err := proto.Unmarshal(data.([]byte), m)
	// if err == nil {
	// 	// fid := m.GetFishID()
	// 	//fmt.Printf("delte fish:%v\n", fid)
	// 	delete(p.fishes, fid)
	// }
}

func (p *FireTool) onP3080010(data interface{}) {
	m := &plr.P3080010{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		fid := m.GetFishID()
		//fmt.Printf("delte fish:%v\n", fid)
		delete(p.fishes, fid)

		fs := m.GetFishes()
		for _, f := range fs {
			delete(p.fishes, f.GetFishID())
		}
	}
}

func (p *FireTool) onP3080002(data interface{}) {
}

func (p *FireTool) onP3080004(data interface{}) {
}

func (p *FireTool) onP3080011(data interface{}) {
}

// C3080001 房间列表
func C3080001(msg *plr.C3080001, con *connector.Connector) *plr.S3080001 {
	m := &plr.S3080001{}
	ch := make(chan int)
	con.Request("C3080001", msg, func(data interface{}) {
		proto.Unmarshal(data.([]byte), m)
		//fmt.Printf("C3080001 response data:%v\n", m)
		ch <- 0
	})
	<-ch
	return m
}

// C3080002 房间列表
func C3080002(msg *plr.C3080002, con *connector.Connector) *plr.S3080002 {
	m := &plr.S3080002{}
	ch := make(chan int)
	con.Request("C3080002", msg, func(data interface{}) {
		proto.Unmarshal(data.([]byte), m)
		//fmt.Printf("C3080002 response data:%v\n", m)
		ch <- 0
	})
	<-ch
	return m
}

// C0000004 游客登陆
//func C0000004(msg *plr.C0000004, con *connector.Connector) *plr.S0000004 {
//	m := &plr.S0000004{}
//	ch := make(chan int)
//	con.Request("C0000004", msg, func(data interface{}) {
//		proto.Unmarshal(data.([]byte), m)
//		//fmt.Printf("C0000004 response data:%v\n", m)
//		ch <- 0
//	})
//	<-ch
//	return m
//}

// C0000001 登陆请求
//func C0000001(msg *plr.C0000001, con *connector.Connector) *plr.S0000001 {
//	m := &plr.S0000001{}
//	ch := make(chan int)
//	con.Request("C0000001", msg, func(data interface{}) {
//		proto.Unmarshal(data.([]byte), m)
//		//fmt.Printf("C0000001 response data:%v\n", m)
//		ch <- 0
//	})
//	<-ch
//	return m
//}
