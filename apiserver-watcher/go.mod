module github.com/yangl900/knet/apiserver-watcher

go 1.15

require (
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
	k8s.io/utils v0.0.0-20201015054608-420da100c033 // indirect
)

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.2
