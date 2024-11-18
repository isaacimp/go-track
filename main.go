package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	dataFile  = "foods_data.json"
	diaryFile = "diary_data.json"
)

type Food struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Calories     float64 `json:"calories"`
	Quantity     int     `json:"quantity"` // in grams
	CalPerDollar float64 `json:"cal_per_dollar"`
	CalPer100g   float64 `json:"cal_per_100g"`
}

type DiaryEntry struct {
	ID       int     `json:"id"`
	Date     string  `json:"date"`
	FoodID   int     `json:"food_id"`
	FoodName string  `json:"food_name"`
	Quantity int     `json:"quantity"` // in grams
	Calories float64 `json:"calories"`
	Cost     float64 `json:"cost"`
}

type DailyDiary struct {
	Entries []DiaryEntry `json:"entries"`
}

var (
	foods      []Food
	dailyDiary DailyDiary
)

// Save and load functions for the food database
func saveToFile() error {
	file, err := os.Create(dataFile)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(foods); err != nil {
		return fmt.Errorf("error encoding data: %v", err)
	}
	return nil
}

func loadFromFile() error {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			foods = make([]Food, 0)
			return nil
		}
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&foods); err != nil {
		return fmt.Errorf("error decoding data: %v", err)
	}
	return nil
}

// Save and load functions for the diary
func saveDiaryToFile() error {
	file, err := os.Create(diaryFile)
	if err != nil {
		return fmt.Errorf("error creating diary file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(dailyDiary); err != nil {
		return fmt.Errorf("error encoding diary data: %v", err)
	}
	return nil
}

func loadDiaryFromFile() error {
	file, err := os.Open(diaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			dailyDiary.Entries = make([]DiaryEntry, 0)
			return nil
		}
		return fmt.Errorf("error opening diary file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&dailyDiary); err != nil {
		return fmt.Errorf("error decoding diary data: %v", err)
	}
	return nil
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func addFoodToDatabase() {
	var food Food

	food.Name = readInput("Food name (e.g. potatoes): ")

	priceStr := readInput("Total Price ($): ")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		fmt.Println("Invalid price. Please enter a number.")
		return
	}
	food.Price = price

	quantityStr := readInput("Quantity (grams): ")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		fmt.Println("Invalid quantity. Please enter a number.")
		return
	}
	food.Quantity = quantity

	caloriesStr := readInput("Total calories: ")
	calories, err := strconv.ParseFloat(caloriesStr, 64)
	if err != nil {
		fmt.Println("Invalid calories. Please enter a number.")
		return
	}
	food.Calories = calories

	// Calculate derived values
	food.CalPerDollar = food.Calories / food.Price
	food.CalPer100g = (food.Calories / float64(food.Quantity)) * 100

	// Set ID based on existing foods
	maxID := 0
	for _, f := range foods {
		if f.ID > maxID {
			maxID = f.ID
		}
	}
	food.ID = maxID + 1

	foods = append(foods, food)

	// Save to file after adding new food
	if err := saveToFile(); err != nil {
		log.Printf("Warning: Failed to save data: %v", err)
	}

	fmt.Printf("\nAdded: %s\n", food.Name)
	fmt.Printf("Price: $%.2f\n", food.Price)
	fmt.Printf("Quantity: %dg\n", food.Quantity)
	fmt.Printf("Calories: %.0f\n", food.Calories)
	fmt.Printf("Calories per Dollar: %.0f\n", food.CalPerDollar)
	fmt.Printf("Calories per 100g: %.0f\n", food.CalPer100g)
}

func searchFoods() {
	query := readInput("Enter food name to search: ")
	query = strings.ToLower(query)

	found := false
	for _, food := range foods {
		if strings.Contains(strings.ToLower(food.Name), query) {
			fmt.Printf("\nFound: %s\n", food.Name)
			fmt.Printf("Price: $%.2f\n", food.Price)
			fmt.Printf("Quantity: %dg\n", food.Quantity)
			fmt.Printf("Calories: %.0f\n", food.Calories)
			fmt.Printf("Calories per Dollar: %.0f\n", food.CalPerDollar)
			fmt.Printf("Calories per 100g: %.0f\n", food.CalPer100g)
			found = true
		}
	}

	if !found {
		fmt.Println("No foods found matching your search.")
	}
}

func viewStats() {
	if len(foods) == 0 {
		fmt.Println("No foods added yet.")
		return
	}

	// Create a copy of foods to sort
	sortedFoods := make([]Food, len(foods))
	copy(sortedFoods, foods)

	fmt.Println("\nFoods ordered by calories per dollar:")
	// Simple bubble sort by CalPerDollar
	for i := 0; i < len(sortedFoods)-1; i++ {
		for j := 0; j < len(sortedFoods)-i-1; j++ {
			if sortedFoods[j].CalPerDollar < sortedFoods[j+1].CalPerDollar {
				sortedFoods[j], sortedFoods[j+1] = sortedFoods[j+1], sortedFoods[j]
			}
		}
	}

	for i, food := range sortedFoods {
		fmt.Printf("%d. %s: %.0f calories/$\n",
			i+1, food.Name, food.CalPerDollar)
	}
}

func addFoodToDiary() {
	query := readInput("Enter food name to search: ")
	query = strings.ToLower(query)

	var matchedFoods []Food
	for _, food := range foods {
		if strings.Contains(strings.ToLower(food.Name), query) {
			matchedFoods = append(matchedFoods, food)
			fmt.Printf("\n%d. %s ($%.2f per %dg, %.0f cal)\n",
				food.ID, food.Name, food.Price, food.Quantity, food.Calories)
		}
	}

	if len(matchedFoods) == 0 {
		fmt.Println("No foods found matching your search.")
		return
	}

	foodIDStr := readInput("Enter the ID of the food you want to add: ")
	foodID, err := strconv.Atoi(foodIDStr)
	if err != nil {
		fmt.Println("Invalid food ID.")
		return
	}

	var selectedFood Food
	found := false
	for _, food := range matchedFoods {
		if food.ID == foodID {
			selectedFood = food
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Food ID not found in search results.")
		return
	}

	quantityStr := readInput("Enter quantity in grams: ")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		fmt.Println("Invalid quantity.")
		return
	}

	// Calculate proportional calories and cost
	ratio := float64(quantity) / float64(selectedFood.Quantity)
	calories := selectedFood.Calories * ratio
	cost := selectedFood.Price * ratio

	// Create diary entry
	entry := DiaryEntry{
		ID:       len(dailyDiary.Entries) + 1,
		Date:     time.Now().Format("2006-01-02"),
		FoodID:   selectedFood.ID,
		FoodName: selectedFood.Name,
		Quantity: quantity,
		Calories: calories,
		Cost:     cost,
	}

	dailyDiary.Entries = append(dailyDiary.Entries, entry)
	if err := saveDiaryToFile(); err != nil {
		log.Printf("Warning: Failed to save diary: %v", err)
	}

	fmt.Printf("\nAdded to today's diary:\n")
	fmt.Printf("%s: %dg\n", selectedFood.Name, quantity)
	fmt.Printf("Calories: %.0f\n", calories)
	fmt.Printf("Cost: $%.2f\n", cost)
}

func viewDiary() {
	fmt.Println("\n=== Diary Viewer ===")
	dateStr := readInput("Enter date (YYYY-MM-DD) or press Enter for today: ")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	var dayEntries []DiaryEntry
	var totalCals float64
	var totalCost float64

	for _, entry := range dailyDiary.Entries {
		if entry.Date == dateStr {
			dayEntries = append(dayEntries, entry)
			totalCals += entry.Calories
			totalCost += entry.Cost
		}
	}

	if len(dayEntries) == 0 {
		fmt.Printf("No entries found for %s\n", dateStr)
		return
	}

	fmt.Printf("\nDiary for %s:\n", dateStr)
	fmt.Println("----------------------------------------")
	for _, entry := range dayEntries {
		fmt.Printf("%s: %dg\n", entry.FoodName, entry.Quantity)
		fmt.Printf("  Calories: %.0f, Cost: $%.2f\n", entry.Calories, entry.Cost)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("Total Calories: %.0f\n", totalCals)
	fmt.Printf("Total Cost: $%.2f\n", totalCost)
}

func showFoodMenu() {
	fmt.Println("\n=== Food Tracker Menu ===")
	fmt.Println("1. Add new food to database")
	fmt.Println("2. Add food to today's diary")
	fmt.Println("3. View diary")
	fmt.Println("4. Search foods")
	fmt.Println("5. View stats")
	fmt.Println("6. Return to Main Menu")
	fmt.Print("Choose an option: ")
}

func showMainMenu() {
	fmt.Println("\n=== Main Menu ===")
	fmt.Println("1. Food Tracking")
	fmt.Println("2. Symptom Tracking")
	fmt.Println("3. Compare")
	fmt.Println("4. Finances")
	fmt.Println("5. Exit")
	fmt.Print("Choose an Option by typing the number: ")
}

func compareTrackMenu() {
	fmt.Println("\n=== Compare Track Menu ===")
	fmt.Println("What would you like to compare:\n ")
	fmt.Println("1. Compare diet and symptoms")
	fmt.Print("Choose an Option by typing the number: ")
}

func compareDietSymptoms() {
	fmt.Println("Compare Diet and Symptoms")
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Text = "Hello World!"
	p.SetRect(0, 0, 25, 5)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}

func handleCompareMenu() {
	for {
		compareTrackMenu()
		choice := readInput("")

		switch choice {
		case "1":
			compareDietSymptoms()
		}
	}
}

func handleFoodMenu() {
	for {
		showFoodMenu()
		choice := readInput("")

		switch choice {
		case "1":
			addFoodToDatabase()
		case "2":
			addFoodToDiary()
		case "3":
			viewDiary()
		case "4":
			searchFoods()
		case "5":
			viewStats()
		case "6":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func main() {
	// Load existing data

	//symptom data
	if err := loadSymptomData(); err != nil {
		log.Printf("Warning: Failed to load existing symptom data: %v", err)
		symptomDiary.Symptoms = make([]Symptom, 0)
	}

	if err := loadSymptomDiaryData(); err != nil {
		log.Printf("Warning: Failed to load existing symptom diary data: %v", err)
		symptomDiary.Entries = make([]SymptomEntry, 0)
	}
	//food data
	if err := loadFromFile(); err != nil {
		log.Printf("Warning: Failed to load existing food data: %v", err)
		foods = make([]Food, 0)
	}

	if err := loadDiaryFromFile(); err != nil {
		log.Printf("Warning: Failed to load existing diary data: %v", err)
		dailyDiary.Entries = make([]DiaryEntry, 0)
	}

	for {
		showMainMenu()
		choice := readInput("")

		switch choice {
		case "1":
			handleFoodMenu()
		case "2":
			HandleSymptomMenu()
		case "3":
			handleCompareMenu()
		case "4":
			HandleFinanceMenu()
		case "5":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}
