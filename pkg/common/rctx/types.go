package rctx

const (
	LabelProdCd    = "prodCode"
	LabelTranCd    = "tranCode"
	LabelRequestNo = "requestNo"
	LabelMerID     = "merId"
	LabelAppID     = "appId"
	LabelMethod    = "method"
	LabelProtocol  = "protocol"
	LabelEndpoint  = "endpoint"

	HTTPProtocol = "restful"
	RPCProtocol  = "gprc"
)

type CommonParams struct {
	RequestNo string `json:"requestNo"`
	TranCode  string `json:"tranCode"`
	TranCd    string `json:"tranCd"`
	ProdCode  string `json:"prodCode"`
	ProdCd    string `json:"prodCd"`
	MerID     string `json:"merId"`
	MerCd     string `json:"merCd"`
	AppID     string `json:"appId"`
	Protocol  string `json:"protocol"` // 请求协议，http/grpc
	Method    string `json:"method"`   // 请求方法
	Endpoint  string `json:"endpoint"`
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
