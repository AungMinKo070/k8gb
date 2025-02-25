/*
Copyright 2021 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/
package dns

import (
	"testing"

	"github.com/AbsaOSS/k8gb/controllers/depresolver"
	"github.com/AbsaOSS/k8gb/controllers/providers/assistant"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/stretchr/testify/assert"
)

var predefinedConfig = depresolver.Config{
	ReconcileRequeueSeconds: 30,
	ClusterGeoTag:           "us-west-1",
	ExtClustersGeoTags:      []string{"us-east-1"},
	EdgeDNSServer:           "8.8.8.8",
	EdgeDNSZone:             "example.com",
	DNSZone:                 "cloud.example.com",
	K8gbNamespace:           "k8gb",
	Infoblox: depresolver.Infoblox{
		Host:     "fakeinfoblox.example.com",
		Username: "foo",
		Password: "blah",
		Port:     443,
		Version:  "0.0.0",
	},
	Override: depresolver.Override{
		FakeInfobloxEnabled: true,
	},
}

func TestCanFilterOutDelegatedZoneEntryAccordingFQDNProvided(t *testing.T) {
	// arrange
	delegateTo := []ibclient.NameServer{
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.1", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.2", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.3", Name: "gslb-ns-za-cloud.example.com"},
	}
	want := []ibclient.NameServer{
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
	}
	customConfig := predefinedConfig
	customConfig.EdgeDNSZone = "example.com"
	customConfig.ExtClustersGeoTags = []string{"za"}
	a := assistant.NewGslbAssistant(nil, customConfig.K8gbNamespace, customConfig.EdgeDNSServer, customConfig.EdgeDNSServerPort)
	provider := NewInfobloxDNS(customConfig, a)
	// act
	extClusters := customConfig.GetExternalClusterNSNames()
	got := provider.filterOutDelegateTo(delegateTo, extClusters["za"])
	// assert
	assert.Equal(t, want, got, "got:\n %q filtered out delegation records,\n\n want:\n %q", got, want)
}
func TestCanSanitizeDelegatedZone(t *testing.T) {
	// arrange
	local := []ibclient.NameServer{
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
	}
	upstream := []ibclient.NameServer{
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.3", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.3", Name: "gslb-ns-cloud-example-com-eu.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-cloud-example-com-eu.example.com"},
		{Address: "10.0.0.1", Name: "gslb-ns-cloud-example-com-eu.example.com"},
		{Address: "10.1.0.2", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.1", Name: "gslb-ns-za-cloud.example.com"},
	}
	want := []ibclient.NameServer{
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.1", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.2", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.3", Name: "gslb-ns-za-cloud.example.com"},
	}
	customConfig := predefinedConfig
	customConfig.EdgeDNSZone = "example.com"
	customConfig.ExtClustersGeoTags = []string{"za"}
	customConfig.ClusterGeoTag = "eu"
	a := assistant.NewGslbAssistant(nil, customConfig.K8gbNamespace, customConfig.EdgeDNSServer, customConfig.EdgeDNSServerPort)
	provider := NewInfobloxDNS(customConfig, a)
	// act
	got := provider.sanitizeDelegateZone(local, upstream)
	// assert
	assert.Equal(t, want, got, "got:\n %q filtered out delegation records,\n\n want:\n %q", got, want)
}

func TestSortNameServer(t *testing.T) {
	delegateTo := []ibclient.NameServer{
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.3", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.2", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.1", Name: "gslb-ns-za-cloud.example.com"},
	}
	want := []ibclient.NameServer{
		{Address: "10.0.0.1", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.2", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.0.0.3", Name: "gslb-ns-eu-cloud.example.com"},
		{Address: "10.1.0.1", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.2", Name: "gslb-ns-za-cloud.example.com"},
		{Address: "10.1.0.3", Name: "gslb-ns-za-cloud.example.com"},
	}
	sortZones(delegateTo)
	assert.Equal(t, want, delegateTo, "got:\n %q \n\n want:\n %q", delegateTo, want)
}

func TestEmptySort(t *testing.T) {
	delegateTo := make([]ibclient.NameServer, 0)

	sortZones(delegateTo)
	assert.Equal(t, 0, len(delegateTo))
}

func TestNilSort(t *testing.T) {
	delegateTo := []ibclient.NameServer(nil)

	sortZones(delegateTo)
	assert.Nil(t, delegateTo)
}
