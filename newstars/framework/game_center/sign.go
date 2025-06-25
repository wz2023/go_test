package game_center

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func sign(params map[string]interface{}, appKey string) string {
	// 根据key的自然顺序排序
	keys := make([]string, len(params))
	j := 0
	for k := range params {
		keys[j] = k
		j++
	}
	sort.Strings(keys)
	// 组合拼凑
	valueStr := ""
	for _, k := range keys {
		if v, ok := params[k]; ok {
			if valueStr == "" {
				valueStr = fmt.Sprintf("%s=%v", k, v)
			} else {
				valueStr = fmt.Sprintf("%s&%s=%v", valueStr, k, v)
			}
		}
	}
	if valueStr == "" {
		valueStr = fmt.Sprintf("%s=%v", "app_key", appKey)
	} else {
		valueStr = fmt.Sprintf("%s&%s=%v", valueStr, "app_key", appKey)
	}
	//log.Debugf("valueStr: %s", valueStr)
	// md5
	h := md5.New()
	h.Write([]byte(valueStr))
	signature := hex.EncodeToString(h.Sum(nil))
	signature = strings.ToLower(signature)
	//fmt.Println("signature:", signature, valueStr)
	//glog.Info("signature:", signature, valueStr)
	//global.GVA_LOG.Sugar().Info("signature:", signature, valueStr)
	return signature
}
