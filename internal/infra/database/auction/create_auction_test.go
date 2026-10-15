package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose_Integration(t *testing.T) {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
	}
	mongoDBName := os.Getenv("MONGODB_DB")
	if mongoDBName == "" {
		mongoDBName = "auctions_test"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skip("MongoDB nao acessivel, pulando teste de integracao:", err)
		return
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("MongoDB nao esta respondendo, pulando teste de integracao:", err)
		return
	}

	duration := 1 * time.Second
	os.Setenv("AUCTION_DURATION", duration.String())
	defer os.Unsetenv("AUCTION_DURATION")

	database := client.Database(mongoDBName)
	repo := NewAuctionRepository(database)

	auctionEntity, internalErr := auction_entity.CreateAuction(
		"Playstation 5",
		"Electronics",
		"Console de videogame em perfeito estado",
		auction_entity.New,
	)
	assert.Nil(t, internalErr)

	createErr := repo.CreateAuction(context.Background(), auctionEntity)
	assert.Nil(t, createErr)

	createdAuction, findErr := repo.FindAuctionById(context.Background(), auctionEntity.Id)
	assert.Nil(t, findErr)
	assert.Equal(t, auction_entity.Active, createdAuction.Status, "Leilao deve estar aberto (Active)")

	time.Sleep(duration + 300*time.Millisecond)

	closedAuction, findErr2 := repo.FindAuctionById(context.Background(), auctionEntity.Id)
	assert.Nil(t, findErr2)
	assert.Equal(t, auction_entity.Completed, closedAuction.Status, "Status alterado automaticamente para Closed")
}
