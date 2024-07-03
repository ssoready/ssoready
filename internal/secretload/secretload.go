package secretload

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func Load(ctx context.Context) error {
	secretID := os.Getenv(envify(fmt.Sprintf("%s_LOAD_AWS_SECRET_ID", filepath.Base(os.Args[0]))))
	if secretID == "" {
		return nil
	}

	return loadSecret(ctx, secretID)
}

func loadSecret(ctx context.Context, secretID string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	res, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretID,
	})

	if err != nil {
		return fmt.Errorf("get secret value: %w", err)
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(*res.SecretString), &data); err != nil {
		return fmt.Errorf("parse secret: %w", err)
	}

	for k, v := range data {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("setenv: %w", err)
		}
	}

	return nil
}
