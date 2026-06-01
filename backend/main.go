package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// --- Struktur Data ---

type User struct {
	Username     string `json:"username" gorm:"primaryKey" binding:"required"`
	Password     string `json:"password" binding:"required"`
	StudyMinutes int    `json:"study_minutes" gorm:"default:0"` // Diubah dari Score menjadi StudyMinutes
}

type Message struct {
	Type         string `json:"type"`
	Sender       string `json:"sender,omitempty"`
	Content      string `json:"content,omitempty"`
	StudyMinutes int    `json:"study_minutes,omitempty"` // Mengirim menit tambahan saat sesi selesai
	Timestamp    string `json:"timestamp,omitempty"`
}

// --- WebSocket Hub ---

type Hub struct {
	sync.RWMutex
	clients    map[*websocket.Conn]string
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	user string
}

var globalHub = &Hub{
	clients:    make(map[*websocket.Conn]string),
	broadcast:  make(chan Message),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client.conn] = client.user
			count := len(h.clients)
			h.Unlock()
			h.broadcastMessage(Message{Type: "online_count", Content: strconv.Itoa(count)})

		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client.conn]; ok {
				delete(h.clients, client.conn)
				client.conn.Close()
			}
			count := len(h.clients)
			h.Unlock()
			h.broadcastMessage(Message{Type: "online_count", Content: strconv.Itoa(count)})

		case message := <-h.broadcast:
			if message.Type == "update_time" { // Nama event diubah
				updateStudyTime(message.Sender, message.StudyMinutes)
				h.broadcastMessage(Message{Type: "refresh_leaderboard"})
			}

			h.RLock()
			for conn := range h.clients {
				conn.WriteJSON(message)
			}
			h.RUnlock()
		}
	}
}

func (h *Hub) broadcastMessage(msg Message) {
	h.RLock()
	defer h.RUnlock()
	for conn := range h.clients {
		conn.WriteJSON(msg)
	}
}

// --- Database Logic ---

func initDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./app.db"
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database:", err)
	}

	// GORM akan otomatis menambahkan kolom 'study_minutes' ke database SQLite kamu
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Gagal melakukan otomatis migrasi:", err)
	}
}

// Fungsi update diubah menyesuaikan kolom baru
func updateStudyTime(username string, addedMinutes int) {
	db.Model(&User{}).Where("username = ?", username).Update("study_minutes", gorm.Expr("study_minutes + ?", addedMinutes))
}

// --- Gin Handlers ---

func registerHandler(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	var existingUser User
	result := db.Where("username = ?", req.Username).First(&existingUser)

	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah terpakai"})
		return
	}

	newUser := User{Username: req.Username, Password: req.Password, StudyMinutes: 0}
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftarkan user baru"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registrasi berhasil, silakan login"})
}

func loginHandler(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	var user User
	result := db.Where("username = ?", req.Username).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Username tidak ditemukan"})
		return
	}

	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login sukses"})
}

func leaderboardHandler(c *gin.Context) {
	var leaderboard []User
	// Ambil 10 teratas berdasarkan total waktu belajar (study_minutes)
	result := db.Order("study_minutes desc").Limit(10).Find(&leaderboard)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data leaderboard"})
		return
	}
	c.JSON(http.StatusOK, leaderboard)
}

func wsHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dibutuhkan"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Gagal mengaktifkan WebSocket:", err)
		return
	}

	client := &Client{hub: globalHub, conn: conn, user: username}
	client.hub.register <- client

	go func() {
		defer func() { client.hub.unregister <- client }()
		for {
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}

			if msg.Type == "chat" {
				msg.Sender = client.user
				msg.Timestamp = time.Now().Format("15:04")
			}
			client.hub.broadcast <- msg
		}
	}()
}

func main() {
	_ = godotenv.Load()
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	initDB()
	go globalHub.run()

	r := gin.Default()
	r.Use(cors.Default())
	r.SetTrustedProxies(nil)

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/leaderboard", leaderboardHandler)
	r.GET("/ws", wsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server berjalan di port :" + port)
	r.Run(":" + port)
}
