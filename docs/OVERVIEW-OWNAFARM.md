# OwnaFarm - Overview Project

> **Blockchain**: Mantle Network (L2 Ethereum)  
> **Docs**: https://docs.mantle.xyz/network  

---

## Apa itu OwnaFarm?

**OwnaFarm** adalah platform yang menghubungkan **Petani** dengan **Investor** menggunakan teknologi blockchain. 

Bayangkan seperti ini:
- **Petani** butuh modal untuk menanam → mendaftarkan invoice/kebunnya ke platform
- **Investor** punya uang → membeli "bibit virtual" yang merepresentasikan invoice petani
- Ketika tanaman panen → Investor mendapat keuntungan, Petani mendapat modal

**Konsep uniknya:** Investor tidak langsung melihat invoice yang ribet, tapi bermain "game pertanian" yang menyenangkan!

---

## Smart Contracts (Mantle Sepolia Testnet)

| Contract | Address |
|----------|---------|
| GoldToken | `0x787c8616d9b8Ccdca3B2b930183813828291dA9c` |
| GoldFaucet | `0x5644F393a2480BE5E63731C30fCa81F9e80277a7` |
| OwnaFarmNFT | `0xC51601dde25775bA2740EE14D633FA54e12Ef6C7` |
| OwnaFarmVault | `0x3b561Df673F08A566A09fEd718f5bdB8018C2CDa` |

> Detail function dan cara penggunaan: [Smart Contract README](https://github.com/OwnaFarm/OwnaFarm-contract/blob/main/README.md)

---

## Struktur Project

```mermaid
graph TB
    subgraph "OwnaFarm Monorepo"
        A[Master-OwnaFarm-frontend]
        B[investor-frontend]
        C[OwnaFarm-contract]
        D[Backend - Coming Soon]
    end
    
    A --> |"Landing Page + Farmer Registration"| E[Petani]
    B --> |"Game Interface"| F[Investor]
    C --> |"Smart Contracts"| G[Mantle Blockchain]
    D --> |"API + Database"| H[PostgreSQL]
    
    style A fill:#22c55e,color:#fff
    style B fill:#3b82f6,color:#fff
    style C fill:#a855f7,color:#fff
    style D fill:#6b7280,color:#fff
```

---

## Penjelasan Setiap Komponen

### 1. Master-OwnaFarm-Frontend (Untuk Petani)

**Fungsi Utama:**
- Landing page utama OwnaFarm (halaman marketing)
- Tempat petani mendaftarkan kebun/invoice mereka

**Halaman yang ada:**
| Halaman | Fungsi |
|---------|--------|
| `/` | Landing page dengan Hero, About, How It Works, Features, Partners |
| `/register-farm` | Form pendaftaran petani (multi-step form) |

**Data yang dikumpulkan dari Petani:**

```mermaid
flowchart LR
    subgraph Step1["Step 1: Personal Info"]
        A1[fullName]
        A2[email]
        A3[phoneNumber]
        A4[idNumber/KTP]
        A5[dateOfBirth]
        A6[address]
        A7[province]
        A8[city]
        A9[district]
        A10[postalCode]
    end
    
    subgraph Step2["Step 2: Business Info"]
        B1[businessName]
        B2[businessType]
        B3[npwp]
        B4[bankName]
        B5[bankAccountNumber]
        B6[bankAccountName]
        B7[yearsOfExperience]
        B8[cropsExpertise]
    end
    
    subgraph Step3["Step 3: Documents"]
        C1[ktpPhoto]
        C2[selfieWithKtp]
        C3[npwpPhoto]
        C4[bankStatement]
        C5[landCertificate]
        C6[businessLicense]
        C7[invoiceFile - COMING SOON]
    end
    
    subgraph Step4["Step 4: Review"]
        D1[Review All Data]
        D2[Submit]
    end
    
    Step1 --> Step2 --> Step3 --> Step4
```

---

### 2. Investor-Frontend (Untuk Investor/Gamers)

**Konsep:** Aplikasi game farming seperti "Hay Day" tapi untuk investasi nyata!

**Halaman yang ada:**
| Halaman | Fungsi |
|---------|--------|
| `/` | Homepage - Menampilkan tanaman aktif, daily reward |
| `/shop` | Marketplace - Beli "bibit" (= investasi ke invoice) |
| `/farm` | My Farm - Lihat semua tanaman milik investor |
| `/leaderboard` | Papan peringkat investor |
| `/profile` | Profil investor |

**Game Mechanics:**

```mermaid
flowchart LR
    A[GOLD] --> B[Buy Seed]
    B --> C[Growing]
    C --> |"Watering +XP"| C
    C --> D[Ready to Harvest]
    D --> E[Harvest]
    E --> F[GOLD + Profit]
    E --> G[XP + Level Up]
    
    style A fill:#fbbf24,color:#000
    style B fill:#22c55e,color:#fff
    style C fill:#84cc16,color:#000
    style D fill:#f97316,color:#fff
    style E fill:#ef4444,color:#fff
    style F fill:#fbbf24,color:#000
    style G fill:#a855f7,color:#fff
```

**Data User (Investor):**
```typescript
interface UserProfile {
  name: string        // Nama user
  avatar: string      // Avatar emoji
  wallet: string      // Wallet address
  level: number       // Level game
  xp: number          // Experience points
  gold: number        // GOLD token (in-game currency)
  water: number       // Water points untuk siram tanaman
}
```

**Data Crop (Tanaman/Investasi):**
```typescript
interface Crop {
  id: string
  name: string            // Nama tanaman (contoh: "Cabai Indofood")
  image: string           // Gambar tanaman
  cctvImage: string       // Gambar CCTV kebun asli
  location: string        // Lokasi kebun
  progress: number        // Progress pertumbuhan (0-100%)
  daysLeft: number        // Sisa hari sampai panen
  yieldPercent: number    // Return yang dijanjikan (contoh: 18%)
  invested: number        // Jumlah GOLD yang diinvestasikan
  status: "growing" | "ready" | "harvested"
  plantedAt: Date         // Tanggal tanam/investasi
}
```

---

## Arsitektur Sistem

### Overview Arsitektur Lengkap

```mermaid
graph TB
    subgraph "Frontend Layer"
        FE1[Master-OwnaFarm<br/>Next.js - Port 3000]
        FE2[Investor-Frontend<br/>Next.js - Port 3001]
    end
    
    subgraph "Backend Layer"
        BE[Backend API<br/>Golang Gin - Port 8080]
        DB[(PostgreSQL)]
        CACHE[(Redis Cache)]
        STORAGE[Cloud Storage<br/>S3/GCS]
    end
    
    subgraph "Blockchain Layer"
        SC1[OwnaFarmNFT<br/>ERC-1155]
        SC2[GoldToken<br/>ERC-20]
        SC3[OwnaFarmVault]
        SC4[GoldFaucet]
        MANTLE[Mantle Network]
    end
    
    FE1 --> |REST API| BE
    FE2 --> |REST API| BE
    FE2 --> |Wagmi/Viem| MANTLE
    
    BE --> DB
    BE --> CACHE
    BE --> STORAGE
    BE --> |ethclient| MANTLE
    
    SC1 --> MANTLE
    SC2 --> MANTLE
    SC3 --> MANTLE
    SC4 --> MANTLE
    
    style FE1 fill:#22c55e,color:#fff
    style FE2 fill:#3b82f6,color:#fff
    style BE fill:#f97316,color:#fff
    style MANTLE fill:#a855f7,color:#fff
```

---

## Alur Lengkap Sistem

### Alur Petani (Farmer Flow)

```mermaid
sequenceDiagram
    participant F as Petani
    participant MF as Master Frontend
    participant BE as Backend
    participant DB as Database
    participant SC as Smart Contract
    
    F->>MF: 1. Buka website, klik "Register Farm"
    MF->>F: 2. Tampilkan form (4 steps)
    F->>MF: 3. Isi data + upload dokumen
    MF->>BE: 4. POST /api/farmers/register
    
    BE->>DB: 5. Simpan Personal Info
    BE->>DB: 6. Simpan Business Info
    BE->>DB: 7. Upload Documents ke Cloud
    BE-->>MF: 8. Return: farmerId, status: "pending"
    
    Note over BE: Admin review & approve
    
    BE->>SC: 9. approveInvoice(tokenId)
    SC-->>BE: 10. Return: txHash
    BE->>DB: 11. Update farmer status: "approved"
    BE-->>F: 12. Notifikasi: "Invoice siap didanai!"
```

### Alur Investor (Investor Flow)

```mermaid
sequenceDiagram
    participant I as Investor
    participant IF as Investor Frontend
    participant BE as Backend
    participant SC as Smart Contract
    participant W as Wallet
    
    I->>IF: 1. Buka game, connect wallet
    IF->>W: 2. Request wallet connection
    W-->>IF: 3. Connected: 0x1234...
    
    IF->>SC: 4. getAvailableInvoices()
    SC-->>IF: 5. Return list of invoices
    
    I->>IF: 6. Pilih seed, klik "Buy"
    IF->>I: 7. Tampilkan modal konfirmasi
    I->>IF: 8. Confirm purchase
    
    IF->>SC: 9. invest(tokenId, amount)
    SC->>W: 10. Request signature
    W-->>SC: 11. Transaction signed
    SC-->>IF: 12. Investment recorded
    
    IF->>BE: 13. POST /api/crops/sync
    BE-->>IF: 14. Game state updated
    
    Note over I,SC: Setelah maturity period...
    
    I->>IF: 15. Klik "Harvest"
    IF->>SC: 16. harvest(investmentId)
    SC->>W: 17. Transfer GOLD + profit
    SC-->>IF: 18. Success!
    IF->>BE: 19. POST /api/game/xp
    BE-->>IF: 20. XP & Level updated
```

---

## Pembagian Data: Database vs Smart Contract

```mermaid
graph LR
    subgraph "DATABASE (PostgreSQL)"
        direction TB
        DB1[User Profiles<br/>- name, email, avatar<br/>- XP, Level, Water points]
        DB2[Farmer Data<br/>- Personal info<br/>- Business info<br/>- Contact details]
        DB3[Documents<br/>- KTP photos<br/>- NPWP photos<br/>- Invoice files]
        DB4[Game State<br/>- Daily rewards<br/>- Login history<br/>- Achievements]
        DB5[Farm Details<br/>- Location<br/>- CCTV links<br/>- Descriptions]
        DB6[Transaction Logs<br/>- Purchase history<br/>- Withdrawal history]
    end
    
    subgraph "SMART CONTRACT (Mantle)"
        direction TB
        SC1[Invoice NFT<br/>- Token ID<br/>- Offtaker ID<br/>- Target fund<br/>- Yield %<br/>- Duration]
        SC2[Investment Record<br/>- Investor wallet<br/>- Amount invested<br/>- Purchase date]
        SC3[Funding Status<br/>- Total funded<br/>- Remaining amount<br/>- Is fully funded]
        SC4[Settlement<br/>- Maturity date<br/>- Is ready to claim<br/>- Claimed status]
        SC5[GOLD Balance<br/>- ERC-20 Token<br/>- In-game currency]
    end
    
    style DB1 fill:#3b82f6,color:#fff
    style DB2 fill:#3b82f6,color:#fff
    style DB3 fill:#3b82f6,color:#fff
    style DB4 fill:#3b82f6,color:#fff
    style DB5 fill:#3b82f6,color:#fff
    style DB6 fill:#3b82f6,color:#fff
    
    style SC1 fill:#a855f7,color:#fff
    style SC2 fill:#a855f7,color:#fff
    style SC3 fill:#a855f7,color:#fff
    style SC4 fill:#a855f7,color:#fff
    style SC5 fill:#a855f7,color:#fff
```

---

## API Communication Diagram

### Master-OwnaFarm ↔ Backend

```mermaid
sequenceDiagram
    participant FE as Master Frontend
    participant BE as Backend API
    participant DB as Database
    participant S3 as Cloud Storage
    
    FE->>BE: POST /api/farmers/register
    Note right of FE: {personalInfo, businessInfo, documents}
    
    BE->>S3: Upload document files
    S3-->>BE: Return file URLs
    
    BE->>DB: INSERT farmer_data
    DB-->>BE: Return farmerId
    
    BE-->>FE: {farmerId, status: "pending"}
```

### Investor-Frontend ↔ Smart Contract

```mermaid
sequenceDiagram
    participant FE as Investor Frontend
    participant W as Wagmi/Viem
    participant SC as Smart Contract
    
    Note over FE,SC: READ Operations (No gas)
    FE->>W: useReadContract()
    W->>SC: getAvailableInvoices()
    SC-->>W: [{tokenId, offtakerId, yield, target}]
    W-->>FE: Data cached by TanStack Query
    
    Note over FE,SC: WRITE Operations (Requires gas)
    FE->>W: useWriteContract()
    W->>SC: invest(tokenId, amount)
    Note right of SC: User signs transaction
    SC-->>W: Transaction hash
    W-->>FE: {success, txHash}
```

### Investor-Frontend ↔ Backend

```mermaid
sequenceDiagram
    participant FE as Investor Frontend
    participant BE as Backend API
    
    FE->>BE: GET /api/crops/{id}/cctv
    BE-->>FE: {cctvUrl, lastUpdated}
    
    FE->>BE: GET /api/user/stats
    BE-->>FE: {xp, level, waterPoints}
    
    FE->>BE: POST /api/daily-reward/claim
    BE-->>FE: {goldAdded: 50}
    
    FE->>BE: POST /api/game/water
    BE-->>FE: {newWaterBalance, xpGained}
```

---

**Created by:** OwnaFarm Team  
**Last Updated:** 2026-01-10  
**Version:** 2.0.0 (Smart Contracts Deployed)
