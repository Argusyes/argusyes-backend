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
	UserName string `json:"userName" bson:"_id"`
	Passwd   string `json:"passwd" bson:"passwd"`
	Salt     string `bson:"salt"`
}

type UserSSH struct {
	Key      string `bson:"key"`
	UserName string `json:"username" bson:"username"`
	Name     string `json:"name" bson:"name"`
	Port     int    `json:"port" bson:"port"`
	Host     string `json:"host" bson:"host"`
	User     string `json:"user" bson:"user"`
	Passwd   string `json:"passwd" bson:"passwd"`
}

type UserSSHUpdater struct {
	OldSSH UserSSH
	NewSSH UserSSH
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
	_, err = userSSHCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("UserSSHKeyIndex"),
		},
	)
	if err != nil {
		log.Fatalf("create index fail : %v", err)
	}

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

func (c *MongoClient) ChangeUserPasswd(user User) interface{} {

	salt := nextSalt()

	filter := bson.D{{"_id", user.UserName}}
	update := bson.D{{"$set", bson.D{{"passwd", MD5V(user.Passwd, salt)}, {"salt", salt}}}}
	result, err := c.userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil || result.UpsertedCount == 1 {
		return errors.New("change passwd fail" + err.Error())
	}
	return nil
}

func (c *MongoClient) InsertUserSSH(userSSH []UserSSH) ([]string, error) {
	r := make([]string, 0)
	errText := ""
	for _, ssh := range userSSH {
		ssh.Key = GeneralSSHId(ssh)
		_, err := c.userSSHCollection.InsertOne(context.TODO(), ssh)
		if err != nil {
			errText += fmt.Sprintf("insert fail %s : %v", ssh.Key, err)
		} else {
			r = append(r, ssh.Key)
		}

	}
	if errText == "" {
		return r, nil
	}
	return r, errors.New(errText)
}

func (c *MongoClient) DeleteUserSSH(userSSH []UserSSH) ([]string, error) {

	r := make([]string, 0)
	errText := ""
	for _, ssh := range userSSH {
		ssh.Key = GeneralSSHId(ssh)
		result, err := c.userSSHCollection.DeleteOne(context.TODO(), bson.M{"key": ssh.Key})
		if err != nil || result.DeletedCount == 0 {
			errText += fmt.Sprintf("delete fail %s : %v", ssh.Key, err)
		} else {
			r = append(r, ssh.Key)
		}

	}
	if errText == "" {
		return r, nil
	}
	return r, errors.New(errText)
}

func (c *MongoClient) UpdateUserSSH(userSSHUpdater []UserSSHUpdater) ([]string, error) {
	r := make([]string, 0)
	errText := ""
	for _, u := range userSSHUpdater {
		u.OldSSH.Key = GeneralSSHId(u.OldSSH)
		u.NewSSH.Key = GeneralSSHId(u.NewSSH)
		result, err := c.userSSHCollection.UpdateOne(context.TODO(), bson.D{{"key", u.OldSSH.Key}}, bson.D{{"$set", u.NewSSH}})
		if err != nil || result.ModifiedCount == 0 {
			errText += fmt.Sprintf("update fail %s : %v", u.OldSSH.Key, err)
		}

	}
	if errText == "" {
		return r, nil
	}
	return r, errors.New(errText)
}

func (c *MongoClient) SelectUserSSH(username string, name string) ([]UserSSH, error) {
	filter := bson.D{{"username", username}}
	if name != "" {
		filter = append(filter, bson.E{Key: "name", Value: name})
	}
	result, err := c.userSSHCollection.Find(context.TODO(), filter)
	if err != nil {
		errText := fmt.Sprintf("Select fail %s : %v", username, err)
		return nil, errors.New(errText)
	}
	var userSSH []UserSSH
	if err = result.All(context.TODO(), &userSSH); err != nil {
		errText := fmt.Sprintf("Select fail %s : %v", username, err)
		return nil, errors.New(errText)
	}
	return userSSH, nil
}
