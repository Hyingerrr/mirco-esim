package meta

const (
	// 兼容http和grpc中的metadata，所以小写
	ProdCd      = "prodcode"
	TranCd      = "trancode"
	RequestNo   = "requestno"
	MerID       = "merid"
	AppID       = "appid"
	Method      = "method"
	Protocol    = "protocol"
	Endpoint    = "endpoint"
	Uri         = "uri" // method
	ServiceName = "servicename"
	TermNO      = "termno"
	TranSeq     = "transeq"
	SrcSysId    = "srcsysid"
	DstSysId    = "dstsysid"
	TraceID     = "traceid"

	StatusCode = "statuscode"

	HTTPProtocol = "restful"
	RPCProtocol  = "gprc"
)

type CommonParams struct {
	RequestNo   string `json:"requestNo"`
	TranCode    string `json:"tranCode"`
	TranCd      string `json:"tranCd"`
	ProdCode    string `json:"prodCode"`
	ProdCd      string `json:"prodCd"`
	MerID       string `json:"merId"`
	MerCd       string `json:"merCd"`
	AppID       string `json:"appId"`
	Protocol    string `json:"_"` // 请求协议，http/grpc
	Method      string `json:"_"` // 请求方法
	Endpoint    string `json:"_"`
	URI         string `json:"-"`
	ServiceName string `json:"-"`
}

func (cp CommonParams) ParseProdCd() string {
	if cp.ProdCode != "" {
		return cp.ProdCode
	}

	return cp.ProdCd
}

func (cp CommonParams) ParseTranCd() string {
	if cp.TranCode != "" {
		return cp.TranCode
	}

	return cp.TranCd
}

func (cp CommonParams) ParseMerID() string {
	if cp.MerCd != "" {
		return cp.MerCd
	}

	return cp.MerID
}

type CommonHeader struct {
	Head *InternalHeader `json:"head"`
}

type InternalHeader struct {
	AppId    string `json:"app_id"`     // 机构号
	TermNo   string `json:"term_no"`    // 终端号
	MerchNo  string `json:"merch_no"`   // 商户号
	MerID    string `json:"merId"`      // 商户号
	DstSysId string `json:"dst_sys_id"` // 服务方子系统id
	SrcSysId string `json:"src_sys_id"` // 调用方子系统id
	ProdCd   string `json:"prod_cd"`    // 产品码
	ProdCode string `json:"prodCode"`   // 产品码
	TranCd   string `json:"tran_cd"`    // 交易码
	TranCode string `json:"tranCode"`   // 交易码
	TranSeq  string `json:"tran_seq"`   // 流水号， 即订单
	TraceId  string `json:"trace_id"`   // 系统跟踪号
}

func (head InternalHeader) ParseMchNo() string {
	if head.MerchNo != "" {
		return head.MerchNo
	}

	return head.MerID
}

func (head InternalHeader) ParseProdCd() string {
	if head.ProdCode != "" {
		return head.ProdCode
	}

	return head.ProdCd
}

func (head InternalHeader) ParseTranCd() string {
	if head.TranCode != "" {
		return head.TranCode
	}

	return head.TranCd
}
