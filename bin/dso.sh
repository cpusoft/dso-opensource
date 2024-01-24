#!/bin/bash


curpath=$(pwd)
source ./read-conf.sh
echo "curpath:" $curpath

export CGO_ENABLED=1
configFile="../conf/project.conf"

serverHost=`ReadINIfile "$configFile" "dns-server" "serverHost" `
serverHttpsPort=`ReadINIfile "$configFile" "dns-server" "serverHttpsPort" `
serverHttpPort=`ReadINIfile "$configFile" "dns-server" "serverHttpPort" `
clientHost=`ReadINIfile "$configFile" "dns-client" "clientHost" `
clientHttpsPort=`ReadINIfile "$configFile" "dns-client" "clientHttpsPort" `
clientHttpPort=`ReadINIfile "$configFile" "dns-client" "clientHttpPort" `

echo "server:" $serverHost:$serverHttpsPort
echo "client:" $clientHost:$clientHttpsPort

function buildFunc()
{
  goutilversion="c54d99a"
  echo "github.com/cpusoft/goutil@"$goutilversion
  echo "buildFunc" $@

  dns_program_name="dns-"$2
  dso_program_name="dso-"$2
  program_dir=$(ReadINIfile $configFile $dns_program_name programDir)
  echo "program_dir:" $program_dir

  # git update
  # save local project.conf
  oldConfigFile=$(date +%Y%m%d%H%M%S)
  echo "save current conf/project.conf to conf/project.conf.$oldConfigFile.bak"
  cp ../conf/project.conf ../conf/project.conf.$oldConfigFile.bak
  
  # go mod
  echo "go update"
  chmod +x *
  cd ../src
  go get -u github.com/cpusoft/goutil@$goutilversion
  go mod tidy

  # go install: go tool compile -help
  echo "go install linux"
  export CGO_ENABLED=1
  export GOOS=linux
  export GOARCH=amd64
  go build -v -ldflags "-X main.CompileParamStr=$dso_program_name" . 
  chmod +x dns-dso 
  mv ./dns-dso $program_dir/bin/$dso_program_name
  chmod +x $program_dir/bin/$dso_program_name

  cd $program_dir/bin/
  echo -e "build $dso_program_name compete\n"
  return 0
}


function startServerFunc()
{
  nohup ./dso-server  >> ../log/servernohup.log 2>&1 &
  sleep 2
  pidhttp=`ps -ef|grep 'dso-server'|grep -v grep|grep -v 'dns.sh' |awk '{print $2}'`
  if [ "$pidhttp" = ""  ]; then
    echo "Start failed"
    echo -e "\nYou can check the failure reason through the log file in ../log/.\n"
  else
    echo "Start successful"
  fi
  return 0
}

function stopServerFunc()
{
  pidhttp=`ps -ef|grep 'dso-server'|grep -v grep|grep -v 'dns.sh' |awk '{print $2}'`
  echo "The current dso-server process id is $pidhttp"
  for pid in $pidhttp
  do
    if [ "$pid" = "" ]; then
      echo "dso-server is not running"
    else
      kill  $pid
      echo "shutdown dso-server success"
 	fi
  done
  return 0
}



function startClientFunc()
{
  nohup ./dso-client  >> ../log/clientnohup.log 2>&1 &
  sleep 2
  pidhttp=`ps -ef|grep 'dso-client'|grep -v grep|grep -v 'dns.sh' |awk '{print $2}'`
  if [ ! $pidhttp ]; then
    echo "Start failed"
    echo -e "\nYou can check the failure reason through the log file in ../log/.\n"
  else
    echo "Start successful"
  fi
  return 0
}

function stopClientFunc()
{
  pidhttp=`ps -ef|grep 'dso-client'|grep -v grep|grep -v 'dns.sh' |awk '{print $2}'`
  echo "The current dso-client process id is $pidhttp"
  for pid in $pidhttp
  do
    if [ "$pid" = "" ]; then
      echo "dso-client is not running"
    else
      kill  $pid
      echo "shutdown dso-client success"
 	fi
  done
  return 0
}

function checkFile()
{
    if [ $# != 1 ] ; then
        echo "file is empty"
        exit 1;
    fi

    if [ ! -f $1 ]; then
        echo "$1 does not exist"
        exit 1;
    fi
}


function helpFunc()
{
    echo "./dso.sh help:"
    echo -e ""
    
    ### for build
    echo -e "for build"
    echo -e "./dso.sh build server\t\tupdate dso-server"     
    echo -e "./dso.sh build client\t\tupdate dso-client"
    echo -e ""

    ### for dso-server
    echo -e "for dso-server"
    echo -e "./dso.sh startserver\t\t\tstart dso-server"
    echo -e "./dso.sh stopserver\t\t\tstop dso-server"
    echo -e "./dso.sh getallsubscribedrrs \t\tget all current subscribed dns rrs in dso-server"
    echo -e ""

    ### for dso-client
    echo -e "for dso-client"
    echo -e "./dso.sh startclient\t\t\tstart dso-client"
    echo -e "./dso.sh stopclient\t\t\tstop dso-client"  
    echo -e "./dso.sh creatednsconnect\t\tcreate connection to dso-server"
    echo -e "./dso.sh closednsconnect\t\tclose connection to dso-server"
    echo -e "./dso.sh startkeepalive\t\t\tcreate dso session between dns-lient and dns-erver"
    echo -e "./dso.sh subscribednsrr @filename\tsubscribe the domainname to dso-server"
    echo -e "./dso.sh unsubscribednsrr @filename\tunsubscribe the domainname to dso-server"
    echo -e "./dso.sh adddnsrrs @filename\t\tadd dns rrs to dso-server"
    echo -e "./dso.sh deldnsrrs @filename\t\tdel dns rrs to dso-server"
    echo -e "./dso.sh queryclientdnsrrss @filename\tquery dns rrs from local dso-client cache"
    echo -e "./dso.sh queryclientalldnsrrs\t\tquery all dns rrs from local dso-client cache"
    echo -e "./dso.sh clearclientalldnsrrs\t\tclear all dns rrs in local dso-client cache"
    echo -e ""

    ### for help
    echo -e "for help"
    echo -e "./dso.sh\t\t\t\tshow this help"
}



case $1 in
  build)
    echo "update dso-server or dso-client"
    buildFunc $@
    ;; 
  startserver)
    echo "start dso-server"
    startServerFunc
    ;;
  stopserver)
    echo "stop dso-server"
    stopServerFunc
    ;;
  getallsubscribedrrs)
    echo "getallsubscribedrrs"
    #curl -s -k -d ''  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/push/getallsubscribedrrs
    curl -k -d ''  -H "Content-type: application/json" -X POST https://$serverHost:$serverHttpsPort/push/getallsubscribedrrs
    echo -e "\n"
    ;;
  startclient)
    echo "start dso-client"
    startClientFunc
    ;;
  stopclient)
    echo "stop dso-client"
    stopClientFunc
    ;;
  creatednsconnect)
    curl -s -k -d ''  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/creatednsconnect
    echo -e "\n"
    ;;
  closednsconnect)
    curl -s -k -d ''  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/closednsconnect
    echo -e "\n"
    ;;  
  startkeepalive)
    curl -s -k -d '{"inactivityTimeout":1200,"keepaliveInterval":1200}'  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/startkeepalive
    echo -e "\n"
    ;; 
  adddnsrrs)
    echo "dns.sh adddnsrrs ../data/adddnsrrs_example.json"
    checkFile $2
    json=$(cat $2)
    echo $json
    curl -s -k -d $json  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/adddnsrrs
    echo -e "\n"
    ;;
  deldnsrrs)
    echo "dns.sh deldnsrrs ../data/deldnsrrs_class_none_example.json: will delete the values that match the domain name, type and data"
    echo "dns.sh deldnsrrs ../data/deldnsrrs_class_any_example.json: will delete the values that match the domain name and type"
    echo "dns.sh deldnsrrs ../data/deldnsrrs_class_any_type_any_example.json: will delete the values that match the domain name"
    echo "ps: type should not be SOA or CNAME temporarily"
    checkFile $2
    json=$(cat $2)
    echo $json
    curl -s -k -d $json  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/deldnsrrs
    echo -e "\n"
    ;;
  subscribednsrr)
    echo "dns.sh subscribednsrr ../data/subscribednsrr_example.json"
    checkFile $2
    json=$(cat $2)
    echo $json
    curl -s -k -d $json  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/subscribednsrr
    echo -e "\n"
    ;;
  unsubscribednsrr)
    echo "dns.sh unsubscribednsrr ../data/unsubscribednsrr_example.json"
    checkFile $2
    json=$(cat $2)
    echo $json
    curl -s -k -d $json  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/unsubscribednsrr
    echo -e "\n"
    ;;
  queryclientdnsrrs)
    echo "dns.sh queryclientdnsrrs ../data/queryclientdnsrrs_example.json"
    checkFile $2
    json=$(cat $2)
    echo $json
    curl -s -k -d $json  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/queryclientdnsrrs
    echo -e "\n"
    ;;
  queryclientalldnsrrs)
    curl -s -k -d ''  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/queryclientalldnsrrs
    echo -e "\n"
    ;; 
  clearclientalldnsrrs)
    curl -s -k -d ''  -H "Content-type: application/json" -X POST https://$clientHost:$clientHttpsPort/client/clearclientalldnsrrs
    echo -e "\n"
    ;;     
  help)
    helpFunc
    ;;      
  *)
    helpFunc
    ;;
esac
echo -e "\n"
