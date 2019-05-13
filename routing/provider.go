package routing

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/openfaas/faas/gateway/requests"
	log "github.com/sirupsen/logrus"
)

const federationProviderNameConstraint = "com.openfaas.federation.gatewayx"

// ProviderLookup allows the federation to determine which provider
// is currently responsible for a given function
type ProviderLookup interface {
	Resolve(functionName string) (providerURI *url.URL, err error)
	AddFunction(f *requests.CreateFunctionRequest)
	GetFunction(name string) (*requests.CreateFunctionRequest, bool)
	GetFunctions() []*requests.CreateFunctionRequest
}

type defaultProviderRouting struct {
	cache           map[string]*requests.CreateFunctionRequest
	providers       map[string]*url.URL
	defaultProvider *url.URL
	lock            sync.RWMutex
}

// NewDefaultProviderRouting creates a default way to resolve providers currently based
// on name constraint
func NewDefaultProviderRouting(providers []string, defaultProvider string) (ProviderLookup, error) {
	providerMap := map[string]*url.URL{}

	for _, v := range providers {
		pURL, err := url.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL using value %s. %v", v, err)
		}
		providerMap[getHostNameWithoutPorts(pURL)] = pURL
	}

	d, err := url.Parse(defaultProvider)
	if err != nil {
		return nil, fmt.Errorf("error parsing default provider URL using value %s. %v", defaultProvider, err)
	}

	return &defaultProviderRouting{
		cache:           make(map[string]*requests.CreateFunctionRequest),
		providers:       providerMap,
		defaultProvider: d,
	}, nil
}

func (d *defaultProviderRouting) Resolve(functionName string) (providerURI *url.URL, err error) {
	f, ok := d.GetFunction(functionName)
	if !ok {
		return nil, fmt.Errorf("can not find function %s in cache map", functionName)
	}

	c, ok := (*f.Annotations)[federationProviderNameConstraint]
	if !ok {
		log.Infof("%s constraint not found using default provider %s", federationProviderNameConstraint, d.defaultProvider.String())
		return d.defaultProvider, nil
	}

	pURL := d.matchBasedOnName(c)
	if pURL == nil {
		log.Infof("%s constraint value found but does not exist in provider list, using default provider %s", c, d.defaultProvider.String())

		return d.defaultProvider, nil
	}

	return pURL, nil
}

func (d *defaultProviderRouting) matchBasedOnName(v string) *url.URL {
	for _, u := range d.providers {
		if strings.EqualFold(getHostNameWithoutPorts(u), v) {
			return u
		}
	}

	return nil
}

func getHostNameWithoutPorts(v *url.URL) string {
	return strings.Split(v.Host, ":")[0]
}

func (d *defaultProviderRouting) AddFunction(f *requests.CreateFunctionRequest) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.cache[f.Service] = f
}

func (d *defaultProviderRouting) GetFunction(name string) (*requests.CreateFunctionRequest, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	v, ok := d.cache[name]

	return v, ok
}

func (d *defaultProviderRouting) GetFunctions() []*requests.CreateFunctionRequest {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*requests.CreateFunctionRequest
	for _, v := range d.cache {
		result = append(result, v)
	}

	return result
}
