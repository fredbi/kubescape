package resourcehandler

import (
	"testing"

	"github.com/kubescape/k8s-interface/k8sinterface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestSelector(t *testing.T) {
	k8sinterface.InitializeMapResourcesMock()

	t.Run("with private methods", func(t *testing.T) {
		t.Run("should get namespace selector", func(t *testing.T) {
			assert.Equal(t, "metadata.namespace==default", getNamespacesSelector(&schema.GroupVersionResource{Version: "v1", Resource: "pods"}, "default", "=="))
			assert.Equal(t, "", getNamespacesSelector(&schema.GroupVersionResource{Version: "v1", Resource: "nodes"}, "default", "=="))
		})
	})

	t.Run("with exclude selector", func(t *testing.T) {
		es := NewExcludeSelector("default,ingress")

		selectors := es.GetNamespacesSelectors(&schema.GroupVersionResource{Resource: "pods"})
		assert.Equal(t, 1, len(selectors))
		assert.Equal(t, "metadata.namespace!=default,metadata.namespace!=ingress,", selectors[0])

		selectors2 := es.GetNamespacesSelectors(&schema.GroupVersionResource{Resource: "namespaces"})
		assert.Equal(t, 1, len(selectors2))
		assert.Equal(t, "metadata.name!=default,metadata.name!=ingress,", selectors2[0])

		require.True(t, es.GetClusterScope(&schema.GroupVersionResource{Resource: "namespaces"}))
		require.False(t, es.GetClusterScope(&schema.GroupVersionResource{Resource: "pods"}))
	})

	t.Run("with include selector", func(t *testing.T) {
		is := NewIncludeSelector("default,ingress")

		selectors := is.GetNamespacesSelectors(&schema.GroupVersionResource{Resource: "pods"})
		assert.Equal(t, 2, len(selectors))
		assert.Equal(t, "metadata.namespace==default", selectors[0])
		assert.Equal(t, "metadata.namespace==ingress", selectors[1])

		selectors2 := is.GetNamespacesSelectors(&schema.GroupVersionResource{Resource: "namespaces"})
		assert.Equal(t, 2, len(selectors2))
		assert.Equal(t, "metadata.name==default", selectors2[0])
		assert.Equal(t, "metadata.name==ingress", selectors2[1])

		require.True(t, is.GetClusterScope(&schema.GroupVersionResource{Resource: "namespaces"}))
		require.False(t, is.GetClusterScope(&schema.GroupVersionResource{Resource: "pods"}))
	})

	t.Run("with empty selector", func(t *testing.T) {
		empty := &EmptySelector{}

		selectors := empty.GetNamespacesSelectors(&schema.GroupVersionResource{Resource: "pods"})
		assert.Equal(t, 1, len(selectors))
		assert.Equal(t, "", selectors[0])

		require.True(t, empty.GetClusterScope(&schema.GroupVersionResource{Resource: "namespaces"}))
	})
}
