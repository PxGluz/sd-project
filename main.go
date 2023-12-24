package main

import (
	"fmt"
	"os"
)

func main() {
	for {
		fmt.Println("Write one of the following command numbers:\n" +
			"1. create processor\n2. delete processor\n" +
			"3. rename processor\n4. send message\n5. exit program")

		command := 0
		_, err := fmt.Scan(&command)
		if err != nil {
			return
		}
		switch command {
		case 1:
			// TODO: replace with actual renaming function
			NewProcessor(len(processors))
		case 2:
			for {
				printProcessorsIds()
				id := -1
				_, err := fmt.Scan(&id)
				if err != nil {
					fmt.Println(err)
				} else {
					err = deleteProcessorById(id)
					if err != nil {
						fmt.Printf("%s, select a valid input.\n", err)
					} else {
						break
					}
				}
			}
		case 3:
		case 4:
			if len(processors) < 2 {
				fmt.Println("You need at least 2 processors to send and receive messages!")
				break
			}
			for {
				printProcessorsIds()

				fmt.Println("Select sender ID")
				sender_id := -1
				_, err := fmt.Scan(&sender_id)
				if err != nil {
					fmt.Println(err)
				} else {
					for {
						sender, err := getProcessorById(sender_id)

						if err != nil {
							fmt.Printf("%s, select valid input.\n", err)
							break
						}

						fmt.Println("Select receiver ID")
						receiver_id := -1
						_, err = fmt.Scan(&receiver_id)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Println("Input message")
							var content string
							_, err := fmt.Scan(&content)
							if err != nil {
								fmt.Println(err)
							} else {
								msg := Message{
									SenderID:    sender.ID,
									MessageType: 3,
									Content:     content,
								}

								sender.SendMessage(receiver_id, msg)
							}
						}
					}
				}

			}
		case 5:
			fmt.Println("Exiting application...")
			os.Exit(0)
		default:
			fmt.Println("Invalid input!")
		}
	}
}
