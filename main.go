package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Coffee представляет продукт
type Coffee struct {
	ID          int
	ImageURL    string
	Title       string
	Description string
	Cost        string
	Article     string
}
type Order struct {
	OrderID   int       `db:"order_id" json:"order_id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Total     float64   `db:"total" json:"total"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

var dbPool *pgxpool.Pool

// Пример списка продуктов
var coffees = []Coffee{
	{ID: 0, ImageURL: "https://media.leverans.ru/product_images_inactive/moscow/shou-restoran-lalalend/капучино.jpg", Title: "Капучино", Description: "Кофейный напиток итальянской кухни на основе эспрессо с добавлением в него подогретого до 65 градусов вспененного молока.", Cost: "120", Article: "31343356"},
	{ID: 1, ImageURL: "https://i.pinimg.com/originals/46/fc/c2/46fcc2767aed83789f346dd310f29da3.jpg", Title: "Эспрессо", Description: "Это популярный способ приготовления кофе, который отличается небольшим размером порции и характерными слоями: тёмной массой, покрытой более светлой пенкой, называемой сливками.", Cost: "150", Article: "25142265"},
	{ID: 2, ImageURL: "https://avatars.mds.yandex.net/i?id=11af79cd4db45ddba235a09c46a16926_l-5858967-images-thumbs&n=13", Title: "Латте", Description: "Кофейный напиток на основе молока, представляющий собой трёхслойную смесь из молочной пены", Cost: "130", Article: "86551930"},
	{ID: 3, ImageURL: "https://scanformenu.ru/compiled/uploads/item_images/2f2d768c95a1a943a0d4d8b1b4b31992.jpg", Title: "Раф", Description: "Это кофейный напиток, который готовится из эспрессо, сливок и сахара. Его можно назвать кофейно-молочным коктейлем или десертом, так как он очень вкусный, сладкий и нежный, в чём-то напоминает крем-брюле.", Cost: "160", Article: "42472068"},
	{ID: 4, ImageURL: "https://lafoy.ru/photo_l/foto-2426-2.jpg", Title: "Американо", Description: "Американо готовится из одной или двух порций эспрессо, в который добавляется от 30 до 470 мл горячей воды. В процессе приготовления горячую воду можно брать как из специальной эспрессомашины, так и из отдельного чайника или подогревателя. Для обогащения вкуса в американо могут добавляться сливки или молоко, разнообразные сиропы, корица, шоколад.", Cost: "100", Article: "64816553"},
	{ID: 5, ImageURL: "https://avatars.mds.yandex.net/get-entity_search/4759071/952720682/S600xU_2x", Title: "Доппио", Description: "Кофейный напиток, который готовится как двойная порция эспрессо с помощью кофейного фильтра или эспрессо-машины.", Cost: "80", Article: "43223687"},
	{ID: 6, ImageURL: "https://avatars.mds.yandex.net/get-entity_search/1528499/952453330/S600xU_2x", Title: "Аффогато", Description: "Итальянский кофейный десерт. Его готовят так: шарик джелато (молочного мороженого) заливают чашечкой горячего эспрессо (30 мл). Бариста часто экспериментируют с ингредиентами и добавляют коньяк, ликёр или сироп. В качестве топпинга используют горький шоколад, какао-порошок, орехи, ягоды, мёд.", Cost: "200", Article: "41337060"},
	{ID: 7, ImageURL: "https://cafecentral.wien/wp-content/uploads/einspaenner_cafecentral.jpg", Title: "Венский кофе", Description: "Это сочетание крепкого чёрного кофе и пенки из взбитых сливок. Последняя аккуратно размещается на поверхности кофе без размешивания.", Cost: "230", Article: "84544447"},
	{ID: 8, ImageURL: "https://i.pinimg.com/736x/03/d9/d2/03d9d27010057294eded352af161340f.jpg", Title: "Моккачино", Description: "Это кофейный напиток, который напоминает капучино или латте, но с добавлением шоколадного соуса.", Cost: "190", Article: "89370190"},
	{ID: 9, ImageURL: "https://lafoy.ru/photo_l/foto-2426-19.jpg", Title: "Бомбон", Description: "Он состоит из эспрессо и сгущённого молока. Этот напиток прекрасно подойдёт на завтрак и зарядит бодростью и хорошим настроением на целый день.", Cost: "175", Article: "59247291"},
}

// GetOrders обрабатывает GET-запрос для получения заказов пользователя
func GetOrders(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем user_id из параметров маршрута
		idStr := r.URL.Path[len("/orders/"):]
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Выполняем запрос к базе данных
		rows, err := db.Query(context.Background(), `SELECT order_id, user_id, total, status, created_at FROM orders WHERE user_id = $1`, userID)
		if err != nil {
			http.Error(w, "Failed to query database", http.StatusInternalServerError)
			log.Println("Query error:", err)
			return
		}
		defer rows.Close()

		// Собираем список заказов
		var orders []Order
		for rows.Next() {
			var order Order
			if err := rows.Scan(&order.OrderID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt); err != nil {
				http.Error(w, "Failed to parse orders", http.StatusInternalServerError)
				log.Println("Scan error:", err)
				return
			}
			orders = append(orders, order)
		}

		// Возвращаем заказы в формате JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}

// CreateOrder обрабатывает POST-запрос для создания нового заказа
func CreateOrder(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем user_id из параметров маршрута
		idStr := r.URL.Path[len("/orders/"):]
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Декодируем тело запроса
		var order Order
		err = json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Устанавливаем user_id для заказа
		order.UserID = userID
		order.CreatedAt = time.Now()

		// Выполняем запрос на добавление заказа
		err = db.QueryRow(
			context.Background(),
			`INSERT INTO orders (user_id, total, status, created_at) 
			 VALUES ($1, $2, $3, $4) RETURNING order_id`,
			order.UserID, order.Total, order.Status, order.CreatedAt,
		).Scan(&order.OrderID)
		if err != nil {
			http.Error(w, "Failed to insert order", http.StatusInternalServerError)
			log.Println("Insert error:", err)
			return
		}

		// Возвращаем созданный заказ
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

// обработчик для GET-запроса, возвращает список продуктов
func getCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовки для правильного формата JSON
	w.Header().Set("Content-Type", "application/json")
	// Преобразуем список заметок в JSON
	json.NewEncoder(w).Encode(coffees)
}

// обработчик для POST-запроса, добавляет продукт
func createCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var newCoffee Coffee
	err := json.NewDecoder(r.Body).Decode(&newCoffee)
	if err != nil {
		fmt.Println("Error decoding request body:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Received new Coffee: %+v\n", newCoffee)
	var lastID int = len(coffees)

	for _, productCoffee := range coffees {
		if productCoffee.ID > lastID {
			lastID = productCoffee.ID
		}
	}
	newCoffee.ID = lastID + 1
	coffees = append(coffees, newCoffee)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCoffee)
}

//Добавление маршрута для получения одного продукта

func getCoffeeByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL
	idStr := r.URL.Path[len("/Coffee/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Coffee ID", http.StatusBadRequest)
		return
	}

	// Ищем продукт с данным ID
	for _, Coffee := range coffees {
		if Coffee.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Coffee)
			return
		}
	}

	// Если продукт не найден
	http.Error(w, "Coffee not found", http.StatusNotFound)
}

// удаление продукта по id
func deleteCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL
	idStr := r.URL.Path[len("/Coffee/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Coffee ID", http.StatusBadRequest)
		return
	}

	// Ищем и удаляем продукт с данным ID
	for i, Coffee := range coffees {
		if Coffee.ID == id {
			// Удаляем продукт из среза
			coffees = append(coffees[:i], coffees[i+1:]...)
			w.WriteHeader(http.StatusNoContent) // Успешное удаление, нет содержимого
			return
		}
	}

	// Если продукт не найден
	http.Error(w, "Coffee not found", http.StatusNotFound)
}

// Обновление продукта по id
func updateCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL
	idStr := r.URL.Path[len("/Coffee/update/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Coffee ID", http.StatusBadRequest)
		return
	}

	// Декодируем обновлённые данные продукта
	var updatedCoffee Coffee
	err = json.NewDecoder(r.Body).Decode(&updatedCoffee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ищем продукт для обновления
	for i, Coffee := range coffees {
		if Coffee.ID == id {

			coffees[i].ImageURL = updatedCoffee.ImageURL
			coffees[i].Title = updatedCoffee.Title
			coffees[i].Description = updatedCoffee.Description
			coffees[i].Cost = updatedCoffee.Cost

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(coffees[i])
			return
		}
	}

	// Если продукт не найден
	http.Error(w, "Coffee not found", http.StatusNotFound)
}

func main() {
	// Подключение к базе данных PostgreSQL
	db, err := pgxpool.New(context.Background(), "postgresql://postgres:03082004Lisa@localhost:5432/shop")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()
	dbPool = db
	http.HandleFunc("/orders/", GetOrders(db))
	http.HandleFunc("/orders/create/", CreateOrder(db))
	http.HandleFunc("/coffees", getCoffeeHandler)           // Получить все продукты
	http.HandleFunc("/coffee/create", createCoffeeHandler)  // Создать продукт
	http.HandleFunc("/coffee/", getCoffeeByIDHandler)       // Получить продукт по ID
	http.HandleFunc("/coffee/update/", updateCoffeeHandler) // Обновить продукт
	http.HandleFunc("/coffee/delete/", deleteCoffeeHandler) // Удалить продукт
	fmt.Println("Server is running on port 8080!")
	http.ListenAndServe(":8080", nil)
}
