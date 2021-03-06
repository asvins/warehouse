package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/asvins/common_db/postgres"
	"github.com/asvins/router"
	"github.com/asvins/utils/config"
	"github.com/jinzhu/gorm"
)

var (
	_headers map[string]string
	products map[string]Product
	testdb   *gorm.DB
)

func _addProduct(p Product) {

	products[p.Name] = p
}

func initDatabase() {
	err := config.Load("warehouse_config.gcfg", ServerConfig)
	if err != nil {
		log.Fatal(err)
	}

	DatabaseConfig := postgres.NewConfig(ServerConfig.Database.User, ServerConfig.Database.DbName, ServerConfig.Database.SSLMode)
	testdb = postgres.GetDatabase(DatabaseConfig)
}

func clean() {
	testdb.Delete(Withdrawal{})
	testdb.Delete(Purchase{})
	testdb.Delete(PurchaseProduct{})
	testdb.Delete(Product{})
	testdb.Delete(Order{})
}

func populateProducts() {
	_addProduct(Product{Name: "coke", Description: "From Coke", Type: 1, CurrQuantity: 60, MinQuantity: 50})
	_addProduct(Product{Name: "h2oh", Description: "From AmBev", Type: 2, CurrQuantity: 100, MinQuantity: 50})
	_addProduct(Product{Name: "pepsi", Description: "From Pepsico", Type: 3, CurrQuantity: 10, MinQuantity: 20})
	_addProduct(Product{Name: "original", Description: "From AmBev", Type: 4, CurrQuantity: 70, MinQuantity: 50})
	_addProduct(Product{Name: "kuat", Description: "From Coke", Type: 5, CurrQuantity: 80, MinQuantity: 90})
	_addProduct(Product{Name: "guarana", Description: "From From AmBev", Type: 6, CurrQuantity: 5, MinQuantity: 10})
	_addProduct(Product{Name: "mate", Description: "From Sei la", Type: 7, CurrQuantity: 60, MinQuantity: 50})
	_addProduct(Product{Name: "soda", Description: "From Coke company", Type: 8, CurrQuantity: 70, MinQuantity: 60})
	_addProduct(Product{Name: "juice", Description: "From Mother Nature", Type: 9, CurrQuantity: 90, MinQuantity: 110})
}

func getBytes(p Product) []byte {
	bjson, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return bjson
}

func setup() {
	_headers = make(map[string]string)
	products = make(map[string]Product)
	populateProducts()

	initDatabase()
}

func makeRequest(httpMethod string, url string, requestObj []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(requestObj))
	addHeaders(req, headers)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// get - http://127.0.0.1:8080/api/inventory/product/:id
func productExists(id int) bool {
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/product/"+strconv.Itoa(id), make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] productExists: ", string(body), " StatusCode: ", response.StatusCode)

	return response.StatusCode == http.StatusOK

}

func getPurchaseByOrderId(orderId int) *Purchase {
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/purchase/order/"+strconv.Itoa(orderId), make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] purchaseByOrderId: ", string(body))

	purchase := Purchase{}
	if err := json.Unmarshal(body, &purchase); err != nil {
		panic(err)
	}

	return &purchase
}

func getWithdrawals() []Withdrawal {
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/withdrawal", make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] response from get Withdrawal: ", string(body))

	ws := []Withdrawal{}
	if err := json.Unmarshal(body, &ws); err != nil {
		panic(err)
	}

	return ws
}

func getPurchaseProducts() []PurchaseProduct {
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/purchaseProduct", make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] response from get Withdrawal: ", string(body))

	pp := []PurchaseProduct{}
	if err := json.Unmarshal(body, &pp); err != nil {
		panic(err)
	}

	return pp
}

func getOpenOrder() *Order {
	orderResponse, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/order/open", make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer orderResponse.Body.Close()
	orderBody, _ := ioutil.ReadAll(orderResponse.Body)

	order := Order{}
	if err := json.Unmarshal(orderBody, &order); err != nil {
		panic(err)
	}

	return &order
}

// GET - http://127.0.0.1:8080/api/inventory/order/open
func openOrderExists() bool {
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/order/open", make([]byte, 1), _headers)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] getOpenOrder: ", string(body))

	return response.StatusCode == http.StatusOK
}

///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////// MAIN //////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
func TestMain(m *testing.M) {
	flag.Parse()
	setup()
	clean()
	exitStatus := m.Run()
	clean()
	os.Exit(exitStatus)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////// DON'T CHANGE THE TESTS ORDER //////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////

// POST - http://127.0.0.1:8080/api/inventory/product
func TestAddProduct(t *testing.T) {
	fmt.Println("[INFO] -- TestAddProduct begin -- ")
	fmt.Println("[INFO] TestAddProduct Should just add a new product to database")

	postProductResponse, err := makeRequest(router.POST, "http://127.0.0.1:8080/api/inventory/product", getBytes(products["coke"]), _headers)
	if err != nil {
		t.Error(err)
	}

	if postProductResponse.StatusCode != http.StatusOK {
		t.Error("[ERROR] Status code should be: ", http.StatusOK, " Got: ", postProductResponse.StatusCode)
	}

	defer postProductResponse.Body.Close()
	body, _ := ioutil.ReadAll(postProductResponse.Body)
	fmt.Println("[INFO] Response: ", string(body))

	coke := &Product{}
	err = json.Unmarshal(body, coke)
	if err != nil {
		t.Error(err)
		return
	}

	products["coke"] = *coke
	if products["coke"].ID == 0 {
		t.Error("[ERROR] Coke id not updated")
	}

	if openOrderExists() {
		t.Error("[ERROR] Order should not have been created!")
	}

	fmt.Println("[INFO] -- TestAddProduct end --\n")
}

// PUT http://127.0.0.1:8080/api/inventory/product/:id
func TestUpdateProductAndCreateOrder(t *testing.T) {
	fmt.Println("[INFO] -- TestUpdateProductAndCreateOrder start --")

	coke := products["coke"]
	coke.CurrQuantity = 30
	products["coke"] = coke

	response, err := makeRequest(router.PUT, "http://127.0.0.1:8080/api/inventory/product/"+strconv.Itoa(products["coke"].ID), getBytes(products["coke"]), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] Status code should be: ", http.StatusOK, " Got: ", response.StatusCode)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] Response: ", string(body))

	if !openOrderExists() {
		t.Error("[ERROR] Order should have been created!")
	}

	fmt.Println("[INFO] -- TestUpdateProductAndCreateOrder end --\n")
}

// DELETE http://127.0.0.1:8080/api/inventory/product/:id
func TestDeleteProduct(t *testing.T) {
	fmt.Println("[INFO] -- TestDeleteProduct start --")

	response, err := makeRequest(router.DELETE, "http://127.0.0.1:8080/api/inventory/product/"+strconv.Itoa(products["coke"].ID), getBytes(products["coke"]), _headers)
	if err != nil {
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] Response: ", string(body))

	if productExists(products["coke"].ID) {
		t.Error("[ERROR] Product with id", products["coke"].ID, " Should have been deleted")
	}

	fmt.Println("[INFO] -- TestDeleteProduct end --\n")
}

func TestAddMultipleProducts(t *testing.T) {
	fmt.Println("[INFO] -- TestDeleteProduct start --")

	for _, value := range products {
		postProductResponse, err := makeRequest(router.POST, "http://127.0.0.1:8080/api/inventory/product", getBytes(value), _headers)
		if err != nil {
			t.Error(err)
		}

		if postProductResponse.StatusCode != http.StatusOK {
			t.Error("[ERROR] Status code should be: ", http.StatusOK, " Got: ", postProductResponse.StatusCode)
		}

		defer postProductResponse.Body.Close()
		body, _ := ioutil.ReadAll(postProductResponse.Body)

		currProd := &Product{}
		json.Unmarshal(body, currProd)
		products[value.Name] = *currProd
	}

	fmt.Println("[INFO] -- TestDeleteProduct end --\n")
}

// PUT http://127.0.0.1:8080/api/inventory/product/:id
func TestUpdateProductAndRemoveFromOrder(t *testing.T) {
	fmt.Println("[INFO] -- TestUpdateProductAndRemoveFromOrder start --")

	juice := products["juice"]
	juice.CurrQuantity = 200
	products["juice"] = juice

	response, err := makeRequest(router.PUT, "http://127.0.0.1:8080/api/inventory/product/"+strconv.Itoa(products["juice"].ID), getBytes(products["juice"]), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] Status code should be: ", http.StatusOK, " Got: ", response.StatusCode)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] Response: ", string(body))

	order := getOpenOrder()
	for _, prod := range order.Pproducts {
		if prod.ProductId == products["juice"].ID {
			t.Error("[ERROR] product should have been deleted!")
		}
	}

	fmt.Println("[INFO] -- TestUpdateProductAndRemoveFromOrde end --\n")
}

// GET http://127.0.0.1:8080/api/inventory/product/:id/consume/:quantity
func TestConsumeProduct(t *testing.T) {
	fmt.Println("[INFO] -- TestConsumeProduct start --")

	id := strconv.Itoa(products["h2oh"].ID)
	quantity := "60"
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/product/"+id+"/consume/"+quantity, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] getOpenOrder: ", string(body))

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] statudCode should be: ", http.StatusOK, " Got: ", response.StatusCode)
	}

	order := getOpenOrder()

	// Verify if purchase product was added to the existing open order
	productAdded := false
	for _, prod := range order.Pproducts {
		if prod.ProductId == products["h2oh"].ID {
			productAdded = true
		}
	}

	if !productAdded {
		t.Error("[EROR] Product NOT added after consume request")
	}

	//verify if a withdrawal entry was created
	ws := getWithdrawals()
	if len(ws) == 0 {
		t.Error("[ERROR] There should have been Withdrawals on database")
	}

	wCount := 0
	for _, w := range ws {
		if w.ProductId == products["h2oh"].ID {
			wCount++
		}
	}

	if wCount != 1 {
		t.Error("[ERROR] Expected number of Withdrawals: 1, Got: " + strconv.Itoa(wCount))
	}

	// verify if the purchase products were created/updated correctly
	purchaseProductCount := 0
	pps := getPurchaseProducts()
	for _, pp := range pps {
		if (pp.ProductId == products["h2oh"].ID) && (pp.OrderId == order.ID) {
			purchaseProductCount++
		}
	}

	if purchaseProductCount != 1 {
		t.Error("[ERROR] PurchaseProductCount should be: 1, Got: " + strconv.Itoa(purchaseProductCount))
	}

	fmt.Println("[INFO] -- TestConsumeProduct end --\n")
}

func TestConsumeProduct2(t *testing.T) {
	fmt.Println("[INFO] -- TestConsumeProduct2 start --")

	id := strconv.Itoa(products["h2oh"].ID)
	quantity := "10"
	response, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/product/"+id+"/consume/"+quantity, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] getOpenOrder: ", string(body))

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] statudCode should be: ", http.StatusOK, " Got: ", response.StatusCode)
	}

	order := getOpenOrder()

	// Verify if purchase product was added to the existing open order
	productAdded := false
	for _, prod := range order.Pproducts {
		if prod.ProductId == products["h2oh"].ID {
			productAdded = true
		}
	}

	if !productAdded {
		t.Error("[EROR] Product NOT added after consume request")
	}

	//verify if a withdrawal entry was created
	ws := getWithdrawals()
	if len(ws) == 0 {
		t.Error("[ERROR] There should have been Withdrawals on database")
	}

	wCount := 0
	for _, w := range ws {
		if w.ProductId == products["h2oh"].ID {
			wCount++
		}
	}

	if wCount != 2 {
		t.Error("[ERROR] Expected number of Withdrawals: 2, Got: " + strconv.Itoa(wCount))
	}

	// verify if the purchase products were created/updated correctly
	purchaseProductCount := 0
	pps := getPurchaseProducts()
	for _, pp := range pps {
		if (pp.ProductId == products["h2oh"].ID) && (pp.OrderId == order.ID) {
			purchaseProductCount++
		}
	}

	if purchaseProductCount != 1 {
		t.Error("[ERROR] PurchaseProductCount should be: 1, Got: " + strconv.Itoa(purchaseProductCount))
	}

	fmt.Println("[INFO] -- TestConsumeProduct2 end --\n")
}

// PUT http://127.0.0.1:8080/api/inventory/order/:id/approve
func TestApproveOrder(t *testing.T) {
	fmt.Println("[INFO] -- TestApproveOrder start --")

	order := getOpenOrder()
	id := strconv.Itoa(order.ID)
	response, err := makeRequest(router.PUT, "http://127.0.0.1:8080/api/inventory/order/"+id+"/approve", make([]byte, 1), _headers)

	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] Status code should be: ", http.StatusOK, " Got: ", response.StatusCode)
	}

	if openOrderExists() {
		t.Error("[ERROR] Open order shouldn't exist")
	}

	purchase := getPurchaseByOrderId(order.ID)
	if purchase.OrderId != order.ID || purchase.PurschaseOrder.CreatedAt != order.CreatedAt || len(purchase.PurschaseOrder.Pproducts) != len(order.Pproducts) {
		t.Error("[ERROR] No Purchase created when order was confirmed")
	}

	fmt.Println("[INFO] -- TestApproveOrder end --\n")
}

func TestTryConcludePurchaseBeforeConfirme(t *testing.T) {
	fmt.Println("[INFO] -- TestTryConcludePurchaseBeforeConfirme start --")

	openResponse, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/purchase/query/open", make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	defer openResponse.Body.Close()
	openBody, _ := ioutil.ReadAll(openResponse.Body)
	fmt.Println("[INFO] openPurchase: ", string(openBody))

	purchases := []Purchase{}
	if err := json.Unmarshal(openBody, &purchases); err != nil {
		t.Error(err)
	}

	if len(purchases) != 1 {
		t.Error("[ERROR] Number of purchases should be 1, Got: " + strconv.Itoa(len(purchases)))
	}

	id := purchases[0].ID
	// Must return 400 - BadRequest
	response, err := makeRequest(router.PUT, "http://127.0.0.1:8080/api/inventory/purchase/"+strconv.Itoa(id)+"/conclude", make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 400, Got: ", response.StatusCode)
	}

	fmt.Println("[INFO] -- TestTryConcludePurchaseBeforeConfirm end --\n")
}

func TestUpdateQuantityAndValue(t *testing.T) {
	fmt.Println("[INFO] -- TestUpdateQuantityAndValue start --")
	purchProducts := getPurchaseProducts()

	id := purchProducts[0].ID
	quantity := 1000

	// updateQuantity
	url := "http://127.0.0.1:8080/api/inventory/purchaseProduct/" + strconv.Itoa(id) + "/updateQuantity/" + strconv.Itoa(quantity)
	response, err := makeRequest(router.PUT, url, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 200, Got: ", response.StatusCode)
	}

	value := 127.27
	// updateValue
	url = "http://127.0.0.1:8080/api/inventory/purchaseProduct/" + strconv.Itoa(id) + "/updateValue/" + strconv.FormatFloat(value, 'f', 6, 64)
	response, err = makeRequest(router.PUT, url, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 200, Got: ", response.StatusCode)
	}

	fmt.Println("[INFO] -- TestUpdateQuantityAndValue end --\n")
}

func TestConfirmePurchase(t *testing.T) {
	fmt.Println("[INFO] -- TestConfirmPurchase start --")

	openResponse, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/purchase/query/open", make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	defer openResponse.Body.Close()
	openBody, _ := ioutil.ReadAll(openResponse.Body)
	fmt.Println("[INFO] openPurchase: ", string(openBody))

	purchases := []Purchase{}
	if err := json.Unmarshal(openBody, &purchases); err != nil {
		t.Error(err)
	}

	if len(purchases) != 1 {
		t.Error("[ERROR] Number of purchases should be 1, Got: " + strconv.Itoa(len(purchases)))
	}

	id := purchases[0].ID
	response, err := makeRequest(router.PUT, "http://127.0.0.1:8080/api/inventory/purchase/"+strconv.Itoa(id)+"/confirm", make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 200, Got: ", response.StatusCode)
	}

	openResponse2, err := makeRequest(router.GET, "http://127.0.0.1:8080/api/inventory/purchase/query/open", make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	defer openResponse2.Body.Close()
	openBody2, _ := ioutil.ReadAll(openResponse2.Body)
	fmt.Println("[INFO] openPurchase: ", string(openBody2))

	purchases2 := []Purchase{}
	if err := json.Unmarshal(openBody2, &purchases2); err != nil {
		t.Error(err)
	}

	if len(purchases2) != 0 {
		t.Error("[ERROR] Number of purchases should be 0, Got: " + strconv.Itoa(len(purchases)))
	}

	fmt.Println("[INFO] -- TestConfirmePurchase end --\n")
}

func TestTryUpdateQuantityAndValueWithPurchaseAlreadyConfirmed(t *testing.T) {
	fmt.Println("[INFO] -- TestTryUpdateQuantityAndValueWithPurchaseAlreadyConfirmed start --")

	purchProducts := getPurchaseProducts()

	id := purchProducts[0].ID
	quantity := 1000

	// updateQuantity
	url := "http://127.0.0.1:8080/api/inventory/purchaseProduct/" + strconv.Itoa(id) + "/updateQuantity/" + strconv.Itoa(quantity)
	response, err := makeRequest(router.PUT, url, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 400, Got: ", response.StatusCode)
	}

	value := 127.27
	// updateValue
	url = "http://127.0.0.1:8080/api/inventory/purchaseProduct/" + strconv.Itoa(id) + "/updateValue/" + strconv.FormatFloat(value, 'f', 6, 64)
	response, err = makeRequest(router.PUT, url, make([]byte, 1), _headers)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Error("[ERROR] /purchase/:id/conclude should have received status 400, Got: ", response.StatusCode)
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("[INFO] tryingUpdateAfterConfirmedResponse: ", string(body))

	fmt.Println("[INFO] -- TestTryUpdateQuantityAndValueWithPurchaseAlreadyConfirmed end --\n")
}

func TestWithdrawalBuildQuery(t *testing.T) {
	fmt.Println("[INFO] -- TestWithdrawalBuildQuery start --")
	querystring := map[string][]string{"gte": {"quantity|200", "approved_at|123456789"}, "eq": {"product_id|45", "order_id|100", "name|h2oh", "2"}}
	w := Withdrawal{Query: querystring}

	query := w.BuildQuery()
	fmt.Println("[INFO] query = ", query)

	if !strings.Contains(query, "quantity>=200") {
		t.Error("[ERROR] Withdrawal BuildQuery() 'gte'  is broken")
	}

	if !strings.Contains(query, "approved_at>=123456789") {
		t.Error("[ERROR] Withdrawal BuildQuery() 'gte' is broken")
	}

	if !strings.Contains(query, "product_id=45") {
		t.Error("[ERROR] Withdrawal BuildQuery() 'eq' is broken")
	}

	if !strings.Contains(query, "order_id=100") {
		t.Error("[ERROR] Withdrawal BuildQuery() 'eq' is broken")
	}

	if !strings.Contains(query, "name='h2oh'") {
		t.Error("[ERROR] Withdrawal BuildQuery() 'eq' is broken")
	}

	fmt.Println("[INFO] -- TestWithdrawalBuildQuery end --\n")
}
