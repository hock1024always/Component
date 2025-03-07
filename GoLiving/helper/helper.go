package helper

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
	"online_meeting/define"
)

type UserClaims struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	IsAdmin int    `json:"is_admin"`
	jwt.StandardClaims
}

// GetMd5
// 生成 md5
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
} //加密

func GetUUID() string {
	return uuid.NewV4().String()
	//NewV4()：生成一个新的随机 UUID
} //获取随即令牌 功能函数

// 生成和解析token
// 生成 token
func GenerateToken(id uint, name string) (string, error) {
	UserClaim := &UserClaims{
		Id:             id,
		Name:           name,
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim) //jwt类型的指针
	tokenString, err := token.SignedString(define.MyKey)          //转化成token字符串，之后返回给前端
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AnalyseToken
// 解析 token
func AnalyseToken(tokenString string) (*UserClaims, error) {
	userClaim := new(UserClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return define.MyKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("analyse Token Error:%v", err)
	}
	return userClaim, nil
}

// Encode函数用于将传入的obj对象进行编码，返回一个字符串
func Encode(obj interface{}) string {
	// 将obj对象转换为json格式
	b, err := json.Marshal(obj)
	// 如果转换失败，则抛出异常
	if err != nil {
		panic(err)
	}
	// 将json格式转换为base64编码
	return base64.StdEncoding.EncodeToString(b)
}

// 解码函数，将字符串解码为对象
func Decode(in string, obj interface{}) {
	// 将字符串解码为字节切片
	b, err := base64.StdEncoding.DecodeString(in)
	// 如果解码失败，则抛出异常
	if err != nil {
		panic(err)
	}
	// 将字节切片解码为对象
	err = json.Unmarshal(b, obj)
	// 如果解码失败，则抛出异常
	if err != nil {
		panic(err)
	}
}
