package cli

import (
	"flag"
	"fmt"
	"log"
	"time"

	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createtokoin -address ADDRESS - create a tokoin for ADDRESS")
	fmt.Println("  editpolicy -address ADDRESS -txid TXID -time TIME -id ID -gps GPS -temperature TEMPERATURE - edit parameters for a tokoin ")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexurpo - Rebuilds the URPO set")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println("  listtokoins -address ADDRESS - List all tokoins belonging to ADDRESS")
	fmt.Println("  deposit -address ADDRESS -holder HOLDER -txid TXID - set a holder for a tokoin")
	fmt.Println("  revocat -address ADDRESS -txid TXID - revocat a tokoin")
	fmt.Println("  redeem -holder HOLDER -owner OWNER -txid TXID -time TIME -id ID -gps GPS -temper TEMPERATURE - redeem a tokoin with the holder address and current condition")
	fmt.Println("  test -flag FLAG -owner OWNER -holder HOLDER -txid TXID -time TIME -id ID -gps GPS -temper TEMPERATURE")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("NODE_ID env. var is not set!")
		os.Exit(1)
	}

	createTokoinCmd := flag.NewFlagSet("createtokoin", flag.ExitOnError)
	editPolicyCmd := flag.NewFlagSet("editpolicy", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexURPOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	listTokoinsCmd := flag.NewFlagSet("listtokoins", flag.ExitOnError)
	depositCmd := flag.NewFlagSet("deposit", flag.ExitOnError)
	revocatCmd := flag.NewFlagSet("revocat", flag.ExitOnError)
	redeemCmd := flag.NewFlagSet("redeem", flag.ExitOnError)
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)

	createTokoinAddress := createTokoinCmd.String("address", "", "The address to mint")
	editPolicyAddress := editPolicyCmd.String("address", "", "The address to edit")
	editPolicyTxId := editPolicyCmd.String("txid", "", "The txid of the edited tokoin")
	editPolicyTime := editPolicyCmd.String("time", "", "The new time for the tokoin")
	editPolicyId := editPolicyCmd.String("id", "", "The new ID for the tokoin")
	editPolicyGPS := editPolicyCmd.String("gps", "", "The new GPS for the tokoin")
	editPolicyTemper := editPolicyCmd.String("temperature", "", "The new temperature for the tokoin")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	listTokoinsAddress := listTokoinsCmd.String("address", "", "The address to list tokoins for")
	depositAddress := depositCmd.String("address", "", "The address of tokoin holder")
	depositHolder := depositCmd.String("holder", "", "The address of tokoin holder")
	depositTxId := depositCmd.String("txid", "", "The txid of the deposited tokoin")
	revocatAddress := revocatCmd.String("address", "", "The address of the tokoin owner")
	revocatTxId := revocatCmd.String("txid", "", "The txid of the revocated tokoin")
	redeemHolder := redeemCmd.String("address", "", "The address of the tokoin holder")
	redeemOwner := redeemCmd.String("owner", "", "The address of the tokoin owner")
	redeemTxId := redeemCmd.String("txid", "", "The txid of the redeemed tokoin")
	redeemTime := redeemCmd.String("time", "", "The time condition of redemption")
	redeemId := redeemCmd.String("id", "", "The ID condition of redemption")
	redeemGPS := redeemCmd.String("gps", "", "The GPS condition of redemption")
	redeemTemper := redeemCmd.String("temper", "", "The temperature condition of redemption")
	testFlag := testCmd.String("flag", "", "The type of the test")
	testOwner := testCmd.String("owner", "", "The owner of the tokoin")
	testHolder := testCmd.String("holder", "", "The holder of the tokoin")
	testTxid := testCmd.String("txid", "", "The txid of the tokoin")
	testTime := testCmd.String("time", "", "The target/current time condition")
	testID := testCmd.String("id", "", "The tartget/current ID condition")
	testGPS := testCmd.String("gps", "", "The target/current GPS condition")
	testTemper := testCmd.String("temper", "", "The target/current temperature condition")

	fmt.Println("cli in - ", time.Now())
	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexurpo":
		err := reindexURPOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createtokoin":
		err := createTokoinCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "editpolicy":
		err := editPolicyCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listtokoins":
		err := listTokoinsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "deposit":
		err := depositCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "revocat":
		err := revocatCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "redeem":
		err := redeemCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if reindexURPOCmd.Parsed() {
		cli.reindexURPO(nodeID)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID, *startNodeMiner)
	}

	if createTokoinCmd.Parsed() {
		if *createTokoinAddress == "" {
			createTokoinCmd.Usage()
			os.Exit(1)
		}
		cli.createTokoin(*createTokoinAddress, nodeID)
	}

	if editPolicyCmd.Parsed() {
		if *editPolicyTxId == "" || *editPolicyAddress == "" {
			editPolicyCmd.Usage()
			os.Exit(1)
		}
		cli.editPolicy(*editPolicyAddress, *editPolicyTxId, nodeID, *editPolicyTime, *editPolicyId, *editPolicyGPS, *editPolicyTemper)
	}

	if listTokoinsCmd.Parsed() {
		if *listTokoinsAddress == "" {
			listTokoinsCmd.Usage()
			os.Exit(1)
		}
		cli.listTokoins(*listTokoinsAddress, nodeID)
	}

	if depositCmd.Parsed() {
		if *depositAddress == "" || *depositHolder == "" || *depositTxId == "" {
			depositCmd.Usage()
			os.Exit(1)
		}
		cli.deposit(*depositAddress, *depositHolder, *depositTxId, nodeID)
	}

	if revocatCmd.Parsed() {
		if *revocatAddress == "" || *revocatTxId == "" {
			revocatCmd.Usage()
			os.Exit(1)
		}
		cli.revocat(*revocatAddress, *revocatTxId, nodeID)
	}

	if redeemCmd.Parsed() {
		if *redeemHolder == "" || *redeemOwner == "" || *redeemTxId == "" {
			redeemCmd.Usage()
			os.Exit(1)
		}
		cli.redeem(*redeemHolder, *redeemOwner, *redeemTxId, nodeID, *redeemTime, *redeemId, *redeemGPS, *redeemTemper)
	}

	if testCmd.Parsed() {
		cli.test(nodeID, *testFlag, *testOwner, *testHolder, *testTxid, *testTime, *testID, *testGPS, *testTemper)
	}

	fmt.Println("cli out - ", time.Now())
}
