package model

import (
	"database/sql"
	"fmt"
	"newstars/Server/fish/consts"
	"newstars/framework/glog"
)

type FishKind struct {
	ID       int32
	KindType int32
	KindName string
	Score    float64
	KindDesc string
	Paths    string
	Interval int32
	BossRoom int32
}

type FishPath struct {
	ID       int32
	Deadtime int64
	Rate     int32
}

type FishRecord struct {
	FishID    int32
	KindID    int32
	KindType  int32
	KindName  string
	Score     float64
	KindDesc  string
	Path      int32
	Interval  int32
	StartTime int64
	SpawnTime int64
	BBoss     bool
	IsPushed  bool //是否已经推送客户端，只在鱼潮时有效
}

type FishUserExtend struct {
	ID              int32
	UserID          string
	CurrCannonID    int32
	CurrCannonRatio float64
	Onlinetimes     int64
	Inven           float64
	HisRevenue      float64
	IsNewplayer     int32
}

type BulletInfo struct {
	BulletID  int32
	TableID   int32
	SeatNo    int32
	VectorX   float64
	VectorY   float64
	Speed     int32
	StartTime int64
	UserID    string
	Wealth    float64
}

type FishSkill struct {
	ID         int32
	SkillType  int32
	SkillName  string
	FreezeTime int32
	GameRoomID int32
	CostWealth float64
}

type SkillRecord struct {
	SkillType  int32
	FreezeTime int32
	UserID     string
	TableID    int32
	SeatNo     int32
	StartTime  int64
}

type Capturerate struct {
	RoomID int32
	Rate   float64
}

func QueryCommonFishKinds(db *sql.DB) ([]*FishKind, error) {
	rows, err := db.Query(`select id,kindtype,kindname,score,kinddesc,paths,intervalsec,bossroom from fish_kind_t where bossroom=0`)
	if err != nil {
		glog.SErrorf(`query fishkind failed,err:%v`, err)
		return nil, err
	}
	kinds := make([]*FishKind, 0)
	defer rows.Close()
	for rows.Next() {
		kind := &FishKind{}
		rows.Scan(&kind.ID, &kind.KindType, &kind.KindName, &kind.Score, &kind.KindDesc, &kind.Paths, &kind.Interval, &kind.BossRoom)
		kinds = append(kinds, kind)
	}
	return kinds, nil
}

func QueryAllFishKinds(db *sql.DB) ([]*FishKind, error) {
	rows, err := db.Query(`select id,kindtype,kindname,score,kinddesc,paths,intervalsec,bossroom from fish_kind_t`)
	if err != nil {
		glog.SErrorf(`query fishkind failed,err:%v`, err)
		return nil, err
	}
	kinds := make([]*FishKind, 0)
	defer rows.Close()
	for rows.Next() {
		kind := &FishKind{}
		rows.Scan(&kind.ID, &kind.KindType, &kind.KindName, &kind.Score, &kind.KindDesc, &kind.Paths, &kind.Interval, &kind.BossRoom)
		kinds = append(kinds, kind)
	}
	return kinds, nil
}

func QueryAllGameID(db *sql.DB) ([]int, error) {
	rows, err := db.Query(`select game_id from address_t`)
	if err != nil {
		glog.SErrorf(`query gameid failed,err:%v`, err)
		return nil, err
	}
	gameids := make([]int, 0)
	defer rows.Close()
	for rows.Next() {
		var gid int
		err = rows.Scan(&gid)
		if err != nil {
			glog.SErrorf(`scan gameid failed,err:%v`, err)
			continue
		}
		gameids = append(gameids, gid)
	}
	return gameids, nil
}

func QueryBossFishKinds(db *sql.DB) ([]*FishKind, error) {
	rows, err := db.Query(`select id,kindtype,kindname,score,kinddesc,paths,intervalsec,bossroom from fish_kind_t where bossroom!=0`)
	if err != nil {
		glog.SErrorf(`query fishkind failed,err:%v`, err)
		return nil, err
	}
	kinds := make([]*FishKind, 0)
	defer rows.Close()
	for rows.Next() {
		kind := &FishKind{}
		rows.Scan(&kind.ID, &kind.KindType, &kind.KindName, &kind.Score, &kind.KindDesc, &kind.Paths, &kind.Interval, &kind.BossRoom)
		kinds = append(kinds, kind)
	}
	return kinds, nil
}

func UpdateUserCannonRate(db *sql.DB, ratio float64, uid string) error {
	_, err := db.Exec(`update fish_user_extend_t set curr_cannon_ratio=? where userid=?`, ratio, uid)
	if err != nil {
		glog.SErrorf("update curr_cannon_ratio failed err:%v", err)
		return err
	}
	return nil
}

func UpdateUserCannonId(db *sql.DB, cannonid int32, uid string) error {
	_, err := db.Exec(`update fish_user_extend_t set curr_cannon_id=? where userid=?`, cannonid, uid)
	if err != nil {
		glog.SErrorf("update curr_cannon_id failed err:%v", err)
		return err
	}
	return nil
}

func QueryFishPath(db *sql.DB) ([]*FishPath, error) {
	rows, err := db.Query(`select id,deadtime,rate from fish_path_t`)
	if err != nil {
		glog.SErrorf(`query fishpath failed,err:%v`, err)
		return nil, err
	}
	paths := make([]*FishPath, 0)
	defer rows.Close()
	for rows.Next() {
		path := &FishPath{}
		rows.Scan(&path.ID, &path.Deadtime, &path.Rate)
		paths = append(paths, path)
	}
	return paths, nil
}

func QueryUserExtend(uid string, db *sql.DB) (*FishUserExtend, error) {
	user := &FishUserExtend{}
	err := db.QueryRow(`select id,userid,curr_cannon_id,curr_cannon_ratio,onlinetimes,inven,hisrevenue,isnewplayer from fish_user_extend_t where 
		userid=?`, uid).Scan(&user.ID, &user.UserID, &user.CurrCannonID, &user.CurrCannonRatio, &user.Onlinetimes, &user.Inven, &user.HisRevenue, &user.IsNewplayer)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CheckUserExtendExists(uid string, db *sql.DB) (bool, error) {
	var count int32
	err := db.QueryRow(`select count(1) from fish_user_extend_t where  userid=?`, uid).Scan(&count)
	if err != nil {
		glog.SErrorf(`query fish_user_extend_t failed,err:%v`, err)
		return false, err
	}
	return count > 0, nil
}

func QueryUserCannon(uid string, db *sql.DB) (*FishUserExtend, error) {
	user := &FishUserExtend{}
	err := db.QueryRow(`select id,userid,curr_cannon_id,curr_cannon_ratio from fish_user_extend_t where 
		userid=?`, uid).Scan(&user.ID, &user.UserID, &user.CurrCannonID, &user.CurrCannonRatio)
	if err != nil {
		glog.SErrorf(`query fish_user_extend_t failed,err:%v`, err)
		return nil, err
	}
	return user, nil
}

func InitUserExtend(uid string, db *sql.DB) error {
	_, err := db.Exec(`insert into fish_user_extend_t(userid,curr_cannon_id,curr_cannon_ratio,isnewplayer) values(?,?,?,?) 
		`, uid, consts.DefaultConnonID, consts.DefaultConnonRatio, 1)
	if err != nil {
		glog.SErrorf(`init fish_user_extend_t failed,err:%v`, err)
		return err
	}
	return nil
}

func QueryFishSkill(db *sql.DB) ([]*FishSkill, error) {
	rows, err := db.Query(`select id,skill_type,skill_name,freeze_time,gameroomid,cost_wealth from fish_skill_t`)
	if err != nil {
		glog.SErrorf(`query fish_skill_t failed,err:%v`, err)
		return nil, err
	}
	skills := make([]*FishSkill, 0)
	defer rows.Close()
	for rows.Next() {
		skill := &FishSkill{}
		rows.Scan(&skill.ID, &skill.SkillType, &skill.SkillName, &skill.FreezeTime, &skill.GameRoomID, &skill.CostWealth)
		skills = append(skills, skill)
	}
	return skills, nil
}

// QueryRoomCaptureRate 查询房间捕获概率
func QueryRoomCaptureRate(db *sql.DB, roomid int32) (float64, float64, error) {
	var capturerate, poolamount float64
	err := db.QueryRow(`select capturerate,poolamount from fish_room_config where gameroomid=?`, roomid).Scan(&capturerate, &poolamount)
	if err != nil {
		glog.SErrorf(`query fish_room_config failed,roomid:%v,err:%v`, roomid, err)
		return 0, 0, err
	}
	return capturerate, poolamount, nil
}

// QueryExcludeFishkinds 查询房间排除的鱼的种类
func QueryExcludeFishkinds(db *sql.DB, roomid int32) (string, error) {
	var excludefishkinds string
	err := db.QueryRow(`select excludefishkinds from fish_room_config where gameroomid=?`, roomid).Scan(&excludefishkinds)
	if err != nil {
		glog.SErrorf(`query fish_room_config failed,roomid:%v,err:%v`, roomid, err)
		return excludefishkinds, err
	}
	return excludefishkinds, nil
}

// QueryAllRoomCaptureRate 查询房间捕获概率
func QueryAllRoomCaptureRate(db *sql.DB) ([]*Capturerate, error) {
	rows, err := db.Query(`select gameroomid,capturerate from fish_room_config`)
	if err != nil {
		glog.SErrorf(`query fish_room_config failed,err:%v`, err)
		return nil, err
	}
	rates := make([]*Capturerate, 0)
	defer rows.Close()
	for rows.Next() {
		rate := &Capturerate{}
		rows.Scan(&rate.RoomID, &rate.Rate)
		rates = append(rates, rate)
	}
	return rates, nil
}

func UpdateRoomConfig(db *sql.DB, roomid int32, rate float64) error {
	var count int32
	err := db.QueryRow(`select count(*) from fish_room_config where gameroomid=?`, roomid).Scan(&count)
	if err != nil {
		glog.SErrorf(`query fish_room_config failed,roomid:%v,err:%v`, roomid, err)
		return err
	}
	if count != 1 {
		return fmt.Errorf("invalid fish_room_config count:%v", count)
	}

	_, err = db.Exec(`update fish_room_config set capturerate=? where gameroomid=?`, rate, roomid)
	if err != nil {
		glog.SErrorf("update fish_room_config failed err:%v", err)
		return err
	}
	return nil
}

func InitRoomConfig(db *sql.DB, roomid int32, rate float64) error {
	_, err := db.Exec(`insert into fish_room_config(gameroomid,capturerate) values(?,?) 
		`, roomid, consts.CaptureRate)
	if err != nil {
		glog.SErrorf(`init fish_room_config failed,err:%v`, err)
		return err
	}
	return nil
}

func CheckRoomConfigExists(db *sql.DB, roomid int32) (bool, error) {
	var count int32
	err := db.QueryRow(`select count(*) from fish_room_config where gameroomid=?`, roomid).Scan(&count)
	if err != nil {
		glog.SErrorf(`query fish_room_config failed,roomid:%v,err:%v`, roomid, err)
		return false, err
	}
	exists := count > 0
	return exists, nil
}
