package functions

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/mr-tron/base58"
	"log"
	"strings"
)

// SendSOL 1 SOL = 1 000 000 000 Lamports
func SendSOL(amount uint64, privateKeyString string, recipientAddress string, rpcURL string) {
	if amount < 10000 {
		log.Fatalf("Cannot send less than 1/100_000 SOL")
	}

	rpcEndpoint := RecognizeRPC(rpcURL)

	c := client.NewClient(rpcEndpoint)

	privateKeyBytes, _ := base58.Decode(privateKeyString)
	feePayer, _ := types.AccountFromBytes(privateKeyBytes)

	//recipientBytes, _ := base58.Decode(recipientAddress)
	//recipientAccount, _ := types.AccountFromBytes(recipientBytes)
	recipientAccount := common.PublicKeyFromString(recipientAddress)

	resp, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feePayer.PublicKey,
			RecentBlockhash: resp.Blockhash,
			Instructions: []types.Instruction{
				system.Transfer(system.TransferParam{
					From:   feePayer.PublicKey,
					To:     recipientAccount,
					Amount: amount,
				}),
			},
		}),
		Signers: []types.Account{feePayer},
	})
	if err != nil {
		log.Fatalf("An error: %v", err)
	}
	sig, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to send tx: %v", err)
	}
	fmt.Println(sig)
}

func getAirdrop(address string, endpoint string) {
	var c *client.Client
	switch endpoint {
	case "1":
		c = client.NewClient(rpc.TestnetRPCEndpoint)
	case "2":
		c = client.NewClient(rpc.DevnetRPCEndpoint)
	default:
		log.Fatalf("Invalid mode, choose either 1 or 2.")
	}

	sig, err := c.RequestAirdrop(
		context.TODO(),
		address,
		1e9)
	if err != nil {
		log.Fatalf("An error occured (airdrop request), error: %v ", err)
	} else {
		fmt.Printf("Airdropped, sig %v ", sig)
	}
}

func GetAirdrops(wallets []string, mode string) {
	for _, addy := range wallets {
		getAirdrop(addy, mode)
	}
}

// convertLamports Mode 1 = SOL to lamports, mode 2 = lamports to SOL
func convertLamportsToSOL(amount float64) float64 {
	// 1 SOL = 1 000 000 000 Lamports
	const lamports = 1_000_000_000
	return amount / lamports
}

func convertSOLtoLamports(amount float64) uint64 {
	const lamports = 1_000_000_000
	return uint64(amount * lamports)
}

func DistributeSOL(amount float64, mainWallets []wallet, burners []string, rpcURL string) {
	burnersAmount := len(burners)
	amountPerWallet := convertSOLtoLamports(amount / float64(burnersAmount))

	iter := 1
	var choices []int
	for _, aliasPKpair := range mainWallets {
		fmt.Printf("[%v] %v", iter, aliasPKpair.Alias)
		fmt.Println("")

		choices = append(choices, iter)
		iter++
	}

	var option int
	for {

		fmt.Print("Enter your choice: ")
		_, err := fmt.Scan(&option)
		if err != nil {
			log.Fatalf("An error occured: %v", err)
		}
		if isInList(choices, option) == false {
			continue
		} else {
			break
		}
	}
	alias := mainWallets[option-1].Alias
	privateKey, err := findPrivateKeyByAlias(mainWallets, alias)
	if err != nil {
		log.Fatalf("Error, cannot find Private Key by Alias")
	}

	for _, burner := range burners {
		SendSOL(amountPerWallet, privateKey, burner, rpcURL)
	}
}

func RecognizeRPC(input string) string {
	lowered := strings.ToLower(input)
	switch lowered {
	case "testnet":
		return "https://api.testnet.solana.com"
	case "devnet":
		return "https://api.devnet.solana.com"
	default:
		return input
	}
}
