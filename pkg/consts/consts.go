package consts

const (
	DELIVERY = "delivery"
	PICKUP   = "pick-up"

	CURRENCY_ARS = "ARS"

	PAYMENT_METHOD_CASH         = 1
	PAYMENT_METHOD_MERCADO_PAGO = 2
	PAYMENT_METHOD_TRANSFER     = 3
	PAYMENT_METHOD_DEBIT        = 4
	PAYMENT_METHOD_CREDIT       = 5
	PAYMENT_METHOD_OTHER        = 6

	ORDER_STATUS_PENDING    = "pendiente"
	ORDER_STATUS_PREPARING  = "preparándose"
	ORDER_STATUS_TO_RETIRE  = "para retirar"
	ORDER_STATUS_ON_THE_WAY = "en camino"
	ORDER_STATUS_DELIVERED  = "entregado"
	ORDER_STATUS_CANCELED   = "cancelado"
	ORDER_STATUS_NONE       = "sin estado"
)
