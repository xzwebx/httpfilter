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
	"strings"
)

type http struct {
	ModuleMap map[string]Module
	ApiMap map[string]map[string]map[string]Api
	FieldMap map[string]interface{}
	ResCodeMap map[string]ResCode
	TipsMap map[string]Tips
	isCheckedRes bool
	c *gin.Context
	dtoObjMap map[string]interface{}
	moduleObjMap map[string]interface{}
}
type Module struct {
	AppUri string `json:"appUri"`
	FilePath string `json:"filePath"`
	Id string `json:"id"`
	RouteDesc string `json:"routeDesc"`
	Uri string `json:"uri"`
}
type Api struct {
	FunDesc string `json:"funDesc"`
	Method string `json:"method"`
	ModuleFun string `json:"moduleFun"`
	ModuleId string `json:"moduleId"`
	ReqMsgId string `json:"reqMsgId"`
	RspMsgId string `json:"rspMsgId"`
	SubUri string `json:"subUri"`
	Meta string `json:"meta"`
	AppMeta string `json:"appMeta"`
	ServiceMeta string `json:"serviceMeta"`
	ModuleMeta string `json:"moduleMeta"`
}
type filedCfg struct {
	Id string `json:"id"`
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
type Rule struct {
	FieldId string `json:"fieldId"`
	CheckType string `json:"checkType"`
	ExprVal []interface{}
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

func SetModuleMap(port uint64, m map[string]interface{}) bool {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.ModuleMap)
	if err != nil {
		return false
	}
	return true
}
func SetApiMap(port uint64, m map[string]interface{}) bool {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.ApiMap)
	if err != nil {
		return false
	}
	return true
}
func SetFieldMap(port uint64, m map[string]interface{}) bool {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.FieldMap)
	if err != nil {
		return false
	}
	return true
}
func SetResCodeMap(port uint64, m map[string]interface{}) bool {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.ResCodeMap)
	if err != nil {
		return false
	}
	return true
}
func SetTipsMap(port uint64, m map[string]interface{}) bool {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	dataType , _ := json.Marshal(m)
	dataString := string(dataType)
	err := json.Unmarshal([]byte(dataString), &p.TipsMap)
	if err != nil {
		return false
	}
	return true
}
func SetIsCheckedRes(port uint64, isCheckedRes uint32) {
	p, exist := httpMap[port]
	if !exist {
		p = &http{}
		httpMap[port] = p
	}
	if (isCheckedRes == 1) {
		p.isCheckedRes = true
	} else {
		p.isCheckedRes = false
	}
}
func SetRoutes(port uint64, pGin *gin.Engine, moduleObjMap map[string]interface{}, dtoObjMap map[string]interface{}) {
	p, _ := httpMap[port]
	p.dtoObjMap = dtoObjMap
	p.moduleObjMap = moduleObjMap
	for moduleId, subUriMap := range p.ApiMap {
		var module = Module{}
		for _, module = range p.ModuleMap {
			if module.Id == moduleId {
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
				inputs[1] = reflect.ValueOf(p.webCommResponse)
				//if !reflect.ValueOf(moduleObjMap[module.Uri]).IsValid() {
				//	inputs[1] = reflect.ValueOf(webCommResponse)
				//} else {
				//	if !reflect.ValueOf(moduleObjMap[module.Uri]).MethodByName(strings.Title(subUri)).IsValid() {
				//		inputs[1] = reflect.ValueOf(webCommResponse)
				//	} else {
				//		inputs[1] = reflect.ValueOf(moduleObjMap[module.Uri]).MethodByName(strings.Title(subUri))
				//	}
				//}
				fun.Call(inputs)
			}
		}
	}
}
func SetFirstFilter(port uint64, pGin *gin.Engine) {
	p, _ := httpMap[port]
	pGin.Use(p.checkReq)
}
func GetInterfaceInfo(c *gin.Context) *Api {
	val, exist := c.Get("interfaceInfo")
	if exist {
		i, ok := val.(Api)
		if ok {
			return &i
		}
	}
	return nil
}
func Response(c *gin.Context, codeKey string, msg interface{}, data interface{}) {
	retData := RstType{CodeKey: codeKey, Msg: msg, Data: data}
	obj, _ := c.Get("_HTTP_CFG")
	p, _ := obj.(*http)
	p.msg(retData)
}

var httpMap = make(map[uint64]*http)
func (p *http)checkReq(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Abort()
		return
	}

	c.Set("_HTTP_CFG", p)
	p.c = c

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
		Response(c, "CLT_ERR", []string{"WEBX_ERR_URL"}, "[]")
		c.Abort()
		return
	}

	if p.ApiMap[module.Id] == nil ||
		p.ApiMap[module.Id][subUri] == nil {
		_, ok := p.ApiMap[module.Id][subUri][strings.ToLower(c.Request.Method)]
		if !ok {
			Response(c, "CLT_ERR", []string{"WEBX_ERR_URL"}, "[]")
			c.Abort()
			return
		}
	}

	c.Set("interfaceInfo", p.ApiMap[module.Id][subUri][strings.ToLower(c.Request.Method)])

	reqMsgId := p.ApiMap[module.Id][subUri][strings.ToLower(c.Request.Method)].ReqMsgId
	if len(reqMsgId) <= 0 || reqMsgId == "0" {
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
	retMsgData := p.cycleCheckParams(p.FieldMap[reqMsgId].(map[string]interface{}), body)
	if retMsgData != nil {
		rstType.Msg = retMsgData
		Response(c, "CLT_ERR", retMsgData, "[]")
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
			return p.getCustomTips(tips)
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
				if len(s) >= int(math.Floor(exprVal[0].(float64))) && len(s) <= int(math.Floor(exprVal[1].(float64))) {
					isPass = true
				}
			}

			if !isPass {
				tips := fmt.Sprint(rule["ruleDesc"])
				if len(tips) > 0 {
					return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
			return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
			return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
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
			return p.getCustomTips(tips)
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
					return p.getCustomTips(tips)
				}
				resByte, _ := json.Marshal(exprVal)
				return []string{"WEBX_WRONG_RANGE", fCfgItem.FieldUrl, string(resByte)}
			}
		}
	}

	return nil
}
func (p *http)msg(params RstType) {
	if p.isCheckedRes {
		val, exist := p.c.Get("interfaceInfo")
		if exist {
			api, ok := val.(Api)
			if ok && len(api.RspMsgId) > 0 && api.RspMsgId != "0" && params.CodeKey == "SUCC" {
				var obj interface{}
				switch ret := params.Data.(type) {
				case string:
					json.Unmarshal([]byte(ret), &obj)
				default:
					obj = params.Data
				}
				retMsgData := p.cycleCheckParams(p.FieldMap[api.RspMsgId].(map[string]interface{}), obj)
				if retMsgData != nil {
					params.CodeKey = "SVC_ERR"
					params.Msg = retMsgData
				}
			}
		}
	}

	code := ""
	msg := ""
	if v, ok := p.ResCodeMap[params.CodeKey]; ok {
		code = v.RstCode
		msg = v.CodeDesc
	}

	if params.Msg == nil || params.Msg == "" {
		if v, ok := p.TipsMap[params.CodeKey]; ok {
			msg = v.Tips
		}
	} else {
		switch ret := params.Msg.(type) {
		case string:
			msg = ret
		case []string:
			msg = p.forMatMsg(ret)
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

	p.c.JSON(200, gin.H{
		"code": code,
		"msg": msg,
		"data": data,
	})
}
func (p *http)forMatMsg(msgList []string) string {
	msg := ""
	if len(msgList) > 0 {
		if v, ok := p.TipsMap[msgList[0]]; ok {
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
func (p *http)getCustomTips(tips string) string {
	if len(tips) > 3 && tips[0:2] == "${" && fmt.Sprintf("%c", tips[len(tips)-1]) == "}" {
		tipsKey := tips[2:len(tips)-1]
		tipsKey = strings.TrimSpace(tipsKey)
		if v, ok := p.TipsMap[tipsKey]; ok {
			return v.Tips
		}
	}
	return tips
}
func (p *http)webCommResponse(c *gin.Context) {
	baseUrlList := strings.Split(c.Request.URL.Path, "/")
	subUri := baseUrlList[len(baseUrlList) - 1]
	fullUrl := ""
	baseUrl := ""
	for idx := range baseUrlList {
		if baseUrlList[idx] == "" {
			continue
		}
		if idx<len(baseUrlList) - 1 {
			baseUrl += "/" + baseUrlList[idx]
		}
		fullUrl += "/" + baseUrlList[idx]
	}

	implsObj, exist := p.moduleObjMap[baseUrl]
	if !exist {
		Response(c, "SUCC", "", "[]")
		return
	}

	reqObj, exist := p.dtoObjMap[fullUrl]
	if !exist {
		Response(c, "SUCC", "", "[]")
		return
	}
	dynamicType := reflect.TypeOf(reqObj)
	newReq := reflect.New(dynamicType.Elem()).Interface()

	err := c.BindJSON(&newReq)
	if err != nil {
		Response(c, "CLT_ERR", err.Error(), "[]")
		return
	}

	module, exist := p.ModuleMap[baseUrl]
	if !exist {
		Response(c, "SUCC", "", "[]")
		return
	}
	api, exist := p.ApiMap[module.Id][subUri][strings.ToLower(c.Request.Method)]
	if !exist {
		Response(c, "SUCC", "", "[]")
		return
	}
	implsObjValue := reflect.ValueOf(implsObj)
	fun := implsObjValue.MethodByName(strings.Title(api.ModuleFun))
	if fun.IsValid() {
		inputs := make([]reflect.Value, 2)
		inputs[0] = reflect.ValueOf(c)
		inputs[1] = reflect.ValueOf(newReq)
		results := fun.Call(inputs)
		codeKey := results[0].Interface().(string)
		if len(codeKey) == 0 {
			codeKey = "SUCC"
		}
		msg := results[1].Interface().(string)
		data := results[2].Interface()
		Response(c, codeKey, msg, data)
		return
	}

	Response(c, "SUCC", "", "[]")
}
