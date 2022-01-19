// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/Shaad7/bookstore-sample-controller/pkg/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Bookstores returns a BookstoreInformer.
	Bookstores() BookstoreInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Bookstores returns a BookstoreInformer.
func (v *version) Bookstores() BookstoreInformer {
	return &bookstoreInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}