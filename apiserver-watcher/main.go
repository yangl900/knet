package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const triggerName = "trigger"

func main() {
	notify := make(chan bool, 1000)
	go watchChanges(notify)
	makeChanges(notify)
}

func makeChanges(notify chan bool) {
	interval := time.Minute * time.Duration(10)
	if ci, ok := os.LookupEnv("CHANGE_INTERVAL_SECONDS"); ok {
		info("Defined env interval: %s", ci)
		if i, err := strconv.Atoi(ci); err == nil {
			info("Use interval in seconds: %d", i)
			interval = time.Second * time.Duration(i)
		}
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ns, err := currentNamespace()
	if err != nil {
		panic(err.Error())
	}

	for {
		for len(notify) > 0 {
			<-notify
		}

		cm, err := clientset.CoreV1().ConfigMaps(ns).Get(context.TODO(), triggerName, metav1.GetOptions{})
		if err != nil {
			info("Failed to get config map %s: %s", triggerName, err)
		}
		cm.SetLabels(map[string]string{"updated": fmt.Sprint(time.Now().Unix())})

		_, err = clientset.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			info("Failed to update config map %s: %s", triggerName, err)
		}

		info("Updated config map %s", triggerName)

		select {
		case <-time.After(time.Second * time.Duration(10)):
			info("Did not receive watch changes in 10s!! FAILED.")
			recoardFailure(clientset, ns)
		case <-notify:
			info("Received watch changes. TEST PASSED.")
		}
		time.Sleep(interval)
	}
}

func recoardFailure(clientset *kubernetes.Clientset, ns string) {
	fn := fmt.Sprintf("faliure-%d", time.Now().Unix())
	fcm := v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fn}, Data: map[string]string{"failed-time": fmt.Sprint(time.Now().UTC())}}
	_, err := clientset.CoreV1().ConfigMaps(ns).Create(context.TODO(), &fcm, metav1.CreateOptions{})
	if err != nil {
		info("Failed to record failure config map: %s", err)
	}
}

func watchChanges(notify chan bool) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ns, err := currentNamespace()
	if err != nil {
		panic(err.Error())
	}

	info("Starting watch in namespace %s", ns)

	for {
		watch, err := clientset.CoreV1().ConfigMaps(ns).Watch(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for event := range watch.ResultChan() {
			c, ok := event.Object.(*v1.ConfigMap)
			if !ok {
				fmt.Println("unexpected type")
			}
			info("%s - %s - version %s", event.Type, c.Name, c.ResourceVersion)

			if c.Name == triggerName {
				notify <- true
			}
		}
	}
}

func info(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

func currentNamespace() (string, error) {
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", fmt.Errorf("failed to read serviceaccount namespace: %w", err)
	}

	return string(b), nil
}
