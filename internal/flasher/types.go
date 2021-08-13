package flasher

import (
	"fmt"
	"strconv"

	"github.com/spacysky322/flasher2-src/internal/util"
)

type (
	JsonObj      = map[string]interface{}
	JsonArr      = []interface{}
	Address      map[string]interface{}
	User         map[string]interface{}
	Item         map[string]interface{}
	Model        map[string]interface{}
	CartItem     map[string]interface{}
	ShippingInfo map[string]interface{}
	UpcomingFS   map[string]interface{}
	Payment      struct {
		Name      string
		ChannelId int
		Option    string
		Version   int
		TxnFee    int64
	}
)

func (a Address) City() string {
	return a["city"].(string)
}

func (a Address) District() string {
	return a["district"].(string)
}

func (a Address) Id() int64 {
	return util.MustInt64(a["id"])
}

func (a Address) State() string {
	return a["state"].(string)
}

func (u User) SayHello() {
	fmt.Println("Hewwo ^_^")
}

func (u User) UserId() int64 {
	return util.MustInt64(u["userid"])
}

func (u User) ShopId() int64 {
	return util.MustInt64(u["shopid"])
}

func (u User) Username() string {
	return u["username"].(string)
}

func (u User) Address() Address {
	if x := u["default_address"]; x != nil {
		return x.(JsonObj)
	}
	return nil
}

func (i Item) AddOnDealId() *int {
	if x := i["add_on_deal_info"]; x != nil {
		num := util.MustInt64(x.(JsonObj)["add_on_deal_id"])
		ii := int(num)
		return &ii
	}
	return nil
}

func (i Item) IsFlashSale() bool {
	return i["flash_sale"] != nil
}

func (i Item) ItemId() int64 {
	return util.MustInt64(i["itemid"])
}

func (i Item) Models() []Model {
	models := []Model{}
	for _, model := range i["models"].(JsonArr) {
		models = append(models, (Model)(model.(JsonObj)))
	}
	return models
}

func (i Item) Name() string {
	return i["name"].(string)
}

func (i Item) Price() int64 {
	return util.MustInt64(i["price"])
}

func (i Item) ShopId() int64 {
	return util.MustInt64(i["shopid"])
}

func (i Item) Stock() int {
	num := util.MustInt64(i["stock"])
	return int(num)
}

func (i Item) UpcomingFS() UpcomingFS {
	if x := i["upcoming_flash_sale"]; x != nil {
		return x.(JsonObj)
	}
	return nil
}

func (m Model) ItemId() int64 {
	return util.MustInt64(m["itemid"])
}

func (m Model) ModelId() int64 {
	return util.MustInt64(m["modelid"])
}

func (m Model) Name() string {
	return m["name"].(string)
}

func (m Model) Price() int64 {
	return util.MustInt64(m["price"])
}

func (m Model) Stock() int {
	num := util.MustInt64(m["stock"])
	return int(num)
}

func (c CartItem) AddOnDealId() *int {
	// nil check must be done beforehand, by Item.AddOnDealId()
	return c["add_on_deal_id"].(*int)
}

func (c CartItem) GroupId() *string {
	if x := c["item_group_id"]; x != nil {
		i := util.MustInt64(x)
		s := strconv.FormatInt(i, 10)
		return &s
	}
	return nil
}

func (c CartItem) ItemId() int64 {
	return util.MustInt64(c["itemid"])
}

func (c CartItem) ModelId() int64 {
	return util.MustInt64(c["modelid"])
}

func (c CartItem) Price() int64 {
	return util.MustInt64(c["price"])
}

func (c CartItem) ShopId() int64 {
	// conversion must be done beforehand, by Item.ShopId()
	return c["shopid"].(int64)
}

func (s ShippingInfo) Name() string {
	return s["channel"].(JsonObj)["name"].(string)
}

func (s ShippingInfo) ChannelId() int {
	return int(util.MustInt64(s["channel"].(JsonObj)["channelid"]))
}

func (s ShippingInfo) Cost() int64 {
	return util.MustInt64(s["original_cost"])
}

func (s ShippingInfo) Warning() string {
	return s["warning"].(string)
}

func (u UpcomingFS) StartTime() int64 {
	return u["start_time"].(int64)
}

func (u UpcomingFS) EndTime() int64 {
	return u["end_time"].(int64)
}
