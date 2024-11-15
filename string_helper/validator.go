package string_helper

import "regexp"

// ValidateIDList 验证字符串是否为逗号分隔的id列表
func ValidateIDList(str string) bool {
	// 定义正则表达式
	regex := regexp.MustCompile(`^(\d+,)*\d+$`)

	// 使用正则表达式验证字符串
	return regex.MatchString(str)
}
