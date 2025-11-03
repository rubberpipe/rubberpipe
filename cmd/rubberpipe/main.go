package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/rubberpipe/rubberpipe/internal/hub"
	"github.com/rubberpipe/rubberpipe/internal/orchestration"
	"github.com/rubberpipe/rubberpipe/internal/persistence"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/rubberpipe/rubberpipe/adapters/destinations"
	_ "github.com/rubberpipe/rubberpipe/adapters/sources"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Println("Usage: rubberpipe <command> [args...]")
		fmt.Println("Commands: backup, restore, list, config")
		return
	}

	db, err := sql.Open("sqlite3", "./rubberpipe.db")
	if err != nil {
		log.Fatalf("FATAL: Failed to open database: %v", err)
	}
	defer db.Close()

	configManager, err := persistence.NewConfigManagerService(db)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize config manager: %v", err)
	}

	adapterConfigs, err := configManager.GetAdapterConfigs()
	if err != nil {
		log.Fatalf("FATAL: Failed to load adapter configs from DB: %v", err)
	}

	executionHub, err := hub.NewHub(adapterConfigs)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize execution hub: %v", err)
	}

	orchestrator, err := orchestration.NewOrchestratorService(executionHub, configManager)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize orchestrator: %v", err)
	}

	cmd := os.Args[1]

	switch cmd {
	case "backup":
		if len(os.Args) < 4 {
			log.Fatal("Usage: rubberpipe backup <source> <destination>")
		}
		source := os.Args[2]
		dest := os.Args[3]

		file, err := orchestrator.Backup(ctx, source, dest)
		if err != nil {
			fmt.Println("Backup failed:", err)
		} else {
			fmt.Println("Backup successful:", file)
		}

	case "restore":
		if len(os.Args) < 3 {
			log.Fatal("Usage: rubberpipe restore <backup_id>")
		}
		backupIdString := os.Args[2]

		backupId, err := strconv.Atoi(backupIdString)
		if err != nil {
			log.Fatalf("Invalid backup_id: %v", err)
		}

		err = orchestrator.Restore(ctx, backupId)

		if err != nil {
			fmt.Println("Restore failed:", err)
		} else {
			fmt.Printf("Restore successful for backup ID: %d\n", backupId)
		}

	case "list":
		records, err := orchestrator.ListBackupHistory(context.Background())
		if err != nil {
			log.Fatalf("Failed to list history: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		fmt.Fprintln(w, "ID\tTIMESTAMP\tSOURCE\tDESTINATION\tSTATUS\tERROR")
		fmt.Fprintln(w, "--\t---------\t------\t-----------\t------\t-----")

		for _, r := range records {
			errMsg := r.ErrorMsg
			if errMsg == "" {
				errMsg = "—"
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
				r.ID,
				r.Timestamp.Format("2006-01-02 15:04:05"),
				r.Source,
				r.Destination,
				r.Status,
				errMsg)
		}

		w.Flush()

	case "config":
		if len(os.Args) < 3 {
			log.Fatal("Usage: rubberpipe config <list|add|remove>")
		}
		sub := os.Args[2]

		switch sub {
		case "list":
			cfgs, err := orchestrator.ListAdapterConfigs()
			if err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tCONFIG PREVIEW")
			fmt.Fprintln(w, "----\t----\t--------------")

			for _, cfg := range cfgs {
				var prettyCfg map[string]interface{}

				json.Unmarshal([]byte(cfg.ConfigJSON), &prettyCfg)

				if _, ok := prettyCfg["password"]; ok {
					prettyCfg["password"] = "*****"
				}

				compactBytes, _ := json.Marshal(prettyCfg)
				preview := string(compactBytes)

				if len(preview) > 50 {
					preview = preview[:47] + "..."
				}

				fmt.Fprintf(w, "%s\t%s\t%s\n", cfg.Name, cfg.Type, preview)
			}
			w.Flush()

		case "add":
			if len(os.Args) < 6 {
				log.Fatal("Usage: rubberpipe config add <name> <type> <json>")
			}
			name := os.Args[3]
			typ := os.Args[4]
			cfgJSON := os.Args[5]

			err := orchestrator.AddAdapterConfig(name, typ, cfgJSON)
			if err != nil {
				log.Fatalf("Failed to add config: %v", err)
			}
			fmt.Printf("Adapter config '%s' added.\n", name)

		case "remove":
			if len(os.Args) < 4 {
				log.Fatal("Usage: rubberpipe config remove <name>")
			}
			name := os.Args[3]
			err := orchestrator.RemoveAdapterConfig(name)
			if err != nil {
				log.Fatalf("Failed to remove config: %v", err)
			}
			fmt.Printf("Adapter config '%s' removed.\n", name)

		default:
			fmt.Println("Unknown config subcommand. Use list, add, or remove.")
		}

	default:
		fmt.Println("Unknown command:", cmd)
	}
}
