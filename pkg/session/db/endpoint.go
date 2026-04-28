package db

import (
	"net/url"
	"path"
	"strconv"
	"strings"
)

// ParseEndpoint 把连接串解析为 Endpoint。
func ParseEndpoint(text string) (*Endpoint, error) {
	u, err := url.ParseRequestURI(text)
	if err != nil {
		return nil, err
	}

	endpoint := &Endpoint{
		Scheme: u.Scheme,
	}

	query := u.Query()

	if len(query) > 0 {
		endpoint.Extra = u.Query()
	}

	endpoint.Path = u.Path

	endpoint.Hostname = u.Hostname()

	i, err := strconv.ParseUint(u.Port(), 10, 16)
	if err == nil {
		endpoint.Port = uint16(i)
	}

	if u.User != nil {
		endpoint.Username = u.User.Username()
		endpoint.Password, _ = u.User.Password()
	}

	return endpoint, nil
}

// Endpoint 表示数据库连接端点。
// openapi:strfmt endpoint
type Endpoint struct {
	Scheme   string
	Hostname string
	Port     uint16
	Path     string
	Username string
	Password string
	Extra    url.Values
}

// Base 返回路径对应的不带扩展名基础名。
func (e Endpoint) Base() string {
	if e.Path != "" {
		return strings.Split(
			path.Base(e.Path),
			".",
		)[0]
	}
	return ""
}

// IsZero 判断端点是否尚未配置 scheme。
func (e Endpoint) IsZero() bool {
	return e.Scheme == ""
}

// SecurityString 返回打码后的端点字符串。
func (e Endpoint) SecurityString() string {
	e.Password = strings.Repeat("-", len(e.Password))
	return e.String()
}

// Host 返回主机与端口组合后的地址。
func (e Endpoint) Host() string {
	if e.Port != 0 {
		return e.Hostname + ":" + strconv.FormatUint(uint64(e.Port), 10)
	}
	return e.Hostname
}

// String 把端点格式化为连接串。
func (e Endpoint) String() string {
	u := url.URL{}
	u.Scheme = e.Scheme
	u.Host = e.Host()

	if e.Extra != nil {
		u.RawQuery = e.Extra.Encode()
	}

	u.Path = e.Path

	if e.Username != "" || e.Password != "" {
		u.User = url.UserPassword(e.Username, e.Password)
	}

	return u.String()
}

// IsTLS 判断当前 scheme 是否表示 TLS 连接。
func (e *Endpoint) IsTLS() bool {
	if e.Scheme == "" {
		return false
	}
	return e.Scheme[len(e.Scheme)-1] == 's'
}

// UnmarshalText 从文本解析端点。
func (e *Endpoint) UnmarshalText(text []byte) error {
	endpoint, err := ParseEndpoint(string(text))
	if err != nil {
		return err
	}
	*e = *endpoint
	return nil
}

// MarshalText 把端点编码为文本。
func (e Endpoint) MarshalText() (text []byte, err error) {
	return []byte(e.String()), nil
}
