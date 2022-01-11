package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rkapps/go_finance/utils"
)

const (
	sep = ";"
)

func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Invalid arguements")
		return
	}

	inpFile := args[1]
	outFile := args[2]
	convertFile(inpFile, outFile)

}

func convertFile(inpFile string, outFile string) {

	var newLines [][]string
	lines := utils.LoadFromFile(inpFile, ",")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if len(line[0]) == 0 {
			continue
		}
		if strings.Compare(line[0], "Date") == 0 {
			continue
		}

		date, err := utils.ParseMintDate(line[0])
		if err != nil {
			fmt.Printf("Error parsing date: %v - Error: %v", line[0], err)
			break
		}

		group, category, merchant := getActivityGroupCategory(line[1], line[5], line[4])
		if strings.Compare(group, "Ignore") == 0 ||
			strings.Compare(group, "Investments") == 0 {
			continue
		}
		if strings.Compare(group, "Others") == 0 {
			fmt.Println(line)
			fmt.Printf("Date: %v group: %s category: %s\n", date, group, category)
			break
		}
		// if strings.Compare(category, "Wireless") != 0 {
		// 	continue
		// }
		var newLine []string
		newLine = append(newLine, date.Format("2006-01-02 15:04:05Z"), group, category, merchant, line[2], line[3], line[4])
		newLines = append(newLines, newLine)
	}

	fmt.Printf("Converted %d lines\n", len(newLines))
	utils.WriteToFile(outFile, newLines, sep)

}

func getActivityGroupCategory(merchant string, category string, dbcr string) (string, string, string) {
	var group string
	switch category {
	case "Gift", "Charity":
		category = "Gifts & Donations"
		group = "Miscellaneous"
		if strings.Compare("All That Glitters", merchant) == 0 {
			category = "Birthday"
			group = "Shopping"
		}
		if strings.Contains(merchant, "TIINGO.COM") {
			category = "Business Services"
			group = "Miscellaneous"
		}

	case "Air Travel", "Hotel", "Travel", "Rental Car & Taxi":
		group = "Entertainment"

		if strings.Contains(merchant, "COX & KINGS") {
			category = "Visa"
			group = "Miscellaneous"
		}

	case "Kids", "Baby Supplies", "Toys", "Piercing":

		group = "Shopping"
		if strings.Contains(merchant, "Babies") {
			category = "Toys"
		}
		if strings.Contains(merchant, "Children's") {
			category = "Stores"
		}

		if strings.Contains(merchant, "PHOTO CHECKOUT") {
			category = "Photo"
		}

	// case "Food & Dining":
	// 	group = category
	// 	category = "Pizza"

	case "Gas & Fuel", "Parking", "Public Transportation", "Service & Parts", "Auto & Transport":

		group = "Auto & Transport"
		if strings.Contains(merchant, "DMV") {
			category = "Taxes"
			group = "Miscellaneous"
		}

	case "Coffee Shops", "Fast Food", "Groceries", "Restaurants", "Alcohol & Bars", "Cake":
		group = "Food & Dining"

	case "Amusement", "Arts", "Movies & DVDs", "Music", "Newspapers & Magazines", "Television", "Streaming", "Entertainment":

		if strings.Compare(category, "Television") == 0 ||
			strings.Compare(merchant, "YouTube TV") == 0 ||
			strings.Compare(merchant, "WTVI PBS CHARLOTTE") == 0 ||
			strings.Compare(merchant, "Peacock") == 0 ||
			strings.Compare(merchant, "The Roku Channel") == 0 {
			category = "Streaming"
		}
		group = "Entertainment"
		if strings.Contains(merchant, "SMALL HANDS BIG ART") {
			group = "Activities"
		}
		if strings.Compare(merchant, "CKO*Patreon* Membership") == 0 {
			category = "Hobbies"
			group = "Activities"
		}

		if strings.Contains(merchant, "CLOUD") {
			group = "Miscellaneous"
		}

	case "Swimming", "Kids Activities", "Hobbies", "Piano", "Zoo", "Museum", "Aquarium", "Photos", "Education":

		if strings.Contains(merchant, "Young Rembrandts") {
			category = "Arts"
		}
		group = "Activities"
		// if strings.Contains(merchant, "Aqua Tots Swim") {
		// 	group = "Kids"
		// }
	case "Health & Fitness":
		category = "Doctor"
		group = "Health & Fitness"

	case "Doctor", "Dentist", "Eyecare", "Gym", "Pharmacy", "Physical Therapy", "Sports", "Health Insurance", "Naturopathy", "Labs":
		group = "Health & Fitness"

	case "Hair", "Spa & Massage", "Personal Care":

		if strings.Contains(merchant, "Pizzazz") {
			category = "Salon"
		}
		category = "Personal Care"
		group = "Health & Fitness"

	case "Home Improvement", "Home Insurance", "Home Supplies", "Home Services", "Furnishings", "Lawn & Garden":

		group = "Shopping"
		if strings.Compare("Home Services", category) == 0 {
			group = "Bills & Utilities"
		}
		if strings.Compare("Home Improvement", category) == 0 {
			group = "Miscellaneous"
		}

		// if strings.Compare("Charlotte Webbs", merchant) == 0 {
		// 	category = "Water & Sewage"
		// 	group = "Bills & Utilities"
		// }

	case "Shopping", "Books & Supplies", "Books", "Clothing", "Electronics & Software", "Sporting Goods", "Office Supplies", "Kids Bikes", "Bikes":

		if strings.Compare("Shopping", category) == 0 {
			category = "Stores"
		}
		group = "Shopping"
		if strings.Compare("Google Cloud", merchant) == 0 {
			category = "Business Services"
			group = "Miscellaneous"
		} else if strings.Compare("Scholastic", merchant) == 0 {
			// group = "Kids"
		}

	case "Home Phone", "Mobile Phone", "Phone", "Utilities", "Mortgage & Rent", "Internet", "Insurance", "Life Insurance",
		"Auto Insurance", "Auto Payment", "Babysitter & Daycare", "Water & Sewage", "Electricity", "Natural Gas":

		// if strings.Compare("AT&T", merchant) == 0 {
		// 	category = "Wireless"
		// } else if strings.Compare("Avista Corp", merchant) == 0 ||
		// 	strings.Compare("Piedmont N G", merchant) == 0 {
		// 	category = "Natural Gas"
		// } else if strings.Compare("City of Klamath Falls", merchant) == 0 ||
		// 	strings.Compare("City of Klamath Utility", merchant) == 0 ||
		// 	strings.Compare("City of Klamath Ut", merchant) == 0 ||
		// 	strings.Compare("Charlotte Webbs", merchant) == 0 {
		// 	category = "Water & Sewage"
		// } else if strings.Compare("Duke Energy", merchant) == 0 {
		// 	category = "Electricity"
		// }
		group = "Bills & Utilities"

		if strings.Contains(merchant, "Google Store") {
			category = "Devices"
			group = "Shopping"
		}

	case "Credit Card Payment", "Transfer", "Deposit", "Cash & ATM":
		group = "Ignore"

	case "Uncategorized":
		group = "Uncategorized"
		if strings.Compare("Assoc For Oral", merchant) == 0 {
			category = "Dentist"
			group = "Health & Fitness"
		}

	case "Income", "Federal Tax", "State Tax", "Paycheck", "Interest Income", "Reimbursement", "Rental Income":
		if strings.Compare("debit", dbcr) == 0 {
			group = "Miscellaneous"
		} else {
			group = "Income"
		}

		if strings.Compare("Mobile Desposit Ref", merchant) == 0 {
			group = "Ignore"
		} else if strings.Compare("Natl Fin Svc", merchant) == 0 {
			group = "Ignore"
		} else if strings.Compare("NC DES", merchant) == 0 {
			category = "Unemployment"
		} else if strings.Compare("NC State Tax", merchant) == 0 {
			// category = "State Tax Refund"
		} else if strings.Compare("Internal Revenue Service", merchant) == 0 {

		} else if strings.Compare("Dish Network", merchant) == 0 {
			category = "Reimbursement"
		} else {
			category = "Paycheck"
		}

	case "Taxes", "Tuition", "Check", "Service Fee", "Late Fee", "Cloud Services", "Business Services", "Bank Fee", "Fees & Charges", "Legal", "Shipping", "Advertising", "Printing", "Misc Expenses", "Phone Repair",
		"HSA Contribution":

		group = "Miscellaneous"

		if strings.Contains(merchant, "Google Store") {
			category = "Devices"
			group = "Shopping"
		} else if strings.Contains(merchant, "Google") {
			category = "Business Services"
			group = "Miscellaneous"
		} else if strings.Contains(merchant, "CLOUD") {
			category = "Business Services"
			group = "Miscellaneous"
		} else if strings.Contains(merchant, "TIINGO.COM") {
			category = "Business Services"
			group = "Miscellaneous"
		} else if strings.Contains(merchant, "PAY4SCHOOL") {
			category = "School Stuff"
			// group = "Kids"
		}

	case "Investments":
		if strings.Contains(merchant, "Coinbase") {
			category = "Crypto"
			group = "Investments"
		}
	default:
		log.Printf("Merchant: %s Category: %s", merchant, category)
		group = "Others"
	}

	if strings.Contains(merchant, "JELD WEN") {
		merchant = "Jeld Wen"
	} else if strings.Contains(merchant, "TRIVENI FOODCOURT") {
		merchant = "Triveni Foodcourt"
	} else if strings.Contains(merchant, "TRIVENI SUPERMARKET") {
		merchant = "Triveni Supermarket"
	} else if strings.Contains(merchant, "LOAN SRVC CNTR AUTO DRAFT") {
		merchant = "Select Portfolio"
	} else if strings.Contains(merchant, "INVITATIONHOMES") {
		merchant = "Invitation Homes"
	} else if strings.Contains(merchant, "NCDES") {
		merchant = "NCDES"
	} else if strings.Contains(merchant, "SAFE BOX ANNUAL FEE") {
		merchant = "Safe Box Annual"
	}

	return group, category, merchant
}
