package server

import "testing"

func TestLookupMachineCostProviders(t *testing.T) {
	liveCloudPrices = false
	t.Cleanup(func() { liveCloudPrices = true })
	empty := HardwareConfig{}
	if p, ok := catalogPrice("azure", "Standard_NC8as_T4_v3"); !ok || p != 0.752 {
		t.Fatalf("azure catalog %v %v", p, ok)
	}
	aws := lookupMachineCost(empty, "g5.xlarge", "us-east-1", "aws")
	if aws.HourlyUSD != 1.006 || aws.Kind != "catalog" {
		t.Fatalf("aws catalog: %+v", aws)
	}
	ibm := lookupMachineCost(empty, "gx3-24x120x1l40s", "us-south", "ibmcloud")
	if ibm.HourlyUSD != 3.88 || ibm.Kind != "catalog" {
		t.Fatalf("ibm catalog: %+v", ibm)
	}
	manual := lookupMachineCost(HardwareConfig{
		Machines: []ManualMachine{{InstanceType: "g5.xlarge", HourlyUSD: 0.4}},
	}, "g5.xlarge", "eu-west-1", "aws")
	if manual.Kind != "manual" || manual.HourlyUSD != 0.4 {
		t.Fatalf("manual override: %+v", manual)
	}
}

func TestNormalizeProvider(t *testing.T) {
	if normalizeProvider("IBMCloud") != "ibmcloud" {
		t.Fatal(normalizeProvider("IBMCloud"))
	}
	if normalizeProvider("AWS") != "aws" {
		t.Fatal(normalizeProvider("AWS"))
	}
	if !providerSupported("aws") || !providerSupported("ibmcloud") || !providerSupported("Azure") || !providerSupported("baremetal") {
		t.Fatal("expected azure/aws/ibmcloud/baremetal supported")
	}
	if normalizeProvider("BareMetal") != "baremetal" || normalizeProvider("None") != "baremetal" {
		t.Fatal(normalizeProvider("BareMetal"), normalizeProvider("None"))
	}
}
