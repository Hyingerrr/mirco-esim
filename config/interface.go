package config

import (
	"time"

	"github.com/spf13/viper"
)

var _conf Config

type Config interface {
	Get(key string) interface{}

	GetString(key string) string

	GetBool(key string) bool

	GetInt(key string) int

	GetInt32(key string) int32

	GetInt64(key string) int64

	GetUint(key string) uint

	GetUint32(key string) uint32

	GetUint64(key string) uint64

	GetFloat64(key string) float64

	GetTime(key string) time.Time

	GetDuration(key string) time.Duration

	// GetIntSlice(key string) []int { return viper.GetIntSlice(key) }

	GetStringSlice(key string) []string

	GetStringMap(key string) map[string]interface{}

	GetStringMapString(key string) map[string]string

	GetStringMapStringSlice(key string) map[string][]string

	GetSizeInBytes(key string) uint

	UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error

	Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error

	Set(key string, value interface{})
}

func Get(key string) interface{} {
	return _conf.Get(key)
}

func GetString(key string) string {
	return _conf.GetString(key)
}

func GetBool(key string) bool {
	return _conf.GetBool(key)
}

func GetInt(key string) int {
	return _conf.GetInt(key)
}

func GetInt32(key string) int32 {
	return _conf.GetInt32(key)
}

func GetInt64(key string) int64 {
	return _conf.GetInt64(key)
}

func GetUint(key string) uint {
	return _conf.GetUint(key)
}

func GetUint32(key string) uint32 {
	return _conf.GetUint32(key)
}

func GetUint64(key string) uint64 {
	return _conf.GetUint64(key)
}

func GetFloat64(key string) float64 {
	return _conf.GetFloat64(key)
}

func GetTime(key string) time.Time {
	return _conf.GetTime(key)
}

func GetDuration(key string) time.Duration {
	return _conf.GetDuration(key)
}

func GetStringSlice(key string) []string {
	return _conf.GetStringSlice(key)
}

func GetStringMap(key string) map[string]interface{} {
	return _conf.GetStringMap(key)
}

func GetStringMapString(key string) map[string]string {
	return _conf.GetStringMapString(key)
}

func GetStringMapStringSlice(key string) map[string][]string {
	return _conf.GetStringMapStringSlice(key)
}

func GetSizeInBytes(key string) uint {
	return _conf.GetSizeInBytes(key)
}

func UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return _conf.UnmarshalKey(key, rawVal, opts...)
}

func Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return _conf.Unmarshal(rawVal, opts...)
}

func Set(key string, value interface{}) {
	_conf.Set(key, value)
}
