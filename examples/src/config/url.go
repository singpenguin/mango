package config


import (
	. "controllers"
)

var Urls = map[string]interface{} {
	"/": &Index{},
	"/news/(\\d+)": &News{},
}
