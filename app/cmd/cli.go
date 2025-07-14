package cmd

import (
	"context"
	"log"
	"os"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/models/migrations"
	"github.com/urfave/cli/v3"
)

func RunCli() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "migrate",
				Usage: "Run database migration",
				Action: func(ctx context.Context, c *cli.Command) error {
					db, err := configs.OpenConnection()
					if err != nil {
						return err
					}
					if err := migrations.AutoMigrate(db); err != nil {
						return err
					}
					log.Println("✅ Migration complete")
					return nil
				},
			},
			{
				Name:  "generate-keys",
				Usage: "Generate new session authentication and encryption keys for .env",
				Action: func(ctx context.Context, c *cli.Command) error {

					if err := configs.GenerateAndPrintSessionKeys(); err != nil {
						return err
					}
					log.Println("✅ Key generation complete. Please copy the keys to your .env file.")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
