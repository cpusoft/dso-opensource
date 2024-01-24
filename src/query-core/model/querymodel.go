package model

import (
	"bytes"
	"encoding/binary"
	"errors"

	"dns-model/common"
	dnsconvert "dns-model/convert"
	"dns-model/packet"
	"dns-model/rr"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/dnsutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/guregu/null"
)

type QueryModel struct {
	HeaderForQueryModel common.HeaderForQueryModel `json:"headerForQueryModel"`
	CountQANAModel      common.CountQANAModel      `json:"countQANAModel"`
	QueryDataModel      QueryDataModel             `json:"queryDataModel"`
}

func (c QueryModel) GetHeaderModel() common.HeaderModel {
	return c.HeaderForQueryModel
}
func (c QueryModel) GetCountModel() common.CountModel {
	return c.CountQANAModel
}
func (c QueryModel) GetDataModel() interface{} {
	return c.QueryDataModel
}
func (c QueryModel) GetDnsModelType() string {
	return "packet"
}

type QueryDataModel struct {
	// QdCount
	QuestionModels []*packet.PacketModel `json:"questionModels"`
	// AnCount
	AnswerModels []*packet.PacketModel `json:"answerModels"`
	// NsCount --> ns type
	AuthorityModels []*packet.PacketModel `json:"authorityModels"`
	// ArCount
	AdditonalModels []*packet.PacketModel `json:"additonalModels"`
}

func NewQueryModelByHeaderAndCount(headerModel common.HeaderModel, countModel common.CountModel) (queryModel *QueryModel, err error) {
	belogs.Debug("NewQueryModelByHeaderAndCount():headerModel:", jsonutil.MarshalJson(headerModel),
		"  countModel:", jsonutil.MarshalJson(countModel))
	c := &QueryModel{}
	headerJson := jsonutil.MarshalJson(headerModel)
	countJson := jsonutil.MarshalJson(countModel)
	belogs.Debug("NewQueryModelByHeaderAndCount(): headerJson:", headerJson, "   countJson:", countJson)

	err = jsonutil.UnmarshalJson(headerJson, &c.HeaderForQueryModel)
	if err != nil {
		belogs.Error("NewQueryModelByHeaderAndCount():UnmarshalJson HeaderForQueryModel fail:", headerJson, err)
		return nil, err
	}
	err = jsonutil.UnmarshalJson(countJson, &c.CountQANAModel)
	if err != nil {
		belogs.Error("NewQueryModelByHeaderAndCount():UnmarshalJson CountQANAModel fail:", countJson, err)
		return nil, err
	}

	c.QueryDataModel.QuestionModels = make([]*packet.PacketModel, 0)
	c.QueryDataModel.AnswerModels = make([]*packet.PacketModel, 0)
	c.QueryDataModel.AuthorityModels = make([]*packet.PacketModel, 0)
	c.QueryDataModel.AdditonalModels = make([]*packet.PacketModel, 0)
	return c, nil
}

func NewQueryModelByParameters(id uint16, qr, aa, tc, rd, ra, rCode uint8) (queryModel *QueryModel, err error) {
	parameter := uint16(common.ComposeQrOpCodeAaTcRdRaZRCode(qr, dnsutil.DNS_OPCODE_QUERY, aa, tc, rd, ra, rCode))
	headerModel, _ := common.NewHeaderModel(id, parameter, common.DNS_HEADER_TYPE_QUERY)
	countModel, _ := common.NewCountModel(0, 0, 0, 0, common.DNS_COUNT_TYPE_QANA)
	return NewQueryModelByHeaderAndCount(headerModel, countModel)
}

// id ,qr, rCode -> header/count
// zTtl: just 0
// questionRrModel
func NewQueryModelForQuestionByParametersAndRrModels(id uint16, zTtl int64,
	questionRrModel *rr.RrModel) (queryModel *QueryModel, err error) {
	belogs.Debug("NewQueryModelForQuestionByParametersAndRrModels(): id:", id, "  zTtl:", zTtl,
		"  questionRrModel:", jsonutil.MarshalJson(questionRrModel))

	// header/count --> queryModel
	qr := dnsutil.DNS_QR_REQUEST
	rCode := dnsutil.DNS_RCODE_NOERROR
	queryModel, err = NewQueryModelByParameters(id, qr, 0, 0, 0, 0, rCode)
	if err != nil {
		belogs.Error("NewQueryModelForQuestionByParametersAndRrModels(): NewQueryModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewQueryModelForQuestionByParametersAndRrModels(): queryModel:", jsonutil.MarshalJson(queryModel))

	// question
	questionModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), questionRrModel, 0, packet.DNS_PACKET_NAME_TYPE_CLASS)
	if err != nil {
		belogs.Error("NewQueryModelForQuestionByParametersAndRrModels(): ConvertRrToPacket fail,  zTtl:", zTtl,
			"  questionRrModel:", jsonutil.MarshalJson(questionRrModel), err)
		return nil, err
	}
	belogs.Debug("NewQueryModelForQuestionByParametersAndRrModels(): questionModel:", jsonutil.MarshalJson(questionModel))

	queryModel.AddQuestionModel(questionModel)
	belogs.Debug("NewQueryModelForQuestionByParametersAndRrModels(): ok queryModel:", jsonutil.MarshalJson(queryModel))
	belogs.Info("#生成QUERY的'查询请求'类型数据包, Question部分: " + osutil.GetNewLineSep() +
		"{'域名':'" + questionRrModel.RrFullDomain +
		"','Type':'" + questionRrModel.RrType +
		"','Class':'" + questionRrModel.RrClass +
		"','Ttl':" + convert.ToString(questionRrModel.RrTtl.ValueOrZero()) + "}")
	return queryModel, nil
}

// id ,qr, rCode -> header/count
// zTtl: just 0
// questionRrModel
func NewQueryModelForAnswerByParametersAndRrModels(id uint16, zTtl int64,
	questionRrModel *rr.RrModel, answerRrModels []*rr.RrModel,
	authorityRrModels []*rr.RrModel, additonalRrModels []*rr.RrModel) (queryModel *QueryModel, err error) {
	belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): id:", id, "  zTtl:", zTtl,
		"  questionRrModel:", jsonutil.MarshalJson(questionRrModel))

	// header/count --> queryModel
	qr := dnsutil.DNS_QR_RESPONSE
	rCode := dnsutil.DNS_RCODE_NOERROR
	queryModel, err = NewQueryModelByParameters(id, qr, 0, 0, 0, 0, rCode)
	if err != nil {
		belogs.Error("NewQueryModelForAnswerByParametersAndRrModels(): NewQueryModelByParameters fail:", err)
		return nil, err
	}
	belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): queryModel:", jsonutil.MarshalJson(queryModel))

	// question
	questionModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), questionRrModel, 0, packet.DNS_PACKET_NAME_TYPE_CLASS)
	if err != nil {
		belogs.Error("NewQueryModelForAnswerByParametersAndRrModels(): add question, ConvertRrToPacket fail,  zTtl:", zTtl,
			"  questionRrModel:", jsonutil.MarshalJson(questionRrModel), err)
		return nil, err
	}
	belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): questionModel:", jsonutil.MarshalJson(questionModel))
	queryModel.AddQuestionModel(questionModel)
	belogs.Info("#生成QUERY的'查询响应'类型数据包, Question部分: " + osutil.GetNewLineSep() +
		"{'域名':'" + questionRrModel.RrFullDomain +
		"','Type':'" + questionRrModel.RrType +
		"','Class':'" + questionRrModel.RrClass +
		"','Ttl':" + convert.ToString(questionRrModel.RrTtl.ValueOrZero()) + "}")

	for i := range answerRrModels {
		answerModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), answerRrModels[i], 0, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
		if err != nil {
			belogs.Error("NewQueryModelForAnswerByParametersAndRrModels(): add answer, ConvertRrToPacket fail,  zTtl:", zTtl,
				"  answerRrModels[i]:", jsonutil.MarshalJson(answerRrModels[i]), err)
			return nil, err
		}
		belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): answerModel:", jsonutil.MarshalJson(answerModel))
		queryModel.AddAnswerModel(answerModel)
		belogs.Info("#生成QUERY的'查询响应'类型数据包, Answer部分: " + osutil.GetNewLineSep() +
			"{'域名':'" + answerRrModels[i].RrFullDomain +
			"','Type':'" + answerRrModels[i].RrType +
			"','Class':'" + answerRrModels[i].RrClass +
			"','Ttl':" + convert.ToString(answerRrModels[i].RrTtl.ValueOrZero()) +
			",'Data':" + answerRrModels[i].RrData + "'}")
	}

	for i := range authorityRrModels {
		authorityModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), authorityRrModels[i], 0, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
		if err != nil {
			belogs.Error("NewQueryModelForAnswerByParametersAndRrModels(): add authority, ConvertRrToPacket fail,  zTtl:", zTtl,
				"  authorityRrModels[i]:", jsonutil.MarshalJson(authorityRrModels[i]), err)
			return nil, err
		}
		belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): authorityModel:", jsonutil.MarshalJson(authorityModel))
		queryModel.AddAuthorityModel(authorityModel)
		belogs.Info("#生成QUERY的'查询响应'类型数据包, Authority部分: " + osutil.GetNewLineSep() +
			"{'域名':'" + authorityRrModels[i].RrFullDomain +
			"','Type':'" + authorityRrModels[i].RrType +
			"','Class':'" + authorityRrModels[i].RrClass +
			"','Ttl':" + convert.ToString(authorityRrModels[i].RrTtl.ValueOrZero()) +
			",'Data':" + authorityRrModels[i].RrData + "'}")
	}

	for i := range additonalRrModels {
		additonalModel, _, err := dnsconvert.ConvertRrToPacket(null.IntFrom(zTtl), additonalRrModels[i], 0, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA)
		if err != nil {
			belogs.Error("NewQueryModelForAnswerByParametersAndRrModels(): add additonal, ConvertRrToPacket fail,  zTtl:", zTtl,
				"  additonalRrModels[i]:", jsonutil.MarshalJson(additonalRrModels[i]), err)
			return nil, err
		}
		belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): additonalModel:", jsonutil.MarshalJson(additonalModel))
		queryModel.AddAdditonalModel(additonalModel)
		belogs.Info("#生成QUERY的'查询响应'类型数据包, Additional部分: " + osutil.GetNewLineSep() +
			"{'域名':'" + additonalRrModels[i].RrFullDomain +
			"','Type':'" + additonalRrModels[i].RrType +
			"','Class':'" + additonalRrModels[i].RrClass +
			"','Ttl':" + convert.ToString(additonalRrModels[i].RrTtl.ValueOrZero()) +
			",'Data':" + additonalRrModels[i].RrData + "'}")
	}

	belogs.Debug("NewQueryModelForAnswerByParametersAndRrModels(): ok queryModel:", jsonutil.MarshalJson(queryModel))
	return queryModel, nil
}

func (c *QueryModel) AddQuestionModel(questionModel *packet.PacketModel) {
	c.QueryDataModel.QuestionModels = append(c.QueryDataModel.QuestionModels, questionModel)
	c.CountQANAModel.QdCount += 1
}
func (c *QueryModel) AddAnswerModel(answerModel *packet.PacketModel) {
	c.QueryDataModel.AnswerModels = append(c.QueryDataModel.AnswerModels, answerModel)
	c.CountQANAModel.AnCount += 1
}

func (c *QueryModel) AddAuthorityModel(authorityModel *packet.PacketModel) {
	c.QueryDataModel.AuthorityModels = append(c.QueryDataModel.AuthorityModels, authorityModel)
	c.CountQANAModel.NsCount += 1
}

func (c *QueryModel) AddAdditonalModel(additonalModel *packet.PacketModel) {
	c.QueryDataModel.AdditonalModels = append(c.QueryDataModel.AdditonalModels, additonalModel)
	c.CountQANAModel.ArCount += 1
}

// queryData:=receiveData[offsetFromStart:]
func ParseBytesToQueryModel(headerModel common.HeaderModel, countModel common.CountModel,
	queryData []byte, offsetFromStart uint16) (receiveQueryModel *QueryModel, newOffsetFromStart uint16, err error) {
	belogs.Debug("ParseBytesToQueryModel(): headerModel:", jsonutil.MarshalJson(headerModel),
		"  countModel:", jsonutil.MarshalJson(countModel),
		"  queryData:", convert.PrintBytesOneLine(queryData), "   offsetFromStart:", offsetFromStart)

	// query bytes
	if len(queryData) == 0 {
		belogs.Error("ParseBytesToQueryModel(): recv byte's length is too small, fail:",
			" len(queryData):", len(queryData))
		return nil, 0, errors.New("Received bytes is too small for legal UPDATE format")
	}

	// header+count
	packetDecompressionLabel := packet.NewPacketDecompressionLabel()
	receiveQueryModel, err = NewQueryModelByHeaderAndCount(headerModel, countModel)
	if err != nil {
		belogs.Error("ParseBytesToQueryModel(): NewQueryModelByHeaderAndCount fail:",
			" headerModel:", jsonutil.MarshalJson(headerModel), " countModel:", jsonutil.MarshalJson(countModel), err)
		return nil, 0, err
	}
	belogs.Info("ParseBytesToQueryModel():NewQueryModelByHeaderAndCount, receiveQueryModel:", jsonutil.MarshalJson(*receiveQueryModel))

	// question
	belogs.Debug("ParseBytesToQueryModel(): will get question packetmodels,  offsetFromStart:", offsetFromStart)
	questionPacketModels, newOffsetFromStart, err := packet.ParseBytesToPacketModels(queryData, 0, offsetFromStart, packet.DNS_PACKET_NAME_TYPE_CLASS, 1, packetDecompressionLabel)
	if err != nil {
		belogs.Error("ParseBytesToQueryModel(): ParseBytesToPacketModels to questionPacketModels fail:",
			" queryData:", convert.PrintBytesOneLine(queryData), "  offsetFromStart:", offsetFromStart, err)
		return nil, 0, errors.New("Received bytes is failed to get questionPacketModels")
	}
	belogs.Debug("ParseBytesToQueryModel(): ParseBytesToPacketModels,len(questionPacketModels):", len(questionPacketModels),
		"  questionPacketModels:", jsonutil.MarshalJson(questionPacketModels), "  newOffsetFromStart:", newOffsetFromStart)
	qdCount := countModel.GetCount(0)
	if int(qdCount) != len(questionPacketModels) {
		belogs.Error("ParseBytesToQueryModel(): qdCount is not equal to len(questionPacketModels), fail:",
			" qdCount:", qdCount, "  len(questionPacketModels):", len(questionPacketModels))
		return nil, 0, errors.New("Received bytes is failed because dqCount is not equal to len(questionPacketModels)")
	}
	receiveQueryModel.QueryDataModel.QuestionModels = questionPacketModels
	for i := range receiveQueryModel.QueryDataModel.QuestionModels {
		questionModel := receiveQueryModel.QueryDataModel.QuestionModels[i]
		questionRrModel, _ := dnsconvert.ConvertPacketToRr("", questionModel)
		belogs.Info("#解析得到QUERY的'查询响应'类型数据包, Question部分: " + osutil.GetNewLineSep() +
			"{'域名':'" + questionRrModel.RrFullDomain +
			"','Type':'" + questionRrModel.RrType +
			"','Class':'" + questionRrModel.RrClass +
			"','Ttl':" + convert.ToString(questionRrModel.RrTtl.ValueOrZero()) + "}")
	}

	// should sub offsetFromStart,because queryData start from offsetFromStart
	belogs.Debug("ParseBytesToQueryModel(): will get answer packetmodels, newOffsetFromStart:", newOffsetFromStart,
		"  offsetFromStart:", offsetFromStart)
	queryData = queryData[newOffsetFromStart-offsetFromStart:]
	if len(queryData) > 0 {
		answerPacketModels, offsetFromStartTmp, err := packet.ParseBytesToPacketModels(queryData, 0, newOffsetFromStart, packet.DNS_PACKET_NAME_TYPE_CLASS_TTL_RDLENGTH_RDATA, -1, packetDecompressionLabel)
		if err != nil {
			belogs.Error("ParseBytesToQueryModel(): ParseBytesToPacketModels to answerPacketModels fail:",
				" queryData:", convert.PrintBytesOneLine(queryData), "  offsetFromStart:", offsetFromStart, err)
			return nil, 0, errors.New("Received bytes is failed to get answerPacketModels")
		}
		newOffsetFromStart = offsetFromStartTmp
		belogs.Debug("ParseBytesToQueryModel(): ParseBytesToPacketModels,len(answerPacketModels):", len(answerPacketModels),
			"  packetModels:", jsonutil.MarshalJson(answerPacketModels), "  newOffsetFromStart:", newOffsetFromStart)

		anCount := countModel.GetCount(1)
		nsCount := countModel.GetCount(2)
		adCount := countModel.GetCount(3)
		belogs.Debug("ParseBytesToQueryModel(): qdCount:", qdCount, "  anCount:", anCount, " nsCount:", nsCount, "  adCount:", adCount)

		if int(anCount+nsCount+adCount) != len(answerPacketModels) {
			belogs.Error("ParseBytesToQueryModel(): sum of anCount+nsCount+adCount is not equal to len(answerPacketModels),",
				" fail:", " anCount:", anCount, " nsCount:", nsCount, "  adCount:", adCount,
				"  len(answerPacketModels):", len(answerPacketModels))
			return nil, 0, errors.New("Received bytes is failed because count is not equal to len(answerPacketModels)")
		}

		receiveQueryModel.QueryDataModel.AnswerModels = answerPacketModels[:anCount]
		receiveQueryModel.QueryDataModel.AuthorityModels = answerPacketModels[anCount : anCount+nsCount]
		receiveQueryModel.QueryDataModel.AdditonalModels = answerPacketModels[anCount+nsCount : anCount+nsCount+adCount]

		for i := range receiveQueryModel.QueryDataModel.AnswerModels {
			answerModel := receiveQueryModel.QueryDataModel.AnswerModels[i]
			answerRrModel, _ := dnsconvert.ConvertPacketToRr("", answerModel)
			belogs.Info("#解析得到QUERY的'查询响应'类型数据包, Answer部分: " + osutil.GetNewLineSep() +
				"{'域名':'" + answerRrModel.RrFullDomain +
				"','Type':'" + answerRrModel.RrType +
				"','Class':'" + answerRrModel.RrClass +
				"','Ttl':" + convert.ToString(answerRrModel.RrTtl.ValueOrZero()) +
				",'Data':" + answerRrModel.RrData + "'}")
		}
		for i := range receiveQueryModel.QueryDataModel.AuthorityModels {
			authorityModel := receiveQueryModel.QueryDataModel.AuthorityModels[i]
			authorityRrModel, _ := dnsconvert.ConvertPacketToRr("", authorityModel)
			belogs.Info("#解析得到QUERY的'查询响应'类型数据包, Authority部分: " + osutil.GetNewLineSep() +
				"{'域名':'" + authorityRrModel.RrFullDomain +
				"','Type':'" + authorityRrModel.RrType +
				"','Class':'" + authorityRrModel.RrClass +
				"','Ttl':" + convert.ToString(authorityRrModel.RrTtl.ValueOrZero()) +
				",'Data':" + authorityRrModel.RrData + "'}")
		}
		for i := range receiveQueryModel.QueryDataModel.AdditonalModels {
			additonalModel := receiveQueryModel.QueryDataModel.AdditonalModels[i]
			additonalRrModel, _ := dnsconvert.ConvertPacketToRr("", additonalModel)
			belogs.Info("#解析得到QUERY的'查询响应'类型数据包, Additonal部分: " + osutil.GetNewLineSep() +
				"{'域名':'" + additonalRrModel.RrFullDomain +
				"','Type':'" + additonalRrModel.RrType +
				"','Class':'" + additonalRrModel.RrClass +
				"','Ttl':" + convert.ToString(additonalRrModel.RrTtl.ValueOrZero()) +
				",'Data':" + additonalRrModel.RrData + "'}")
		}

	}
	belogs.Debug("ParseBytesToQueryModel(): receiveQueryModel:", jsonutil.MarshalJson(receiveQueryModel), "  newOffsetFromStart:", newOffsetFromStart)
	return receiveQueryModel, newOffsetFromStart, nil
}

func (c *QueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, c.HeaderForQueryModel.Bytes())
	binary.Write(wr, binary.BigEndian, c.CountQANAModel.Bytes())

	for i := range c.QueryDataModel.QuestionModels {
		binary.Write(wr, binary.BigEndian, c.QueryDataModel.QuestionModels[i].Bytes())
	}
	for i := range c.QueryDataModel.AnswerModels {
		binary.Write(wr, binary.BigEndian, c.QueryDataModel.AnswerModels[i].Bytes())
	}
	for i := range c.QueryDataModel.AuthorityModels {
		binary.Write(wr, binary.BigEndian, c.QueryDataModel.AuthorityModels[i].Bytes())
	}
	for i := range c.QueryDataModel.AdditonalModels {
		binary.Write(wr, binary.BigEndian, c.QueryDataModel.AdditonalModels[i].Bytes())
	}
	return wr.Bytes()
}
