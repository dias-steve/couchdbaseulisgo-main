package couchdbUtils

import "encoding/json"

type DtoConverter[Entity any, Dto any] interface {
	ToDto(entity Entity) Dto
	ToEntity(dto Dto) Entity
	ConvertListToDto(list []Entity) []Dto
	ConvertListToEntity(list []Dto) []Entity
	ConvertListToDtoWithPagination(list ResponseListWithPagination[[]Entity]) ResponseListWithPagination[[]Dto]
	ConvertListToEntityWithPagination(list ResponseListWithPagination[[]Dto]) ResponseListWithPagination[[]Entity]
}

type dtoConverter[Entity any, Dto any] struct {
}

func NewDtoConverter[Entity any, Dto any]() DtoConverter[Entity, Dto] {
	return &dtoConverter[Entity, Dto]{}
}

func (c *dtoConverter[Entity, Dto]) ToDto(entity Entity) Dto {
	dto := MapperObject[Entity, Dto](entity)
	return dto

}
func (c *dtoConverter[Entity, Dto]) ToEntity(dto Dto) Entity {
	entity := MapperObject[Dto, Entity](dto)
	return entity
}

func (c *dtoConverter[Entity, Dto]) ConvertListToDto(list []Entity) []Dto {
	var result []Dto
	for _, entity := range list {
		dto := c.ToDto(entity)
		result = append(result, dto)
	}
	return result
}

func (c *dtoConverter[Entity, Dto]) ConvertListToEntity(list []Dto) []Entity {
	var result []Entity
	for _, dto := range list {
		entity := c.ToEntity(dto)
		result = append(result, entity)
	}
	return result
}

func (c *dtoConverter[Entity, Dto]) ConvertListToDtoWithPagination(list ResponseListWithPagination[[]Entity]) ResponseListWithPagination[[]Dto] {
	var result ResponseListWithPagination[[]Dto]
	result.Data = c.ConvertListToDto(list.Data)
	result.Pagination = list.Pagination
	return result
}

func (c *dtoConverter[Entity, Dto]) ConvertListToEntityWithPagination(list ResponseListWithPagination[[]Dto]) ResponseListWithPagination[[]Entity] {
	var result ResponseListWithPagination[[]Entity]
	result.Data = c.ConvertListToEntity(list.Data)
	result.Pagination = list.Pagination
	return result
}

func MapperObject[T any, V any](objectToMappe T) (objectMapped V) {
	objectByte, _ := json.Marshal(objectToMappe)
	json.Unmarshal(objectByte, &objectMapped)

	return objectMapped
}

func MapList[T any, V any](listToMappe []T) (listMapped []V) {

	for _, object := range listToMappe {
		objectMapped := MapperObject[T, V](object)
		listMapped = append(listMapped, objectMapped)
	}

	return listMapped
}
