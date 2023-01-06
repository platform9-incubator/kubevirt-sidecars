package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/kubevirt/pkg/hooks"

	hooksInfo "kubevirt.io/kubevirt/pkg/hooks/info"
	hooksV1alpha1 "kubevirt.io/kubevirt/pkg/hooks/v1alpha1"
	domainSchema "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

const (
	hookName = "vmxnet3-hook"
)

type infoServer struct{}

func (s infoServer) Info(ctx context.Context, params *hooksInfo.InfoParams) (*hooksInfo.InfoResult, error) {
	log.Println("Hook's Info method has been called")

	return &hooksInfo.InfoResult{
		Name: "smbios",
		Versions: []string{
			hooksV1alpha1.Version,
		},
		HookPoints: []*hooksInfo.HookPoint{
			&hooksInfo.HookPoint{
				Name:     hooksInfo.OnDefineDomainHookPointName,
				Priority: 0,
			},
		},
	}, nil
}

type v1alpha1Server struct{}

func (s v1alpha1Server) OnDefineDomain(ctx context.Context, params *hooksV1alpha1.OnDefineDomainParams) (*hooksV1alpha1.OnDefineDomainResult, error) {
	log.Println("Hook's OnDefineDomain callback method has been called")

	vmiJSON := params.GetVmi()
	vmiSpec := v1.VirtualMachineInstance{}
	err := json.Unmarshal(vmiJSON, &vmiSpec)
	if err != nil {
		log.Printf("Failed to unmarshal given VMI spec: %s %s", err, vmiJSON)
		panic(err)
	}

	domainXML := params.GetDomainXML()
	domainSpec := domainSchema.DomainSpec{}
	err = xml.Unmarshal(domainXML, &domainSpec)
	if err != nil {
		log.Printf("Failed to unmarshal given domain spec: %s", domainXML)
		panic(err)
	}

	convertNicModel(&domainSpec)

	newDomainXML, err := xml.Marshal(domainSpec)
	if err != nil {
		log.Printf("Failed to marshal updated domain spec: %s", err.Error())
		panic(err)
	}

	log.Printf("Successfully updated original domain spec with requested attributes")

	return &hooksV1alpha1.OnDefineDomainResult{
		DomainXML: newDomainXML,
	}, nil
}

func convertNicModel(domainSpec *domainSchema.DomainSpec) {
	if domainSpec.Devices.Interfaces == nil {
		return
	}

	for _, nicDevice := range domainSpec.Devices.Interfaces {
		if nicDevice.Model != nil && nicDevice.Model.Type != "vmxnet3" {
			nicDevice.Model.Type = "vmxnet3"
		}
	}
}

func main() {
	// Start listening on /var/run/kubevirt-hooks/osx-hook.sock,
	// and register an infoServer (to expose information about this
	// hook) and a callback server (which does the heavy lifting).

	socketPath := hooks.HookSocketsSharedDirectory + "/" + hookName + ".sock"
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Printf("Failed to initialized socket on path: $s %s", socket)
		log.Printf("Check whether given directory exists and socket name is not already taken by other file")
		panic(err)
	}
	defer os.Remove(socketPath)

	server := grpc.NewServer([]grpc.ServerOption{}...)
	hooksInfo.RegisterInfoServer(server, infoServer{})
	hooksV1alpha1.RegisterCallbacksServer(server, v1alpha1Server{})
	log.Printf("Starting hook server exposing 'info' and 'v1alpha1' services on socket %s", socketPath)
	server.Serve(socket)
}
