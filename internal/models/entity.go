package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Entity define el contrato que deben cumplir los modelos como Brand y ProductType
// para poder ser gestionados por el servicio y repositorio genérico.
type Entity interface {
	SetID(id primitive.ObjectID)
	SetName(name string)
	SetSlug(slug string)
	SetStatus(status string)
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
	// GetID() primitive.ObjectID // Esta interfaz podría crecer a futuro
}

// Para que Brand implemente la interfaz Entity.
func (b *Brand) SetID(id primitive.ObjectID) { b.ID = id }
func (b *Brand) SetName(name string)         { b.Name = name }
func (b *Brand) SetSlug(slug string)         { b.Slug = slug }
func (b *Brand) SetStatus(status string)     { b.Status = status }
func (b *Brand) SetCreatedAt(t time.Time)    { b.CreatedAt = t }
func (b *Brand) SetUpdatedAt(t time.Time)    { b.UpdatedAt = t }

// Para que ProductType implemente la interfaz Entity.
func (pt *ProductType) SetID(id primitive.ObjectID) { pt.ID = id }
func (pt *ProductType) SetName(name string)         { pt.Name = name }
func (pt *ProductType) SetSlug(slug string)         { pt.Slug = slug }
func (pt *ProductType) SetStatus(status string)     { pt.Status = status }
func (pt *ProductType) SetCreatedAt(t time.Time)    { pt.CreatedAt = t }
func (pt *ProductType) SetUpdatedAt(t time.Time)    { pt.UpdatedAt = t }
