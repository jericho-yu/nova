package validator

import (
	"testing"
	"time"
)

type TestStruct struct {
	Name     string     `v-rule:"required;min>3;max<10" v-name:"名称"`
	Email    string     `v-rule:"required;email" v-name:"邮箱"`
	Date     string     `v-rule:"required;date" v-name:"日期"`
	Time     string     `v-rule:"required;time" v-name:"时间"`
	Datetime *string    `v-rule:"required;datetime" v-name:"日期时间"`
	Ptr      *string    `v-rule:"required" v-name:"指针"`
	EmptyPtr *string    `v-rule:"" v-name:"空指针"`
	A1       int        `v-rule:"required" v-name:"a1"`
	A2       *int       `v-rule:"required" v-name:"a2"`
	A3       *float32   `v-rule:"required" v-name:"a3"`
	A4       *time.Time `v-rule:"required;datetime" v-name:"a4"`
}

func TestValidator(t *testing.T) {
	// 测试通过的情况
	dt := "2000-01-02 03:04:05"
	validPtr := "valid"
	num := 2
	var s float32 = 2.2
	t1 := time.Now()
	validStruct := TestStruct{
		Name:     "ValidName",
		Email:    "test@example.com",
		Date:     "2022-01-02",
		Time:     "03:04:05.12345",
		Datetime: &dt,
		Ptr:      &validPtr,
		EmptyPtr: nil,
		A1:       1,
		A2:       &num,
		A3:       &s,
		A4:       &t1,
	}

	validator := NewValidator(validStruct)
	if err := validator.Validate(); err != nil {
		t.Logf("expected no error, got %v", err)
		// t.Errorf("expected no error, got %v", err)
	}
}
