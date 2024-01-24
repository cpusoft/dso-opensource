package drive

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestConvertChainRepossToRrDependentModels(t *testing.T) {
	chainRepos := make([]ChainRepos, 0)
	json := `[{"repos":["rrdp.apnic.net","sakuya.nat.moe"]},{"repos":["rrdp.arin.net","repo.kagl.me"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app"]},{"repos":["rrdp.arin.net","cloudie.rpki.app"]},{"repos":["rrdp.apnic.net","rpki-repository.nic.ad.jp"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net"]},{"repos":["rrdp.ripe.net","rpki1.rpki-test.sit.fraunhofer.de"]},{"repos":["rrdp.apnic.net"]},{"repos":["rrdp.lacnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rrdp.krill.cloud"]},{"repos":["rrdp.afrinic.net"]},{"repos":["rrdp.ripe.net"]},{"repos":["rrdp.apnic.net","rpki.cnnic.cn"]},{"repos":["rrdp.ripe.net","ca.rg.net"]},{"repos":["rrdp.apnic.net","rrdp.twnic.tw"]},{"repos":["rrdp.arin.net","dev.tw"]},{"repos":["rrdp.ripe.net","dev.tw"]},{"repos":["rrdp.arin.net","repo.kagl.me","dev.tw"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","dev.tw"]},{"repos":["rrdp.apnic.net","rrdp-rps.arin.net","dev.tw"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net"]},{"repos":["rrdp.apnic.net","rrdp-rps.arin.net"]},{"repos":["rrdp.arin.net"]},{"repos":["rrdp.lacnic.net","rpki-repo.registro.br"]},{"repos":["rrdp-as0.apnic.net"]},{"repos":["rrdp.apnic.net","repo-rpki.idnic.net"]},{"repos":["rrdp.arin.net","rpki.tools.westconnect.ca"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","rpki.qs.nu"]},{"repos":["rrdp.arin.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.ripe.net","rrdp.rp.ki"]},{"repos":["rrdp.ripe.net","repo.kagl.me"]},{"repos":["rrdp.arin.net","ca.nat.moe"]},{"repos":["rrdp.apnic.net","rpki.rand.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.pedjoeang.group"]},{"repos":["rrdp.apnic.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.ripe.net","chloe.sobornost.net"]},{"repos":["chloe.sobornost.net"]},{"repos":["rrdp.ripe.net","rpki-rrdp.us-east-2.amazonaws.com"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rpki.cc"]},{"repos":["rrdp.sub.apnic.net"]},{"repos":["rrdp.apnic.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rpki.komorebi.network:3030"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki.komorebi.network:3030"]},{"repos":["rrdp.arin.net","rrdp.roa.tohunet.com"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki-publication.haruue.net"]},{"repos":["rrdp.apnic.net","rpki.akrn.net"]},{"repos":["rrdp.arin.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rpki-01.pdxnet.uk"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rpki-01.pdxnet.uk"]},{"repos":["rrdp.ripe.net","rpki.pedjoeang.group"]},{"repos":["rrdp.arin.net","rpki.qs.nu"]},{"repos":["rrdp.apnic.net","rrdp.rp.ki"]},{"repos":["rrdp.ripe.net","rpki.owl.net"]},{"repos":["sakuya.nat.moe","rpki-rrdp.mnihyc.com"]},{"repos":["rrdp.ripe.net","rpki.folf.systems"]},{"repos":["rrdp.apnic.net","rpki.roa.net"]},{"repos":["rrdp.arin.net","rpki.roa.net"]},{"repos":["rrdp.ripe.net","rrdp.rpki.tianhai.link"]},{"repos":["rrdp.apnic.net","rrdp.rpki.tianhai.link"]},{"repos":["rrdp.arin.net","rpki.multacom.com"]},{"repos":["rrdp.ripe.net","rpki.roa.net"]},{"repos":["sakuya.nat.moe","rpki.apernet.io"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki.apernet.io"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","magellan.ipxo.com"]},{"repos":["rrdp.ripe.net","x-0100000000000011.p.u9sv.com"]},{"repos":["rrdp.ripe.net","rrdp.paas.rpki.ripe.net","rrdp-rps.arin.net"]},{"repos":["rrdp.ripe.net","cloudie.rpki.app","rrdp-rps.arin.net"]},{"repos":["rrdp.arin.net","rpki.akrn.net"]},{"repos":["rrdp.ripe.net","rpki.multacom.com"]},{"repos":["rrdp.apnic.net","rpki.owl.net"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.pedjoeang.group"]},{"repos":["rrdp.ripe.net","rrdp.krill.cloud","rov-measurements.nlnetlabs.net"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki.luys.cloud"]},{"repos":["sakuya.nat.moe","rpki.luys.cloud"]},{"repos":["rrdp.ripe.net","0.sb"]},{"repos":["rrdp.arin.net","rrdp.rp.ki"]},{"repos":["rrdp.apnic.net","0.sb"]},{"repos":["rrdp.apnic.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.apnic.net","sakuya.nat.moe","rpki-rrdp.mnihyc.com"]},{"repos":["rrdp.arin.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.ripe.net","rpki.admin.freerangecloud.com"]},{"repos":["rrdp.ripe.net","rpki.zappiehost.com"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.zappiehost.com"]},{"repos":["rrdp.ripe.net","rpki-01.pdxnet.uk"]},{"repos":["rrdp.arin.net","rrdp.sub.apnic.net"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","rpki.cc"]},{"repos":["rrdp.ripe.net","rrdp-rps.arin.net","cloudie.rpki.app","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","rrdp-rps.arin.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.ripe.net","ca.rg.net","rrdp.paas.rpki.ripe.net"]},{"repos":["rrdp.arin.net","cloudie.rpki.app","rrdp.paas.rpki.ripe.net"]},{"repos":["rpki-rrdp.us-east-2.amazonaws.com"]}]`
	err := jsonutil.UnmarshalJson(json, &chainRepos)
	fmt.Println(chainRepos, err)
	if err != nil {
		return
	}
	rrDependentModels, err := convertChainRepossToRrDependentModels(chainRepos)
	fmt.Println(rrDependentModels, err)
}