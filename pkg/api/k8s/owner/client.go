package owner

import (
	"context"
	"fmt"
	"sync"

	"github.com/milosgajdos/kraph/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

const (
	source = "k8s"
)

// API discovery results
type result struct {
	apiRes  api.Resource
	resList *unstructured.UnstructuredList
	err     error
}

// topMap stores topology map
type topMap struct {
	top *Top
	err error
}

type client struct {
	// ctx is client context
	ctx context.Context
	// disc is kubernetes discovery client
	disc discovery.DiscoveryInterface
	// dyn is kubernetes dynamic client
	dyn dynamic.Interface
	// source is API source
	source api.Source
	// opts are client options
	opts Options
}

// NewClient creates a new kubernetes API client and returns it
func NewClient(ctx context.Context, disc discovery.DiscoveryInterface, dyn dynamic.Interface, opts ...Option) *client {
	copts := Options{}
	for _, apply := range opts {
		apply(&copts)
	}

	return &client{
		ctx:    ctx,
		disc:   disc,
		dyn:    dyn,
		source: NewSource(source),
		opts:   copts,
	}
}

// Discover discovers kubernetes APIs and returns them in a single API object.
// It returns error if it fails to read the resources of if it fails to parse their versions
func (k *client) Discover() (api.API, error) {
	srvPrefResList, err := k.disc.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API groups: %w", err)
	}

	a := NewAPI(k.source)

	for _, srvPrefRes := range srvPrefResList {
		gv, err := schema.ParseGroupVersion(srvPrefRes.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("failed parsing %s into GroupVersion: %w", srvPrefRes.GroupVersion, err)
		}

		for _, ar := range srvPrefRes.APIResources {
			if !stringIn("list", ar.Verbs) {
				continue
			}

			resource := NewResource(ar, gv, api.Options{})

			if err := a.Add(resource, api.AddOptions{}); err != nil {
				return nil, err
			}
		}
	}

	return a, nil
}

// processResults processes API call request results.
// It builds API topology map from the received results.
func (k *client) processResults(a api.API, resChan <-chan result, doneChan chan struct{}, topChan chan<- topMap) {
	var err error

	top := NewTop(a)

	for result := range resChan {
		if result.err != nil {
			err = result.err
			close(doneChan)
			break
		}

		for _, raw := range result.resList.Items {
			object := NewObject(result.apiRes, raw)

			if terr := top.Add(object, api.AddOptions{}); terr != nil {
				err = terr
				close(doneChan)
				break
			}
		}
	}

	topChan <- topMap{
		top: top,
		err: err,
	}
}

// Map builds a map of API resources in a given client namespace
// If the namespace is empty it queries API groups across all namespaces.
// It returns error if any of the API calls fails with error.
func (k *client) Map(a api.API) (api.Top, error) {
	var wg sync.WaitGroup

	resChan := make(chan result, 250)
	doneChan := make(chan struct{})

	for _, resource := range a.Resources() {
		// if particular namespace is required and the resource is not namespaced, skip
		if len(k.opts.Namespace) > 0 && !resource.Namespaced() {
			continue
		}

		gvResClient := k.dyn.Resource(schema.GroupVersionResource{
			Group:    resource.Group(),
			Version:  resource.Version(),
			Resource: resource.Name(),
		})

		var client dynamic.ResourceInterface
		switch k.opts.Namespace {
		case "":
			client = gvResClient
		default:
			client = gvResClient.Namespace(k.opts.Namespace)
		}

		wg.Add(1)
		go func(r api.Resource) {
			defer wg.Done()
			var cont string
			for {
				res, err := client.List(k.ctx, metav1.ListOptions{
					Limit:    100,
					Continue: cont,
				})

				select {
				case resChan <- result{apiRes: r, resList: res, err: err}:
				case <-doneChan:
					return
				}

				if err != nil {
					return
				}

				cont = res.GetContinue()
				if cont == "" {
					break
				}
			}
		}(resource)
	}

	topChan := make(chan topMap, 1)
	go k.processResults(a, resChan, doneChan, topChan)

	wg.Wait()
	close(resChan)

	t := <-topChan

	if t.err != nil {
		return nil, t.err
	}

	return t.top, nil
}
