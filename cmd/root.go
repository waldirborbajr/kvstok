package cmd

import (
	"log"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
	"github.com/waldirborbajr/kvstok/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kvstok",
	Short:   "KVStoK is a CLI-based KEY VALUE storage.",
	Long:    ``,
	Version: version.AppVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	must.Must(rootCmd.Execute(), "Execute() on parsing commands.")
}

func init() {
	// Import config
	initConfig()

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.DisableSuggestions = true

	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.DelCmd)
	rootCmd.AddCommand(commands.GetCmd)
	rootCmd.AddCommand(commands.LstCmd)
	rootCmd.AddCommand(commands.ExpCmd)
	rootCmd.AddCommand(commands.ImpCmd)
	rootCmd.AddCommand(commands.TtlCmd)
}

func initConfig() {
	homePath := kvpath.GetKVHomeDir() + "/.config/kvstok/" + database.DBName

	var err error

	opt := nutsdb.DefaultOptions
	opt.SegmentSize = 8 * nutsdb.MB
	opt.CommitBufferSize = 4 * nutsdb.MB
	opt.MaxBatchSize = (15 * opt.SegmentSize / 4) / 100
	opt.MaxBatchCount = (15 * opt.SegmentSize / 4) / 100 / 100

	database.DB, err = nutsdb.Open(opt, nutsdb.WithDir(homePath))
	if err != nil {
		log.Fatal(err.Error())
	}

	database.DB.Update(func(tx *nutsdb.Tx) error {
		// you should call Bucket with data structure and the name of bucket first
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	})

	// defer database.DB.Close()
}

func preRun(cmd *cobra.Command, args []string) error {
	store, err := database.NewStore("")
	if err != nil {
		return err
	}
	defer store.Close()

	// Tenta carregar salt se existir
	_ = store.LoadMasterSalt()

	// Se o usuário passou --master, define automaticamente
	if masterPassword != "" {
		if err := store.GetMasterKey().SetMasterPassword(masterPassword); err != nil {
			return fmt.Errorf("senha mestra inválida")
		}
		return nil
	}

	// Se ainda não está configurado, força o init
	if !store.sec.IsMasterPasswordSet() {
		if cmd.Use != "init" {
			fmt.Println("⚠️  kvstok ainda não foi inicializado.")
			fmt.Println("   Execute: kvstok init")
			os.Exit(1)
		}
	}

	return nil
}

// GetStore retorna o store já com master password carregada (usado nos comandos)
func GetStore() (*database.Store, error) {
	store, err := database.NewStore("")
	if err != nil {
		return nil, err
	}

	if err := store.LoadMasterSalt(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}