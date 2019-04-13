package main

import (
	"flag"
	"net/url"
	"os"
	"fmt"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"math/rand"

	"github.com/Shopify/toxiproxy/client"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/testing_frameworks/integration"
)

func main() {
	klog.InitFlags(nil)

	var (
		flags = pflag.NewFlagSet("", pflag.ExitOnError)
	)

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	apiServerURL, err := url.Parse("http://0.0.0.0:8080")
	if err != nil {
		klog.Fatal(err)
	}

	proxyAURL, err := url.Parse("http://0.0.0.0:8081")
	if err != nil {
		klog.Fatal(err)
	}

	proxyBURL, err := url.Parse("http://0.0.0.0:8082")
	if err != nil {
		klog.Fatal(err)
	}

	cp := &integration.ControlPlane{
		APIServer: &integration.APIServer{
			URL: apiServerURL,
			Out: os.Stdout,
			Err: os.Stderr,
		},
	}

	err = cp.Start()
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof("Starting toxyproxy server")
	go startToxyproxyServer()
	time.Sleep(5 * time.Second)

	klog.Infof("Starting toxyproxy client")
	tp := toxiproxy.NewClient("localhost:8474")

	proxyA, err := tp.CreateProxy("a", proxyAURL.Host, cp.APIServer.URL.Host)
	if err != nil {
		if err != nil {
			klog.Fatal(err)
		}
	}
	proxyA.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": 3000,
	})

	proxyB, err := tp.CreateProxy("b", proxyBURL.Host, cp.APIServer.URL.Host)
	if err != nil {
		if err != nil {
			klog.Fatal(err)
		}
	}
	proxyB.AddToxic("latency_down", "latency", "downstream", 1.0, toxiproxy.Attributes{
		"latency": 5000,
	})

	ns := "default"
	eID := "election-id-string"

	client, err := createApiserverClient(proxyAURL.Host, "")
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof("Starting leader election a")
	go newLeaderElection("pod-a", ns, eID, client)

	client, err = createApiserverClient(proxyBURL.Host, "")
	if err != nil {
		klog.Fatal(err)
	}

	klog.Infof("Starting leader election b")
	go newLeaderElection("pod-b", ns, eID, client)

	go func() {
		for {
			r:=rand.Intn(100)
			time.Sleep(time.Duration(r) * time.Second)
			patch:=fmt.Sprintf(`{"metadata":{"annotations":{"update":"%v"}}}`, time.Now())
			client.CoreV1().ConfigMaps(ns).Patch(eID, types.MergePatchType, []byte(patch), "")
		}
	}()

	handleSigterm(cp)
}

func newLeaderElection(podName, namespace, electionID string, client clientset.Interface) {
	setupLeaderElection(&leaderElectionConfig{
		Client:     client,
		ElectionID: electionID,
		OnStartedLeading: func(stopCh chan struct{}) {
			klog.Infof("I am the new leader (%s)", podName)
		},
		OnStoppedLeading: func() {
			klog.Infof("I am not the leader anymore (%s)", podName)
		},
		PodName:      podName,
		PodNamespace: namespace,
	})
}

func handleSigterm(cp *integration.ControlPlane) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Exiting")
	cp.Stop()
	os.Exit(0)
}

func createApiserverClient(apiserverHost, kubeConfig string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(apiserverHost, kubeConfig)
	if err != nil {
		return nil, err
	}

	klog.Infof("Creating API client for %s", cfg.Host)
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func startToxyproxyServer() {
	toxiproxyPath := os.Getenv("TEST_ASSET_TOXIPROXY")
	if toxiproxyPath == "" {
		klog.Fatal("TEST_ASSET_TOXIPROXY env variable is not optional")
	}

	cmd := exec.Command(toxiproxyPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		klog.Fatalf("Toxiproxy error: %v", err)
	}
}
