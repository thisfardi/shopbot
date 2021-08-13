package flasher

type Client interface {
	GetUser() User
	ParseUrl(url string) (itemid, shopid int64, err error)
	FetchItem(itemid, shopid int64) (Item, error)
	AddToCart(item Item, selectedModel int) (CartItem, error)
	Checkout(item CartItem, info ShippingInfo, payment Payment) error
}
