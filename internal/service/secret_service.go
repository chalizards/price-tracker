package service

import (
	"context"
	"log"
	"sync"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var (
	secretCache = make(map[string]string)
	secretMu    sync.RWMutex
)

func GetGeminiSecret(secretName string) string {
	secretMu.RLock()
	if cached, ok := secretCache[secretName]; ok {
		secretMu.RUnlock()
		return cached
	}
	secretMu.RUnlock()

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Falha ao criar cliente: %v", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName + "/versions/latest",
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("Falha ao acessar versão do segredo: %v", err)
	}

	value := string(result.Payload.Data)

	secretMu.Lock()
	secretCache[secretName] = value
	secretMu.Unlock()

	return value
}
