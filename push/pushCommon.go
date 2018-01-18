package push

import (
	"Taonet/common"
	"Taonet/taonetBase"
)

type RuleData struct {

	DataType 	common.PushType

	DataDetail	string

	DataSource	taonetBase.ITaoSession
}

func BuildRuleData(t common.PushType, detail string, source taonetBase.ITaoSession)*RuleData{
	return &RuleData{
		DataType:t,
		DataDetail:detail,
		DataSource:source,
	}
}
