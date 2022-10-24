package mongoDB

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"time"
)

type User struct {
	UserName string `json:"user_name" bson:"_id"`
	Passwd   string `json:"passwd" bson:"passwd"`
	Salt     string `bson:"salt"`
	Token    string `json:"token" bson:"token"`
}

type UserSSH struct {
	Key      string `bson:"_id"`
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Passwd   string `json:"passwd"`
}

type MongoClient struct {
	mongoCli          *mongo.Client
	userSSHCollection *mongo.Collection
	userCollection    *mongo.Collection
}

var Client *MongoClient

func init() {
	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read Config File Fail %e", err)
	}
	host := conf.Get("mongo.Host").(string)
	clientOptions := options.Client().ApplyURI(host).SetSocketTimeout(3 * time.Second)

	// 连接到MongoDB
	mgoCli, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("connect to mongo DB fail : %v", err)
	}

	// 检查连接
	err = mgoCli.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("connect to mongo DB fail : %v", err)
	}

	userSSHCollection := mgoCli.Database("Argusyes").Collection("UserSSH")
	userCollection := mgoCli.Database("Argusyes").Collection("User")

	Client = &MongoClient{
		mongoCli:          mgoCli,
		userSSHCollection: userSSHCollection,
		userCollection:    userCollection,
	}
	log.Println("MongoDB connect success")
}

func GeneralSSHId(ssh UserSSH) string {
	return fmt.Sprintf("%s:%s@%s:%d", ssh.UserName, ssh.User, ssh.Host, ssh.Port)
}

func MD5V(str string, salt string) string {
	b := []byte(str)
	s := []byte(salt)
	h := md5.New()
	h.Write(s)
	h.Write(b)
	var res []byte
	res = h.Sum(nil)
	for i := 0; i < 3; i++ {
		h.Reset()
		h.Write(res)
		res = h.Sum(nil)
	}
	return hex.EncodeToString(res)
}

func nextSalt() string {
	rand.Seed(time.Now().UnixNano())
	size := rand.Intn(26) + 1
	warehouse := []int{97, 122}
	result := make([]byte, 26)
	for i := 0; i < size; i++ {
		result[i] = uint8(warehouse[0] + rand.Intn(26))
	}
	return string(result)
}

func (c *MongoClient) Close() {
	err := c.mongoCli.Disconnect(context.TODO())
	if err != nil {
		log.Fatalf("Close MongoDB fail : %v", err)
	}
	log.Println("Connection to MongoDB closed.")
}

func (c *MongoClient) InsertUser(user User) error {
	user.Salt = nextSalt()
	user.Passwd = MD5V(user.Passwd, user.Salt)
	_, err := c.userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		errText := fmt.Sprintf("Insert fail : %v", err)
		return errors.New(errText)
	}
	return nil
}

func (c *MongoClient) CheckUserPasswd(user User) error {
	var result User
	if err := c.userCollection.FindOne(context.TODO(), bson.D{{"_id", user.UserName}}).Decode(&result); err != nil {
		return err
	} else if result.Passwd != MD5V(user.Passwd, result.Salt) {
		return errors.New("pass word diff")
	}
	return nil
}

func (c *MongoClient) UpdateUserToken(username, token string) error {
	filter := bson.D{{"_id", username}}
	update := bson.D{{"$set", bson.D{{"token", token}}}}
	result, err := c.userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil || result.UpsertedCount == 1 {
		return errors.New("update token fail" + err.Error())
	}
	return nil
}

func (c *MongoClient) CheckUserToken(username, token string) error {
	var result User
	if err := c.userCollection.FindOne(context.TODO(), bson.D{{"_id", username}}).Decode(&result); err != nil {
		return err
	} else if result.Token != token {
		return errors.New("token diff")
	}
	return nil
}

func (c *MongoClient) InsertUserSSH(userSSH []UserSSH) ([]string, error) {
	docs := make([]interface{}, 0)
	for _, ssh := range userSSH {
		ssh.Key = GeneralSSHId(ssh)
		docs = append(docs, ssh)
	}
	ordered := false
	result, err := c.userSSHCollection.InsertMany(context.TODO(), docs, &options.InsertManyOptions{Ordered: &ordered})
	if result == nil {
		errText := fmt.Sprintf("Insert fail : %v", err)
		return nil, errors.New(errText)
	}
	r := make([]string, 0)
	for _, id := range result.InsertedIDs {
		r = append(r, id.(string))
	}
	if err != nil {
		errText := fmt.Sprintf("Insert fail : %v", err)
		return r, errors.New(errText)
	}
	return r, nil
}
