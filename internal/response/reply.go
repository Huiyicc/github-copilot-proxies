package response

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
	HTTP状态码约定：
	服务器访问正常始终200,错误交给code
*/

func BindStruct(c *gin.Context, bind interface{}) error {
	if err := c.ShouldBindJSON(bind); err != nil {
		FailJson(c, NoIntactParameters, false, "结构体绑定错误")
		return errors.New("BindError")
	}
	return nil
}

func SuccessJson(c *gin.Context, msg string, data ...interface{}) {
	var tmps interface{}
	if len(data) > 0 {
		tmps = data[0]
	}
	c.JSON(http.StatusOK, Message{
		ErrCode: 0,
		Data:    tmps,
		Msg:     msg,
	})

}

func FailJson(c *gin.Context, load FailStruct, WriteLog bool, logMsh ...string) {
	if WriteLog {
		var werrmsg string
		for _, v := range logMsh {
			werrmsg += v + "\n"
		}
	}
	c.JSON(http.StatusOK, Message{
		ErrCode: load.Code,
		Msg:     load.Msg,
	})
}
func FailJsonAndStatusCode(c *gin.Context, code int, load FailStruct, WriteLog bool, logMsh ...string) {
	if WriteLog {
		var werrmsg string
		for _, v := range logMsh {
			werrmsg += v + "\n"
		}
	}
	c.JSON(code, Message{
		ErrCode: load.Code,
		Msg:     load.Msg,
	})
}

func SuccessByte(c *gin.Context, data []byte) {
	c.Writer.Write(data)
}
