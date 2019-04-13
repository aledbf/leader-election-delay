module github.com/aledbf/leader-election-delay

go 1.12

require (
	github.com/Shopify/toxiproxy v2.1.4+incompatible
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/spf13/pflag v1.0.1
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/api v0.0.0-20190409092523-d687e77c8ae9
	k8s.io/apimachinery v0.0.0-20190409092423-760d1845f48b
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.0
	sigs.k8s.io/testing_frameworks v0.1.1
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190411052641-7a6b4715b709
