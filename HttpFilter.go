package httpfilter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Module struct {
	AppUri string `json:"appUri"`
	FilePath string `json:"filePath"`
	Id uint32 `json:"id"`
	RouteDesc string `json:"routeDesc"`
	Uri string `json:"uri"`
}
type Api struct {
	FunDesc string `json:"funDesc"`
	Method string `json:"method"`
	ModuleFun string `json:"moduleFun"`
	ModuleId uint32 `json:"moduleId"`
	ReqMsgId uint32 `json:"reqMsgId"`
	RspMsgId uint32 `json:"rspMsgId"`
	SubUri string `json:"subUri"`
}
type http struct {
	ModuleMap map[string]Module
	ApiMap map[string]map[string]map[string]Api
	FieldMap map[string]interface{}
}
type Rule struct {
	FieldId string `json:"fieldId"`
	CheckType string `json:"checkType"`
	ExprVal []interface{}
}
type filedCfg struct {
	Id uint32 `json:"id"`
	FieldUrl string `json:"fieldUrl"`
	FieldType string `json:"fieldType"`
	ExpVal string `json:"expVal"`
	LenLimit uint32 `json:"lenLimit"`
	Rules interface{} `json:"rules"`
	CheckType string `json:"checkType"`
	ExprVal []interface{} `json:"exprVal"`
	FieldDesc string `json:"fieldDesc"`
	IfMust string `json:"ifMust"`
	KeyType string `json:"keyType"`
}

type RstType struct {
	CodeKey string
	Msg interface{}
	Data interface{}
}
type ResCode struct {
	RstKey string `json:"rstKey"`
	RstCode string `json:"rstCode"`
	CodeDesc string `json:"codeDesc"`
}
type Tips struct {
	Key string `json:"key"`
	Tips string `json:"tips"`
}
type Response struct {
	ResCodeMap map[string]ResCode
	TipsMap map[string]Tips
}

func (res *Response)SetResCodeMap(m map[string]interface{}) bool {
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &res.ResCodeMap)
	if err != nil {
		return false
	}
	return true
}
func (res *Response)SetTipsMap(m map[string]interface{}) bool {
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &res.TipsMap)
	if err != nil {
		return false
	}
	return true
}

func (res *Response)MSG(params RstType) gin.H{
	code := ""
	msg := ""
	if v, ok := res.ResCodeMap[params.CodeKey]; ok {
		code = v.RstCode
		msg = v.CodeDesc
	}

	if params.Msg == nil || params.Msg == "" {
		if v, ok := res.TipsMap[params.CodeKey]; ok {
			msg = v.Tips
		}
	} else {
		switch ret := params.Msg.(type) {
		case string:
			msg = ret
		case []string:
			msg = res.ForMatMsg(ret)
		default:
		}
	}

	var data interface{}
	switch ret := params.Data.(type) {
	case string:
		json.Unmarshal([]byte(ret), &data)
	default:
		data = params.Data
	}

	return gin.H{
		"code": code,
		"msg": msg,
		"data": data,
	}
}
func (res *Response)ForMatMsg(msgList []string) string {
	msg := ""
	if len(msgList) > 0 {
		if v, ok := res.TipsMap[msgList[0]]; ok {
			msg = v.Tips
			if len(msgList) > 1 {
				cutMsg := msg
				msgTemp := ""
				for i:=1; i<=len(msgList) -1; i++ {
					tplStr := cutMsg[0: len(cutMsg)]
					idx := strings.Index(tplStr, "%")
					if idx == -1 {
						return msg
					}

					eIdx := idx +2
					if i>=len(msgList) -1 {
						eIdx = len(cutMsg)
					}
					tMsg := tplStr[0 : eIdx]
					msgTemp += fmt.Sprintf(tMsg, msgList[i])
					cutMsg = cutMsg[idx+2 : len(cutMsg)]
				}
				msg = msgTemp
			}
		} else {
			msg = msgList[0]
		}
	}

	return msg
}
var R = &Response{}

var H = &http{}
func WebCommResponse(c *gin.Context) {
	rstType := RstType{CodeKey: "SYSTEM_TIPS", Msg: "", Data: "[]"}

	rstType.CodeKey = "SUCCESS"
	retData := R.MSG(rstType)
	c.JSON(200, retData)
}
func (p *http)SetRoutes(pGin *gin.Engine, moduleObjMap map[string]interface{}) {
	for moduleId, subUriMap := range p.ApiMap {
		var module = Module{}
		for _, module = range p.ModuleMap {
			if strconv.Itoa(int(module.Id)) == moduleId {
				break
			}
		}
		grp := pGin.Group(module.Uri)

		for subUri, methodMap := range subUriMap {
			for method, _ := range methodMap {
				value := reflect.ValueOf(grp)
				fun := value.MethodByName(strings.ToUpper(method))
				inputs := make([]reflect.Value, 2)
				inputs[0] = reflect.ValueOf(subUri)
				inputs[1] = reflect.ValueOf(moduleObjMap[module.Uri]).MethodByName(strings.Title(subUri))
				defer func() {
					if p := recover(); p != nil{
						fmt.Println(p)
						inputs[1] = reflect.ValueOf(WebCommResponse)
						fun.Call(inputs)
					}
				}()
				fun.Call(inputs)
			}
		}
	}
}
func (p *http)SetModuleMap(m map[string]interface{}) bool {
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.ModuleMap)
	if err != nil {
		return false
	}
	return true
}
func (p *http)SetApiMap(m map[string]interface{}) bool {
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.ApiMap)
	if err != nil {
		return false
	}
	return true
}
func (p *http)SetFieldMap(m map[string]interface{}) bool {
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.FieldMap)
	if err != nil {
		return false
	}
	return true
}
func (p *http)CheckReq(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Abort()
		return
	}
	baseUrlList := strings.Split(c.Request.URL.Path, "/")
	subUri := baseUrlList[len(baseUrlList) - 1]
	baseUrl:= ""
	for idx:=0; idx<len(baseUrlList) - 1; idx++ {
		if baseUrlList[idx] != "" {
			baseUrl += "/" + baseUrlList[idx]
		}
	}

	module, ok := p.ModuleMap[baseUrl]
	rstType := RstType{CodeKey: "SYSTEM_TIPS", Msg: []string{"ERR_URL"}, Data: "[]"}
	if !ok {
		retData := R.MSG(rstType)
		c.JSON(200, retData)
		c.Abort()
		return
	}

	if p.ApiMap[strconv.Itoa(int(module.Id))] == nil ||
		p.ApiMap[strconv.Itoa(int(module.Id))][subUri] == nil {
		_, ok := p.ApiMap[strconv.Itoa(int(module.Id))][subUri][strings.ToLower(c.Request.Method)]
		if !ok {
			retData := R.MSG(rstType)
			c.JSON(200, retData)
			c.Abort()
			return
		}
	}

	reqMsgId := p.ApiMap[strconv.Itoa(int(module.Id))][subUri][strings.ToLower(c.Request.Method)].ReqMsgId
	if reqMsgId <= 0 {
		c.Next()
		return
	}

	data, err := c.GetRawData()
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	var body interface{}
	_ = json.Unmarshal(data, &body)

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	retMsgData := p.cycleCheckParams(p.FieldMap[strconv.Itoa(int(reqMsgId))].(map[string]interface{}), body)
	if retMsgData != nil {
		rstType.Msg = retMsgData
		retData := R.MSG(rstType)
		c.JSON(200, retData)
		c.Abort()
		return
	}

	c.Next()
}
func (p *http)cycleCheckParams(msgFieldMap map[string]interface{}, data interface{}) interface{}{
	if msgFieldMap == nil || msgFieldMap["__FieldCfg"] == nil {
		return nil
	}

	resByte, _ := json.Marshal(msgFieldMap["__FieldCfg"])
	var fCfgItem filedCfg
	json.Unmarshal(resByte, &fCfgItem)
	if fCfgItem.IfMust == "" {
		return nil
	}

	for key, value := range msgFieldMap {
		if key == "__FieldCfg" {
			if fCfgItem.FieldUrl == "" {
				if fCfgItem.IfMust == "YES" {
					if data == nil {
						return []string{"NULL_MSG_BODY"}
					}

					if fCfgItem.FieldType == "OBJ" {
						m, ok := data.(map[string]interface{})
						if !ok {
							return []string{"WRONG_OBJ_BODY"}
						}

						if len(m) == 0 {
							return []string{"NULL_MSG_BODY"}
						}
					}

					if fCfgItem.FieldType == "LIST" {
						l, ok := data.([]interface{})
						if !ok {
							return []string{"WRONG_OBJ_BODY"}
						}

						if len(l) == 0 {
							return []string{"NULL_MSG_BODY"}
						}
					}
				}
			} else {
				if fCfgItem.IfMust == "NO" && (data == nil || data == "") {
					return nil
				}

				var retMsgData interface{} = nil
				if fCfgItem.FieldType == "STR" {
					retMsgData = p.isStringOk(fCfgItem, data)
				} else if fCfgItem.FieldType == "INT" {
					retMsgData = p.isIntOk(fCfgItem, data)
				} else if fCfgItem.FieldType == "OBJ" {
					retMsgData = p.isObjOk(fCfgItem, data)
				} else if fCfgItem.FieldType == "LIST" {
					retMsgData = p.isListOk(fCfgItem, data)
				}

				if retMsgData != nil {
					return retMsgData
				}
			}
			continue
		}

		if fCfgItem.FieldType == "LIST" || (fCfgItem.FieldType == "OBJ" && fCfgItem.KeyType == "VOBJ") {
			l, _ := data.([]interface{})
			for _, v := range l {
				retMsgData := p.cycleCheckParams(value.(map[string]interface{}), v)
				if retMsgData != nil {
					return retMsgData
				}
			}
		} else {
			m, _ := data.(map[string]interface{})
			retMsgData := p.cycleCheckParams(value.(map[string]interface{}), m[key])
			if retMsgData != nil {
				return retMsgData
			}
		}
	}

	return nil
}
func (p *http)isStringOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if paramValue == nil || paramValue == "" {
		return []string{"NULL_STR_FIELD", fCfgItem.FieldUrl}
	}

	s, ok := paramValue.(string)
	if !ok {
		return []string{"NOT_STR_VALUE", fCfgItem.FieldUrl}
	}

	isPass := true
	rules, ok := fCfgItem.Rules.([]interface{})
	if !ok {
		return nil
	}

	for _, item := range rules {
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		exprVal, ok := rule["exprVal"].([]interface{})
		if !ok {
			continue
		}
		if rule["checkType"] == "RANGE" {
			_, ok = exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if len(s) < int(math.Floor(v[0].(float64))) && len(s) > int(math.Floor(v[1].(float64))) {
						isPass = false
						break
					}
				}
			} else {
				if len(s) < int(math.Floor(exprVal[0].(float64))) || len(s) > int(math.Floor(exprVal[1].(float64))) {
					isPass = false
				}
			}

			if !isPass {
				resByte, _ := json.Marshal(exprVal)
				return []string{"WRONG_STR_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "ENU" {
			isPass = false
			if rule["isCaseSensitive"] == 1 {
				for _, value := range exprVal {
					v, _ := value.(string)
					if s == v {
						isPass = true
						break
					}
				}
			} else {
				for _, value := range exprVal {
					v, _ := value.(string)
					if strings.ToUpper(s) == strings.ToUpper(v) {
						isPass = true
						break
					}
				}
			}

			if !isPass {
				resByte, _ := json.Marshal(exprVal)
				return []string{"WRONG_ENU_VALUE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "REGEX" {
			isPass = false
			resByte, _ := json.Marshal(exprVal)
			str := string(resByte)
			for _, value := range exprVal {
				v, ok := value.(string)
				if !ok {
					return []string{"SYSTEM_ERR"}
				}

				str := strings.Trim(v, "/")
				reg, err := regexp.Compile(str)
				if err != nil {
					return []string{"SYSTEM_ERR"}
				}

				found := reg.MatchString(s)
				if rule["matchType"] == "AND" {
					if !found {
						str = v
						break
					}
				} else {
					if found {
						isPass = true
						break
					}
				}
			}
			if !isPass {
				return []string{"WRONG_REGEX_VALUE", fCfgItem.FieldUrl, str}
			}
		}
	}

	return nil
}
func (p *http)isIntOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if paramValue == nil {
		return []string{"NULL_INT_FIELD", fCfgItem.FieldUrl}
	}

	i, ok := paramValue.(float64)
	if !ok {
		return []string{"NOT_INT_VALUE", fCfgItem.FieldUrl}
	}

	isPass := true
	rules, ok := fCfgItem.Rules.([]interface{})
	if !ok {
		return nil
	}
	for _, item := range rules {
		rule, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		exprVal, ok := rule["exprVal"].([]interface{})
		if !ok {
			continue
		}

		if rule["checkType"] == "RANGE" {
			_, ok := exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if i < v[0].(float64) && i > v[1].(float64) {
						isPass = false
						break
					}
				}
			} else {
				if i < exprVal[0].(float64) || i > exprVal[1].(float64) {
					isPass = false
				}
			}

			if !isPass {
				resByte, _ := json.Marshal(exprVal)
				return []string{"WRONG_INT_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "ENU" {
			isPass = false
			for _, value := range exprVal {
				v, _ := value.(float64)
				if i == v {
					isPass = true
					break
				}
			}

			if !isPass {
				resByte, _ := json.Marshal(exprVal)
				return []string{"WRONG_INT_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "REGEX" {
			isPass = false
			resByte, _ := json.Marshal(exprVal)
			str := string(resByte)
			for _, value := range exprVal {
				v, ok := value.(string)
				if !ok {
					return []string{"SYSTEM_ERR"}
				}

				reg, err := regexp.Compile(v)
				if err != nil {
					return []string{"SYSTEM_ERR"}
				}

				s := fmt.Sprintf("%f", i)
				found := reg.MatchString(s)
				if rule["matchType"] == "AND" {
					if !found {
						str = v
						break
					}
				} else {
					if found {
						isPass = true
						break
					}
				}
			}
			if !isPass {
				return []string{"WRONG_REGEX_VALUE", fCfgItem.FieldUrl, str}
			}
		}
	}

	return nil
}
func (p *http)isObjOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if paramValue == nil {
		return []string{"NULL_FIELD", fCfgItem.FieldUrl}
	}

	resByte, err := json.Marshal(paramValue)
	if err != nil {
		return []string{"WRONG_OBJ_VALUE", fCfgItem.FieldUrl}
	}
	var m map[string]interface{}
	err = json.Unmarshal(resByte, &m)
	if err != nil {
		return []string{"WRONG_OBJ_VALUE", fCfgItem.FieldUrl}
	}

	if len(m) == 0 {
		return []string{"NULL_VALUE", fCfgItem.FieldUrl}
	}

	return nil
}
func (p *http)isListOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if paramValue == nil {
		return []string{"NULL_FIELD", fCfgItem.FieldUrl}
	}

	resByte, err := json.Marshal(paramValue)
	if err != nil {
		return []string{"WRONG_LIST_VALUE", fCfgItem.FieldUrl}
	}
	var l []interface{}
	err = json.Unmarshal(resByte, &l)
	if err != nil {
		return []string{"WRONG_LIST_VALUE", fCfgItem.FieldUrl}
	}

	if len(l) == 0 {
		return []string{"NULL_VALUE", fCfgItem.FieldUrl}
	}

	return nil
}
