package response

import (
	"net"
)

type Response struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	IP     string      `json:"ip"`
	Result interface{} `json:"result"`
}

type IpResult struct {
	Version     string `json:"version"`
	Continent   string `json:"continent"`
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	District    string `json:"district"`
	Isp         string `json:"isp"`
	CountryCode string `json:"country_code"`
	Fields      []string `json:"fields"`
	Raw         string `json:"raw"`
}

const (
	CodeSuccess = 0

	ErrCodeInvalidIP      = 4001
	ErrCodeIPDataNotFound = 4002
	ErrCodeInternalError = 5001
)

func Success(ip string, result *IpResult) Response {
	return Response{
		Code:   CodeSuccess,
		Msg:    "success",
		IP:     ip,
		Result: result,
	}
}

func Error(ip string, code int, msg string) Response {
	return Response{
		Code:   code,
		Msg:    msg,
		IP:     ip,
		Result: nil,
	}
}

func InvalidIP(ip string) Response {
	return Error(ip, ErrCodeInvalidIP, "invalid ip address")
}

func NotFound(ip string) Response {
	return Error(ip, ErrCodeIPDataNotFound, "ip data not found")
}

func InternalError(ip string) Response {
	return Error(ip, ErrCodeInternalError, "internal error")
}

func IsIPv6(ip string) bool {
	return net.ParseIP(ip).To4() == nil
}
