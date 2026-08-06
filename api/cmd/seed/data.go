package main

// seedUser is a demo account to insert directly into the DB (bypassing
// register + email verification, since Resend rejects fake domains).
type seedUser struct {
	fullName string
	email    string
	isAdmin  bool
}

const seedPassword = "password1234"

var seedUserList = []seedUser{
	{fullName: "Admin Toko", email: "admin@go-commerce.local", isAdmin: true},
	{fullName: "Budi Santoso", email: "customer1@go-commerce.local"},
	{fullName: "Siti Aminah", email: "customer2@go-commerce.local"},
	{fullName: "Rizky Pratama", email: "customer3@go-commerce.local"},
}

// seedProduct is a demo catalog entry. sku doubles as the idempotency key
// (products.sku has a unique index for active rows) and as the picsum.photos
// seed, so each product gets a stable, distinct placeholder image.
type seedProduct struct {
	sku      string
	name     string
	desc     string
	price    string
	stock    int32
	category string
	inactive bool // soft-deleted right after creation, for admin demo
}

var seedProductList = []seedProduct{
	// Kaos
	{sku: "KAOS-001", name: "Kaos Polos Hitam", desc: "Kaos katun combed 30s, nyaman dipakai harian.", price: "89000", stock: 40, category: "Kaos"},
	{sku: "KAOS-002", name: "Kaos Polos Putih", desc: "Kaos katun combed 30s, warna putih bersih.", price: "89000", stock: 35, category: "Kaos"},
	{sku: "KAOS-003", name: "Kaos Oversize Navy", desc: "Potongan oversize, bahan tebal tidak menerawang.", price: "125000", stock: 18, category: "Kaos"},
	{sku: "KAOS-004", name: "Kaos Grafis Streetwear", desc: "Desain grafis eksklusif, bahan cotton combed 24s.", price: "149000", stock: 0, category: "Kaos"},
	{sku: "KAOS-005", name: "Kaos Polo Pique", desc: "Kerah polo bahan pique, cocok untuk semi-formal.", price: "179000", stock: 22, category: "Kaos"},
	{sku: "KAOS-006", name: "Kaos Lengan Panjang Rib", desc: "Lengan panjang rib fit, hangat dan stylish.", price: "135000", stock: 15, category: "Kaos"},

	// Celana
	{sku: "CLN-001", name: "Celana Chino Slim Fit", desc: "Chino slim fit, cocok untuk kerja maupun santai.", price: "219000", stock: 25, category: "Celana"},
	{sku: "CLN-002", name: "Celana Jeans Straight", desc: "Denim straight cut, awet dan tidak mudah pudar.", price: "259000", stock: 30, category: "Celana"},
	{sku: "CLN-003", name: "Celana Jogger Katun", desc: "Jogger katun stretch, elastis di pinggang dan pergelangan.", price: "189000", stock: 20, category: "Celana"},
	{sku: "CLN-004", name: "Celana Cargo Tactical", desc: "Banyak kantong, bahan ripstop tahan lama.", price: "249000", stock: 12, category: "Celana"},
	{sku: "CLN-005", name: "Celana Pendek Denim", desc: "Denim pendek casual, cocok dipakai musim panas.", price: "159000", stock: 28, category: "Celana"},
	{sku: "CLN-006", name: "Celana Kulot Wanita", desc: "Kulot bahan wolfis, jatuh dan tidak panas.", price: "199000", stock: 10, category: "Celana", inactive: true},

	// Sepatu
	{sku: "SPT-001", name: "Sepatu Sneakers Canvas", desc: "Sneakers canvas ringan, cocok untuk aktivitas harian.", price: "279000", stock: 33, category: "Sepatu"},
	{sku: "SPT-002", name: "Sepatu Running Mesh", desc: "Mesh breathable, sol empuk untuk lari jarak jauh.", price: "349000", stock: 27, category: "Sepatu"},
	{sku: "SPT-003", name: "Sepatu Slip-On Casual", desc: "Slip-on praktis tanpa tali, mudah dipakai lepas.", price: "229000", stock: 19, category: "Sepatu"},
	{sku: "SPT-004", name: "Sepatu Boots Kulit", desc: "Kulit asli, cocok untuk gaya rugged.", price: "459000", stock: 0, category: "Sepatu"},
	{sku: "SPT-005", name: "Sandal Selop Pria", desc: "Selop kulit sintetis, ringan dan anti selip.", price: "99000", stock: 45, category: "Sepatu"},
	{sku: "SPT-006", name: "Sepatu Sekolah Hitam", desc: "Sepatu hitam polos, sesuai standar seragam sekolah.", price: "189000", stock: 16, category: "Sepatu"},

	// Aksesoris
	{sku: "AKS-001", name: "Topi Baseball Cap", desc: "Topi baseball adjustable, bahan katun twill.", price: "79000", stock: 50, category: "Aksesoris"},
	{sku: "AKS-002", name: "Tas Selempang Kanvas", desc: "Kanvas tebal, muat dompet dan gadget harian.", price: "159000", stock: 24, category: "Aksesoris"},
	{sku: "AKS-003", name: "Ikat Pinggang Kulit", desc: "Kulit sapi asli, gesper anti karat.", price: "129000", stock: 21, category: "Aksesoris"},
	{sku: "AKS-004", name: "Kacamata Hitam UV400", desc: "Lensa UV400, melindungi mata dari sinar matahari.", price: "99000", stock: 38, category: "Aksesoris"},
	{sku: "AKS-005", name: "Dompet Kulit Pria", desc: "Dompet kulit lipat tiga, banyak slot kartu.", price: "149000", stock: 17, category: "Aksesoris"},
	{sku: "AKS-006", name: "Jam Tangan Digital", desc: "Jam tangan digital tahan air, tampilan minimalis.", price: "199000", stock: 14, category: "Aksesoris", inactive: true},
}
