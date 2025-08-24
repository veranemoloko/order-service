package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	model "order/internal/entity"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	outputDir := "send_get_scripts/sample_data"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	goodOrders := generateGoodOrders(6, r)
	saveJSON(filepath.Join(outputDir, "orders_1.json"), goodOrders)
	appendUID(filepath.Join(outputDir, "uids.txt"), goodOrders)

	updateOrders := generateUpdateOrders(goodOrders, r)
	saveJSON(filepath.Join(outputDir, "orders_update.json"), updateOrders)

	badOrders := generateBadOrders(goodOrders)
	saveJSON(filepath.Join(outputDir, "orders_bad.json"), badOrders)
	appendUID(filepath.Join(outputDir, "uids.txt"), badOrders)
}

func generateGoodOrders(n int, r *rand.Rand) []model.Order {
	orders := make([]model.Order, 0, n)
	for i := 0; i < n; i++ {
		orders = append(orders, newOrder(r))
	}
	return orders
}

func generateUpdateOrders(baseOrders []model.Order, r *rand.Rand) []model.Order {
	updates := make([]model.Order, 0, len(baseOrders))
	for _, order := range baseOrders {
		order.Items[0].Price += r.Intn(200)
		order.DateCreated = time.Now()
		updates = append(updates, order)
	}
	return updates
}

func generateBadOrders(goodOrders []model.Order) []model.Order {
	badOrders := make([]model.Order, 0, len(goodOrders))
	invalidUIDCounter := 1

	for _, order := range goodOrders {
		bad := order
		bad.OrderUID = fmt.Sprintf("invalid_uid_%03d", invalidUIDCounter)
		bad.Payment.OrderUID = bad.OrderUID
		bad.Payment.Transaction = bad.OrderUID
		for i := range bad.Items {
			bad.Items[i].OrderUID = bad.OrderUID
			bad.Items[i].Rid = bad.OrderUID + "_item1"
		}
		bad.Payment.Amount = -100
		badOrders = append(badOrders, bad)
		invalidUIDCounter++
	}
	return badOrders
}

func newOrder(r *rand.Rand) model.Order {
	uid := "b563" + randString(12, r)
	track := "WBILM" + randString(8, r)
	sizes := []string{"S", "M", "L", "XL"}

	return model.Order{
		OrderUID:          uid,
		TrackNumber:       track,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: randString(10, r),
		CustomerID:        "customer" + randString(5, r),
		DeliveryService:   "meest",
		Shardkey:          fmt.Sprintf("%d", r.Intn(10)+1),
		SmID:              r.Intn(100),
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: model.Delivery{
			OrderUID: uid,
			Name:     "Test " + randString(5, r),
			Phone:    fmt.Sprintf("+972%07d", r.Intn(10000000)),
			Zip:      fmt.Sprintf("%07d", r.Intn(10000000)),
			City:     "City" + randString(3, r),
			Address:  "Street " + randString(4, r),
			Region:   "Region " + randString(3, r),
			Email:    "test" + randString(5, r) + "@gmail.com",
		},
		Payment: model.Payment{
			OrderUID:     uid,
			Transaction:  uid,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       r.Intn(2000) + 200,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: r.Intn(1000) + 100,
			GoodsTotal:   r.Intn(1000) + 100,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				OrderUID:    uid,
				ChrtID:      r.Intn(1000000),
				TrackNumber: track,
				Price:       r.Intn(500) + 100,
				Rid:         uid + "_item1",
				Name:        "Product " + randString(5, r),
				Sale:        r.Intn(30),
				Size:        sizes[r.Intn(len(sizes))],
				TotalPrice:  r.Intn(500) + 100,
				NmID:        r.Intn(1000000),
				Brand:       "Brand" + randString(3, r),
				Status:      202,
			},
		},
	}
}

func saveJSON(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

func appendUID(filename string, orders []model.Order) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening UID file:", err)
		return
	}
	defer file.Close()
	for _, order := range orders {
		file.WriteString(order.OrderUID + "\n")
	}
}

func randString(n int, r *rand.Rand) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
