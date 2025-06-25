package util

import (
	"testing"
)

func TestNewJWT(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEyNSwicGxhdGZvcm0iOjIsInBrZ19uYW1lIjoiYSIsIkJ1ZmZlclRpbWUiOjg2NDAwLCJpc3MiOiJxbVBsdXMiLCJhdWQiOlsiR1ZBIl0sImV4cCI6MTc1MDE1NjUzOCwibmJmIjoxNzQ5NTUxNzM4fQ.7S-dqPiBaBMOF406wZtFHSszzZFjm_E_58WOja61XUc"
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbiI6IjEwMV94aWFvamllIiwidGltZSI6MTc0OTcyMzM1OSwibmlja05hbWUiOiJ4aWFvamllIiwiYWNjb3VudE5hbWUiOiJ4aWFvamllIiwibWVyY2hhbnRJZCI6IjEwMSIsImN1cnJlbmN5IjoiR0MiLCJleHAiOjE3NTAzMjgxNTksImlhdCI6MTc0OTcyMzM1OX0.kLtm-DiGJtTcJbUbrE2CixTMkNbHLbKL_07yPk_vAR0"

	jwt := NewJWT().SetSigningKey([]byte("tBS@#3ixECKe5yC")).SetIssuer("tBS@#3ixECKe5yC").SetBufferTime("1d").SetExpiresTime("7d")

	customToken, err := jwt.ParseToken(token)
	if err != nil {
		t.Errorf("err2:%v", err)
		return
	}
	t.Logf("customToken2: %v", customToken)
}
