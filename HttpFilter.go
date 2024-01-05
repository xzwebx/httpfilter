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
	NullTips string `json:"nullTips"`
}

type RstType struct {
	CodeKey string
	Msg interface{}
	Data interface{}
	RspMsgId uint32
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
	isCheckedRes bool
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
func (res *Response)SetIsCheckedRes(isCheckedRes uint32) {
	if (isCheckedRes == 1) {
		res.isCheckedRes = true
	} else {
		res.isCheckedRes = false
	}
}

func (res *Response)MSG(params RstType) gin.H{
	if res.isCheckedRes && params.RspMsgId > 0 {
		retMsgData := H.cycleCheckParams(H.FieldMap[strconv.Itoa(int(params.RspMsgId))].(map[string]interface{}), params.Data)
		if retMsgData != nil {
			rstType := RstType{CodeKey: "SVC_ERR", Msg: retMsgData, Data: params.Data}
			return res.MSG(rstType)
		} else {
			params.RspMsgId = 0
			return res.MSG(params)
		}
	} else {
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
func (res *Response)GetRspMsgId(c *gin.Context) uint32 {
	val, exist := c.Get("__rspMsgId")
	if exist {
		i, ok := val.(uint32)
		if ok {
			return i
		}
	}
	return 0
}
var R = &Response{isCheckedRes: true}

var H = &http{}
func WebCommResponse(c *gin.Context) {
	rstType := RstType{CodeKey: "CLT_ERR", Msg: "", Data: "[]", RspMsgId: R.GetRspMsgId(c)}
	rstType.CodeKey = "SUCC"
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
				if !reflect.ValueOf(moduleObjMap[module.Uri]).IsValid() {
					inputs[1] = reflect.ValueOf(WebCommResponse)
				} else {
					if !reflect.ValueOf(moduleObjMap[module.Uri]).MethodByName(strings.Title(subUri)).IsValid() {
						inputs[1] = reflect.ValueOf(WebCommResponse)
					} else {
						inputs[1] = reflect.ValueOf(moduleObjMap[module.Uri]).MethodByName(strings.Title(subUri))
					}
				}
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
	rstType := RstType{CodeKey: "CLT_ERR", Msg: []string{"WEBX_ERR_URL"}, Data: "[]"}
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

	rspMsgId := p.ApiMap[strconv.Itoa(int(module.Id))][subUri][strings.ToLower(c.Request.Method)].RspMsgId
	if (rspMsgId > 0) {
		c.Set("__rspMsgId", rspMsgId)
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
	var retMsgData interface{} = nil

	if msgFieldMap == nil {
		return nil
	}

	var fCfgItem filedCfg
	isRoot := true
	if msgFieldMap["__FieldCfg"] != nil {
		isRoot = false
		resByte, _ := json.Marshal(msgFieldMap["__FieldCfg"])
		json.Unmarshal(resByte, &fCfgItem)
	}

	for key, value := range msgFieldMap {
		if isRoot {
			resByte, _ := json.Marshal(((value.(map[string]interface{}))["__FieldCfg"]))
			json.Unmarshal(resByte, &fCfgItem)
		}

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

		if key == "__FieldCfg" {
			continue
		}

		if fCfgItem.FieldType == "LIST" || (fCfgItem.FieldType == "OBJ" && fCfgItem.KeyType == "VOBJ") {
			if isRoot {
				retMsgData = p.cycleCheckParams(value.(map[string]interface{}), data)
				if retMsgData != nil {
					return retMsgData
				}
			} else {
				l, _ := data.([]interface{})
				for _, v := range l {
					retMsgData = p.cycleCheckParams(value.(map[string]interface{}), v)
					if retMsgData != nil {
						return retMsgData
					}
				}
			}
		} else {
			if isRoot {
				retMsgData = p.cycleCheckParams(value.(map[string]interface{}), data)
			} else {
				m, _ := data.(map[string]interface{})
				retMsgData = p.cycleCheckParams(value.(map[string]interface{}), m[key])
			}
			if retMsgData != nil {
				return retMsgData
			}
		}
	}

	return nil
}

func getCustomTips(tips string) string {
	if len(tips) > 3 && tips[0:2] == "${" && fmt.Sprintf("%c", tips[len(tips)-1]) == "}" {
		tipsKey := tips[2:len(tips)-1]
		tipsKey = strings.TrimSpace(tipsKey)
		if v, ok := R.TipsMap[tipsKey]; ok {
			return v.Tips
		}
	}
	return tips
}

func (p *http)isStringOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if fCfgItem.IfMust == ""{
		return nil
	}

	if fCfgItem.IfMust == "NO" && (paramValue == nil || paramValue == "") {
		return nil
	}

	s, ok := paramValue.(string)
	if !ok || paramValue == "" {
		tips := fmt.Sprint(fCfgItem.NullTips)
		if len(tips) > 0 {
			return getCustomTips(tips)
		}
		return []string{"WEBX_NULL_FIELD", "string", fCfgItem.FieldUrl}
	}

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
			isPass := false
			_, ok = exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if len(s) >= int(math.Floor(v[0].(float64))) && len(s) <= int(math.Floor(v[1].(float64))) {
						isPass = true
						break
					}
				}
			} else {
				if len(s) >= int(math.Floor(exprVal[0].(float64))) || len(s) <= int(math.Floor(exprVal[1].(float64))) {
					isPass = true
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				return []string{"WEBX_WRONG_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "ENU" {
			isPass := false
			if rule["isCaseSensitive"].(float64) == 1 {
				if rule["isMatched"].(float64) == 1 {
					for _, value := range exprVal {
						v, _ := value.(string)
						if s == v {
							isPass = true
							break
						}
					}
				} else {
					isPass = true
					for _, value := range exprVal {
						v, _ := value.(string)
						if s == v {
							isPass = false
							break
						}
					}
				}
			} else {
				if rule["isMatched"].(float64) == 1 {
					for _, value := range exprVal {
						v, _ := value.(string)
						if strings.ToUpper(s) == strings.ToUpper(v) {
							isPass = true
							break
						}
					}
				} else {
					isPass = true
					for _, value := range exprVal {
						v, _ := value.(string)
						if strings.ToUpper(s) == strings.ToUpper(v) {
							isPass = false
							break
						}
					}
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				if rule["isMatched"].(float64) == 1 {
					return []string{"WEBX_WRONG_ENU_VALUE", fCfgItem.FieldUrl, string(resByte)}
				} else {
					return []string{"WEBX_EXCLUSION_ENU_VALUE", fCfgItem.FieldUrl, string(resByte)}
				}
			}
		}

		if rule["checkType"] == "REGEX" {
			isPass := false
			resByte, _ := json.Marshal(exprVal)
			str := string(resByte)
			for _, value := range exprVal {
				v, ok := value.(string)
				if !ok {
					return []string{"SVC_ERR"}
				}

				str := strings.Trim(v, "/")
				reg, err := regexp.Compile(str)
				if err != nil {
					return []string{"SVC_ERR"}
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
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				return []string{"WEBX_WRONG_REGEX_VALUE", fCfgItem.FieldUrl, str}
			}
		}
	}

	return nil
}
func (p *http)isIntOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if fCfgItem.IfMust == ""{
		return nil
	}

	if fCfgItem.IfMust == "NO" && (paramValue == nil || paramValue == "") {
		return nil
	}

	i, ok := paramValue.(float64)
	if !ok || paramValue == nil {
		tips := fmt.Sprint(fCfgItem.NullTips)
		if len(tips) > 0 {
			return getCustomTips(tips)
		}
		return []string{"WEBX_NULL_FIELD", "number", fCfgItem.FieldUrl}
	}

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
			isPass := false
			_, ok := exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if i >= v[0].(float64) && i <= v[1].(float64) {
						isPass = true
						break
					}
				}
			} else {
				if i >= exprVal[0].(float64) && i <= exprVal[1].(float64) {
					isPass = true
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				return []string{"WEBX_WRONG_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}

		if rule["checkType"] == "ENU" {
			isPass := false
			if rule["isMatched"].(float64) == 1 {
				for _, value := range exprVal {
					v, _ := value.(float64)
					if i == v {
						isPass = true
						break
					}
				}
			} else {
				isPass = true
				for _, value := range exprVal {
					v, _ := value.(float64)
					if i == v {
						isPass = false
						break
					}
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				if rule["isMatched"].(float64) == 1 {
					return []string{"WEBX_WRONG_ENU_VALUE", fCfgItem.FieldUrl, string(resByte)}
				} else {
					return []string{"WEBX_EXCLUSION_ENU_VALUE", fCfgItem.FieldUrl, string(resByte)}
				}
			}
		}

		if rule["checkType"] == "REGEX" {
			isPass := false
			resByte, _ := json.Marshal(exprVal)
			str := string(resByte)
			for _, value := range exprVal {
				v, ok := value.(string)
				if !ok {
					return []string{"SVC_ERR"}
				}

				reg, err := regexp.Compile(v)
				if err != nil {
					return []string{"SVC_ERR"}
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
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				return []string{"WEBX_WRONG_REGEX_VALUE", fCfgItem.FieldUrl, str}
			}
		}
	}

	return nil
}
func (p *http)isObjOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if fCfgItem.IfMust == ""{
		return nil
	}

	if fCfgItem.IfMust == "NO" && (paramValue == nil || paramValue == "") {
		return nil
	}

	resByte, err1 := json.Marshal(paramValue)
	var m map[string]interface{}
	err2 := json.Unmarshal(resByte, &m)
	if paramValue == nil || err1 != nil || err2 != nil || len(m) == 0 {
		tips := fmt.Sprint(fCfgItem.NullTips)
		if len(tips) > 0 {
			return getCustomTips(tips)
		}
		return []string{"WEBX_NULL_FIELD", "map", fCfgItem.FieldUrl}
	}

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
			isPass := false
			_, ok = exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if len(m) >= int(math.Floor(v[0].(float64))) && len(m) <= int(math.Floor(v[1].(float64))) {
						isPass = true
						break
					}
				}
			} else {
				if len(m) >= int(math.Floor(exprVal[0].(float64))) && len(m) <= int(math.Floor(exprVal[1].(float64))) {
					isPass = true
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				return []string{"WEBX_WRONG_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}
	}

	return nil
}
func (p *http)isListOk(fCfgItem filedCfg, paramValue interface{}) interface{}{
	if fCfgItem.IfMust == ""{
		return nil
	}

	if fCfgItem.IfMust == "NO" && (paramValue == nil || paramValue == "") {
		return nil
	}

	resByte, err1 := json.Marshal(paramValue)
	var l []interface{}
	err2 := json.Unmarshal(resByte, &l)
	if paramValue == nil || err1 != nil || err2 != nil || len(l) == 0 {
		tips := fmt.Sprint(fCfgItem.NullTips)
		if len(tips) > 0 {
			return getCustomTips(tips)
		}
		return []string{"WEBX_NULL_FIELD", "list", fCfgItem.FieldUrl}
	}

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
			isPass := false
			_, ok = exprVal[0].([]interface{})
			if ok {
				for _, value := range exprVal {
					v, _ := value.([]interface{})
					if len(l) >= int(math.Floor(v[0].(float64))) && len(l) <= int(math.Floor(v[1].(float64))) {
						isPass = true
						break
					}
				}
			} else {
				if len(l) >= int(math.Floor(exprVal[0].(float64))) || len(l) <= int(math.Floor(exprVal[1].(float64))) {
					isPass = true
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				return []string{"WEBX_WRONG_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}
	}

	return nil
}
