REPO=docker.io/yangl

all: build-images
	pwd && cd apiserver-watcher && ko apply -f ./deploy-local.yaml

build-images: 
	cd apiserver-watcher/python && docker build -t $(REPO)/apiserver-watcher-py .
	docker push $(REPO)/apiserver-watcher-py
