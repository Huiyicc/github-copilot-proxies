package response

/*
	code约定:
	code代表错误，无错误始终为0
	200 OK - [GET]：服务器成功返回用户请求的数据；
	201 CREATED - [POST/PUT/PATCH]：用户新建或修改数据成功；
	202 Accepted - [*]：表示一个请求已经进入后台排队（异步任务）；
	204 NO CONTENT - [DELETE]：用户删除数据成功；
	400 INVALID REQUEST - [POST/PUT/PATCH]：用户发出的请求有错误，服务器没有进行新建或修改数据的操作；
	401 Unauthorized - [*]：表示用户没有权限（令牌、用户名、密码错误）；
	403 Forbidden - [*] 表示用户得到授权（与401错误相对），但是访问是被禁止的；
	404 NOT FOUND - [*]：用户发出的请求针对的是不存在的记录，服务器没有进行操作；
	406 Not Acceptable - [GET]：用户请求的格式不可得；
	410 Gone -[GET]：用户请求的资源被永久删除，且不会再得到的；
	422 Unprocesable entity - [POST/PUT/PATCH] 当创建一个对象时，发生一个验证错误；
	500 INTERNAL SERVER ERROR - [*]：服务器发生错误，用户将无法判断发出的请求是否成功。
*/

var (
	NoAccess = FailStruct{
		Code: 401,
		Msg:  "无权访问，请先登录",
	}
	TokenWrongful = FailStruct{
		Code: 401,
		Msg:  "Token不合法，请重新登录",
	}
	TokenOverdue = FailStruct{
		Code: 401,
		Msg:  "Token过期，请重新登录",
	}
	NoIntactParameters = FailStruct{
		Code: -10001,
		Msg:  "参数提交不完整，请重试",
	}
	UserError = FailStruct{
		Code: 10001,
		Msg:  "帐号或密码错误",
	}
	SignError = FailStruct{
		Code: 10002,
		Msg:  "",
	}
	CaptchaError = FailStruct{
		Code: 10003,
		Msg:  "生成验证码错误",
	}
	CaptchaVefError = FailStruct{
		Code: 10004,
		Msg:  "验证码错误",
	}
	WechatLoginError = FailStruct{
		Code: 10005,
		Msg:  "微信登陆错误",
	}
)

type FailStruct struct {
	Code int
	Msg  string
}

type Message struct {
	ErrCode int         `json:"error"`
	Data    interface{} `json:"data"`
	Msg     string      `json:"message"`
}

type Token struct {
	Token string `json:"token"`
}
