package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"time"

	"order/internal/model"
)

const (
	outputDir        = "send_get_scripts/sample_data"
	goodOrdersFile   = "orders_1.json"
	updateOrdersFile = "orders_update.json"
	badOrdersFile    = "orders_bad.json"
	uidFile          = "uids.txt"
	numGoodOrders    = 6
)

func main() {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	if err := os.WriteFile(filepath.Join(outputDir, uidFile), []byte{}, 0644); err != nil {
		fmt.Println("Error clearing UID file:", err)
		return
	}

	goodOrders := generateGoodOrders(numGoodOrders, r)
	saveJSON(filepath.Join(outputDir, goodOrdersFile), goodOrders)
	appendUID(filepath.Join(outputDir, uidFile), goodOrders)

	updateOrders := generateUpdateOrders(goodOrders, r)
	saveJSON(filepath.Join(outputDir, updateOrdersFile), updateOrders)
	appendUID(filepath.Join(outputDir, uidFile), updateOrders)

	badOrders := generateBadOrders(goodOrders)
	saveJSON(filepath.Join(outputDir, badOrdersFile), badOrders)
	appendUID(filepath.Join(outputDir, uidFile), badOrders)
}

func generateGoodOrders(n int, r *mathrand.Rand) []model.Order {
	orders := make([]model.Order, 0, n)
	for i := 0; i < n; i++ {
		orders = append(orders, newOrder(r))
	}
	return orders
}

func generateUpdateOrders(baseOrders []model.Order, r *mathrand.Rand) []model.Order {
	updates := make([]model.Order, 0, len(baseOrders))
	for _, order := range baseOrders {
		order.Items[0].Price += r.Intn(200)
		order.DateCreated = fmt.Sprintf("2021-11-2%d06:22:19Z", r.Intn(9))
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
		bad.Payment.Transaction = bad.OrderUID
		bad.Payment.Amount = -100
		badOrders = append(badOrders, bad)
		invalidUIDCounter++
	}
	return badOrders
}

func newOrder(r *mathrand.Rand) model.Order {
	uid := "b563" + randString(12)
	track := "WBILM" + randString(8)
	sizes := []string{"S", "M", "L", "XL"}

	return model.Order{
		OrderUID:          uid,
		TrackNumber:       track,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: randString(10),
		CustomerID:        "customer" + randString(5),
		DeliveryService:   "meest",
		Shardkey:          fmt.Sprintf("%d", r.Intn(10)+1),
		SmID:              r.Intn(100),
		DateCreated:       fmt.Sprintf("2021-11-2%d06:22:19Z", r.Intn(9)),
		OofShard:          "1",
		Delivery: model.Delivery{
			Name:    "Test " + randString(5),
			Phone:   fmt.Sprintf("+972%07d", r.Intn(10000000)),
			Zip:     fmt.Sprintf("%07d", r.Intn(10000000)),
			City:    "City" + randString(3),
			Address: "Street " + randString(4),
			Region:  "Region " + randString(3),
			Email:   "test" + randString(5) + "@gmail.com",
		},
		Payment: model.Payment{
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
				ChrtID:      r.Intn(1000000),
				TrackNumber: track,
				Price:       r.Intn(500) + 100,
				Rid:         uid + "_item1",
				Name:        "Product " + randString(5),
				Sale:        r.Intn(30),
				Size:        sizes[r.Intn(len(sizes))],
				TotalPrice:  r.Intn(500) + 100,
				NmID:        r.Intn(1000000),
				Brand:       "Brand" + randString(3),
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
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening UID file:", err)
		return
	}
	defer file.Close()
	for _, order := range orders {
		file.WriteString(order.OrderUID + "\n")
	}
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)
}
