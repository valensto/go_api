package api_apbp

import (
	"fmt"
	"time"

	"github.com/valensto/api_apbp/internal/repo/order"
	"github.com/valensto/api_apbp/internal/repo/user"
	"github.com/valensto/api_apbp/pkg/mailer"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JsonOrder struct {
	ID            primitive.ObjectID `json:"-"`
	CreatedAt     time.Time          `json:"created_at,omitempty"`
	ModifiedAt    time.Time          `json:"modified_at,omitempty"`
	Ref           string             `json:"ref,omitempty" validate:"required"`
	RecoveryAt    time.Time          `json:"recovery_at,omitempty"`
	RelationShip  relationShip       `json:"relationShip,omitempty"`
	ProductsLines []ProductLine      `json:"products,omitempty" validate:"required,unique,min=1,dive,required"`
	Status        string             `json:"status,omitempty" validate:"required,oneof=waiting confirm ready delivered"`
}

type relationShip struct {
	Customer primitive.ObjectID `json:"customer"`
	Editor   primitive.ObjectID `json:"editor,omitempty"`
	Included *included          `json:"included,omitempty"`
}

type included struct {
	Customer JsonUser `json:"customer"`
	Editor   JsonUser `json:"editor"`
}

type ProductLine struct {
	Quantity float32 `json:"quantity,omitempty" validate:"required"`
	Unit     string  `json:"unit,omitempty" validate:"required,oneof=gr p"`
	Ref      string  `json:"ref,omitempty" validate:"required,ref,len=8"`
	Name     string  `json:"name,omitempty" validate:"required"`
	AUW      float32 `json:"auw,omitempty" validate:"required,numeric"`
}

type forecastProduct struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

type JsonForecast struct {
	Product  forecastProduct `json:"product"`
	Quantity int             `json:"quantity"`
}

func MapForecastToJson(fo order.Forecast) JsonForecast {
	fp := forecastProduct{
		Name: fo.Product.Name,
		Ref:  fo.Product.Ref,
	}

	return JsonForecast{
		Product:  fp,
		Quantity: fo.Quantity,
	}
}

func MapIncludeToJSON(icld *order.Included) *included {
	if icld == nil {
		return nil
	}
	return &included{
		Customer: MapUserToJSON(icld.Customer),
		Editor:   MapUserToJSON(icld.Customer),
	}
}

func MapOrderToJSON(o order.Order) JsonOrder {
	productLines := make([]ProductLine, len(o.ProductsLines))
	for i, pl := range o.ProductsLines {
		productLines[i] = ProductLine{
			Quantity: pl.Quantity,
			Unit:     pl.Unit,
			Ref:      pl.Ref,
			Name:     pl.Name,
			AUW:      pl.AUW,
		}
	}

	return JsonOrder{
		Ref:        o.Ref,
		CreatedAt:  o.CreatedAt,
		ModifiedAt: o.ModifiedAt,
		RecoveryAt: o.RecoveryAt,
		RelationShip: relationShip{
			Customer: o.RelationShip.Customer,
			Editor:   o.RelationShip.Editor,
			Included: MapIncludeToJSON(o.RelationShip.Included),
		},
		ProductsLines: productLines,
		Status:        o.Status,
	}
}

func (o JsonOrder) NewOrderMail(u user.User) mailer.Mail {
	mail := mailer.NewMail()

	included := included{
		Customer: MapUserToJSON(u),
	}
	o.RelationShip.Included = &included

	mail.ParseTemplate("web/templates/mail/recap.html", o)

	mail.Subject = "Nouvelle commande"
	mail.To = []string{"v.e.brochard@gmail.com"}

	return mail
}

func (o JsonOrder) NewStatusMail() mailer.Mail {
	mail := mailer.NewMail()

	mail.ParseTemplate("web/templates/mail/statur.html", o)

	mail.Subject = "Commande ready"
	mail.To = []string{"v.e.brochard@gmail.com"}

	return mail
}

func GenerateRef() string {
	now := time.Now()
	return fmt.Sprintf("%s", now.Format("060201150405"))
}
