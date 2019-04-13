
DIR := ${CURDIR}

.PHONY: run
run: script
	TEST_ASSET_KUBE_APISERVER=$(DIR)/assets/bin/kube-apiserver \
	TEST_ASSET_ETCD=$(DIR)/assets/bin/etcd \
	TEST_ASSET_KUBECTL=$(DIR)/assets/bin/kubectl \
	TEST_ASSET_TOXIPROXY=$(DIR)/assets/bin/toxiproxy-server \
	GO111MODULE=on go run main.go election.go

.PHONY: script
script:
	$(DIR)/scripts/download-binaries.sh $(DIR)
