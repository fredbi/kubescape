package resourcehandler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileResourceHandler(t *testing.T) {
	ctx := context.Background()
	t.Run("", func(t *testing.T) {
		inputs := []string{""}
		adaptors := &RegistryAdaptors{}
		h := NewFileResourceHandler(ctx, inputs, adaptors)

		t.Run("GetClusterAPIServerInfo is not implemented", func(t *testing.T) {
			require.Nil(t, h.GetClusterAPIServerInfo(ctx))
		})
		//func (fileHandler *FileResourceHandler) GetResources(ctx context.Context, sessionObj *cautils.OPASessionObj, _ *armotypes.PortalDesignator) (*cautils.K8SResources, map[string]workloadinterface.IMetadata, *cautils.KSResources, error) {
	})
	//func getResourcesFromPath(ctx context.Context, path string) (map[string]reporthandling.Source, []workloadinterface.IMetadata, error) {
}
