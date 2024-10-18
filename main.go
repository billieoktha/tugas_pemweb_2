// Tugas Pemweb 2

package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

// Struktur untuk model User
type User struct {
    ID   int    `json:"id"`   // Field ID dengan tipe data int
    Name string `json:"name"` // Field Name dengan tipe data string
}

type Message struct {
    Message string `json:"Message"` // Struktur untuk pesan saat user dihapus
}

var db *sql.DB // Variabel global untuk koneksi database

// Middleware BasicAuth untuk autentikasi dasar
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth() // Mendapatkan username dan password dari request

        // Jika autentikasi gagal, kembalikan respon Unauthorized
        if !ok || !validateCredentials(username, password) {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Jika berhasil, lanjutkan ke handler berikutnya
        next.ServeHTTP(w, r)
    })
}

// Fungsi untuk memvalidasi username dan password (bisa diganti dengan cek ke database)
func validateCredentials(username, password string) bool {
    return username == "admin" && password == "tugas3" // Validasi kredensial sederhana
}

// Mendapatkan semua user dari database (READ)
func getUsers(w http.ResponseWriter, r *http.Request) {
    // Menjalankan query SQL untuk mengambil data user
    rows, err := db.Query("SELECT id, name FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close() // Pastikan untuk menutup koneksi setelah selesai

    var users []User // Slice untuk menyimpan daftar user
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name) // Scan hasil query ke dalam struct User
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        users = append(users, user) // Tambahkan user ke slice
    }

    // Mengembalikan hasil dalam format JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

// Membuat user baru (CREATE)
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    // Decode data JSON yang dikirimkan ke dalam struct User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Menjalankan query SQL untuk menambahkan user baru
    result, err := db.Exec("INSERT INTO users (name) VALUES (?)", user.Name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mendapatkan ID user yang baru ditambahkan
    id, err := result.LastInsertId()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user.ID = int(id) // Set ID user yang baru dibuat
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user) // Kembalikan data user dalam format JSON
}

// Memperbarui data user (UPDATE)
func updateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)  // Mendapatkan parameter dari URL
    id := vars["id"]     // Mendapatkan ID user dari URL

    var user User
    // Decode data JSON yang dikirimkan ke dalam struct User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Menjalankan query SQL untuk memperbarui user berdasarkan ID
    _, err = db.Exec("UPDATE users SET name = ? WHERE id = ?", user.Name, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user.ID, _ = strconv.Atoi(id) // Set ID user yang diperbarui
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user) // Kembalikan data user yang diperbarui dalam format JSON
}

// Menghapus user (DELETE)
func deleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)  // Mendapatkan parameter dari URL
    id := vars["id"]     // Mendapatkan ID user dari URL

    // Menjalankan query SQL untuk menghapus user berdasarkan ID
    _, err := db.Exec("DELETE FROM users WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Kembalikan pesan konfirmasi dalam format JSON
    var message Message
    message = Message{Message: "Berhasil Di Hapus"}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(message)
}

// Fungsi utama untuk menyiapkan router dan menjalankan server
func main() {
    // String koneksi database untuk XAMPP
    dsn := "root:@tcp(127.0.0.1:3306)/tugas_pemweb" // Koneksi ke MySQL di XAMPP
    var err error
    db, err = sql.Open("mysql", dsn) // Buka koneksi ke database
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }
    defer db.Close() // Pastikan koneksi database ditutup setelah selesai

    // Menyiapkan router
    router := mux.NewRouter()

    // Rute untuk pengecekan status server
    router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }).Methods("GET")

    // Melindungi rute dengan middleware basicAuth
    router.HandleFunc("/pengguna", basicAuth(getUsers)).Methods("GET")
    router.HandleFunc("/pengguna", basicAuth(createUser)).Methods("POST")
    router.HandleFunc("/pengguna/{id}", basicAuth(updateUser)).Methods("PUT")
    router.HandleFunc("/pengguna/{id}", basicAuth(deleteUser)).Methods("DELETE")

    // Menjalankan server di port 8080
    fmt.Println("Server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
