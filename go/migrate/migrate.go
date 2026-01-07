package migrate

import (
	"fmt"
	"log"

	"sotsukenn/go/database"
	"sotsukenn/go/models"

	"github.com/spf13/cobra"
	"gorm.io/gorm/logger"
)

func MigrateModelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "db",
		Short: "Migrate models to database",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := database.GetDBWithLogger(logger.Silent)
			if err != nil {
				log.Fatalf("Failed to get database instance: %v", err)
			}

			fmt.Println("Migrating models to the database...")

			err = db.AutoMigrate(
				&models.User{},
				&models.FrigateConnect{},
				&models.FCMToken{},
				&models.DetectionEvent{},
			)

			if err != nil {
				log.Fatalf("Migration failed: %v", err)
			}

			fmt.Println("Migration completed successfully.")
		},
	}
}

func MigrateMarkdownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "md",
		Short: "Migrate markdown files to database",
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			force, _ := cmd.Flags().GetBool("force")
			update, _ := cmd.Flags().GetBool("update")

			if path == "" {
				log.Fatal("Path is required. Use --path flag to specify the directory.")
			}

			fmt.Printf("Migrating markdown files from: %s\n", path)
			fmt.Printf("Force: %v, Update: %v\n", force, update)

			// TODO: 实现实际的 markdown 文件迁移逻辑
			fmt.Println("Markdown migration feature is not implemented yet.")
		},
	}

	cmd.Flags().StringP("path", "p", "", "Markdown file directory location")
	cmd.Flags().Bool("force", false, "Force overwrite database record")
	cmd.Flags().Bool("update", false, "Update database record")

	return cmd
}
