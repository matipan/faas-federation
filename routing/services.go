package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

// ReadServicesResult list of deployed functions by provider
type ReadServicesResult struct {
	Providers map[string][]*requests.Function
}

// ReadServices queries each of the given providers to list deployed functions
func ReadServices(providers []string) (*ReadServicesResult, error) {
	var urls []*url.URL

	for _, v := range providers {
		u, _ := url.Parse(v)
		u.Path = "/system/functions"
		urls = append(urls, u)
	}

	results := Get(urls, len(providers))
	serviceResult := &ReadServicesResult{Providers: map[string][]*requests.Function{}}
	for _, v := range results {
		if v.Err != nil {
			log.Errorf("error fetching function list for %s. %v", providers[v.Index], v.Err)
			break
		}

		if v.Response.StatusCode > 399 {
			log.Errorf("unexpected error code %d while fetching function list for %s. %v", v.Response.StatusCode, providers[v.Index], v.Err)
			break
		}

		var function []*requests.Function
		functionBytes, err := ioutil.ReadAll(v.Response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response for %s. %v", providers[v.Index], err)
		}

		_ = v.Response.Body.Close()
		err = json.Unmarshal(functionBytes, &function)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling response for %s. %v", providers[v.Index], err)
		}

		serviceResult.Providers[providers[v.Index]] = append(serviceResult.Providers[providers[v.Index]], function...)
	}

	return serviceResult, nil
}

func createToRequest(request *requests.CreateFunctionRequest) *requests.Function {
	return &requests.Function{
		Name:              request.Service,
		Annotations:       request.Annotations,
		EnvProcess:        request.EnvProcess,
		Image:             request.Image,
		Labels:            request.Labels,
		AvailableReplicas: 1,
		Replicas:          1,
	}
}

func requestToCreate(f *requests.Function) *requests.CreateFunctionRequest {
	return &requests.CreateFunctionRequest{
		Service:     f.Name,
		Annotations: f.Annotations,
		Labels:      f.Labels,
		Image:       f.Image,
		EnvProcess:  f.EnvProcess,
	}
}
