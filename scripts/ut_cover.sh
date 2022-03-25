#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
function ut_cover() {
  cd ${cm}/$1
  echo "cd ${cm}/$1"
  go test -coverprofile cover.out ./...
  total=$(go tool cover -func=cover.out | tail -1)
  echo ${total}
  rm cover.out
  coverage=$(echo ${total} | grep -P '\d+\.\d+(?=\%)' -o) #如果macOS 不支持grep -P选项，可以通过brew install grep更新grep
  #计算注释覆盖率，需要安装gocloc： go install github.com/hhatto/gocloc/cmd/gocloc@latest
  comment_coverage=$(gocloc --include-lang=Go --output-type=json . | jq '(.total.comment-.total.files*6)/(.total.code+.total.comment)*100')
  echo "注释率：${comment_coverage}%"

  # 如果测试覆盖率低于N，认为ut执行失败
  (( $(awk "BEGIN {print (${coverage} >= $2)}") )) || (echo "$1 单测覆盖率: ${coverage} 低于 $2%"; exit 1)
  (( $(awk "BEGIN {print (${comment_coverage} >= $3)}") )) || (echo "$1 注释覆盖率: ${comment_coverage} 低于 $3%"; exit 1)
  if test -z "$GIT_COMMITTER"; then
      echo "no committer, ignore sql insert"
  else
    echo "insert into ut_result (committer, commit_id, branch, module_name, cover_rate, ci_pass,comment_rate) values ('${GIT_COMMITTER}','${GIT_COMMIT}','${GIT_BRACH_NAME}','${1}','${coverage}',true,${comment_coverage})" | mysql -h192.168.1.121 -P3316 -uroot -p${JENKINS_MYSQL_PWD} -Dchainmaker_test
  fi

}
set -e

cm=$(pwd)

if [[ $cm == *"scripts" ]] ;then
  cm=$cm/..
fi

if [ -n "$1" ] ;then
  echo "check UT cover: $1."
  ut_cover "$1" 40 10
else
   ut_cover "module/accesscontrol" 47 2.3
#  ut_cover "module/blockchain" 2
#  ut_cover "module/conf/chainconf" 26
#  ut_cover "module/conf/localconf" 11
  ut_cover "module/consensus" 30 1
  #ut_cover "module/consensus/dpos" 50
  #ut_cover "module/consensus/raft" 0
#  ut_cover "module/consensus/solo" 0
  #ut_cover "module/consensus/tbft" 10
  ut_cover "module/core" 2.3 23
  ut_cover "module/net" 29 11
  ut_cover "module/rpcserver" 0 3.1
  ut_cover "module/snapshot" 47 7
  ut_cover "module/sync" 71 4.3
  ut_cover "module/subscriber" 70 0.9
  ut_cover "tools/cmc" 10 2.3
fi
