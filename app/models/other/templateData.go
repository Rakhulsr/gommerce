package other

import (
	"net/url"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
)

type UserForTemplate struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Role      string
	Addresses []models.Address
}

type BasePageData struct {
	Title                   string
	IsLoggedIn              bool
	User                    *UserForTemplate
	UserID                  string
	CartCount               int
	CSRFToken               string
	Message                 string
	MessageStatus           string
	Query                   url.Values
	Breadcrumbs             []breadcrumb.Breadcrumb
	IsAuthPage              bool
	IsAdminPage             bool
	HideAdminWelcomeMessage bool
	CurrentPath             string
	IsAdminRoute            bool
	OrderID                 string
	Order                   *models.Order
	Orders                  []models.Order
	Payment                 *models.Payment
}
