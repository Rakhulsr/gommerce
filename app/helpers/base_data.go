package helpers

import (
	"fmt"
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/middlewares"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
)

func FormatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %.0f", amount)
}

func GetBaseData(r *http.Request, pageSpecificData map[string]interface{}) map[string]interface{} {
	if pageSpecificData == nil {
		pageSpecificData = make(map[string]interface{})
	}

	if cartCountVal := r.Context().Value(middlewares.CartCountKey); cartCountVal != nil {
		if count, ok := cartCountVal.(int); ok {
			pageSpecificData["CartCount"] = count
		} else {
			pageSpecificData["CartCount"] = 0
		}
	} else {
		pageSpecificData["CartCount"] = 0
	}

	if _, exists := pageSpecificData["breadcrumbs"]; !exists {
		pageSpecificData["breadcrumbs"] = []breadcrumb.Breadcrumb{}
	}

	return pageSpecificData
}
