package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type Report struct {
	Date       string
	ItemsSold  []CartItem
	TotalSales float64
}

type CartItem struct {
	Item     Item
	Quantity int
}

// === Fungsi Utilitas yang sudah ada ===

func loadItems() ([]Item, error) {
	data, err := os.ReadFile("items.json")
	if err != nil {
		return nil, err
	}

	var items []Item
	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func loadReports() ([]Report, error) {
	data, err := os.ReadFile("report.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []Report{}, nil
		}
		return nil, err
	}

	var reports []Report
	err = json.Unmarshal(data, &reports)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func saveReports(reports []Report) {
	data, _ := json.MarshalIndent(reports, "", "  ")
	_ = os.WriteFile("report.json", data, 0644)
}

func showPurchaseHistory() {
	clearScreen()

	reports, err := loadReports()
	if err != nil || len(reports) == 0 {
		fmt.Println("Belum ada riwayat pembelian.")
		fmt.Println("\nTekan ENTER untuk kembali ke menu...")
		fmt.Scanln()
		fmt.Scanln()
		return
	}

	var grandTotal float64
	for _, r := range reports {
		grandTotal += r.TotalSales
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("Total Penjualan: Rp.%.0f\n", grandTotal)
	fmt.Println("--------------------------------")

	for _, r := range reports {
		for _, item := range r.ItemsSold {
			fmt.Printf(
				"%s (%s) x%d - [Rp.%.0f]\n",
				r.Date,
				item.Item.Name,
				item.Quantity,
				item.Item.Price*float64(item.Quantity),
			)
		}
		fmt.Println()
	}

	fmt.Println("--------------------------------")
	fmt.Println("Tekan ENTER untuk kembali ke menu...")
	fmt.Scanln()
}




func saveItems(items []Item) {
	data, _ := json.MarshalIndent(items, "", "  ")
	_ = os.WriteFile("items.json", data, 0644)
}

func showMenu(items []Item) {
	fmt.Println("===== LIST BARANG =====")
	for _, item := range items {
		fmt.Printf("%d. %s - Rp %.0f (Stock: %d)\n", item.ID, item.Name, item.Price, item.Stock)
	}
	fmt.Println("========================")
}

func showCart(cart []CartItem) {
	if len(cart) == 0 {
		fmt.Println("Keranjang kosong.")
		return
	}
	fmt.Println("--- Keranjang Saat Ini ---")
	for i, c := range cart {
		sub := c.Item.Price * float64(c.Quantity)
		fmt.Printf("[%d] %s x%d = Rp %.0f\n", i+1, c.Item.Name, c.Quantity, sub)
	}
	fmt.Println("--------------------------")
}

func findItemByID(items []Item, id int) (*Item, bool) {
	for i := range items {
		if items[i].ID == id {
			return &items[i], true
		}
	}
	return nil, false
}

func saveReceipt(cart []CartItem, total float64) {
	now := time.Now()
	// Gunakan format yang lebih aman untuk nama file
	fileName := "receipt_" + now.Format("20060102_150405") + ".txt"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Gagal menyimpan struk: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("===== STRUK BELANJA =====\n")
	file.WriteString(now.Format("02/01/2006 15:04:05\n"))
	file.WriteString("-------------------------\n")

	for _, c := range cart {
		line := fmt.Sprintf("%s x%d = Rp %.0f\n", c.Item.Name, c.Quantity, c.Item.Price*float64(c.Quantity))
		file.WriteString(line)
	}

	file.WriteString(fmt.Sprintf("\nTOTAL: Rp %.0f\n", total))
	file.WriteString("==========================\n")
	fmt.Println("Struk berhasil disimpan sebagai:", fileName)
}

// === FUNGSI BARU: Hapus Item ===

// removeItem menghapus item dari keranjang berdasarkan indeks
// dan mengembalikan stok item yang dihapus ke daftar item utama.
func removeItem(cart *[]CartItem, items []Item, index int) {
	// Pastikan indeks valid
	if index < 0 || index >= len(*cart) {
		fmt.Println("Indeks item tidak valid!")
		return
	}

	itemToRemove := (*cart)[index]
	
	// 1. Kembalikan stok yang dikurangi sebelumnya
	// Cari item yang sesuai di daftar barang utama (items)
	for i := range items {
		if items[i].ID == itemToRemove.Item.ID {
			items[i].Stock += itemToRemove.Quantity
			fmt.Printf("Stok %s dikembalikan sebanyak %d.\n", items[i].Name, itemToRemove.Quantity)
			break
		}
	}

	// 2. Hapus item dari slice cart
	// [0:index] -> elemen sebelum indeks, [index+1:] -> elemen setelah indeks
	*cart = append((*cart)[:index], (*cart)[index+1:]...)
	fmt.Println("Item berhasil dihapus dari keranjang.")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}


// === Fungsi Utama ===

func main() {
	for{
		clearScreen()
		fmt.Println("===== MENU =====")
		fmt.Println("[a] Aplikasi Kasir")
		fmt.Println("[b] Lihat Riwayat Pembelian")
		fmt.Println("[q] Keluar")

		fmt.Print("Pilih menu: ")

		var menu string
		fmt.Scan(&menu)

		if menu == "b" {
			showPurchaseHistory()
			continue
		}

		if menu == "q" {
			fmt.Println("Terima kasih!")
			return
		}

		if menu != "a" {
			fmt.Println("Menu tidak valid!")
			time.Sleep(1 * time.Second)
			continue
		}

		items, err := loadItems()
		if err != nil {
			fmt.Println("Gagal memuat items.json:", err)
			return
		}

		var cart []CartItem
		// itemMap digunakan untuk memudahkan update stock item asli di loop
		// PtrMap akan menunjuk ke item *di dalam slice items*
		itemMap := make(map[int]*Item)
		for i := range items {
			itemMap[items[i].ID] = &items[i]
		}

		// Loop utama untuk penambahan barang
		for {
			showMenu(items)

			var id int
			fmt.Print("Pilih ID barang (0 untuk Selesai, -1 untuk Hapus Item): ")
			fmt.Scan(&id)

			// A. Pilihan Selesai Transaksi
			if id == 0 {
				break
			}
			
			// B. Pilihan Hapus Item
			if id == -1 {
				if len(cart) == 0 {
					fmt.Println("Keranjang kosong, tidak ada yang bisa dihapus.")
					continue
				}
				showCart(cart)
				var index int
				fmt.Print("Masukkan nomor item di keranjang yang ingin dihapus (1, 2, dst.): ")
				fmt.Scan(&index)
				// Panggil fungsi hapus, index - 1 karena tampilan mulai dari 1
				removeItem(&cart, items, index-1)
				continue
			}

			// C. Pilihan Tambah Item (ID > 0)
			
			itemPtr, found := itemMap[id] // Gunakan map untuk mendapatkan pointer
			if !found {
				fmt.Println("ID barang tidak ditemukan!")
				continue
			}

			if itemPtr.Stock <= 0 {
				fmt.Println("Stock habis!")
				continue
			}

			var qty int
			fmt.Print("Jumlah: ")
			fmt.Scan(&qty)

			if qty > itemPtr.Stock {
				fmt.Println("Stock tidak cukup!")
				continue
			}

			// Kurangi stock (Menggunakan pointer ke item asli)
			itemPtr.Stock -= qty
			saveItems(items)


			// Cek apakah item sudah ada di keranjang untuk digabungkan
			itemFoundInCart := false
			for i := range cart {
				if cart[i].Item.ID == id {
					cart[i].Quantity += qty
					itemFoundInCart = true
					break
				}
			}

			if !itemFoundInCart {
				// Tambahkan item baru ke keranjang
				// Catatan: *itemPtr di-dereference untuk membuat salinan Item.
				cart = append(cart, CartItem{Item: *itemPtr, Quantity: qty})
			}
			
			fmt.Println("Barang ditambahkan ke keranjang!")
			showCart(cart) // Tampilkan keranjang setelah penambahan
			fmt.Println()
		}

		// 1. Ringkasan & Total
		fmt.Println("\n===== PROSES FINAL TRANSAKSI =====")
		if len(cart) == 0 {
			fmt.Println("Keranjang kosong. Transaksi dibatalkan.")
			// Pastikan stock tidak tersimpan jika transaksi batal (walaupun sudah diurus di removeItem)
			return
		}

		fmt.Println("\n===== RINGKASAN BELANJA =====")
		var total float64
		for _, c := range cart {
			sub := c.Item.Price * float64(c.Quantity)
			fmt.Printf("%s x%d = Rp %.0f\n", c.Item.Name, c.Quantity, sub)
			total += sub
		}
		fmt.Printf("\nTOTAL: Rp %.0f\n", total)

		// 2. Simpan stock terbaru
		saveItems(items)

		// 3. Simpan struk
		now := time.Now()
		saveReceipt(cart, total)
		reports, _ := loadReports()
		newReport := Report{
			Date:       now.Format("02/01/2006 15:04:05"),
			ItemsSold:  cart,
			TotalSales: total,
		}

		reports = append(reports, newReport)
		saveReports(reports)


		fmt.Println("\nTransaksi Selesai. Terima kasih!")
		fmt.Scanln()
		fmt.Scanln()
	}
}