package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func HashText(first, second string) string {

	secret := sha256.Sum256([]byte(first + "." + second))

	return fmt.Sprintf("%x", secret)
}

func GetDateUniqueText() string {

	n := time.Now()
	timeText := strings.ReplaceAll(n.Format("2006-01-02-15-04-05"), "-", "")
	r := math.Min(float64(rand.Int63n(n.UnixNano())%100000+10000), 99999)

	return fmt.Sprintf("%s%d", timeText, int64(r))
}

// Sha256 Sha256加密
func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src + time.Now().String()))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

// Sha256By Sha256加密 不加时间戳
func Sha256By(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

// Sha256Short Sha256加密
func Sha256Short(src string) string {
	res := Sha256(src)
	return res[len(res)-8:]
}

func GenerateUuid(length uint) (res string) {
	// 0-9在ASCII表中为48-57
	// a-z为97-122
	createRangeChar := func() rune {

		rand.New(rand.NewSource(time.Now().UnixNano()))
		randNum := rand.Int31n(36)
		if randNum < 10 {

			return randNum + 48
		} else {

			return randNum + 87
		}
	}

	for i := 0; i < int(length); i++ {

		res += string(createRangeChar())
	}
	return res
}

func GenerateRandomString(length int) string {
	characters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var result strings.Builder
	rand.Int63n(time.Now().UnixNano())

	for i := 0; i < length; i++ {
		index := rand.Intn(len(characters))
		result.WriteByte(characters[index])
	}

	return result.String()
}

// FormatWithComma 使用自定义函数实现千分位格式化
func FormatWithComma(amount float64, decimal int, split rune) string {
	// 构建格式字符串
	format := fmt.Sprintf("%%.%df", decimal)
	amountText := fmt.Sprintf(format, amount)

	// 分离整数部分和小数部分
	parts := strings.Split(amountText, ".")
	intPart := parts[0]

	n := len(intPart)
	if n <= 3 {
		return amountText
	}

	var result strings.Builder
	for i, digit := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(split)
		}
		result.WriteRune(digit)
	}

	// 如果有小数部分，则添加回结果
	if len(parts) > 1 {
		result.WriteString(".")
		result.WriteString(parts[1])
	}

	return result.String()
}

// FormatStrWithComma 使用自定义函数实现千分位格式化
func FormatStrWithComma(value string, split rune, limit int) string {

	n := len(value)
	if n <= limit {
		return value
	}

	var result strings.Builder
	for i, digit := range value {
		if i > 0 && i%limit == 0 {
			result.WriteRune(split)
		}
		result.WriteRune(digit)
	}

	return result.String()
}

// RemoveDiacritics removes diacritics from a given string.
func RemoveDiacritics(input string) string {
	replacements := map[rune]rune{
		'Ắ': 'A', 'Ằ': 'A', 'Ẳ': 'A', 'Ẵ': 'A', 'Ặ': 'A',
		'Ấ': 'A', 'Ầ': 'A', 'Ẩ': 'A', 'Ẫ': 'A', 'Ậ': 'A',
		'É': 'E', 'È': 'E', 'Ẻ': 'E', 'Ẽ': 'E', 'Ẹ': 'E',
		'Ế': 'E', 'Ề': 'E', 'Ể': 'E', 'Ễ': 'E', 'Ệ': 'E',
		'Í': 'I', 'Ì': 'I', 'Ỉ': 'I', 'Ĩ': 'I', 'Ị': 'I',
		'Ó': 'O', 'Ò': 'O', 'Ỏ': 'O', 'Õ': 'O', 'Ọ': 'O',
		'Ố': 'O', 'Ồ': 'O', 'Ổ': 'O', 'Ỗ': 'O', 'Ộ': 'O',
		'Ớ': 'O', 'Ờ': 'O', 'Ở': 'O', 'Ỡ': 'O', 'Ợ': 'O',
		'Ú': 'U', 'Ù': 'U', 'Ủ': 'U', 'Ũ': 'U', 'Ụ': 'U',
		'Ứ': 'U', 'Ừ': 'U', 'Ử': 'U', 'Ữ': 'U', 'Ự': 'U',
		'Ý': 'Y', 'Ỳ': 'Y', 'Ỷ': 'Y', 'Ỹ': 'Y', 'Ỵ': 'Y',
		'Ă': 'A', 'Â': 'A', 'Đ': 'D', 'Ê': 'E', 'Ô': 'O', 'Ơ': 'O', 'Ư': 'U',
		'!': 'I', // Specific replacement for "TH!"
	}

	// Normalize the input string to NFD (Normalization Form D)
	nfd := norm.NFD.String(input)
	var sb strings.Builder

	// Iterate over the normalized string and remove/replace characters
	for _, runeValue := range nfd {
		if replacement, exists := replacements[runeValue]; exists {
			sb.WriteRune(replacement)
		} else if !unicode.IsMark(runeValue) {
			sb.WriteRune(runeValue)
		}
	}

	return sb.String()
}

func CoverCyrillicToLatin(input string) string {

	var cyrillicToLatin = map[rune]string{
		'А': "A", 'Б': "B", 'В': "B", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "E", 'Ж': "ZH", 'З': "3", 'И': "I", 'Й': "Y",
		'К': "K", 'Л': "L", 'М': "M", 'Н': "H", 'О': "O", 'П': "P", 'Р': "P", 'С': "C", 'Т': "T", 'У': "U", 'Ф': "F",
		'Х': "X", 'Ц': "TS", 'Ч': "CH", 'Ш': "SH", 'Щ': "SHCH", 'Ы': "Y", 'Э': "E", 'Ю': "YU", 'Я': "YA",
		'а': "a", 'б': "6", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i", 'й': "y",
		'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f",
		'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch", 'ы': "y", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	return strings.Map(
		func(r rune) rune {
			if l, ok := cyrillicToLatin[r]; ok {
				return []rune(l)[0]
			}
			return r
		},
		input,
	)
}

type TransferInfo struct {
	Account string
	Balance uint64

	Amount       uint64
	AmountMethod string
}

func ExtractTransferInfo(content string) (info *TransferInfo, err error) {

	// 转换重音
	content = RemoveDiacritics(content)

	// 定义正则表达式，用于匹配金额和余额
	amountPrefix := `(\+|\-|So tien GD:?)(\s|\) ?|VND|)`
	amountRegex := amountPrefix + `+(\d{1,17}(?:,|\.\d{2})?\d{1,3}(?:\d{3})*(?:,|\.\d{3})*(?:,|\d{3})?)`
	amountRe := regexp.MustCompile(amountRegex)

	balancePrefix := `(?i)(So du:? ?\+?|So du cuoi:? ?|So du hien tai:? ?|SD:? ?|Balance:? ?|So du kha dung:? ?|Han muc kha dung:? ?)`
	balanceRegex := balancePrefix + `(\s|\)|:|VND|)+(\d{1,17}(?:,|\.\d{2})?\d{1,3}(?:\d{3})*(?:,|\.\d{3})*(?:,|\d{3})?)`
	balanceRe := regexp.MustCompile(balanceRegex)

	accountPrefix := `(?i)(TK|Tai khoan thanh toan|Tai khoan|Account):?\s*(VCB\s*)?(\s|\(|\)|:|VND|)`
	accountRegex := accountPrefix + `(\d{1,17})`
	accountRe := regexp.MustCompile(accountRegex)

	// 使用正则表达式匹配 content 中的金额和余额
	amountMatches := amountRe.FindStringSubmatch(content)
	balanceMatches := balanceRe.FindStringSubmatch(content)
	accountMatches := accountRe.FindStringSubmatch(content)

	info = &TransferInfo{}

	// 提取金额部分并转换为 float64
	if amountMatches != nil && len(amountMatches) > 3 {
		amountStr := strings.ReplaceAll(amountMatches[3], ",", "")

		var amount uint64
		amount, err = strconv.ParseUint(amountStr, 10, 64)

		if err != nil {
			return
		}

		info.AmountMethod = amountMatches[1]
		info.Amount = amount
	}

	// 提取余额部分并转换为 float64
	if balanceMatches != nil && len(balanceMatches) > 3 {
		balanceStr := strings.ReplaceAll(balanceMatches[3], ",", "")

		var balance uint64
		balance, err = strconv.ParseUint(balanceStr, 10, 64)
		if err != nil {
			return
		}

		info.Balance = balance
	}

	// 提取账号部分
	if accountMatches != nil && len(accountMatches) > 3 {

		info.Account = accountMatches[len(accountMatches)-1]
	}

	// 如果所有信息都为空，则返回 nil
	if info.Amount == 0 && info.Balance == 0 && info.Account == "" {
		return nil, nil
	}

	return
}
