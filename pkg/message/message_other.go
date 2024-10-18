//go:build !windows

package message

import (
	"log"
)

func ShowAppLaunchMessage() {
	log.Printf("%s: %s\n", "运行成功", "服务已经启动.")
}
