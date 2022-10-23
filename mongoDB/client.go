package mongoDB

import (
	"context"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type UserSSH struct {
	UserID    string `json:"user_id"`
	SSHConfig []SSH  `json:"ssh_config"`
}

type SSH struct {
	Port   int    `json:"port"`
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type MongoClient struct {
	mongoCli          *mongo.Client
	userSSHCollection *mongo.Collection
	ctx               context.Context
	cancel            context.CancelFunc
}

var Client *MongoClient

func init() {
	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}
	host := conf.Get("mongo.Host").(string)
	clientOptions := options.Client().ApplyURI(host).SetSocketTimeout(3 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// 连接到MongoDB
	mgoCli, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("connect to mongo DB fail : %v", err)
	}

	// 检查连接
	err = mgoCli.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("connect to mongo DB fail : %v", err)
	}

	collection := mgoCli.Database("Argusyes").Collection("UserSSH")

	Client = &MongoClient{
		mongoCli:          mgoCli,
		userSSHCollection: collection,
		ctx:               ctx,
		cancel:            cancel,
	}
	log.Println("MongoDB connect success")
}

func generalId(userId string, ssh SSH) string {
	return fmt.Sprintf("%s:%s@%s:%d", userId, ssh.User, ssh.Host, ssh.Port)
}

func (c *MongoClient) Close() {
	err := c.mongoCli.Disconnect(context.TODO())
	if err != nil {
		log.Fatalf("Close MongoDB fail : %v", err)
	}
	log.Println("Connection to MongoDB closed.")
	c.cancel()
}

func (c *MongoClient) InsertUserSSH(userSSH UserSSH) error {
	docs := make([]interface{}, 0)
	for _, ssh := range userSSH.SSHConfig {
		docs = append(docs, bson.M{"_id": generalId(userSSH.UserID, ssh), "port": ssh.Port, "host": ssh.Host, "user": ssh.User, "passwd": ssh.Passwd})
	}
	ordered := false
	_, err := c.userSSHCollection.InsertMany(context.TODO(), docs, &options.InsertManyOptions{Ordered: &ordered})
	if err != nil {
		errText := fmt.Sprintf("Insert fail : %v", err)
		return errors.New(errText)
	}
	return nil
}
