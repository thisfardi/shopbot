package flasher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spacysky322/flasher2-src/internal/constants"
	"github.com/spacysky322/flasher2-src/internal/transportwrapper"
	"github.com/spacysky322/flasher2-src/internal/util"
)

var oempty = make(JsonObj)
var aempty = make(JsonArr, 0)

type FlasherClient struct {
	user   User
	client *http.Client
}

func (f *FlasherClient) GetUser() User {
	return f.user
}

func (f *FlasherClient) ParseUrl(url string) (itemid, shopid int64, err error) {
	re := regexp.MustCompile("/(\\d+)/(\\d+)")
	match := re.FindStringSubmatch(url)
	if len(match) != 0 {
		itemid, err = strconv.ParseInt(match[2], 10, 64)
		shopid, err = strconv.ParseInt(match[1], 10, 64)
		return
	}

	re = regexp.MustCompile("\\.(\\d+)\\.(\\d+)")
	match = re.FindStringSubmatch(url)
	if len(match) != 0 {
		itemid, err = strconv.ParseInt(match[2], 10, 64)
		shopid, err = strconv.ParseInt(match[1], 10, 64)
		return
	}

	err = fmt.Errorf("invalid url")
	return
}

func (f *FlasherClient) FetchItem(itemid, shopid int64) (item Item, err error) {
	req, err := http.NewRequest("GET", "https://mall.shopee.co.id/api/v4/item/get", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("itemid", strconv.FormatInt(itemid, 10))
	q.Add("shopid", strconv.FormatInt(shopid, 10))
	req.URL.RawQuery = q.Encode()
	resp, err := f.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsondata, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}
	if errcode := jsondata["error"]; errcode != nil {
		code, _ := errcode.(json.Number).Int64()
		err = fmt.Errorf("error: %d", code)
		return
	}
	if jsondata["data"] == nil {
		err = fmt.Errorf("item not found")
		return
	}

	item = jsondata["data"].(JsonObj)
	return
}

func (f *FlasherClient) AddToCart(item Item, selectedModel int) (cartItem CartItem, err error) {
	if item.Models()[selectedModel].Stock() == 0 {
		return nil, fmt.Errorf("out of stock")
	}
	req, err := http.NewRequest("POST", "https://mall.shopee.co.id/api/v4/cart/add_to_cart", util.JsonEncodeBuff(
		JsonObj{
			"quantity":          1,
			"donot_add_quality": false,
			"client_source":     5,
			"shopid":            item.ShopId(),
			"itemid":            item.ItemId(),
			"modelid":           item.Models()[selectedModel].ModelId(),
		},
	))
	if err != nil {
		return
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}

	if errcode := util.MustInt64(data["error"]); errcode != 0 {
		return nil, fmt.Errorf("error: %d", errcode)
	}
	cartData := data["data"].(JsonObj)["cart_item"].(JsonObj)
	cartData["add_on_deal_id"] = item.AddOnDealId()
	cartData["shopid"] = item.ShopId()

	return cartData, nil
}

func (f *FlasherClient) Checkout(item CartItem, info ShippingInfo, payment *Payment) (err error) {
	// TODO: Fix error_params
	if strings.TrimSpace(info.Warning()) != "" {
		return fmt.Errorf("can't use this shipping info because of %v, please choose another one", info.Warning())
	}

	ts := time.Now().Unix()
	if err = f.checkoutGet(item, ts); err != nil {
		return
	}
	totalPayable := info.Cost() + item.Price() + payment.TxnFee
	req, err := http.NewRequest("POST", "https://mall.shopee.co.id/api/v2/checkout/place_order", util.JsonEncodeBuff(
		JsonObj{
			"status":    200,
			"headers":   oempty,
			"cart_type": 0,
			"dropshipping_info": JsonObj{
				"phone_number": "",
				"enabled":      false,
				"name":         "",
			},
			"shipping_orders": JsonArr{
				JsonObj{
					"is_fsv_applied":              false,
					"selected_logistic_channelid": info.ChannelId(),
					"cod_fee":                     0,
					"order_total":                 0,
					"shipping_id":                 1,
					"shopee_shipping_discount_id": 0,
					"selected_logistic_channelid_with_warning":   nil,
					"shipping_fee_discount":                      0,
					"shipping_group_description":                 "",
					"selected_preferred_delivery_time_option_id": 0,
					"buyer_remark":                               "",
					"buyer_address_data": JsonObj{
						"tax_address":  "",
						"address_type": 0,
						"addressid":    f.user.Address().Id(),
					},
					"order_total_without_shipping": item.Price(),
					"tax_payable":                  0,
					"amount_detail": JsonObj{
						"BASIC_SHIPPING_FEE":                  info.Cost(),
						"COD_FEE":                             0,
						"SHOPEE_OR_SELLER_SHIPPING_DISCOUNT":  -1500000000,
						"VOUCHER_DISCOUNT":                    0,
						"SHIPPING_DISCOUNT_BY_SELLER":         0,
						"SELLER_ESTIMATED_INSURANCE_FEE":      0,
						"SELLER_ESTIMATED_BASIC_SHIPPING_FEE": 0,
						"SHIPPING_DISCOUNT_BY_SHOPEE":         1500000000,
						"INSURANCE_FEE":                       0,
						"ITEM_TOTAL":                          item.Price(),
						"SELLER_ONLY_SHIPPING_DISCOUNT":       0,
						"shop_promo_only":                     true,
						"TAX_FEE":                             0,
						"TAX_EXEMPTION":                       0,
					},
					"fulfillment_info": JsonObj{
						"managed_by_sbs":         false,
						"order_fulfillment_type": 2,
						"fulfillment_flag":       64,
						"fulfillment_source":     "",
						"warehouse_address_id":   0,
					},
					"voucher_wallet_checking_channel_ids": JsonArr{info.ChannelId()},
					"shoporder_indexes":                   JsonArr{0},
					"shipping_fee":                        info.Cost(),
					"tax_exemption":                       0,
					"shipping_group_icon":                 "",
					"buyer_ic_number":                     "",
				},
			},
			"selected_payment_channel_data": JsonObj{
				"channel_id":               payment.ChannelId,
				"version":                  payment.Version,
				"channel_item_option_info": JsonObj{"option_info": payment.Option},
				"text_info":                oempty,
			},
			"fsv_selection_infos": aempty,
			"disabled_checkout_info": JsonObj{
				"auto_popup":  false,
				"description": "",
				"error_infos": oempty,
			},
			"timestamp": ts,
			"checkout_price_data": JsonObj{
				"shipping_subtotal":                 info.Cost(),
				"shipping_discount_subtotal":        0,
				"shipping_subtotal_before_discount": info.Cost(),
				"bundle_deals_discount":             nil,
				"group_buy_discount":                0,
				"merchandise_subtotal":              item.Price(),
				"tax_payable":                       0,
				"buyer_txn_fee":                     payment.TxnFee,
				"credit_card_promotion":             nil,
				"promocode_applied":                 nil,
				"shopee_coins_redeemed":             nil,
				"total_payable":                     totalPayable,
				"tax_exemption":                     0,
			},
			"shoporders": JsonArr{
				JsonObj{
					"buyer_remark": "",
					"cod_fee":      0,
					"shipping_fee": info.Cost(),
					"order_total":  totalPayable,
					"amount_detail": JsonObj{
						"BASIC_SHIPPING_FEE":                  info.Cost(),
						"COD_FEE":                             0,
						"SHOPEE_OR_SELLER_SHIPPING_DISCOUNT":  -1500000000,
						"VOUCHER_DISCOUNT":                    0,
						"SHIPPING_DISCOUNT_BY_SELLER":         0,
						"SELLER_ESTIMATED_INSURANCE_FEE":      0,
						"SELLER_ESTIMATED_BASIC_SHIPPING_FEE": 0,
						"SHIPPING_DISCOUNT_BY_SHOPEE":         1500000000,
						"INSURANCE_FEE":                       0,
						"ITEM_TOTAL":                          item.Price(),
						"SELLER_ONLY_SHIPPING_DISCOUNT":       0,
						"shop_promo_only":                     true,
						"TAX_FEE":                             0,
						"TAX_EXEMPTION":                       0,
					},
					"shop": JsonObj{
						"is_official_shop": false,
						"shopid":           item.ShopId(),
						"shop_name":        "",
						"remark_type":      0,
						"support_ereceipt": false,
						"images":           "",
						"cb_option":        false,
					},
					"items": JsonArr{
						JsonObj{
							"itemid":             item.ItemId(),
							"is_add_on_sub_item": false,
							"image":              "",
							"shopid":             item.ShopId(),
							"opc_extra_data": JsonObj{
								"slash_price_activity_id": 0,
							},
							"promotion_id":               0,
							"add_on_deal_id":             item.AddOnDealId(),
							"add_on_deal_label":          "",
							"modelid":                    item.ModelId(),
							"offerid":                    0,
							"source":                     "",
							"checkout":                   true,
							"item_group_id":              item.GroupId(),
							"service_by_shopee_flag":     false,
							"addon_deal_sub_type":        0,
							"is_streaming_price":         false,
							"non_shippable_err":          "",
							"none_shippable_full_reason": "",
							"price":                      item.Price(),
							"is_flash_sale":              true,
							"categories": JsonArr{
								JsonObj{"catids": aempty},
							},
							"shippable":             true,
							"name":                  "",
							"none_shippable_reason": "",
							"is_pre_order":          false,
							"stock":                 0,
							"model_name":            "",
							"quantity":              1,
						},
					},
					"selected_preferred_delivery_time_option_id": 0,
					"selected_logistic_channelid":                info.ChannelId(),
					"tax_payable":                                0,
					"buyer_address_data": JsonObj{
						"tax_address":  "",
						"address_type": 0,
						"addressid":    f.user.Address().Id(),
					},
					"shipping_fee_discount": 0,
					"tax_info": JsonObj{
						"use_new_custom_tax_msg": false,
						"custom_tax_msg":         "",
						"custom_tax_msg_short":   "",
						"remove_custom_tax_hint": false,
					},
					"order_total_without_shipping": item.Price(),
					"tax_exemption":                0,
					"shipping_id":                  1,
					"buyer_ic_number":              "",
					"ext_ad_info_mappings":         aempty,
				},
			},
			"can_checkout":      true,
			"order_update_info": oempty,
			"buyer_txn_fee_info": JsonObj{
				"learn_more_url": "https://shopee.co.id/events3/code/1177477559/",
				"description":    "Besar biaya penanganan adalah Rp2.500 dari total transaksi.",
				"title":          "Biaya Penanganan",
			},
			"client_id": 0,
			"promotion_data": JsonObj{
				"promotion_msg":  "",
				"price_discount": 0,
				"can_use_coins":  false,
				"coin_info": JsonObj{
					"coin_offset":          0,
					"coin_earn":            0,
					"coin_earn_by_voucher": 0,
					"coin_used":            0,
				},
				"free_shipping_voucher_info": JsonObj{
					"free_shipping_voucher_id": 0,
					"disabled_reason":          nil,
					"banner_info": JsonObj{
						"msg":            "",
						"learn_more_msg": "",
					},
					"free_shipping_voucher_code": "",
				},
				"card_promotion_id": nil,
				"voucher_code":      nil,
				"shop_voucher_entrances": JsonArr{
					JsonObj{
						"status": true,
						"shopid": item.ShopId(),
					},
				},
				"voucher_info": JsonObj{
					"coin_earned":         0,
					"discount_percentage": 0,
					"discount_value":      0,
					"voucher_code":        nil,
					"reward_type":         0,
					"coin_percentage":     0,
					"used_price":          0,
					"promotionid":         0,
				},
				"applied_voucher_code":   nil,
				"platform_vouchers":      aempty,
				"card_promotion_enabled": false,
				"invalid_message":        "",
				"use_coins":              false,
			},
			"captcha_version": 1,
			"_cft":            JsonArr{1},
		},
	))
	if err != nil {
		return
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := util.JsonDecodeResp(resp)
	if err, ok := data["error"]; ok {
		return fmt.Errorf("checkout error: %v", err)
	}
	return
}

func (f *FlasherClient) FetchShippingInfo(item Item) (result []ShippingInfo, err error) {
	req, err := http.NewRequest("GET", "https://mall.shopee.co.id/api/v4/pdp/get_shipping_info", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("city", f.user.Address().City())
	q.Add("district", f.user.Address().District())
	q.Add("itemid", strconv.FormatInt(item.ItemId(), 10))
	q.Add("shopid", strconv.FormatInt(item.ShopId(), 10))
	q.Add("state", f.user.Address().State())
	req.URL.RawQuery = q.Encode()
	resp, err := f.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}
	if errcode := data["error"]; errcode != nil {
		err = fmt.Errorf("failed to get shipping info, error: %s", errcode.(string))
		return
	}

	result = []ShippingInfo{}
	for _, iinfo := range data["data"].(JsonObj)["shipping_infos"].(JsonArr) {
		result = append(result, iinfo.(JsonObj))
	}
	return
}

func (f *FlasherClient) checkoutGet(item CartItem, ts int64) (err error) {
	req, err := http.NewRequest("POST", "https://mall.shopee.co.id/api/v2/checkout/get_quick", util.JsonEncodeBuff(
		JsonObj{
			"timestamp": ts,
			"shoporders": JsonArr{
				JsonObj{
					"shop": JsonObj{
						"shopid": item.ShopId(),
					},
					"items": JsonArr{
						JsonObj{
							"itemid":             item.ItemId(),
							"modelid":            item.ModelId(),
							"add_on_deal_id":     item.AddOnDealId(),
							"is_add_on_sub_item": nil,
							"item_group_id":      item.GroupId(),
							"quantity":           1,
						},
					},
					"logistics": JsonObj{
						"recommended_channelids": nil,
					},
					"selected_preferred_delivery_time_slot_id": nil,
				},
			},
			"selected_payment_channel_data": oempty,
			"promotion_data": JsonObj{
				"use_coins": false,
				"free_shipping_voucher_info": JsonObj{
					"free_shipping_voucher_id": 0,
					"disabled_reason":          "",
					"description":              "",
				},
				"platform_vouchers":            aempty,
				"shop_vouchers":                aempty,
				"check_shop_voucher_entrances": true,
				"auto_apply_shop_voucher":      false,
			},
			"device_info": JsonObj{
				"device_id":          "",
				"device_fingerprint": "",
				"tongdun_blackbox":   "",
				"buyer_payment_info": oempty,
			},
			"tax_info": JsonObj{
				"tax_id": "",
			},
		},
	))
	if err != nil {
		return
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; !(status >= 200 && status <= 299) {
		err = fmt.Errorf("checkoutGet error, status code: %d", status)
		return
	}

	return
}

func NewFlasher(cookies map[string]*http.Cookie) (*FlasherClient, error) {
	f := &FlasherClient{}

	user, err := getUser(cookies)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	rt := transportwrapper.Wrap(client.Transport)
	rt.Cookies = cookies

	rt.Set("Referer", "https://shopee.co.id")
	rt.Set("If-None-Match-", "*")
	rt.Set("Content-Type", "application/json")
	rt.Set("User-Agent", constants.UserAgent)
	rt.Set("X-Csrftoken", cookies["csrftoken"].Value)

	client.Transport = rt
	f.user = user
	f.client = client
	return f, nil
}
