// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"strings"

	"go.etcd.io/raft/v3/raftpb"
)

func main() {
	// 开始解析参数   集群地址
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	// 节点ID
	id := flag.Int("id", 1, "node ID")
	// 保存数据的 服务器 端口
	kvport := flag.Int("port", 9121, "key-value server port")
	// 是否新的机器加入 ，而不是新的集群
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	// proposeC 信息
	proposeC := make(chan string)
	defer close(proposeC)
	// confChangeC 监听是否配置修改
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// kvs 初始化保存的类型
	var kvs *kvstore
	// 获得快照数据
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	// newRaftNode 获得一个 新的 raft节点
	commitC, errorC, snapshotterReady := newRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	serveHTTPKVAPI(kvs, *kvport, confChangeC, errorC)
}
