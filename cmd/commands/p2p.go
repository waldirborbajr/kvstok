package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	security "github.com/waldirborbajr/kvstok/internal/security"
	syncpkg "github.com/waldirborbajr/kvstok/internal/sync"
)

// P2PCmd represents the peer-to-peer sync command.
var P2PCmd = &cobra.Command{
	Use:     "p2p",
	Short:   "Start P2P synchronization with discovered peers.",
	Aliases: []string{"sync"},
	Long: `Start the kvstok peer-to-peer synchronization agent.
It advertises this host on the local network, discovers other peers via mDNS,
and exchanges key deltas over DTLS.`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceName, err := cmd.Flags().GetString("device")
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read device name: %v\n", err)
			return
		}

		if deviceName == "" {
			deviceName, err = os.Hostname()
			if err != nil {
				deviceName = "kvstok-peer"
			}
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read port: %v\n", err)
			return
		}

		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read interval: %v\n", err)
			return
		}

		once, err := cmd.Flags().GetBool("once")
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read once flag: %v\n", err)
			return
		}

		pubKeyPath := filepath.Join(kvpath.GetKVHomeDir(), ".config", "kvstok", "kvstok.pub")
		privKeyPath := filepath.Join(kvpath.GetKVHomeDir(), ".config", "kvstok", "kvstok.priv")

		pubKeyBytes, err := os.ReadFile(pubKeyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read public key: %v\n", err)
			return
		}

		privKeyBytes, err := os.ReadFile(privKeyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to read private key: %v\n", err)
			return
		}

		pubKey := security.BytesToPublicKey(pubKeyBytes)
		privKey := security.BytesToPrivateKey(privKeyBytes)

		pm, err := syncpkg.NewPeerManager(deviceName, pubKey, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to create peer manager: %v\n", err)
			return
		}

		if err := pm.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to start peer discovery: %v\n", err)
			return
		}
		defer func() {
			if err := pm.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "P2PCmd() - failed to stop peer manager: %v\n", err)
			}
		}()

		store, err := database.GetStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - failed to open store: %v\n", err)
			return
		}
		defer store.Close()

		syncer := syncpkg.NewDTLSSync(pm, store, pubKey, privKey)

		fmt.Printf("Starting P2P sync as %s on port %d\n", deviceName, port)

		if err := syncWithPeers(cmd, pm, syncer, interval, once); err != nil {
			fmt.Fprintf(os.Stderr, "P2PCmd() - %v\n", err)
		}
	},
}

func init() {
	P2PCmd.Flags().String("device", "", "Device name to advertise via mDNS")
	P2PCmd.Flags().Int("port", syncpkg.DefaultPort, "Port for P2P discovery and DTLS connections")
	P2PCmd.Flags().Duration("interval", 30*time.Second, "Interval between peer sync attempts")
	P2PCmd.Flags().Bool("once", false, "Run one sync pass then exit")
}

func syncWithPeers(cmd *cobra.Command, pm *syncpkg.PeerManager, syncer *syncpkg.DTLSSync, interval time.Duration, once bool) error {
	ctx := cmd.Context()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		peers := pm.DiscoverPeers()
		if len(peers) == 0 {
			fmt.Fprintln(os.Stderr, "P2PCmd() - no peers discovered yet")
		} else {
			for _, peer := range peers {
				if err := syncer.SyncWithPeer(peer); err != nil {
					fmt.Fprintf(os.Stderr, "P2PCmd() - sync failed for peer %s: %v\n", peer.DeviceName, err)
					continue
				}
				fmt.Printf("Synchronized with peer %s (%s)\n", peer.DeviceName, peer.ID)
			}
		}

		if once {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
