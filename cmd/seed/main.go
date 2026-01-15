package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/ownafarm/ownafarm-backend/internal/database"
	"github.com/ownafarm/ownafarm-backend/internal/models"
)

var walletAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

func main() {
	fmt.Println("=== Admin User Seeder ===")

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	err := database.Connect(&cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get admin details from user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter admin wallet address (0x...): ")
	walletAddress, _ := reader.ReadString('\n')
	walletAddress = strings.TrimSpace(walletAddress)

	// Validate wallet address format
	if !walletAddressRegex.MatchString(walletAddress) {
		log.Fatal("Invalid wallet address format. Must be 0x followed by 40 hex characters.")
	}

	// Normalize wallet address to checksum format
	walletAddress = strings.ToLower(walletAddress)

	fmt.Print("Enter admin role (default: admin): ")
	role, _ := reader.ReadString('\n')
	role = strings.TrimSpace(role)
	if role == "" {
		role = "admin"
	}

	// Create admin user
	admin := &models.AdminUser{
		WalletAddress: walletAddress,
		Role:          role,
		IsActive:      true,
	}

	// Check if admin with wallet address already exists
	var existingAdmin models.AdminUser
	result := database.DB.Where("wallet_address = ?", walletAddress).First(&existingAdmin)
	if result.Error == nil {
		log.Fatal("Admin with this wallet address already exists")
	}

	// Insert into database
	result = database.DB.Create(admin)
	if result.Error != nil {
		log.Fatal("Failed to create admin user:", result.Error)
	}

	fmt.Println("\n✅ Admin user created successfully!")
	fmt.Printf("   ID:             %s\n", admin.ID)
	fmt.Printf("   Wallet Address: %s\n", admin.WalletAddress)
	fmt.Printf("   Role:           %s\n", admin.Role)
}
