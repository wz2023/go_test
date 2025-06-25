package util

import "github.com/rs/xid"

func GetTraceId() string {
	id := xid.New()
	return id.String()
}
