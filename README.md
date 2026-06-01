# 🚀 Realtime Study App

Sebuah aplikasi produktivitas berbasis **Real-time** yang dirancang untuk membantu pengguna fokus belajar bersama secara virtual. Dilengkapi dengan fitur Timer Pomodoro, Global Chat, dan Leaderboard kompetitif dengan antarmuka bergaya **Neon**.

🌐 **Cobain Live Demo:** [realtime-study-app.wrdt.my.id](https://study-app.wrdt.my.id)

Aplikasi ini menggunakan arsitektur **Monorepo**, di mana Backend (Golang) dan Frontend (Flutter) berada di dalam satu *repository*.

---

## ✨ Fitur Utama

- ⏱️ **Focus Timer (25 Menit)**: Penghitung waktu mundur yang dilengkapi dengan suara alarm (*beep*) ketika sesi fokus selesai.
- 💬 **Global Lounge (Real-time Chat)**: Obrolan langsung antar pengguna tanpa perlu me-*refresh* halaman, didukung oleh protokol WebSocket.
- 🏆 **Live Leaderboard**: Papan peringkat pengguna dengan total waktu belajar tertinggi yang ter-*update* secara instan.
- 🟢 **Online Counter**: Menampilkan jumlah pengguna yang sedang *online* di dalam ruang belajar secara *real-time*.
- 🎨 **Neon UI**: Desain antarmuka *Dark Mode* modern dengan sentuhan gradasi warna *Neon Blue* dan *Neon Pink*.

---

## 🛠️ Teknologi yang Digunakan

**Backend:**
- **[Go (Golang)](https://go.dev/)**: Bahasa pemrograman utama.
- **[Gin Web Framework](https://gin-gonic.com/)**: Untuk *routing* REST API.
- **[Gorilla WebSocket](https://github.com/gorilla/websocket)**: Untuk komunikasi dua arah secara *real-time*.
- **[GORM](https://gorm.io/)**: ORM (*Object Relational Mapping*) untuk mengelola database.
- **SQLite**: Database ringan untuk menyimpan data *user* dan total waktu belajar.

**Frontend:**
- **[Flutter](https://flutter.dev/)**: Framework UI lintas platform.
- **[web_socket_channel](https://pub.dev/packages/web_socket_channel)**: Penghubung klien WebSocket di Flutter.
- **[http](https://pub.dev/packages/http)**: Untuk melakukan *request* ke REST API (Login & Register).
- **[audioplayers](https://pub.dev/packages/audioplayers)**: Untuk memutar suara alarm saat sesi fokus selesai.

---

## 📁 Struktur Direktori (Monorepo)

```text
realtime-study-app/
│
├── backend/                # Source code Golang Server
│   ├── main.go             # Entry point dan logika utama server
│   ├── go.mod              # Dependency manager Go
│   └── app.db              # File database SQLite (ter-generate otomatis)
│
├── frontend/               # Source code Flutter Client
│   ├── lib/
│   │   └── main.dart       # Entry point Flutter & rancangan UI
│   ├── pubspec.yaml        # Dependency manager Flutter
│   └── ...
│
└── README.md
