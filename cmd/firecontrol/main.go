package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fire-equipment-control/internal/domain"
	"fire-equipment-control/internal/service"
	"fire-equipment-control/internal/storage"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		printUsage()
		return nil
	}
	switch args[0] {
	case "demo":
		return runDemo()
	case "init":
		path := "fire-equipment.db"
		if len(args) > 1 && args[1] != "" {
			path = args[1]
		}
		store, err := storage.Open(path)
		if err != nil {
			return err
		}
		defer store.Close()
		fmt.Println(filepath.Clean(path))
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("firecontrol: 消防器材登记、审核、查询和归档")
	fmt.Println("commands: help | init [path] | demo")
}

func runDemo() error {
	store, err := storage.Open("fire-equipment.db")
	if err != nil {
		return err
	}
	defer store.Close()
	manager := service.New(store)
	workflow, err := manager.CreateReviewArchive(domain.EquipmentRecord{Code: "DEMO-001", Type: "灭火器", Building: "A栋", Floor: 1, InspectionDate: "2026-08-23", Owner: "安全员", Status: domain.StatusDraft}, "demo-operator")
	if err != nil {
		return err
	}
	fmt.Printf("%s %s reviews=%d audits=%d\n", workflow.Record.Code, domain.StatusLabel(workflow.Record.Status), workflow.ReviewCount, workflow.AuditCount)
	return nil
}
