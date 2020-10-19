package dtofactory

import (
	"fmt"

	"github.com/turbonomic/kubeturbo/pkg/discovery/repository"
	sdkbuilder "github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"

	"github.com/golang/glog"
)

type businessAppEntityDTOBuilder struct {
	appToEntityMap map[repository.K8sApp][]repository.K8sAppComponent
}

func NewBusinessAppEntityDTOBuilder(appToEntityMap map[repository.K8sApp][]repository.K8sAppComponent) *businessAppEntityDTOBuilder {
	return &businessAppEntityDTOBuilder{
		appToEntityMap: appToEntityMap,
	}
}

// Build entityDTOs based on the given volume to pod mappings.
func (builder *businessAppEntityDTOBuilder) BuildEntityDTOs() []*proto.EntityDTO {
	var result []*proto.EntityDTO

	for app, entities := range builder.appToEntityMap {
		appId := string(app.Uid)
		entityDTOBuilder := sdkbuilder.NewEntityDTOBuilder(proto.EntityDTO_BUSINESS_APPLICATION, appId)
		displayName := fmt.Sprintf("%s/%s", app.Namespace, app.Name)
		entityDTOBuilder.DisplayName(displayName)

		commoditiesSold := builder.getAppCommoditiesBought(app, entities)
		if len(commoditiesSold) > 1 {
			entityDTOBuilder.SellsCommodities(commoditiesSold)
		}

		entityDTOBuilder.WithPowerState(proto.EntityDTO_POWERED_ON)

		// build entityDTO.
		entityDto, err := entityDTOBuilder.Create()
		if err != nil {
			glog.Errorf("Failed to build Application: %s entityDTO: %s", displayName, err)
			continue
		}

		result = append(result, entityDto)
	}

	return result
}

func (builder *businessAppEntityDTOBuilder) getAppCommoditiesBought(app repository.K8sApp, entities []repository.K8sAppComponent) []*proto.CommodityDTO {
	var commoditiesBought []*proto.CommodityDTO

	for _, entity := range entities {
		key := fmt.Sprintf("%s-%s/%s-%s/%s", "App", app.Namespace, app.Name, entity.TurboType.String(), entity.Name)
		commodityBought, err := sdkbuilder.NewCommodityDTOBuilder(proto.CommodityDTO_APPLICATION).
			Key(key).
			Capacity(accessCommodityDefaultCapacity).
			Create()
		if err != nil {
			glog.Errorf("Error creating commodityBought by Business App %s/%s: %v ", app.Namespace, app.Name, err)
		} else {
			// TODO(irfanurrehman): Is there a need to set properties eg movable, resizable, etc.
			commoditiesBought = append(commoditiesBought, commodityBought)
		}
	}

	return commoditiesBought
}
