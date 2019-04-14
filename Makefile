
DIR := ${CURDIR}

.PHONY: run
run: script
	GO111MODULE=on go build -o leader-election-delay main.go election.go

	TEST_ASSET_KUBE_APISERVER=$(DIR)/assets/bin/kube-apiserver \
	TEST_ASSET_ETCD=$(DIR)/assets/bin/etcd \
	TEST_ASSET_KUBECTL=$(DIR)/assets/bin/kubectl \
	TEST_ASSET_TOXIPROXY=$(DIR)/assets/bin/toxiproxy-server \
	./leader-election-delay -v=10

.PHONY: script
script:
	$(DIR)/scripts/download-binaries.sh $(DIR)
