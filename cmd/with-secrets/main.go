package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ucarion/cli"
)

type args struct {
	SecretID string   `cli:"secret-id"`
	Argv     []string `cli:"argv..."`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		client := secretsmanager.NewFromConfig(cfg)
		res, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
			SecretId: &args.SecretID,
		})

		if err != nil {
			return fmt.Errorf("get secret value: %w", err)
		}

		var data map[string]string
		if err := json.Unmarshal([]byte(*res.SecretString), &data); err != nil {
			return fmt.Errorf("parse secret: %w", err)
		}

		envv := os.Environ()
		for k, v := range data {
			envv = append(envv, fmt.Sprintf("%s=%s", k, v))
		}

		if len(args.Argv) == 0 {
			return fmt.Errorf("argv is required")
		}

		argv0, err := exec.LookPath(args.Argv[0])
		if err != nil {
			return fmt.Errorf("look path: %q: %w", args.Argv[0], err)
		}

		if err := syscall.Exec(argv0, args.Argv, envv); err != nil {
			return fmt.Errorf("execve: %w", err)
		}

		return nil
	})
}
