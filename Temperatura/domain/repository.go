// repository.go
package domain

import "github.com/vicpoo/APItemperatura/Temperatura/domain/entities"

type TemperatureRepository interface {
	Save(temp entities.Temperature) error
	GetAll() ([]entities.Temperature, error)
}
