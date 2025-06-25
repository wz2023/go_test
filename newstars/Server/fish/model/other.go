package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"newstars/framework/game_center"
	"newstars/framework/glog"
	"newstars/framework/util"
	"newstars/framework/util/ip17mon"
	"strconv"
	"strings"
	"time"
)

// GameRoom for table
type GameRoom struct {
	Gameroomid     int32
	Gameroomname   string
	Tablescounts   int32
	Minenteramount int32
	Maxenteramount int64
	Gamekindid     int32
	Baseamount     float64
	Gamecommission float64
	GameRoomType   int32
	BRobot         bool
	RobotNumbers   int32
}

// UserInfo 用户信息
type UserInfo struct {
	Userid      string
	Nickname    string
	DisPlayName string
	Faceid      int32
	Sexuality   int32
	Wealth      float64
	Status      int32
	AccType     int32
	GameID      int32
	Profit      float64
	FaceFrameID int32
}

// Gametable 桌台
type Gametable struct {
	Gametableid   int32
	Gametablename string
	Gameroomid    int32
	Minbetamount  int32
	Maxbetamount  int64
	Seatnumbers   int32
	GameID        int32
}

// QueryRoomListByKind list
func QueryRoomListByKind(kindid int32, db *sql.DB) ([]*GameRoom, error) {
	// err := db.Ping()
	// if err != nil {
	// 	return nil, err
	// }

	rows, err := db.Query(`select gameroomid,gameroomname,minenteramount,maxenteramount,baseamount,tablescounts,gamecommission,gameroomtype,robot,robotnumbers
		from gameroom_t where gamekindid = ?`, kindid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rooms := make([]*GameRoom, 0)

	for rows.Next() {
		v := &GameRoom{}
		err = rows.Scan(&v.Gameroomid, &v.Gameroomname, &v.Minenteramount, &v.Maxenteramount, &v.Baseamount, &v.Tablescounts, &v.Gamecommission, &v.GameRoomType, &v.BRobot, &v.RobotNumbers)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, v)
	}
	return rooms, rows.Err()
}

func QueryRoomById(roomid int32, db *sql.DB) (*GameRoom, error) {
	room := &GameRoom{}
	err := db.QueryRow(`select gameroomid,gameroomname,minenteramount,maxenteramount,baseamount,tablescounts,gamecommission,gameroomtype
	from gameroom_t where gameroomid = ?`, roomid).Scan(&room.Gameroomid, &room.Gameroomname, &room.Minenteramount, &room.Maxenteramount, &room.Baseamount, &room.Tablescounts, &room.Gamecommission, &room.GameRoomType)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// QueryUserInfo 查询用户信息
//func QueryUserInfo(uid int32, db *sql.DB) (*UserInfo, error) {
//	// err := db.Ping()
//	// if err != nil {
//	// 	return nil, err
//	// }
//
//	u := &UserInfo{}
//	err := db.QueryRow(`select account_t.userid,account_t.nickname,account_t.status,account_t.faceid,account_t.faceframeid,account_t.iparea,userwealth_t.wealth,account_t.acctype,account_t.sexuatily
//	,game_id,userwealth_t.profit from account_t left join userwealth_t on account_t.userid= userwealth_t.userid
//		where account_t.userid = ?`, uid).Scan(&u.Userid, &u.Nickname, &u.Status, &u.Faceid, &u.FaceFrameID, &u.DisPlayName, &u.Wealth, &u.AccType, &u.Sexuality, &u.GameID, &u.Profit)
//	if u.DisPlayName == "" {
//		index := rand.Intn(len(Area))
//		u.DisPlayName = Area[index]
//	}
//
//	return u, err
//}

// QueryGameTables 查询游戏桌台by rid
func QueryGameTables(rid int32, db *sql.DB) ([]*Gametable, error) {
	// err := db.Ping()
	// if err != nil {
	// 	return nil, err
	// }
	rows, err := db.Query(`select gametableid,gametablename,gameroomid,minbetamount,maxbetamount,seatnumbers,gameid from gametable_t 
		where gameroomid = ?`, rid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ts := make([]*Gametable, 0)
	for rows.Next() {
		v := &Gametable{}
		err = rows.Scan(&v.Gametableid, &v.Gametablename, &v.Gameroomid, &v.Minbetamount, &v.Maxbetamount, &v.Seatnumbers, &v.GameID)
		if err != nil {
			return nil, err
		}
		ts = append(ts, v)
	}
	return ts, rows.Err()
}

// InsertGameTable 插入游戏桌台
func InsertGameTable(t *Gametable, db *sql.DB) error {
	var (
		r   sql.Result
		err error
		id  int64
	)
	// err = db.Ping()
	// if err != nil {
	// 	return err
	// }

	r, err = db.Exec(`insert into gametable_t (gametablename,gameroomid,minbetamount,maxbetamount,seatnumbers,gameid) 
	values (?,?,?,?,?,?)`, t.Gametablename, t.Gameroomid, t.Minbetamount, t.Maxbetamount, t.Seatnumbers, t.GameID)
	if err == nil {
		id, err = r.LastInsertId()
		if err == nil {
			t.Gametableid = int32(id)
		}
	}
	return err
}

// QueryAiInfo get ai
func QueryAiInfo(size int32, db *sql.DB) ([]*UserInfo, error) {
	// err := db.Ping()
	// if err != nil {
	// 	return nil, err
	// }

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(`select account_t.userid,account_t.nickname,account_t.status,account_t.faceid,account_t.iparea,userwealth_t.wealth 
		from account_t left join userwealth_t on account_t.userid= userwealth_t.userid 
		where account_t.acctype = 3 and account_t.status = 1 limit ?`, size)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	uinfs := make([]*UserInfo, 0)
	for rows.Next() {
		u := &UserInfo{}
		err = rows.Scan(&u.Userid, &u.Nickname, &u.Status, &u.Faceid, &u.DisPlayName, &u.Wealth)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		u.Status = 2
		uinfs = append(uinfs, u)
	}
	err = rows.Err()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(uinfs) != int(size) {
		tx.Commit()
		return uinfs, fmt.Errorf("ai not enough")
	}

	for _, v := range uinfs {
		_, err = tx.Exec(`update account_t set status = 2 where userid = ?`, v.Userid)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return uinfs, tx.Commit()
}

// UpdateAiStatus 更新ai状态
func UpdateAiStatus(uid, status int32, db *sql.DB) error {
	_, err := db.Exec(`update account_t set status = ? where userid = ?`, status, uid)
	if err != nil {
		return err
	}
	// _, err = db.Exec(`update userwealth_t set wealth = ? where userid = ?`, rand.Intn(1300)+100, uid)
	return err
}

// QueryRoomLimitAmount 查询房间限额
func QueryRoomLimitAmount(rid int32, db *sql.DB) (float64, error) {
	var min float64
	err := db.QueryRow(`select minenteramount from gameroom_t where gameroomid = ?`, rid).Scan(&min)
	return min, err
}

// QueryRoomComission 查询房间限额
func QueryRoomComission(rid int32, db *sql.DB) (float64, error) {
	var comission float64
	err := db.QueryRow(`select gamecommission from gameroom_t where gameroomid = ?`, rid).Scan(&comission)
	return comission, err
}

// QueryRoomInfo 查询房间信息
func QueryRoomInfo(rid int32, db *sql.DB) (float64, float64, error) {
	var amount, comission float64
	err := db.QueryRow(`select baseamount,gamecommission from gameroom_t where gameroomid = ?`, rid).Scan(&amount, &comission)
	return amount, comission, err
}

// QueryRoomName 查询房间名称
func QueryRoomName(rid int32, db *sql.DB) (string, error) {
	var name string
	err := db.QueryRow(`select gameroomname from gameroom_t where gameroomid = ?`, rid).Scan(&name)
	return name, err
}

// QueryRoomAndroidInfo 房间机器人信息
func QueryRoomAndroidInfo(rid int32, db *sql.DB) (bool, int8, error) {
	var bRobot bool
	var bLevel int8
	err := db.QueryRow(`select robot,robotlevel from gameroom_t where gameroomid = ?`, rid).Scan(&bRobot, &bLevel)
	return bRobot, bLevel, err
}

// SetRoomAndroidInfo 设置房间机器人是否开启
func SetRoomAndroidInfo(rid int32, enable bool, db *sql.DB) error {
	_, err := db.Exec(`update gameroom_t set robot=? where gameroomid=?`, enable, rid)
	return err
}

// QueryAiInfoByLimit get ai
//func QueryAiInfoByLimit(size int, left float64, right float64, kindid int, db *sql.DB) ([]*UserInfo, error) {
//	tx, err := db.Begin()
//	if err != nil {
//		return nil, err
//	}
//	startuid := 4000
//	enduid := 5000
//	if kindid == 1 {
//		startuid = 4000
//		enduid = 4400
//	} else if kindid == 2 {
//		startuid = 4400
//		enduid = 4700
//
//	} else if kindid == 5 {
//		startuid = 4700
//		enduid = 5000
//	}
//	rows, err := tx.Query(`select account_t.userid,account_t.nickname,account_t.status,userwealth_t.wealth
//		from account_t left join userwealth_t on account_t.userid= userwealth_t.userid
//		where account_t.acctype = 3 and account_t.userid > ? and account_t.userid < ? and account_t.status = 1 and userwealth_t.wealth >= ? limit ?`, startuid, enduid, left, size)
//	if err != nil {
//		tx.Rollback()
//		return nil, err
//	}
//	defer rows.Close()
//
//	area := make([]string, len(Area))
//	copy(area, Area)
//
//	uinfs := make([]*UserInfo, 0)
//	for rows.Next() {
//		u := &UserInfo{}
//		err = rows.Scan(&u.Userid, &u.Nickname, &u.Status, &u.Wealth)
//		if err != nil {
//			tx.Rollback()
//			return nil, err
//		}
//		index := rand.Intn(len(area))
//		u.DisPlayName = area[index]
//		u.Faceid = rand.Int31n(30) + 1
//		u.Status = 2
//		uinfs = append(uinfs, u)
//		area = append(area[:index], area[index+1:]...)
//	}
//	err = rows.Err()
//	if err != nil {
//		tx.Rollback()
//		return nil, err
//	}
//
//	if len(uinfs) != int(size) {
//		tx.Rollback()
//		return uinfs, fmt.Errorf("ai not enough")
//	}
//
//	for _, v := range uinfs {
//		_, err = tx.Exec(`UPDATE account_t SET status = 2,faceid=?,iparea=? WHERE userid = ?`, v.Faceid, v.DisPlayName, v.Userid)
//		if err != nil {
//			glog.SErrorf("")
//			tx.Rollback()
//			return nil, err
//		}
//	}
//
//	return uinfs, tx.Commit()
//}

// InitAiStatusByKindid init ai status
func InitAiStatusByKindid(kindid int, db *sql.DB) error {
	startuid := 4000
	enduid := 5000
	if kindid == 1 {
		startuid = 4000
		enduid = 4400
	} else if kindid == 2 {
		startuid = 4400
		enduid = 4700

	} else if kindid == 5 {
		startuid = 4700
		enduid = 5000
	}
	_, err := db.Exec(`update account_t set status=1 where acctype=3 and userid>? and userid<?`, startuid, enduid)
	return err
}

// Agent 代理
type Agent struct {
	Name   string
	QQ     string
	Weixin string
}

// QueryAgentList list agent
func QueryAgentList(db *sql.DB, platformid int32, gameid int32) ([]Agent, error) {
	list := make([]Agent, 0)
	rows, err := db.Query(`SELECT name,q_q,wechat FROM agency_t WHERE work = ? AND is_show = ?  AND platformid=? AND game_id=?`, 1, 1, platformid, gameid)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		item := Agent{}
		err = rows.Scan(&item.Name, &item.QQ, &item.Weixin)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

// UserWealth 财富
type UserWealth struct {
	UserID           int32
	CoinWealth       float64
	BankWealth       float64
	BSetBankPassword int32
}

// QueryUserWealth 查询用户财富
func QueryUserWealth(uid int32, db *sql.DB) (UserWealth, error) {
	uWealth := UserWealth{}
	err := db.QueryRow(`select userwealth_t.userid,userwealth_t.wealth,userbank_t.bankamount,isnull(userbank_t.password) from userwealth_t
		 left join userbank_t on userwealth_t.userid = userbank_t.userid where userwealth_t.userid = ?`, uid).Scan(&uWealth.UserID,
		&uWealth.CoinWealth, &uWealth.BankWealth, &uWealth.BSetBankPassword)
	return uWealth, err
}

// Inventory 库存
type Inventory struct {
	ID         int32
	KindID     int32
	RoomID     int32
	TableID    int32
	PoolAmount float64 // 库存
	UpdateTime int64
	Revenue    float64 // 营收
	Threshold  float64 // 阈值
}

// QueryInventory 查询库存
func QueryInventory(db *sql.DB) ([]*Inventory, error) {
	list := make([]*Inventory, 0)
	rows, err := db.Query(`select id,roomid,tableid,kindid,poolamount,updatetime,revenue,threshold from inventory_t`)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Inventory{}
		err = rows.Scan(&item.ID, &item.RoomID, &item.TableID, &item.KindID, &item.PoolAmount, &item.UpdateTime, &item.Revenue, &item.Threshold)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

// InsertInventory insert 库存数据
func InsertInventory(item *Inventory, db *sql.DB) error {
	result, err := db.Exec(`insert into inventory_t (roomid,tableid,kindid,poolamount,updatetime,revenue) values (?,?,?,?,?,?)`, item.RoomID,
		item.TableID, item.KindID, item.PoolAmount, item.UpdateTime, item.Revenue)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = int32(id)
	return err
}

// UpdateInventory 更新库存
func UpdateInventory(item *Inventory, db *sql.DB) error {
	_, err := db.Exec(`update inventory_t set poolamount = ?,updatetime = ?,revenue = ?, threshold = ? where id = ?`, item.PoolAmount, item.UpdateTime, item.Revenue, item.Threshold, item.ID)
	return err
}

// UpdateInventoryByRid 斗地主更新库存
func UpdateInventoryByRid(rid int32, amount float64, db *sql.DB) error {
	_, err := db.Exec(`update inventory_t set poolamount = ?,updatetime = ? where roomid = ?`, amount, time.Now().Unix(), rid)
	return err
}

// FrozenAccount 冻结账户
func FrozenAccount(uid int32, db *sql.DB) error {
	_, err := db.Exec(`update account_t set status = 3 where userid = ?`, uid)
	return err
}

// UnFrozenAccount 冻结账户
func UnFrozenAccount(uid int32, db *sql.DB) error {
	_, err := db.Exec(`update account_t set status = 1 where userid = ?`, uid)
	return err
}

// GameKindTable 游戏种类
type GameKindTable struct {
	KindID   int32
	KindName string
	Status   int32
	IconType int32
}

// QueryGameKind 查询游戏种类
func QueryGameKind(db *sql.DB) ([]*GameKindTable, error) {
	list := make([]*GameKindTable, 0)
	rows, err := db.Query(`select gamekindid,gamekindname,status,icontype from gamekind_t`)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &GameKindTable{}
		err = rows.Scan(&item.KindID, &item.KindName, &item.Status, &item.IconType)
		if err != nil {
			return list, err
		}
		list = append(list, item)
	}

	return list, err
}

// QueryPlatformTableMap query gameid map to table 游戏分平台
func QueryPlatformTableMap(db *sql.DB, gamekindid int) (map[string]int32, error) {
	var gameidmap sql.NullString
	idmap := make(map[string]int32)
	err := db.QueryRow(`select platformtablemap from gamekind_t where gamekindid=?`, gamekindid).Scan(&gameidmap)
	if err != nil {
		return nil, err
	}
	if gameidmap.String == "" {
		return idmap, nil
	}
	err = json.Unmarshal([]byte(gameidmap.String), &idmap)
	if err != nil {
		return nil, err
	}
	return idmap, nil
}

// RechargeUserCoin 充值
func RechargeUserCoin(uid string, amount float64, orderid string, db *sql.DB) (float64, error) {
	tx, err := db.Begin()
	if err != nil {
		glog.SErrorf("RechargeUserCoin failed error:%v", err)
		return 0, err
	}
	settlewid, err := InsertRecordAmount(uid, RecordFastRecharge, orderid, tx)
	if err != nil {
		tx.Rollback()
		glog.SErrorf("insert recordAmount failed for db.%v", err.Error())
		return 0, err
	}

	_, err = tx.Exec(`update userwealth_t set wealth = wealth + ? where userid = ?`, amount, uid)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = UpdateRecordAmount(settlewid, uid, tx)
	if err != nil {
		tx.Rollback()
		glog.SErrorf("update recordAmount failed for db.%v", err.Error())
		return 0, err
	}

	var coin float64
	err = tx.QueryRow(`select wealth from userwealth_t where userid = ?`, uid).Scan(&coin)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()
	return coin, nil
}

// CheckAppleReceipt check receipt
func CheckAppleReceipt(receipt string, db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`select count(*) from apple_receipt_t where receipt = ?`, receipt).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, errors.New("receipt is exsit")
	}
	return true, nil
}

// InsertAppleReceipt insert into receipt
func InsertAppleReceipt(receipt string, db *sql.DB) error {
	_, err := db.Exec(`insert into apple_receipt_t  (receipt) value (?)`, receipt)
	return err
}

const (
	RecordRegister           = iota + 1000 //账号注册
	RecordBindMobile                       //绑定手机
	RecordToBank                           //存入保险箱
	RecordFromBank                         //从保险箱取出
	RecordRechargeApply                    //兑换申请
	RecRechargeFail                        //退回玩家
	RecordPayConfirm                       //代理充值
	RecordRefoundpayApply                  //撤单申请
	RecordRefoundpayFail                   //撤单失败
	RecordFastRecharge                     //快捷充值
	RecordPurchaseCannon                   //购买炮台
	RecordUseFishSkill                     //使用技能
	RecordSuperSetUserWealth               //后台-修改钱包
	RecordSuperSetUserBank                 //后台-修改保险箱
	RecordMailCoin                         //邮件领取金币
	RecordWagencyAmount                    //全民代理奖励金币
)

// InsertRegisterRecordAmount insert into record_amount when user register
func InsertRegisterRecordAmount(userid, itype int32, tx *sql.Tx) error {
	var endAmount, bankAmount float64
	if userid >= 682500 {
		err := tx.QueryRow(`select wealth from userwealth_t where userid = ?`, userid).Scan(&endAmount)
		if err != nil {
			return err
		}

		err = tx.QueryRow(`select bankamount from userbank_t where userid = ?`, userid).Scan(&bankAmount)
		if err != nil {
			return err
		}

		sql := fmt.Sprintf(`insert into record_amount_%d_t (userid,type,roundcode,startamount,
			endamount,alteramount,bankamount,starttime,endtime) values(?,?,?,?,?,?,?,?,?)`, userid%10)
		_, err = tx.Exec(sql, userid, itype, "", 0, endAmount, endAmount, bankAmount, time.Now().Unix(), time.Now().Unix())
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertRecordAmount insert into record_amount
func InsertRecordAmount(userid string, itype int32, roundcode string, tx *sql.Tx) (int64, error) {
	var startAmount float64
	var lastID int64
	//if userid >= 682500 {
	realType := itype
	if itype < 1000 {
		if roundcode == "" {
			return 0, fmt.Errorf("empty roundcode")
		}
		arr := strings.Split(roundcode, ":")
		//捕鱼局号多了_uid
		if len(arr) != 3 && len(arr) != 4 {
			return 0, fmt.Errorf("invalid roundcode")
		}

		ptype, err := strconv.ParseInt(arr[1], 10, 0)
		if err != nil {
			return 0, err
		}
		realType = int32(ptype)
	}

	//err := tx.QueryRow(`select wealth from userwealth_t where userid = ?`, userid).Scan(&startAmount)
	//if err != nil {
	//	return 0, err
	//}

	uinfo, err := game_center.GetUserInfoByID(userid)
	if err != nil {
		return 0, err
	}
	startAmount = uinfo.Wealth

	sql := fmt.Sprintf(`insert into record_amount_%d_t (userid,type,roundcode,startamount,starttime) values(?,?,?,?,?)`, util.StringToIntHash(userid)%10)
	rs, err := tx.Exec(sql, userid, realType, roundcode, startAmount, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	lastID, err = rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	//}
	return lastID, nil
}

// UpdateRecordAmount update record_amount
func UpdateRecordAmount(id int64, userid string, tx *sql.Tx) error {
	var endAmount, bankAmount float64
	//if userid >= 682500 && id > 0 {
	if id > 0 {
		uinfo, err := game_center.GetUserInfoByID(userid)
		if err != nil {
			return err
		}
		endAmount = uinfo.Wealth

		//err := tx.QueryRow(`select wealth from userwealth_t where userid = ?`, userid).Scan(&endAmount)
		//if err != nil {
		//	return err
		//}
		//err = tx.QueryRow(`select bankamount from userbank_t where userid = ?`, userid).Scan(&bankAmount)
		//if err != nil {
		//	return err
		//}
		sql := fmt.Sprintf(`update record_amount_%d_t set endamount = ?, alteramount=?-startamount,bankamount=? 
			,endtime=? where id = ?`, util.StringToIntHash(userid)%10)
		_, err = tx.Exec(sql, endAmount, endAmount, bankAmount, time.Now().Unix(), id)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateRecordAmountStarttime update record_amount
func UpdateRecordAmountStarttime(id int64, userid int32, startTime int64, tx *sql.Tx) error {
	if userid >= 682500 && id > 0 {
		sql := fmt.Sprintf(`update record_amount_%d_t set starttime=? where id = ?`, userid%10)
		_, err := tx.Exec(sql, startTime, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeteteRecordAmount delete record_amount
func DeteteRecordAmount(id int64, userid string, tx *sql.DB) error {
	//if userid >= 682500 && id > 0 {
	if id > 0 {
		sql := fmt.Sprintf(`delete from record_amount_%d_t  where id = ?`, util.StringToIntHash(userid)%10)
		_, err := tx.Exec(sql, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddVipValue 新增玩家充值金额 vip等级用
func AddVipValue(uid int32, amount float64, tx *sql.Tx) error {

	var count int
	err := tx.QueryRow(`select count(1) from vip_level_t where userid = ?`, uid).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = tx.Exec(`insert into vip_level_t (userid,vipvalue,svipvalue,vvipvalue,viplevel) values(?,?,?,?,?)`, uid, 0, 0, 0, 0)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`update vip_level_t set vipvalue = vipvalue + ? where userid = ?`, amount, uid)
	if err != nil {
		return err
	}
	return nil
}

// ResetVipValue 重置玩家充值金额 vip等级用
func ResetVipValue(uid int32, db *sql.DB) error {
	if !IsRealPlayer(uid) {
		var count int
		err := db.QueryRow(`select count(1) from vip_level_t where userid = ?`, uid).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = db.Exec(`insert into vip_level_t (userid,vipvalue,svipvalue,vvipvalue,viplevel) values(?,?,?,?,?)`, uid, 0, 0, 0, 0)
			if err != nil {
				return err
			}
		}
		_, err = db.Exec(`update vip_level_t set vipvalue =0,viplevel=0, vvipvalue=0 where userid = ?`, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddSVipValue 新增玩家税收金额 vip等级用
func AddSVipValue(uid int32, amount float64, tx *sql.Tx) error {
	if uid >= 682500 {
		var count int
		err := tx.QueryRow(`select count(1) from vip_level_t where userid = ?`, uid).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err = tx.Exec(`insert into vip_level_t (userid,vipvalue,svipvalue,vvipvalue,viplevel) values(?,?,?,?,?)`, uid, 0, 0, 0, 0)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(`update vip_level_t set svipvalue = svipvalue + ? where userid = ?`, amount, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddVVipValue 新增玩家税收金额 vip等级用
func AddVVipValue(uid, size int32, tx *sql.Tx) error {
	var count int
	err := tx.QueryRow(`select count(1) from vip_level_t where userid = ?`, uid).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = tx.Exec(`insert into vip_level_t (userid,vipvalue,svipvalue,vvipvalue,viplevel) values(?,?,?,?,?)`, uid, 0, 0, 0, 0)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`update vip_level_t set vvipvalue = vvipvalue + ? where userid = ?`, size, uid)
	if err != nil {
		return err
	}
	return nil
}

// ComputeVipLevel 计算vip等级
func ComputeVipLevel(uid int32, tx *sql.Tx) (int32, int32, error) {
	var vipvalue float64
	var viplevel, newlevel int32
	var err error
	// if uid >= 682500 {
	err = tx.QueryRow(`select vipvalue,viplevel from vip_level_t where userid = ?`, uid).Scan(&vipvalue, &viplevel)
	if err != nil {
		return viplevel, newlevel, err
	}

	level := make([]bool, 22)
	level[0] = true
	var v1 int32
	switch {
	case vipvalue >= 0 && vipvalue < 30:
		v1 = 0
	case vipvalue >= 30 && vipvalue < 500:
		v1 = 1
	case vipvalue >= 500 && vipvalue < 3000:
		v1 = 2
	case vipvalue >= 3000 && vipvalue < 10000:
		v1 = 3
	case vipvalue >= 10000 && vipvalue < 30000:
		v1 = 4
	case vipvalue >= 30000 && vipvalue < 70000:
		v1 = 5
	case vipvalue >= 70000:
		v1 = 6
	}

	newlevel = v1
	if newlevel != viplevel {
		_, err = tx.Exec(`update vip_level_t set viplevel = ? where userid = ?`, newlevel, uid)
		if err != nil {
			return viplevel, newlevel, err
		}
	}

	// }
	return viplevel, newlevel, err
}

// QueryVipLevel 查询vip等级
func QueryVipLevel(uid int32, db *sql.DB) (int32, float64, error) {
	var count int
	err := db.QueryRow(`select count(1) from vip_level_t where userid = ?`, uid).Scan(&count)
	if err != nil {
		return 0, 0, err
	}
	if count == 0 {
		_, err = db.Exec(`insert into vip_level_t (userid,vipvalue,svipvalue,vvipvalue,viplevel) values(?,?,?,?,?)`, uid, 0, 0, 0, 0)
		if err != nil {
			return 0, 0, err
		}
	}

	var viplevel int32
	var vipvalue float64
	err = db.QueryRow(`select viplevel,vipvalue from vip_level_t where userid = ?`, uid).Scan(&viplevel, &vipvalue)
	if err != nil {
		return 0, 0, err
	}
	return viplevel, vipvalue, nil
}

// PayConfig 充值配置
type PayConfig struct {
	Pay string
	Vip int32
}

// QueryPayConfigByIP 查询充值配置
func QueryPayConfigByIP(ip string, db *sql.DB) (*PayConfig, error) {
	cfg := &PayConfig{}
	if ip == "" {
		return cfg, nil
	}
	loc, err := ip17mon.Find(ip)
	if err != nil {
		return nil, err
	}
	if loc != nil {
		err = db.QueryRow(`select pay,vip from money_config_city_t
			 where city=?`, loc.City).Scan(&cfg.Pay, &cfg.Vip)
		if err == nil {
			return cfg, nil
		}

		err = db.QueryRow(`select pay,vip from money_config_province_t
				 where province=?`, loc.Region).Scan(&cfg.Pay, &cfg.Vip)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	return cfg, nil
}

// QueryExchangeBYIP 查询兑换配置
func QueryExchangeBYIP(ip string, db *sql.DB) (string, error) {
	exchange := ""
	if ip == "" {
		return exchange, nil
	}
	loc, err := ip17mon.Find(ip)
	if err != nil {
		return exchange, err
	}
	if loc != nil {
		err = db.QueryRow(`select exchange from money_config_city_t
			 where city=?`, loc.City).Scan(&exchange)
		if err == nil {
			return exchange, nil
		}
		err = db.QueryRow(`select exchange from money_config_province_t
				 where province=?`, loc.Region).Scan(&exchange)

		if err != nil {
			return "", err
		}
		return exchange, nil
	}

	return "", nil
}

// IsRealPlayer 是否真实玩家
func IsRealPlayer(uid int32) bool {
	return uid >= 682500
}

// AddLuckyProfit 更新玩家排行榜数据
func AddLuckyProfit(uid int32, amount float64, gamekindid int, gids []int32, tx *sql.Tx) error {
	var count int64
	var rs sql.Result
	var err error
	gameids := strings.Replace(strings.Trim(fmt.Sprint(gids), "[]"), " ", ",", -1)
	now := time.Now().Unix()
	if amount > 0 {
		rs, err = tx.Exec(`update lucky_list_t set profit=profit+?,luckyprofit=if(luckyprofit>?,luckyprofit,?),
		gamekindid=if(luckyprofit>?,gamekindid,?),luckytime=if(luckyprofit>?,luckytime,?),gameids=? where userid=?`, amount, amount, amount, amount, gamekindid, amount, now, gameids, uid)
	} else if amount < 0 {
		rs, err = tx.Exec(`update lucky_list_t set profit=profit+?,gameids=? where userid=?`, amount, gameids, uid)
	} else {
		return nil
	}
	if err != nil {
		return err
	}
	count, err = rs.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = tx.Exec(`insert into lucky_list_t(userid,profit,luckyprofit,gamekindid,luckytime,gameids) values(?,?,?,?,?,?)`, uid, amount, amount, gamekindid, now, gameids)
		if err != nil {
			return err
		}
	}
	return nil
}

type VIPConfig struct {
	VipLvl       int32
	VipValue     float64
	FaceFrameids []int32
}

var VIPConf struct {
	VIPConf []*VIPConfig
}

func QueryVipConfig(db *sql.DB) ([]*VIPConfig, error) {
	if VIPConf.VIPConf == nil {
		var cfg sql.NullString
		err := db.QueryRow(`select vipconfig from global_config_t`).Scan(&cfg)
		if err != nil {
			return nil, err
		}

		if cfg.String == "" {
			VIPConf.VIPConf = make([]*VIPConfig, 0)
		} else {
			err = json.Unmarshal([]byte(cfg.String), &VIPConf)
			if err != nil {
				return nil, err
			}
		}
	}
	return VIPConf.VIPConf, nil
}

// RandomFrameID calc robot faceframeid by viplevel
func RandomFrameID(viplevel int32, db *sql.DB) int32 {
	var fid int32
	vipconf, err := QueryVipConfig(db)
	if err == nil {
		if vipconf != nil {
			for _, v := range vipconf {
				if v.VipLvl == viplevel {
					ids := v.FaceFrameids
					if ids != nil && len(ids) > 0 {
						idx := rand.Intn(len(ids))
						fid = ids[idx]
					}
				}
			}
		}
	} else {
		glog.SErrorf("random frameid failed.err:%v", err)
	}
	return fid
}

// UpdateAndroidVipLevel update android viplevel
func UpdateAndroidVipLevel(uid int32, updatevalue float64, db *sql.DB) error {
	if !IsRealPlayer(uid) {
		if updatevalue > 0 {
			tx, err := db.Begin()
			if err != nil {
				glog.SErrorf("Begin transform failed :%v.", err)
				return err
			}
			err = AddVipValue(uid, updatevalue, tx)
			if err != nil {
				tx.Rollback()
				glog.SErrorf("AddVipValue failed :%v.", err)
				return err
			}
			err = AddVVipValue(uid, 1, tx)
			if err != nil {
				tx.Rollback()
				glog.SErrorf("AddVVipValue failed :%v.", err)
				return err
			}
			_, _, err = ComputeVipLevel(uid, tx)
			if err != nil {
				tx.Rollback()
				glog.SErrorf("ComputeVipLevel failed :%v.", err)
				return err
			}
			tx.Commit()
		}
	}
	return nil
}

// QueryInventoryConfig QueryInventoryConfig
func QueryInventoryConfig(db *sql.DB) ([]*InventoryConfig, error) {
	rows, err := db.Query(`select game_kind_id,conotrol_type,low_threshold,high_threshold,percent 
	from inventory_config_t`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	inves := make([]*InventoryConfig, 0)

	for rows.Next() {
		v := &InventoryConfig{}
		err = rows.Scan(&v.GameKindID, &v.ControlType, &v.LowThreshold, &v.HighThreshold, &v.Percent)
		if err != nil {
			return nil, err
		}
		inves = append(inves, v)
	}
	return inves, rows.Err()
}
