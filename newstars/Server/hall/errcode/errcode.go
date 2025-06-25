package errcode

//错误
const (
	CodeOK              = 0
	DBError             = -301
	InvalidParamError   = -302
	MutilLoginError     = -303
	MutilAccNameError   = -304
	MutilMobileError    = -305
	UserNameError       = -306
	PasswordError       = -307
	PlatformError       = -308
	MachineIDNotExsit   = -309
	AccountForbid       = -310
	NickNameDuplicate   = -311
	AuthCodeError       = -312
	AlipayExsitError    = -313
	MobileNotExst       = -314
	AmountlimitError    = -315
	WealthNotEnough     = -316
	ServerStopError     = -317 //停服拒绝登陆
	AppleVerifyError    = -318 //苹果验证失败
	AlipayLimitedError  = -319 //支付宝额度用完(已经兑换的超过额度)
	AlipayCountError    = -320 //支付宝次数限制
	BankCardExsitError  = -321 //银行卡未绑定
	AlipayLimitingError = -322 //支付宝金额不足(已经兑换的+正在提取的超过额度)
	BankCardCountError  = -323 //银行卡次数限制
	BankLimitedError    = -324 //银行卡额度用完(已经兑换的超过额度)
	BankLimitingError   = -325 //银行卡额度不足(已经兑换的+正在提取的超过额度)
	PayBusyError        = -326 //提款过于频繁
	IPForbid            = -329 //IP限制
	IPEmptyIP           = -330 //IP为空
	IPLimit             = -331 //IP限制
	FrequencyLimit      = -332 //请求太快
)
