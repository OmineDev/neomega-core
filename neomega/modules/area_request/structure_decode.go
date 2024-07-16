package area_request

import (
	"fmt"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/utils/structure/mc_structure"
)

type StructureResponse struct {
	raw              *packet.StructureTemplateDataResponse
	decodedStructure neomega.DecodedStructure
}

func newStructureResponse(r *packet.StructureTemplateDataResponse) neomega.StructureResponse {
	return &StructureResponse{
		raw: r,
	}
}

func (sr *StructureResponse) Raw() *packet.StructureTemplateDataResponse {
	return sr.raw
}

func (sr *StructureResponse) Decode() (s neomega.DecodedStructure, err error) {
	if !sr.raw.Success {
		return nil, fmt.Errorf("response get fail result")
	}
	if sr.decodedStructure != nil {
		return sr.decodedStructure, nil
	}
	structureData := sr.raw.StructureTemplate
	structure := &mc_structure.StructureContent{}
	err = structure.FromNBT(structureData)
	if err != nil {
		return nil, err
	}
	decodeStructure := structure.Decode()
	sr.decodedStructure = decodeStructure
	return decodeStructure, nil

}
