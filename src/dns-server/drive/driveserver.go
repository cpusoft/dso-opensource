package drive

import (
	"errors"
	"time"

	pushmodel "dns-model/push"
	"dns-model/rr"
	dnsserver "dns-server/dns"
	dsomodel "dso-core/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
)

func activePushAll() {
	belogs.Debug("activePushAll()")
	start := time.Now()
	path := "https://" + conf.String("dns-server::serverHost") + ":" + conf.String("dns-server::serverHttpsPort") +
		"/push/activepushall"
	pushResultRrModels := make([]*pushmodel.PushResultRrModel, 0)
	belogs.Debug("activePushAll(): path:", path)
	err := httpclient.PostAndUnmarshalResponseModel(path, "", false, &pushResultRrModels)
	if err != nil {
		belogs.Error("activePushAll(): httpclient/push/activepushall, fail:", path, err)
		return
	}
	if len(pushResultRrModels) == 0 {
		belogs.Debug("activePushAll():httpclient/push/subscribe have no results, path:", path,
			" time(s):", time.Since(start))
		return
	}
	belogs.Debug("activePushAll(): pushResultRrModels:", jsonutil.MarshalJson(pushResultRrModels))

	for i := range pushResultRrModels {
		connKey := pushResultRrModels[i].ConnKey
		rrModels := pushResultRrModels[i].RrModels
		dnsModel, err := dsomodel.NewDsoModelWithPushTlvModel(rrModels)
		if err != nil {
			belogs.Error("activePushAll(), NewDsoModelWithPushTlvModel fail, rrModels:", jsonutil.MarshalJson(rrModels), err)
			continue
		}
		belogs.Debug("activePushAll(): dnsModel:", jsonutil.MarshalJson(dnsModel))

		err = dnsserver.SendTcpDnsModel(connKey, dnsModel)
		if err != nil {
			belogs.Error("activePushAll(): SendDsoModel fail: connKey:", connKey, "  dnsModel:", jsonutil.MarshalJson(dnsModel), err)
			continue
		}
		belogs.Info("activePushAll(): SendDsoModel ok, connKey:", connKey, " dnsModel:", jsonutil.MarshalJson(dnsModel),
			"  time(s):", time.Since(start))
	}
}

func queryServerDnsRrs(serverDnsRrModel *ServerDnsRrModel) (resultDnsRrs []*rr.RrModel, err error) {
	belogs.Debug("queryServerDnsRrs(): serverDnsRrModel:", jsonutil.MarshalJson(serverDnsRrModel))
	if len(serverDnsRrModel.RrModels) != 1 {
		belogs.Error("queryServerDnsRrs(): len(serverDnsRrModel.RrModels) isnot 1, serverDnsRrModel:", jsonutil.MarshalJson(serverDnsRrModel))
		return nil, errors.New("len(serverDnsRrModel.RrModels) isnot one")
	}
	return queryRrModelsDb(serverDnsRrModel.RrModels[0])
}

func queryServerAllDnsRrs() (resultDnsRrs []*rr.RrModel, err error) {
	belogs.Debug("queryServerAllDnsRrs():")
	return queryAllRrModelsDb()
}

func queryRpkiRepos() {
	belogs.Debug("queryRpkiRepos():")
	start := time.Now()

	chainReposs := make([]ChainRepos, 0)
	path := "https://" + conf.String("rpstir2-rp::serverHost") + ":" + conf.String("rpstir2-rp::serverHttpsPort") +
		"/sys/exportchainrepos"
	belogs.Debug("queryRpkiRepos(): path:", path)
	err := httpclient.PostAndUnmarshalStruct(path, "", false, &chainReposs)
	if err != nil {
		belogs.Error("queryRpkiRepos(): httpclient/sys/exportchainrepos, fail:", path, err)
		return
	}
	if len(chainReposs) == 0 {
		belogs.Error("queryRpkiRepos(): httpclient/sys/exportchainrepos, chainReposs.Repos is empty, path:", path,
			" time(s):", time.Since(start))
		return
	}
	belogs.Debug("queryRpkiRepos(): chainReposs:", jsonutil.MarshalJson(chainReposs), " time(s):", time.Since(start))
	// `[{"repos":["rrdp.apnic.net","sakuya.nat.moe"]},{"repos":["rrdp.arin.net","repo.kagl.me"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app"]},{"repos":["rrdp.arin.net","cloudie.rpki.app"]},{"repos":["rrdp.apnic.net","rpki-repository.nic.ad.jp"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net"]},{"repos":["rrdp.ripe.net","rpki1.rpki-test.sit.fraunhofer.de"]},{"repos":["rrdp.apnic.net"]},{"repos":["rrdp.lacnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rrdp.krill.cloud"]},{"repos":["rrdp.afrinic.net"]},{"repos":["rrdp.ripe.net"]},{"repos":["rrdp.apnic.net","rpki.cnnic.cn"]},{"repos":["rrdp.ripe.net","ca.rg.net"]},{"repos":["rrdp.apnic.net","rrdp.twnic.tw"]},{"repos":["rrdp.arin.net","dev.tw"]},{"repos":["rrdp.ripe.net","dev.tw"]},{"repos":["rrdp.arin.net","repo.kagl.me","dev.tw"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","dev.tw"]},{"repos":["rrdp.apnic.net","rrdp-rps.arin.net","dev.tw"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net"]},{"repos":["rrdp.apnic.net","rrdp-rps.arin.net"]},{"repos":["rrdp.arin.net"]},{"repos":["rrdp.lacnic.net","rpki-repo.registro.br"]},{"repos":["rrdp-as0.apnic.net"]},{"repos":["rrdp.apnic.net","repo-rpki.idnic.net"]},{"repos":["rrdp.arin.net","rpki.tools.westconnect.ca"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","rpki.qs.nu"]},{"repos":["rrdp.arin.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.ripe.net","rrdp.rp.ki"]},{"repos":["rrdp.ripe.net","repo.kagl.me"]},{"repos":["rrdp.arin.net","ca.nat.moe"]},{"repos":["rrdp.apnic.net","rpki.rand.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.pedjoeang.group"]},{"repos":["rrdp.apnic.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.ripe.net","chloe.sobornost.net"]},{"repos":["chloe.sobornost.net"]},{"repos":["rrdp.ripe.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rpki.cc"]},{"repos":["rrdp.sub.apnic.net"]},{"repos":["rrdp.apnic.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.komorebi.network:3030"]},{"repos":["rrdp.arin.net","rrdp.roa.tohunet.com"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki-publication.haruue.net"]},{"repos":["rrdp.apnic.net","rpki.akrn.net"]},{"repos":["rrdp.arin.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki-01.pdxnet.uk"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rpki.pedjoeang.group"]},{"repos":["rrdp.arin.net","rpki.qs.nu"]},{"repos":["rrdp.apnic.net","rrdp.rp.ki"]},{"repos":["rrdp.ripe.net","rpki.owl.net"]},{"repos":["sakuya.nat.moe","rpki-rrdp.mnihyc.com"]},{"repos":["rrdp.ripe.net","rpki.folf.systems"]},{"repos":["rrdp.apnic.net","rpki.roa.net"]},{"repos":["rrdp.arin.net","rpki.roa.net"]},{"repos":["rrdp.ripe.net","rrdp.rpki.tianhai.link"]},{"repos":["rrdp.apnic.net","rrdp.rpki.tianhai.link"]},{"repos":["rrdp.arin.net","rpki.multacom.com"]},{"repos":["rrdp.ripe.net","rpki.roa.net"]},{"repos":["sakuya.nat.moe","rpki.apernet.io"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki.apernet.io"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","magellan.ipxo.com"]},{"repos":["rrdp.ripe.net","x-0100000000000011.p.u9sv.com"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","rrdp-rps.arin.net"]},{"repos":["rrdp.ripe.net","cloudie.rpki.app","rrdp-rps.arin.net"]},{"repos":["rrdp.arin.net","rpki.akrn.net"]},{"repos":["rrdp.ripe.net","rpki.multacom.com"]},{"repos":["rrdp.apnic.net","rpki.owl.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.pedjoeang.group"]},{"repos":["rrdp.ripe.net","rrdp.krill.cloud","rov-measurements.nlnetlabs.net"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki.luys.cloud"]},{"repos":["sakuya.nat.moe","rpki.luys.cloud"]},{"repos":["rrdp.ripe.net","0.sb"]},{"repos":["rrdp.arin.net","rrdp.rp.ki"]},{"repos":["rrdp.apnic.net","0.sb"]},{"repos":["rrdp.apnic.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki-rrdp.mnihyc.com"]},{"repos":["rrdp.arin.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.ripe.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.ripe.net","rpki.zappiehost.com"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.zappiehost.com"]},{"repos":["rrdp.ripe.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.arin.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.ripe.net","ca.rg.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rrdp.paas.rpki.ripe.net"]},{"repos":["rpki-rrdp.us-east-2.amazonaws.com"]}]`

	rrDependentModels, err := convertChainRepossToRrDependentModels(chainReposs)
	if err != nil {
		belogs.Error("queryRpkiRepos(): convertChainRepossToRrDependentModels, fail:", err)
		return
	}

	err = saveRrDependentModelsDb(rrDependentModels)
	if err != nil {
		belogs.Error("queryRpkiRepos(): saveRrDependentModelsDb, fail:", err)
		return
	}
	belogs.Info("queryRpkiRepos(): saveRrDependentModelsDb, rrDependentModels:", jsonutil.MarshalJson(rrDependentModels), " time(s):", time.Since(start))
	return
}

func convertChainRepossToRrDependentModels(chainReposs []ChainRepos) (rrDependentModels []*RrDependentModel, err error) {
	belogs.Debug("convertChainRepossToRrDependentModels(): len(chainReposs):", len(chainReposs))

	rrDependentModelMap := make(map[string]*RrDependentModel, 0)
	rrDependentModels = make([]*RrDependentModel, 0)
	for _, chainRepos := range chainReposs {
		for _, repo := range chainRepos.Repos {
			_, ok := rrDependentModelMap[repo]
			if !ok {
				id, err := queryByRrFullDomain(repo)
				if err != nil || id == 0 {
					belogs.Debug("convertChainRepossToRrDependentModels(): not found id by rrFullDomain:", repo, err)
					continue
				}
				rrDependentModelTmp := &RrDependentModel{}
				rrDependentModelTmp.Id = id
				rrDependentModelTmp.RrFullDomain = repo
				rrDependentModelTmp.ChildIds = make([]uint64, 0)

				belogs.Debug("convertChainRepossToRrDependentModels(): first range chainRepos.Repos, repo:", repo,
					" rrDependentModelTmp:", jsonutil.MarshalJson(rrDependentModelTmp))
				rrDependentModelMap[repo] = rrDependentModelTmp
				rrDependentModels = append(rrDependentModels, rrDependentModelTmp)
			}
		}
	}
	belogs.Debug("convertChainRepossToRrDependentModels(): rrDependentModelMap:", jsonutil.MarshalJson(rrDependentModelMap))

	for _, chainRepos := range chainReposs {
		topRepo := chainRepos.Repos[0]
		topRepoMap, _ := rrDependentModelMap[topRepo]
		belogs.Debug("convertChainRepossToRrDependentModels(): before add subRepo, topRepo:", topRepo,
			"  topRepoMap:", jsonutil.MarshalJson(topRepoMap), "  len(chainRepos.Repos):", len(chainRepos.Repos))

		for i := 1; i < len(chainRepos.Repos); i++ {
			subRepo := chainRepos.Repos[i]
			subRepoMap, ok := rrDependentModelMap[subRepo]
			if !ok {
				belogs.Error("queryRpkiRepos(): second range chainRepos.Repos, subRepo not found subRepo:", subRepo)
				continue
			}
			belogs.Debug("convertChainRepossToRrDependentModels(): subRepo:", subRepo,
				" subRepoMap.Id:", subRepoMap.Id, "  topRepo:", topRepo)
			topRepoMap.ChildIds = append(topRepoMap.ChildIds, subRepoMap.Id)
		}
		belogs.Debug("convertChainRepossToRrDependentModels(): after add subRepo, topRepo:", topRepo,
			"  topRepoMap:", jsonutil.MarshalJson(topRepoMap))

	}
	belogs.Info("convertChainRepossToRrDependentModels(): rrDependentModels:", jsonutil.MarshalJson(rrDependentModels))
	return rrDependentModels, nil
}
