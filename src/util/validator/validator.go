package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jericho-yu/nova/src/util/common"
	"github.com/jericho-yu/nova/src/util/operation"
	"github.com/jericho-yu/nova/src/util/str"
)

type (
	// Validator 验证器 验证规则 -> [required] [email|datetime|date|time] [min<|min<=] [max>|max=] [range=]
	Validator[T any] struct {
		data           T
		prefixNames    []string
		err            error
		emailFormat    string
		dateFormat     string
		timeFormat     string
		datetimeFormat string
		checkFunctions checkFunctionMap
	}

	checkFunction    func(rule string, fieldName string, value any) error
	checkFunctionMap map[string]checkFunction
)

// New 实例化：验证器
func New[T any](data T, prefixNames ...string) *Validator[T] {
	return NewValidator(data, prefixNames...)
}

// NewValidator 实例化：验证器
//
//go:fix 建议使用New方法
func NewValidator[T any](data T, prefixNames ...string) *Validator[T] {
	p := make([]string, 0)
	if len(prefixNames) > 0 {
		p = prefixNames
	}

	ins := &Validator[T]{
		data:           data,
		prefixNames:    p,
		emailFormat:    `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		dateFormat:     `^\d{4}-\d{2}-\d{2}$`,
		timeFormat:     `^\d{2}:\d{2}:\d{2}\.{0,1}\d+$`,
		datetimeFormat: `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`,
		checkFunctions: make(checkFunctionMap, 0),
	}

	ins.checkFunctions = checkFunctionMap{
		"string":     ins.checkString,
		"*string":    ins.checkString,
		"int":        ins.checkInt,
		"*int":       ins.checkInt,
		"int8":       ins.checkInt8,
		"*int8":      ins.checkInt8,
		"int16":      ins.checkInt16,
		"*int16":     ins.checkInt16,
		"int32":      ins.checkInt32,
		"*int32":     ins.checkInt32,
		"int64":      ins.checkInt64,
		"*int64":     ins.checkInt64,
		"uint":       ins.checkUint,
		"*uint":      ins.checkUint,
		"uint8":      ins.checkUint8,
		"*uint8":     ins.checkUint8,
		"uint16":     ins.checkUint16,
		"*uint16":    ins.checkUint16,
		"uint32":     ins.checkUint32,
		"*uint32":    ins.checkUint32,
		"uint64":     ins.checkUint64,
		"*uint64":    ins.checkUint64,
		"float32":    ins.checkFloat32,
		"*float32":   ins.checkFloat32,
		"float64":    ins.checkFloat64,
		"*float64":   ins.checkFloat64,
		"time.Time":  ins.checkTime,
		"*time.Time": ins.checkTime,
	}

	return ins
}

// Validate 执行验证
func (my *Validator[T]) Validate(exChecks ...func(item T) error) error {
	defer my.clean()

	if my.err != nil {
		return my.err
	}

	my.err = my.validate(my.data)
	if my.err != nil {
		return my.err
	}

	if len(exChecks) > 0 {
		for _, rule := range exChecks {
			if err := rule(my.data); err != nil {
				my.err = err
				return my.err
			}
		}
	}

	return my.err
}

// EmailFormat 设置email默认规则
func (my *Validator[T]) EmailFormat(emailFormat string) *Validator[T] {
	my.emailFormat = emailFormat

	return my
}

// DateFormat 设置日期默认规则
func (my *Validator[T]) DateFormat(dateFormat string) *Validator[T] {
	my.dateFormat = dateFormat

	return my
}

// TimeFormat 设置时间默认规则
func (my *Validator[T]) TimeFormat(timeFormat string) *Validator[T] {
	my.timeFormat = timeFormat

	return my
}

// DatetimeFormat 设置日期+时间默认规则
func (my *Validator[T]) DatetimeFormat(datetimeFormat string) *Validator[T] {
	my.datetimeFormat = datetimeFormat

	return my
}

func (my *Validator[T]) clean() { my.err = nil }

// validate 执行验证
func (my *Validator[T]) validate(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Ptr {
		return ValidateErr.New("不符合结构或指针")
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := range val.NumField() {
		field := val.Type().Field(i)
		if field.Anonymous {
			// 递归验证嵌套字段
			if err := NewValidator(val.Field(i).Interface(), my.prefixNames...).Validate(); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get("v-rule")
		if tag == "" || tag == "-" {
			continue
		}

		fieldName := my.concatFieldName(operation.Ternary(field.Tag.Get("v-name") != "", field.Tag.Get("v-name"), str.NewTransfer(val.Type().Name()).PascalToCamel()))

		for _, rule := range strings.Split(tag, ";") {
			if fn, exist := my.checkFunctions[fmt.Sprintf("%v", reflect.ValueOf(val.Field(i).Interface()).Type())]; exist {
				if err := fn(rule, fieldName, val.Field(i).Interface()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (my *Validator[T]) concatFieldName(fieldName string) string {
	var concatFieldNames = make([]string, len(my.prefixNames)+1)

	if len(my.prefixNames) > 0 {
		copy(concatFieldNames, my.prefixNames)
		concatFieldNames[len(my.prefixNames)] = fieldName

		return strings.Join(concatFieldNames, ".")
	}

	return fieldName
}

func (my *Validator[T]) checkTime(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	if !reflect.DeepEqual(value, time.Time{}) {
		return TimeErr.NewFormat("[%s]必须是时间类型", fieldName)
	}
	return nil
}

// checkString 验证：string -> 支持的规则 required、email、email=、date、date=、time、time=、datetime、datetime=、min<、min<=、max>、max>=、range=、length=
func (my *Validator[T]) checkString(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	if value.(string) == "" {
		return nil
	}

	switch {
	case rule == "required":
		if value == "" {
			return RequiredErr.New(fieldName)
		}
	case rule == "email=":
		emailFormat := strings.TrimPrefix(rule, "email=")
		if matched, _ := regexp.MatchString(emailFormat, value.(string)); !matched {
			return EmailErr.New(fieldName)
		}
	case rule == "email":
		if matched, _ := regexp.MatchString(my.emailFormat, value.(string)); !matched {
			return EmailErr.New(fieldName)
		}
	case strings.HasPrefix(rule, "time"):
		if matched, _ := regexp.MatchString(my.timeFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]时间格式错误，正确格式：%s", fieldName, my.timeFormat)
		}
	case strings.HasPrefix(rule, "time="):
		timeFormat := strings.TrimPrefix(rule, "time=")
		if matched, _ := regexp.MatchString(timeFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]时间格式错误，正确格式：%s", fieldName, timeFormat)
		}
	case strings.HasPrefix(rule, "datetime="):
		datetimeFormat := strings.TrimPrefix(rule, "datetime=")
		if matched, _ := regexp.MatchString(datetimeFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]时间格式错误，正确格式：%s", fieldName, datetimeFormat)
		}
	case strings.HasPrefix(rule, "datetime"):
		if matched, _ := regexp.MatchString(my.datetimeFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]时间格式错误，正确格式：%s", fieldName, my.datetimeFormat)
		}
	case strings.HasPrefix(rule, "date="):
		dateFormat := strings.TrimPrefix(rule, "date=")
		if matched, _ := regexp.MatchString(dateFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]日期格式错误，正确格式：%s", fieldName, dateFormat)
		}
	case strings.HasPrefix(rule, "date"):
		if matched, _ := regexp.MatchString(my.dateFormat, value.(string)); !matched {
			return TimeErr.NewFormat("[%s]日期格式错误，正确格式：%s", fieldName, my.dateFormat)
		}
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if utf8.RuneCountInString(value.(string)) <= common.ToInt(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if utf8.RuneCountInString(value.(string)) < common.ToInt(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if utf8.RuneCountInString(value.(string)) >= common.ToInt(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if utf8.RuneCountInString(value.(string)) > common.ToInt(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, ",")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt(betweens[0])
		max := common.ToInt(betweens[1])
		if utf8.RuneCountInString(value.(string)) < min || utf8.RuneCountInString(value.(string)) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	case strings.HasPrefix(rule, "length="):
		max := strings.TrimPrefix(rule, "length=")
		if utf8.RuneCountInString(value.(string)) != common.ToInt(max) {
			return LengthErr.NewFormat("[%s]长度必须为：%d", fieldName, common.ToInt(max))
		}
	}

	return nil
}

// checkInt 验证：int -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkInt(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(int) <= common.ToInt(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(int) < common.ToInt(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(int) >= common.ToInt(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(int) > common.ToInt(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt(betweens[0])
		max := common.ToInt(betweens[1])
		if value.(int) < min || value.(int) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkInt8 验证：int8 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkInt8(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(int8) <= common.ToInt8(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(int8) < common.ToInt8(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(int8) >= common.ToInt8(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(int8) > common.ToInt8(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt8(betweens[0])
		max := common.ToInt8(betweens[1])
		if value.(int8) < min || value.(int8) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkInt16 验证：int16 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkInt16(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(int16) <= common.ToInt16(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(int16) < common.ToInt16(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(int16) >= common.ToInt16(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(int16) > common.ToInt16(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt16(betweens[0])
		max := common.ToInt16(betweens[1])
		if value.(int16) < min || value.(int16) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkInt32 验证：int32 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkInt32(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(int32) <= common.ToInt32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(int32) < common.ToInt32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(int32) >= common.ToInt32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(int32) > common.ToInt32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt32(betweens[0])
		max := common.ToInt32(betweens[1])
		if value.(int32) < min || value.(int32) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkInt64 验证：int64 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkInt64(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(int64) <= common.ToInt64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(int64) < common.ToInt64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(int64) >= common.ToInt64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(int64) > common.ToInt64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToInt64(betweens[0])
		max := common.ToInt64(betweens[1])
		if value.(int64) < min || value.(int64) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkUint 验证：uint -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkUint(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(uint) <= common.ToUint(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(uint) < common.ToUint(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(uint) >= common.ToUint(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(uint) > common.ToUint(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToUint(betweens[0])
		max := common.ToUint(betweens[1])
		if value.(uint) < min || value.(uint) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkUint8 验证：uint8 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkUint8(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(uint8) <= common.ToUint8(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(uint8) < common.ToUint8(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(uint8) >= common.ToUint8(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(uint8) > common.ToUint8(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToUint8(betweens[0])
		max := common.ToUint8(betweens[1])
		if value.(uint8) < min || value.(uint8) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkUint16 验证：uint16 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkUint16(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(uint16) <= common.ToUint16(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(uint16) < common.ToUint16(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(uint16) >= common.ToUint16(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(uint16) > common.ToUint16(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToUint16(betweens[0])
		max := common.ToUint16(betweens[1])
		if value.(uint16) < min || value.(uint16) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkUint32 验证：uint32 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkUint32(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(uint32) <= common.ToUint32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(uint32) < common.ToUint32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(uint32) >= common.ToUint32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(uint32) > common.ToUint32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToUint32(betweens[0])
		max := common.ToUint32(betweens[1])
		if value.(uint32) < min || value.(uint32) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkUint64 验证：uint64 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkUint64(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(uint64) <= common.ToUint64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(uint64) < common.ToUint64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于：%d", fieldName, common.ToInt(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(uint64) >= common.ToUint64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(uint64) > common.ToUint64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于：%d", fieldName, common.ToInt(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：d~d", fieldName)
		}
		min := common.ToUint64(betweens[0])
		max := common.ToUint64(betweens[1])
		if value.(uint64) < min || value.(uint64) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%d~%d之间", fieldName, min, max)
		}
	}

	return nil
}

// checkFloat32 验证：float32 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkFloat32(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(float32) <= common.ToFloat32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%f", fieldName, common.ToFloat32(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(float32) < common.ToFloat32(min) {
			return LengthErr.NewFormat("[%s]长度不能小于[%f]", fieldName, common.ToFloat32(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(float32) >= common.ToFloat32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%f", fieldName, common.ToFloat32(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(float32) > common.ToFloat32(max) {
			return LengthErr.NewFormat("[%s]长度不能大于[%f]", fieldName, common.ToFloat32(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：f~f", fieldName)
		}
		min := common.ToFloat32(betweens[0])
		max := common.ToFloat32(betweens[1])
		if value.(float32) < min || value.(float32) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%f~%f之间", fieldName, min, max)
		}
	}

	return nil
}

// checkFloat64 验证：float64 -> 支持的规则 required、min<、min<=、max>、max>=、range=
func (my *Validator[T]) checkFloat64(rule, fieldName string, value any) error {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if rule == "required" && reflect.ValueOf(value).IsNil() {
			return RequiredErr.New(fieldName)
		}
		value = reflect.ValueOf(value).Elem().Interface()
	}

	switch {
	case strings.HasPrefix(rule, "min<="):
		min := strings.TrimPrefix(rule, "min<=")
		if value.(float64) <= common.ToFloat64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于等于：%f", fieldName, common.ToFloat64(min))
		}
	case strings.HasPrefix(rule, "min<"):
		min := strings.TrimPrefix(rule, "min<")
		if value.(float64) < common.ToFloat64(min) {
			return LengthErr.NewFormat("[%s]长度不能小于[%f]", fieldName, common.ToFloat64(min))
		}
	case strings.HasPrefix(rule, "max>="):
		max := strings.TrimPrefix(rule, "max>=")
		if value.(float64) >= common.ToFloat64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于等于：%f", fieldName, common.ToFloat64(max))
		}
	case strings.HasPrefix(rule, "max>"):
		max := strings.TrimPrefix(rule, "max>")
		if value.(float64) > common.ToFloat64(max) {
			return LengthErr.NewFormat("[%s]长度不能大于[%f]", fieldName, common.ToFloat64(max))
		}
	case strings.HasPrefix(rule, "range="):
		between := strings.TrimPrefix(rule, "range=")
		betweens := strings.Split(between, "~")
		if len(betweens) != 2 {
			return RuleErr.NewFormat("[%s]规则定义错误，规则定义错误，规则格式：f~f", fieldName)
		}
		min := common.ToFloat64(betweens[0])
		max := common.ToFloat64(betweens[1])
		if value.(float64) < min || value.(float64) > max {
			return LengthErr.NewFormat("[%s]长度必须在：%f~%f之间", fieldName, min, max)
		}
	}

	return nil
}
