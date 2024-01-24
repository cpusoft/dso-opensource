package transact

import (
	"errors"
	"strings"

	dnsconvert "dns-model/convert"
	"dns-model/packet"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func queryQuestionDb(questionModel *packet.PacketModel) (questionRrModel *rr.RrModel, answerRrModels, authorityRrModels, additonalRrModels []*rr.RrModel, err error) {

	belogs.Debug("queryQuestionDb():questionModel:", jsonutil.MarshalJson(questionModel))
	fullDomain := rr.FormatRrFullDomain(string(questionModel.PacketDomain.FullDomain))
	packetType := questionModel.PacketType
	packetRrType := dnsutil.DnsIntTypes[packetType]
	packetClass := questionModel.PacketClass
	packetRrClass := dnsutil.DnsIntClasses[packetClass]
	belogs.Debug("queryQuestionDb(): fullDomain:", fullDomain, "   packetType:", packetType, "  packetRrTyp:", packetRrType,
		"   packetClass:", packetClass, "   packetRrClass:", packetRrClass)

	// found origin
	origins := make([]string, 0)
	orginSql := `select o.origin from lab_dns_origin o order by o.id`
	err = xormdb.XormEngine.SQL(orginSql).Find(&origins)
	if err != nil {
		belogs.Error("queryQuestionDb(): sql orginSql fail, orginSql:", orginSql, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("queryQuestionDb(): Find origins:", jsonutil.MarshalJson(origins))
	var origin string
	for i := range origins {
		if strings.HasSuffix(fullDomain, origins[i]) {
			origin = origins[i]
			break
		}
	}
	belogs.Debug("queryQuestionDb(): Find origin:", origin)
	if len(origin) == 0 {
		belogs.Error("queryQuestionDb(): sql orgin isnot found,origins:", jsonutil.MarshalJson(origins), "   fullDomain:", fullDomain)
		return nil, nil, nil, nil, errors.New("cannot found origin")
	}

	selectSql := `select o.origin, r.rrFullDomain, r.rrType, r.rrClass, IFNULL(r.rrTtl,o.ttl) as rrTtl, r.rrData  
	from lab_dns_rr r,	lab_dns_origin o 
	where r.originId = o.id and   `
	orderSql := ` order by r.id `

	answerRrModels = make([]*rr.RrModel, 0)
	sql := selectSql + ` r.rrFullDomain = ? and r.rrType = ? ` + orderSql
	belogs.Debug("queryQuestionDb(): get answerRrModels, sql:", sql)
	err = xormdb.XormEngine.SQL(sql, fullDomain, packetRrType).Find(&answerRrModels)
	if err != nil {
		belogs.Error("queryQuestionDb(): sql answerRrModels fail, sql:", sql, "  fullDomain:", fullDomain,
			"   rrType:", packetRrType, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("queryQuestionDb(): Find answerRrModels:", jsonutil.MarshalJson(answerRrModels))

	authorityRrModels = make([]*rr.RrModel, 0)
	sql = selectSql + ` o.origin = ? and r.rrType = 'NS' ` + orderSql
	belogs.Debug("queryQuestionDb(): get authorityRrModels, sql:", sql)
	err = xormdb.XormEngine.SQL(sql, origin).Find(&authorityRrModels)
	if err != nil {
		belogs.Error("queryQuestionDb(): sql authorityRrModels fail, sql:", sql, "  fullDomain:", fullDomain,
			"   rrType:", packetRrType, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("queryQuestionDb(): Find authorityRrModels:", jsonutil.MarshalJson(authorityRrModels))

	additonalRrModels = make([]*rr.RrModel, 0)
	if len(authorityRrModels) > 0 {
		var cnNames strings.Builder
		cnNames.WriteString("(")
		for i := range authorityRrModels {
			cnNames.WriteString("'" + authorityRrModels[i].RrData + "'")
			if i < len(authorityRrModels)-1 {
				cnNames.WriteString(",")
			}
		}
		cnNames.WriteString(")")
		belogs.Debug("queryQuestionDb(): for additonalRrModels, cnNames:", cnNames.String())
		sql = selectSql + " r.rrFullDomain in " + cnNames.String() + `  and r.rrType in ('A','AAAA') ` + orderSql
		belogs.Debug("queryQuestionDb(): get additonalRrModels, sql:", sql)
		err = xormdb.XormEngine.SQL(sql).Find(&additonalRrModels)
		if err != nil {
			belogs.Error("queryQuestionDb(): sql additonalRrModels fail, sql:", sql, "  fullDomain:", fullDomain,
				"   rrType:", packetRrType, err)
			return nil, nil, nil, nil, err
		}
	}
	belogs.Debug("queryQuestionDb(): additonalRrModels:", jsonutil.MarshalJson(additonalRrModels))

	// question is packet, convert to rr
	belogs.Debug("queryQuestionDb(): questionModel should ConvertPacketToRr :", jsonutil.MarshalJson(questionModel))
	questionRrModel, err = dnsconvert.ConvertPacketToRr(origin, questionModel)
	if err != nil {
		belogs.Error("queryQuestionDb(): ConvertPacketToRr questionRrModel fail, questionModel:", jsonutil.MarshalJson(questionModel),
			"   origin:", origin, err)
		return nil, nil, nil, nil, err
	}
	belogs.Debug("queryQuestionDb(): questionRrModel:", jsonutil.MarshalJson(questionRrModel))
	return

}
